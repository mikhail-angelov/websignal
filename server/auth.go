package server

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-chi/chi"
	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
)

// Auth service
type Auth struct {
	jwtSectret string
	tokenAuth  *jwtauth.JWTAuth
}

//LoginRequest request
type LoginRequest struct {
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
	r.Get("/login/callback", a.loginCallback)
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
	request := &LoginRequest{}
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
	jwtCookie := &http.Cookie{Name: "jwt", Value: tokenString, HttpOnly: true, Path: "/",
		MaxAge: cookieExpiration, Secure: false}
	http.SetCookie(w, jwtCookie)
	render.Status(r, http.StatusOK)
}
func (a *Auth) loginCallback(w http.ResponseWriter, r *http.Request) {

}

func (a *Auth) logout(w http.ResponseWriter, r *http.Request) {

}
