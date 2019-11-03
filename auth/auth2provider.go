package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/render"
	"github.com/mikhail-angelov/websignal/logger"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/yandex"
)

const (
	urlLoginSuffix    = "/login"
	urlCallbackSuffix = "/callback"
	urlLogoutSuffix   = "/logout"
)

type Oauth2Config struct {
	infoURL     string
	redirectURL string
	endpoint    oauth2.Endpoint
	scopes      []string
	mapUser     func(UserData, []byte) User // map info from InfoURL to User
}

type Provider interface {
	Name() string
	LoginHandler(w http.ResponseWriter, r *http.Request)
	AuthHandler(w http.ResponseWriter, r *http.Request)
	LogoutHandler(w http.ResponseWriter, r *http.Request)
}

// Auth2Provider service
type Auth2Provider struct {
	name   string
	jwt    *JWT
	conf   *oauth2.Config
	config *Oauth2Config
	log    *logger.Log
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// Auth2ProviderParams service parameters
type Auth2ProviderParams struct {
	name   string
	cid    string
	secret string
	jwt    *JWT
	log    *logger.Log
}

//NewAuth2Provider constructor
func NewAuth2Provider(params *Auth2ProviderParams) *Auth2Provider {
	config := getConf(params)
	conf := &oauth2.Config{
		ClientID:     params.cid,
		ClientSecret: params.secret,
		Scopes:       []string{},
		Endpoint:     config.endpoint,
	}
	conf.RedirectURL = config.redirectURL
	return &Auth2Provider{
		name:   params.name,
		jwt:    params.jwt,
		log:    params.log,
		conf:   conf,
		config: config,
	}
}

func getConf(params *Auth2ProviderParams) *Oauth2Config {
	if params.name == "yandex" {
		return &Oauth2Config{
			endpoint:    yandex.Endpoint,
			redirectURL: "http://localhost:9001/auth/yandex/callback",
			scopes:      []string{},
			// See https://tech.yandex.com/passport/doc/dg/reference/response-docpage/
			infoURL: "https://login.yandex.ru/info?format=json",
			mapUser: func(data UserData, _ []byte) User {
				userInfo := User{
					ID:   "yandex_" + "data.HashID(sha1.New()" + data.Value("id"),
					Name: data.Value("display_name"), // using Display Name by default
				}
				if userInfo.Name == "" {
					userInfo.Name = data.Value("real_name") // using Real Name (== full name) if Display Name is empty
				}
				if userInfo.Name == "" {
					userInfo.Name = data.Value("login") // otherwise using login
				}

				if data.Value("default_avatar_id") != "" {
					userInfo.PictureURL = fmt.Sprintf("https://avatars.yandex.net/get-yapic/%s/islands-200", data.Value("default_avatar_id"))
				}
				return userInfo
			},
		}
	} else if params.name == "local" {
		return &Oauth2Config{}
	}
	return nil
}

// Handler main handler
func (a *Auth2Provider) Handler(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, urlLoginSuffix) && a.name == "local" {
		a.localLogin(w, r)
		return
	}
	if strings.HasSuffix(r.URL.Path, urlLoginSuffix) {
		a.LoginHandler(w, r)
		return
	}
	if strings.HasSuffix(r.URL.Path, urlCallbackSuffix) {
		a.AuthHandler(w, r)
		return
	}
	if strings.HasSuffix(r.URL.Path, urlLogoutSuffix) {
		a.LogoutHandler(w, r)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}
func (a *Auth2Provider) Name() string {
	return a.name
}
func (a *Auth2Provider) LoginHandler(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, "invalid request")
		return
	}
	request := &loginRequest{}
	err = json.Unmarshal(body, request)
	rd, err := randToken()
	cid, err := randToken()
	if err != nil {
		a.log.Logf("[DEBUG] failed to make claim's id %v", err)
	}

	claims := Claims{
		Handshake: &Handshake{
			State: rd,
			From:  r.URL.Query().Get("from"),
		},
		SessionOnly: r.URL.Query().Get("session") != "" && r.URL.Query().Get("session") != "0",
		StandardClaims: jwt.StandardClaims{
			Id:        cid,
			Audience:  r.URL.Query().Get("site"),
			ExpiresAt: time.Now().Add(30 * time.Minute).Unix(),
			NotBefore: time.Now().Add(-1 * time.Minute).Unix(),
		},
	}

	if _, err := a.jwt.Set(w, claims); err != nil {
		log.Printf("[DEBUG] failed to set token %v", err)
		return
	}
	loginURL := a.conf.AuthCodeURL(rd)
	log.Printf("[INFO] yandex login %s %s", loginURL, rd)
	http.Redirect(w, r, loginURL, http.StatusFound)
}
func (a *Auth2Provider) AuthHandler(w http.ResponseWriter, r *http.Request) {
	oauthClaims, token, err := a.jwt.Get(r)
	if err != nil {
		log.Printf("[WARN] token parse error %v %v", err, oauthClaims)
		// render.Status(r, http.StatusBadRequest)
		// render.PlainText(w, r, "invalid request")
		return
	}
	if oauthClaims.Handshake == nil {
		log.Printf("[WARN] invalid handshake token %v", err)
		return
	}

	retrievedState := oauthClaims.Handshake.State
	log.Printf("[WARN] token parse %s %s", r.URL.Query().Get("state"), retrievedState)
	if retrievedState == "" || retrievedState != r.URL.Query().Get("state") {
		log.Printf("[WARN] unexpected state %v", err)
		return
	}

	tok, err := a.conf.Exchange(context.Background(), r.URL.Query().Get("code"))
	log.Printf("[DEBUG] token with state %s", tok)
	if err != nil {
		log.Printf("[DEBUG] 1 error %v", err)
	}

	client := a.conf.Client(context.Background(), tok)
	infoURL := a.config.infoURL
	uinfo, err := client.Get(infoURL)
	if err != nil {
		log.Printf("[DEBUG] 2 error %v", err)
	}

	defer func() {
		if e := uinfo.Body.Close(); e != nil {
			log.Printf("[WARN] failed to close response body, %s", e)
		}
	}()

	data, err := ioutil.ReadAll(uinfo.Body)
	if err != nil {
		log.Printf("[DEBUG] failed to read user info %v", err)
		return
	}

	jData := map[string]interface{}{}
	if e := json.Unmarshal(data, &jData); e != nil {
		log.Printf("[DEBUG] failed to unmarshal user info %v", err)
		return
	}
	log.Printf("[DEBUG] got raw user info %+v", jData)

	u := a.config.mapUser(jData, data)

	cid, err := randToken()
	if err != nil {
		log.Printf("[DEBUG] failed to make claim's id %v", err)
		return
	}
	claims := Claims{
		User: &u,
		StandardClaims: jwt.StandardClaims{
			Issuer: Issuer,
			Id:     cid,
		},
	}

	if _, err = a.jwt.Set(w, claims); err != nil {
		log.Printf("[DEBUG] failed to set token %v", err)
		return
	}

	log.Printf("[INFO] login success %v, %s %v", oauthClaims.Handshake, token, u)

	if oauthClaims.Handshake != nil && oauthClaims.Handshake.From != "" {
		http.Redirect(w, r, oauthClaims.Handshake.From, http.StatusTemporaryRedirect)
		return
	}

	render.Status(r, http.StatusOK)
}
func (a *Auth2Provider) LogoutHandler(w http.ResponseWriter, r *http.Request) {
	a.jwt.Clean(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

//Check for test
func (a *Auth2Provider) user(w http.ResponseWriter, r *http.Request) {
	user, err := GetUserInfo(r)
	log.Printf("user %v - %v ", user, err)

	if user.PictureURL == "" {
		pic, err := GenerateAvatar(user.Email)
		if err != nil {
			log.Printf("failed to gen avatar %v", err)
		}
		user.Picture = pic
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, user)
}

func (a *Auth2Provider) localLogin(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, "invalid request")
		return
	}

	email := r.Form.Get("email")
	a.log.Logf("[INFO] login %v %v", r.Form.Get("from"), email)

	//todo, check password and handle validation
	if r.Form.Get("password") == "" {
		render.Status(r, http.StatusUnauthorized)
		render.PlainText(w, r, "invalid login")
		return
	}
	redirect := r.URL.Query().Get("from")
	if redirect == "" {
		redirect = "/"
	}

	u := composeAuth2ProviderUser(email)
	cid, err := randToken()
	if err != nil {
		log.Printf("[DEBUG] failed to make claim's id %v", err)
		return
	}

	claims := Claims{
		User: &u,
		StandardClaims: jwt.StandardClaims{
			Issuer: Issuer,
			Id:     cid,
		},
	}

	log.Printf("-------- %v   %v", claims, redirect)

	if _, err = a.jwt.Set(w, claims); err != nil {
		log.Printf("[DEBUG] failed to set token %v", err)
		return
	}

	http.Redirect(w, r, redirect, http.StatusFound)
}

func composeAuth2ProviderUser(email string) User {
	return User{
		ID:         "local_" + email,
		Name:       email,
		Email:      email,
		PictureURL: "",
	}
}

// Value returns value for key or empty string if not found
func (u UserData) Value(key string) string {
	// json.Unmarshal converts json "null" value to go's "nil", in this case return empty string
	if val, ok := u[key]; ok && val != nil {
		return fmt.Sprintf("%v", val)
	}
	return ""
}
