package auth

import (
	"bytes"
	"context"
	"net/http"

	"github.com/nullrocks/identicon"
	"github.com/pkg/errors"
)

// User is a structure to share user data between backend and front end
type User struct {
	// set by service
	Name       string `json:"name"`
	ID         string `json:"id"`
	Picture    []byte `json:"picture,omitempty"`
	PictureURL string `json:"pictureUrl,omitempty"`
	Audience   string `json:"aud,omitempty"`

	// set by client
	IP         string                 `json:"ip,omitempty"`
	Email      string                 `json:"email,omitempty"`
	Attributes map[string]interface{} `json:"attrs,omitempty"`
}

// UserData - temp type to parse user info
type UserData map[string]interface{}

type contextKey string

// GetUserInfo returns user info from request context
func GetUserInfo(r *http.Request) (user User, err error) {

	ctx := r.Context()
	if ctx == nil {
		return User{}, errors.New("no info about user")
	}
	if u, ok := ctx.Value(contextKey("user")).(User); ok {
		return u, nil
	}

	return User{}, errors.New("user can't be parsed")
}

// SetUserInfo sets user into request context
func SetUserInfo(r *http.Request, user User) *http.Request {
	ctx := r.Context()
	ctx = context.WithValue(ctx, contextKey("user"), user)
	return r.WithContext(ctx)
}

// GenerateAvatar for give user with identicon
func GenerateAvatar(user string) ([]byte, error) {

	iconGen, err := identicon.New("pkgz/auth", 5, 5)
	if err != nil {
		return nil, errors.Wrap(err, "can't create identicon service")
	}

	ii, err := iconGen.Draw(user) // generate an IdentIcon
	if err != nil {
		return nil, errors.Wrapf(err, "failed to draw avatar for %s", user)
	}

	buf := &bytes.Buffer{}
	err = ii.Png(300, buf)
	return buf.Bytes(), err
}
