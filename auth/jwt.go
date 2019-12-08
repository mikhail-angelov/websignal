package auth

import (
	"crypto/rand"
	"crypto/sha1"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/pkg/errors"
)

//Claims jwt
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

//JWT service
type JWT struct {
	jwtSectret string
}

const (
	JWTCookieName = "jwt"
	JWTHeaderKey  = "X-JWT"
	JWTQuery      = "token"
	TokenDuration = 24 * time.Hour
	Issuer        = "websignal"
	SecureCookies = false
)

//NewJWT creates a new JWT service
func NewJWT(jwtSectret string) *JWT {
	return &JWT{jwtSectret}
}

//Get claims from http request
func (j *JWT) Get(r *http.Request) (Claims, string, error) {

	fromCookie := false
	tokenString := ""

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

	claims, err := j.Parse(tokenString)
	if err != nil {
		return Claims{}, "", errors.Wrap(err, "failed to get token")
	}

	if !fromCookie && j.IsExpired(claims) {
		return Claims{}, "", errors.New("token expired")
	}

	return claims, tokenString, nil
}

// Parse token string and verify. Not checking for expiration
func (j *JWT) Parse(tokenString string) (Claims, error) {
	parser := jwt.Parser{SkipClaimsValidation: true} // allow parsing of expired tokens

	token, err := parser.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(j.jwtSectret), nil
	})
	if err != nil {
		return Claims{}, errors.Wrap(err, "can't parse token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return Claims{}, errors.New("invalid token")
	}

	return *claims, j.validate(claims)
}

func (j *JWT) validate(claims *Claims) error {
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
func (j *JWT) IsExpired(claims Claims) bool {
	return !claims.VerifyExpiresAt(time.Now().Unix(), true)
}

//NewJwtToken generate new token from payload
func (j *JWT) NewJwtToken(claims Claims) string {
	token := jwt.New(jwt.GetSigningMethod("HS256"))
	token.Claims = claims
	tokenStr, err := token.SignedString([]byte(j.jwtSectret))
	if err != nil {
		log.Fatal(err)
	}
	return tokenStr
}

// Set creates token cookie and put it to ResponseWriter
// accepts claims and sets expiration if none defined.
// permanent flag means long-living cookie, false makes it session only.
func (j *JWT) Set(w http.ResponseWriter, claims Claims) (Claims, error) {
	if claims.ExpiresAt == 0 {
		claims.ExpiresAt = time.Now().Add(TokenDuration).Unix()
	}

	if claims.Issuer == "" {
		claims.Issuer = Issuer
	}

	tokenString, err := j.Token(claims)
	if err != nil {
		return Claims{}, errors.Wrap(err, "failed to make token token")
	}

	cookieExpiration := 0 // session cookie

	jwtCookie := http.Cookie{Name: JWTCookieName, Value: tokenString, HttpOnly: false, Path: "/",
		MaxAge: cookieExpiration, Secure: SecureCookies}
	http.SetCookie(w, &jwtCookie)

	return claims, nil
}

// Clean jwt auth from response
func (j *JWT) Clean(w http.ResponseWriter) {
	jwtCookie := http.Cookie{Name: JWTCookieName, Value: "", HttpOnly: false, Path: "/",
		MaxAge: -1, Expires: time.Unix(0, 0), Secure: SecureCookies}
	http.SetCookie(w, &jwtCookie)
}

// Token makes token with claims
func (j *JWT) Token(claims Claims) (string, error) {

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString([]byte(j.jwtSectret))
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
