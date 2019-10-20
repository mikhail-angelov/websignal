package server

import (
	"context"
	"crypto/rand"
	"crypto/sha1"
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
	"github.com/pkg/errors"
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
	jwtSectret string
	tokenAuth  *jwtauth.JWTAuth
	conf       oauth2.Config
}

type Claims struct {
	jwt.StandardClaims
	User        *User      `json:"user,omitempty"` // user info
	SessionOnly bool       `json:"sess_only,omitempty"`
	Handshake   *Handshake `json:"handshake,omitempty"` // used for oauth handshake
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
func NewAuth(jwtSectret string) *Auth {
	return &Auth{
		jwtSectret: jwtSectret,
		tokenAuth:  jwtauth.New("HS256", []byte(jwtSectret), nil),
	}
}

// HTTPHandler main handler
func (a *Auth) HTTPHandler(r chi.Router) {
	r.Post("/login", a.login)
	r.Get("/callback", a.loginCallback)
	r.Post("/yandex/login", a.loginYandex)
	r.Get("/yandex/callback", a.loginYandexCallback)
	r.Post("/logout", a.logout)
}

// Verifier Seek, verify and validate JWT tokens
func (a *Auth) Verifier() func(http.Handler) http.Handler {
	return jwtauth.Verifier(a.tokenAuth)
}

// Authenticator handles valid / invalid tokens. In this example, we use
// the provided authenticator middleware, but you can write your
// own very easily, look at the Authenticator method in jwtauth.go
// and tweak it, its not scary.
// r.Use(auth.Authenticator)
func (a *Auth) Authenticator(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, _, err := jwtauth.FromContext(r.Context())
		log.Printf("auth %v - %v", token, err)
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

		if token == nil || !token.Valid {
			http.Error(w, http.StatusText(401), 401)
			return
		}

		// Token is authenticated, pass it through
		next.ServeHTTP(w, r)
	})
}

//NewJwtToken generate new token from payload
func (a *Auth) NewJwtToken(claims ...jwt.MapClaims) string {
	token := jwt.New(jwt.GetSigningMethod("HS256"))
	if len(claims) > 0 {
		token.Claims = claims[0]
	}
	tokenStr, err := token.SignedString([]byte(a.jwtSectret))
	if err != nil {
		log.Fatal(err)
	}
	return tokenStr
}

//Check for test
func (a *Auth) Check(w http.ResponseWriter, r *http.Request) {

}

func (a *Auth) login(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.PlainText(w, r, "invalid request")
		return
	}
	request := &loginRequest{}
	err = json.Unmarshal(body, request)
	log.Printf("[INFO] login %v", request)

	//todo, check password and handle validation
	if request.Password == "" {
		render.Status(r, http.StatusUnauthorized)
		render.PlainText(w, r, "invalid login")
		return
	}
	authPayload := make(map[string]interface{})
	authPayload["email"] = request.Email
	tokenString := a.NewJwtToken(authPayload)
	cookieExpiration := int((time.Hour * 24).Seconds())
	jwtCookie := &http.Cookie{Name: JWTCookieName, Value: tokenString, HttpOnly: true, Path: "/",
		MaxAge: cookieExpiration, Secure: false}
	http.SetCookie(w, jwtCookie)
	render.Status(r, http.StatusOK)
}
func (a *Auth) loginCallback(w http.ResponseWriter, r *http.Request) {

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
		log.Printf("[DEBUG] failed to make claim's id %v", err)
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

	if _, err := a.Set(w, claims); err != nil {
		log.Printf("[DEBUG] failed to set token %v", err)
		return
	}
	loginURL := a.conf.AuthCodeURL(rd)
	log.Printf("[INFO] yandex login %s %s", loginURL, rd)
	http.Redirect(w, r, loginURL, http.StatusFound)
}
func (a *Auth) loginYandexCallback(w http.ResponseWriter, r *http.Request) {
	oauthClaims, token, err := a.Get(r)
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

	u := mapUser(jData, data)

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

	if _, err = a.Set(w, claims); err != nil {
		log.Printf("[DEBUG] failed to set token %v", err)
		return
	}

	log.Printf("[INFO] login success %v, %s %v", claims, token, u)
	render.Status(r, http.StatusOK)
}

func (a *Auth) logout(w http.ResponseWriter, r *http.Request) {

}

