package server

import (
	"net/http"
)

var publicConsolePaths = map[string]bool{
	"/v2/console/base/brand/public": true,
}

func isPublicConsoleHandler(pathPattern string, methods []string) bool {
	if !publicConsolePaths[pathPattern] || len(methods) == 0 {
		return false
	}
	for _, m := range methods {
		if m != http.MethodGet && m != http.MethodHead {
			return false
		}
	}
	return true
}

func publicConsoleHandler(handler func(http.ResponseWriter, *http.Request)) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		r.Header.Del(consoleUsernameHeader)
		r.Header.Del(consoleAdminHeader)
		r.Header.Del("Authorization")
		handler(w, r)
	}
}
