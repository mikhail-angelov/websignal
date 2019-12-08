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

type oauth2Config struct {
	oauth2.Config
	infoURL string
	mapUser func(UserData, []byte) User // map info from InfoURL to User
}

// Auth2Provider service
type Auth2Provider struct {
	name string
	jwt  *JWT
	conf *oauth2Config
	log  *logger.Log
	url  string
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
	url    string
}

//NewAuth2Provider constructor
func NewAuth2Provider(params *Auth2ProviderParams) *Auth2Provider {
	return &Auth2Provider{
		name: params.name,
		jwt:  params.jwt,
		log:  params.log,
		conf: getConf(params),
		url:  params.url,
	}
}

func getConf(params *Auth2ProviderParams) *oauth2Config {
	if params.name == "yandex" {
		return &oauth2Config{
			Config: oauth2.Config{
				ClientID:     params.cid,
				ClientSecret: params.secret,
				Endpoint:     yandex.Endpoint,
				Scopes:       []string{},
			},
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
		return &oauth2Config{}
	}
	return nil
}

// Handler main handler
func (a *Auth2Provider) Handler(w http.ResponseWriter, r *http.Request) {
	if strings.HasSuffix(r.URL.Path, urlLoginSuffix) {
		a.loginHandler(w, r)
		return
	}
	if strings.HasSuffix(r.URL.Path, urlCallbackSuffix) {
		a.authHandler(w, r)
		return
	}
	if strings.HasSuffix(r.URL.Path, urlLogoutSuffix) {
		a.logoutHandler(w, r)
		return
	}
	w.WriteHeader(http.StatusNotFound)
}

//Name return provider name
func (a *Auth2Provider) Name() string {
	return a.name
}

func (a *Auth2Provider) loginHandler(w http.ResponseWriter, r *http.Request) {
	if a.name == "local" {
		a.localLogin(w, r)
		return
	}
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
		a.log.Logf("[WARN] failed to make claim's id %v", err)
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
			ExpiresAt: time.Now().Add(24 * time.Hour).Unix(),
			NotBefore: time.Now().Add(-1 * time.Minute).Unix(),
		},
	}

	if _, err := a.jwt.Set(w, claims); err != nil {
		log.Printf("[WARN] failed to set token %v", err)
		return
	}
	// setting RedirectURL to rootURL/routingPath/provider/callback
	// e.g. http://localhost:8080/auth/github/callback
	a.conf.RedirectURL = a.makeRedirURL(r.URL.Path)

	loginURL := a.conf.AuthCodeURL(rd)
	log.Printf("[INFO] yandex login %s %s", loginURL, rd)
	http.Redirect(w, r, loginURL, http.StatusFound)
}
func (a *Auth2Provider) authHandler(w http.ResponseWriter, r *http.Request) {
	oauthClaims, token, err := a.jwt.Get(r)
	if err != nil {
		log.Printf("[WARN] token parse error %v %v", err, oauthClaims)
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, "invalid request")
		return
	}
	if oauthClaims.Handshake == nil {
		log.Printf("[WARN] invalid handshake token %v", err)
		return
	}

	retrievedState := oauthClaims.Handshake.State
	log.Printf("[INFO] token parse %s %s", r.URL.Query().Get("state"), retrievedState)
	if retrievedState == "" || retrievedState != r.URL.Query().Get("state") {
		log.Printf("[WARN] unexpected state %v", err)
		return
	}

	tok, err := a.conf.Exchange(context.Background(), r.URL.Query().Get("code"))
	if err != nil {
		log.Printf("[WARN] 1 error %v", err)
	}

	client := a.conf.Client(context.Background(), tok)
	uinfo, err := client.Get(a.conf.infoURL)
	if err != nil {
		log.Printf("[WARN] 2 error %v", err)
	}

	defer func() {
		if e := uinfo.Body.Close(); e != nil {
			log.Printf("[WARN] failed to close response body, %s", e)
		}
	}()

	data, err := ioutil.ReadAll(uinfo.Body)
	if err != nil {
		log.Printf("[WARN] failed to read user info %v", err)
		return
	}

	jData := map[string]interface{}{}
	if e := json.Unmarshal(data, &jData); e != nil {
		log.Printf("[WARN] failed to unmarshal user info %v", err)
		return
	}

	u := a.conf.mapUser(jData, data)

	cid, err := randToken()
	if err != nil {
		log.Printf("[WARN] failed to make claim's id %v", err)
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
		log.Printf("[WARN] failed to set token %v", err)
		return
	}

	log.Printf("[INFO] login success %v, %s %v", oauthClaims.Handshake, token, u)

	if oauthClaims.Handshake != nil && oauthClaims.Handshake.From != "" {
		http.Redirect(w, r, oauthClaims.Handshake.From, http.StatusTemporaryRedirect)
		return
	}

	render.Status(r, http.StatusOK)
}
func (a *Auth2Provider) logoutHandler(w http.ResponseWriter, r *http.Request) {
	a.jwt.Clean(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
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

	u := User{
		ID:         "local_" + email,
		Name:       email,
		Email:      email,
		PictureURL: "",
	}
	cid, err := randToken()
	if err != nil {
		log.Printf("[WARN] failed to make claim's id %v", err)
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
		log.Printf("[WARN] failed to set token %v", err)
		return
	}

	http.Redirect(w, r, redirect, http.StatusFound)
}

func (a *Auth2Provider) makeRedirURL(path string) string {
	elems := strings.Split(path, "/")
	newPath := strings.Join(elems[:len(elems)-1], "/")

	return strings.TrimRight(a.url, "/") + strings.TrimRight(newPath, "/") + urlCallbackSuffix
}