func (a *Auth) Get(r *http.Request) (Claims, string, error) {

	fromCookie := false
	tokenString := ""

	log.Printf("[WARN] cb %v %v", r.URL.Query(), r.Header)

	// try to get from "token" query param
	if tkQuery := r.URL.Query().Get(JWTQuery); tkQuery != "" {
		tokenString = tkQuery
	}

	// try to get from JWT header
	if tokenHeader := r.Header.Get(JWTHeaderKey); tokenHeader != "" && tokenString == "" {
		tokenString = tokenHeader
	}

	// try to get from JWT cookie
	if tokenString == "" {
		fromCookie = true
		jc, err := r.Cookie(JWTCookieName)
		if err != nil {
			return Claims{}, "", err
		}
		tokenString = jc.Value
	}

	claims, err := a.Parse(tokenString)
	if err != nil {
		return Claims{}, "", errors.Wrap(err, "failed to get token")
	}

	if !fromCookie && a.IsExpired(claims) {
		return Claims{}, "", errors.New("token expired")
	}

	return claims, tokenString, nil
}

// Parse token string and verify. Not checking for expiration
func (a *Auth) Parse(tokenString string) (Claims, error) {
	parser := jwt.Parser{SkipClaimsValidation: true} // allow parsing of expired tokens

	token, err := parser.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(a.jwtSectret), nil
	})
	if err != nil {
		return Claims{}, errors.Wrap(err, "can't parse token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return Claims{}, errors.New("invalid token")
	}

	return *claims, a.validate(claims)
}

func (a *Auth) validate(claims *Claims) error {
	cerr := claims.Valid()

	if cerr == nil {
		return nil
	}

	if e, ok := cerr.(*jwt.ValidationError); ok {
		if e.Errors == jwt.ValidationErrorExpired {
			return nil // allow expired tokens
		}
	}

	return cerr
}

// IsExpired returns true if claims expired
func (a *Auth) IsExpired(claims Claims) bool {
	return !claims.VerifyExpiresAt(time.Now().Unix(), true)
}

// endpoint: yandex.Endpoint,
// 		scopes:   []string{},
// 		// See https://tech.yandex.com/passport/doc/dg/reference/response-docpage/
// 		infoURL: "https://login.yandex.ru/info?format=json",

type User struct {
	// set by service
	Name     string `json:"name"`
	ID       string `json:"id"`
	Picture  string `json:"picture"`
	Audience string `json:"aud,omitempty"`

	// set by client
	IP         string                 `json:"ip,omitempty"`
	Email      string                 `json:"email,omitempty"`
	Attributes map[string]interface{} `json:"attrs,omitempty"`
}
type UserData map[string]interface{}

// Value returns value for key or empty string if not found
func (u UserData) Value(key string) string {
	// json.Unmarshal converts json "null" value to go's "nil", in this case return empty string
	if val, ok := u[key]; ok && val != nil {
		return fmt.Sprintf("%v", val)
	}
	return ""
}

func mapUser(data UserData, _ []byte) User {
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
		userInfo.Picture = fmt.Sprintf("https://avatars.yandex.net/get-yapic/%s/islands-200", data.Value("default_avatar_id"))
	}
	return userInfo
}

// Set creates token cookie with xsrf cookie and put it to ResponseWriter
// accepts claims and sets expiration if none defined. permanent flag means long-living cookie,
// false makes it session only.
func (a *Auth) Set(w http.ResponseWriter, claims Claims) (Claims, error) {
	if claims.ExpiresAt == 0 {
		claims.ExpiresAt = time.Now().Add(TokenDuration).Unix()
	}

	if claims.Issuer == "" {
		claims.Issuer = Issuer
	}

	tokenString, err := a.Token(claims)
	if err != nil {
		return Claims{}, errors.Wrap(err, "failed to make token token")
	}

	cookieExpiration := 0 // session cookie

	jwtCookie := http.Cookie{Name: JWTCookieName, Value: tokenString, HttpOnly: true, Path: "/",
		MaxAge: cookieExpiration, Secure: SecureCookies}
	http.SetCookie(w, &jwtCookie)

	return claims, nil
}

// Token makes token with claims
func (a *Auth) Token(claims Claims) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(a.jwtSectret))
	if err != nil {
		return "", errors.Wrap(err, "can't sign token")
	}
	return tokenString, nil
}

func randToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", errors.Wrap(err, "can't get random")
	}
	s := sha1.New()
	if _, err := s.Write(b); err != nil {
		return "", errors.Wrap(err, "can't write randoms to sha1")
	}
	return fmt.Sprintf("%x", s.Sum(nil)), nil
}
