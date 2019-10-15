package server

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/go-chi/chi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func startupAuthT(t *testing.T, secret string) (ts *httptest.Server, a *Auth, teardown func()) {

	auth := NewAuth(secret)
	router := chi.NewRouter()
	router.Route("/auth", auth.HTTPHandler)
	router.Route("/auth/check", func(rapi chi.Router) {
		rapi.Group(func(r chi.Router) {
			r.Use(auth.Verifier())
			r.Use(auth.Authenticator)
			r.Get("/", auth.Check)
		})
	})
	ts = httptest.NewServer(router)

	teardown = func() {
		ts.Close()
	}

	return ts, auth, teardown
}

func TestServerJWT(t *testing.T) {
	jwtSectret := "test"
	ts, a, teardown := startupAuthT(t, jwtSectret)
	defer teardown()

	h := http.Header{}
	authPayload := make(map[string]interface{})
	authPayload["user"] = "ma"
	tokenString := a.NewJwtToken(authPayload)
	cookieExpiration := int((time.Hour * 24).Seconds())
	jwtCookie := &http.Cookie{Name: "jwt", Value: tokenString, HttpOnly: true, Path: "/",
		MaxAge: cookieExpiration, Secure: false}

	status, resp := testRequest(t, ts, "GET", "/auth/check", h, jwtCookie, nil)

	log.Printf("test %v - %v", status, resp)
}

func TestLoginAPI(t *testing.T) {
	jwtSectret := "test"
	ts, _, teardown := startupAuthT(t, jwtSectret)
	defer teardown()

	r := strings.NewReader(`{"email":"test","password":"test"}`)
	client := http.Client{}
	req, err := http.NewRequest("POST", ts.URL+"/auth/login", r)
	assert.Nil(t, err)
	resp, err := client.Do(req)

	require.Nil(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	assert.Equal(t, 1, len(resp.Cookies()))
	assert.Equal(t, "jwt", resp.Cookies()[0].Name, "test")

	log.Printf("[INFO] login token: %v ", resp.Cookies())
}

func TestLoginFailedAPI(t *testing.T) {
	jwtSectret := "test"
	ts, _, teardown := startupAuthT(t, jwtSectret)
	defer teardown()

	r := strings.NewReader(`{"email":"test","password":""}`)
	client := http.Client{}
	req, err := http.NewRequest("POST", ts.URL+"/auth/login", r)
	assert.Nil(t, err)
	resp, err := client.Do(req)
	require.Nil(t, err)
	assert.Equal(t, 401, resp.StatusCode)
}

func testRequest(t *testing.T, ts *httptest.Server, method, path string, header http.Header, cookies *http.Cookie, body io.Reader) (int, string) {
	req, err := http.NewRequest(method, ts.URL+path, body)
	if err != nil {
		t.Fatal(err)
		return 0, ""
	}

	if header != nil {
		for k, v := range header {
			req.Header.Set(k, v[0])
		}
	}

	if cookies != nil {
		req.AddCookie(cookies)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
		return 0, ""
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
		return 0, ""
	}
	defer resp.Body.Close()

	return resp.StatusCode, string(respBody)
}
