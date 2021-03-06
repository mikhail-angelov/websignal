package auth

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/go-pkgz/rest"
	"github.com/mikhail-angelov/websignal/logger"
	"github.com/pkg/errors"
	"golang.org/x/oauth2"
)

// Auth service
type Auth struct {
	jwt       *JWT
	conf      oauth2.Config
	providers []*Auth2Provider
	log       *logger.Log
	url       string // root url for the rest service, i.e. http://blah.example.com, required
}

type loginRequest1 struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

//NewAuth constructor
func NewAuth(jwtSectret string, log *logger.Log, url string) *Auth {
	return &Auth{
		jwt: NewJWT(jwtSectret),
		log: log,
		url: url,
	}
}

// Handlers gets http.Handler for all providers
// it process urls: auth/logout, auth/user, auth/<provider name>/<any>
func (a *Auth) Handlers() (authHandler http.Handler) {

	ah := func(w http.ResponseWriter, r *http.Request) {
		elems := strings.Split(r.URL.Path, "/")
		if len(elems) < 2 {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		command := elems[len(elems)-1]

		// allow logout without specifying provider
		if command == "logout" {
			if len(a.providers) == 0 {
				w.WriteHeader(http.StatusBadRequest)
				rest.RenderJSON(w, r, rest.JSON{"error": "provides not defined"})
				return
			}
			a.providers[0].Handler(w, r)
			return
		}

		// show user info
		if command == "user" {
			claims, _, err := a.jwt.Get(r)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				rest.RenderJSON(w, r, rest.JSON{"error": err.Error()})
				return
			}
			if claims.User.PictureURL == "" {
				pic, err := GenerateAvatar(claims.User.Email)
				if err != nil {
					log.Printf("failed to gen avatar %v", err)
				}
				claims.User.Picture = pic
			}
			rest.RenderJSON(w, r, claims.User)
			return
		}

		// regular auth handlers
		provName := elems[len(elems)-2]
		p, err := a.getProviderByName(provName)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			rest.RenderJSON(w, r, rest.JSON{"error": fmt.Sprintf("provider %s not supported", provName)})
			return
		}
		p.Handler(w, r)
	}

	return http.HandlerFunc(ah)
}

// Auth handles valid / invalid tokens. In this example, we use
// the provided authenticator middleware, but you can write your
// own very easily, look at the Authenticator method in jwtauth.go
// and tweak it, its not scary.
// r.Use(auth.Auth)
func (a *Auth) Auth(next http.Handler) http.Handler {
	onError := func(w http.ResponseWriter, r *http.Request, err error) {
		a.log.Logf("[DEBUG] auth failed, %v", err)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
	}

	fn := func(w http.ResponseWriter, r *http.Request) {
		claims, tkn, err := a.jwt.Get(r)
		if err != nil {
			onError(w, r, errors.Wrap(err, "can't get token"))
			return
		}

		if claims.Handshake != nil { // handshake in token indicate special use cases, not for login
			onError(w, r, errors.New("invalid kind of token"))
			return
		}

		if claims.User == nil {
			onError(w, r, errors.New("no user info presented in the claim"))
			return
		}

		if claims.User != nil { // if user in token populate it to context

			if a.jwt.IsExpired(claims) {
				if claims, err = a.refreshExpiredToken(w, claims, tkn); err != nil {
					a.jwt.Clean(w)
					onError(w, r, errors.Wrap(err, "can't refresh token"))
					return
				}
			}

			r = SetUserInfo(r, *claims.User) // populate user info to request context
		}

		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

// ValidateToken for WS and return claims ID
func (a *Auth) ValidateToken(tokenString string) (*User, error) {
	claims, err := a.jwt.Parse(tokenString)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get token")
	}

	if a.jwt.IsExpired(claims) {
		return nil, errors.Wrap(err, "token expired")
	}
	log.Printf("success auth  %v", claims.User.ID)
	if claims.User.PictureURL == "" {
		pic, err := GenerateAvatar(claims.User.Email)
		if err != nil {
			log.Printf("failed to gen avatar %v", err)
		}
		claims.User.Picture = pic
	}

	return claims.User, nil
}

// refreshExpiredToken makes a new token with passed claims
func (a *Auth) refreshExpiredToken(w http.ResponseWriter, claims Claims, tkn string) (Claims, error) {

	// cache refreshed claims for given token in order to eliminate multiple refreshes for concurrent requests
	// if a.RefreshCache != nil {
	// 	if c, ok := a.RefreshCache.Get(tkn); ok {
	// 		// already in cache
	// 		return c.(token.Claims), nil
	// 	}
	// }

	claims.ExpiresAt = 0           // this will cause now+duration for refreshed token
	c, err := a.jwt.Set(w, claims) // Set changes token
	if err != nil {
		return Claims{}, err
	}

	// if a.RefreshCache != nil {
	// 	a.RefreshCache.Set(tkn, c)
	// }

	a.log.Logf("[DEBUG] token refreshed for %+v", claims.User)
	return c, nil
}

// AddProvider add new auth2 provider
func (a *Auth) AddProvider(name, cid, secret string) {
	provider := NewAuth2Provider(&Auth2ProviderParams{name: name, cid: cid, secret: secret, jwt: a.jwt, log: a.log, url: a.url})
	a.providers = append(a.providers, provider)
}

func (a *Auth) getProviderByName(name string) (*Auth2Provider, error) {
	for _, p := range a.providers {
		if p.Name() == name {
			return p, nil
		}
	}
	return nil, errors.Errorf("provider %s not found", name)
}
