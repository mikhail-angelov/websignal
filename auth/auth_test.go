package auth

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/mikhail-angelov/websignal/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func startupAuthT(t *testing.T, secret string) (ts *httptest.Server, a *Auth, teardown func()) {
	log := logger.New()
	auth := NewAuth(secret, log)
	router := chi.NewRouter()
	router.Route("/auth", auth.HTTPHandler)
	router.Get("/auth/check", verifyLogin(t))

	ts = httptest.NewServer(router)

	teardown = func() {
		ts.Close()
	}

	return ts, auth, teardown
}

func TestLoginAPI(t *testing.T) {
	jwtSectret := "test"
	ts, _, teardown := startupAuthT(t, jwtSectret)
	defer teardown()

	r := strings.NewReader(`{"email":"test","password":"test"}`)
	client := http.Client{}
	req, err := http.NewRequest("POST", ts.URL+"/auth/login?from=/auth/check", r)
	assert.Nil(t, err)
	_, err = client.Do(req)
	require.Nil(t, err)

	assert.Equal(t, 200, http.StatusOK)
}

func verifyLogin(t *testing.T) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		cookies := r.Header.Get("Cookie")
		name := strings.Split(cookies, "=")[0]
		assert.Equal(t, "jwt", name)

		log.Printf("[INFO] login token: %v ", cookies)
		render.Status(r, http.StatusOK)
		render.PlainText(w, r, "ok")
	}
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
