package user_server

import (
	"net/http"
)

func AuthenticationRequirement(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// TODO
	})
}

func isTokenValid(token string) bool {
	// TODO
	return true
}
