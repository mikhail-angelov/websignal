package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
	"github.com/mikhail-angelov/websignal/logger"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/yandex"
)

const (
	JWTCookieName = "jwt"
	JWTHeaderKey  = "X-JWT"
	JWTQuery      = "token"
	TokenDuration = time.Hour
	Issuer        = "websignal"
	SecureCookies = false
)

// Auth service
type Auth struct {
	jwt  *JWT
	conf oauth2.Config
	log  *logger.Log
}

type Handshake struct {
	State string `json:"state,omitempty"`
	From  string `json:"from,omitempty"`
	ID    string `json:"id,omitempty"`
}

//LoginRequest request
type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

//NewAuth constructor
func NewAuth(jwtSectret string, log *logger.Log) *Auth {
	return &Auth{
		jwt: NewJWT(jwtSectret),
		log: log,
	}
}

// HTTPHandler main handler
func (a *Auth) HTTPHandler(r chi.Router) {
	r.Post("/login", a.login)
	r.Post("/yandex/login", a.loginYandex)
	r.Get("/yandex/callback", a.loginYandexCallback)
	r.Post("/logout", a.logout)
	r.Group(func(r chi.Router) {
		r.Use(a.Authenticator)
		r.Route("/user", func(ra chi.Router) {
			ra.Get("/", a.user)
		})
	})
}

// Authenticator handles valid / invalid tokens. In this example, we use
// the provided authenticator middleware, but you can write your
// own very easily, look at the Authenticator method in jwtauth.go
// and tweak it, its not scary.
// r.Use(auth.Authenticator)
func (a *Auth) Authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		claims, _, err := a.jwt.Get(r)
		log.Printf("auth %v - %v", claims, err)
		if err != nil {
			switch err {
			default:
				http.Error(w, http.StatusText(401), 401)
				return
			case jwtauth.ErrExpired:
				http.Error(w, "expired", 401)
				return
			case jwtauth.ErrUnauthorized:
				http.Error(w, http.StatusText(401), 401)
				return
			case nil:
				// no error
			}
		}

		// Token is authenticated, pass it through
		rr := SetUserInfo(r, *claims.User)
		next.ServeHTTP(w, rr)
	})
}

//Check for test
func (a *Auth) user(w http.ResponseWriter, r *http.Request) {
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

func (a *Auth) login(w http.ResponseWriter, r *http.Request) {
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

	u := composeAuthUser(email)
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

func composeAuthUser(email string) User {
	return User{
		ID:         "local_" + email,
		Name:       email,
		Email:      email,
		PictureURL: "",
	}
}

func (a *Auth) loginYandex(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, "invalid request")
		return
	}
	request := &loginRequest{}
	err = json.Unmarshal(body, request)

	a.conf = oauth2.Config{
		ClientID:     os.Getenv("YANDEX_OAUTH2_ID"),
		ClientSecret: os.Getenv("YANDEX_OAUTH2_SECRET"),
		Scopes:       []string{},
		Endpoint:     yandex.Endpoint,
	}
	a.conf.RedirectURL = "http://localhost:9001/auth/yandex/callback"

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
func (a *Auth) loginYandexCallback(w http.ResponseWriter, r *http.Request) {
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
	infoURL := "https://login.yandex.ru/info?format=json"
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

	u := mapUserYandex(jData, data)

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

func (a *Auth) logout(w http.ResponseWriter, r *http.Request) {
	jwtCookie := http.Cookie{Name: JWTCookieName, Value: "", HttpOnly: false, Path: "/",
		MaxAge: -1, Expires: time.Unix(0, 0), Secure: SecureCookies}
	http.SetCookie(w, &jwtCookie)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// endpoint: yandex.Endpoint,
// 		scopes:   []string{},
// 		// See https://tech.yandex.com/passport/doc/dg/reference/response-docpage/
// 		infoURL: "https://login.yandex.ru/info?format=json",

// Value returns value for key or empty string if not found
func (u UserData) Value(key string) string {
	// json.Unmarshal converts json "null" value to go's "nil", in this case return empty string
	if val, ok := u[key]; ok && val != nil {
		return fmt.Sprintf("%v", val)
	}
	return ""
}

func mapUserYandex(data UserData, _ []byte) User {
	userInfo := User{
		ID:   "yandex_" + "token.HashID(sha1.New()" + data.Value("id"),
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
}
