package server

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPublicConsoleHandlerAllowList(t *testing.T) {
	cases := []struct {
		path    string
		methods []string
		want    bool
	}{
		{"/v2/console/base/brand/public", []string{http.MethodGet}, true},
		{"/v2/console/base/brand/public", []string{http.MethodGet, http.MethodHead}, true},
		{"/v2/console/base/brand/public", []string{http.MethodPut}, false},
		{"/v2/console/base/brand/public", []string{http.MethodGet, http.MethodPost}, false},
		{"/v2/console/base/brand/public", nil, false},
		{"/v2/console/base/brand", []string{http.MethodGet}, false},
		{"/v2/console/base/brand/public/x", []string{http.MethodGet}, false},
		{"/v2/console/economy/summary", []string{http.MethodGet}, false},
	}
	for _, c := range cases {
		if got := isPublicConsoleHandler(c.path, c.methods); got != c.want {
			t.Errorf("isPublicConsoleHandler(%q, %v) = %v, want %v", c.path, c.methods, got, c.want)
		}
	}
}

func TestPublicConsoleHandlerDropsIdentity(t *testing.T) {
	var user, auth string
	h := publicConsoleHandler(func(w http.ResponseWriter, r *http.Request) {
		user = r.Header.Get(consoleUsernameHeader)
		auth = r.Header.Get("Authorization")
	})
	req := httptest.NewRequest(http.MethodGet, "/v2/console/base/brand/public", nil)
	req.Header.Set(consoleUsernameHeader, "spoofed")
	req.Header.Set("Authorization", "Bearer x")
	h(httptest.NewRecorder(), req)
	if user != "" || auth != "" {
		t.Fatalf("public handler must not see identity headers: %q %q", user, auth)
	}
}
