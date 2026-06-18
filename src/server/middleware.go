package server

import (
	"crypto/subtle"
	"net/http"
)

type Middleware func(http.Handler) http.Handler

func BasicAuth(username, password string) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
			user, pass, ok := r.BasicAuth()
			if !ok ||
				subtle.ConstantTimeCompare([]byte(user), []byte(username)) != 1 ||
				subtle.ConstantTimeCompare([]byte(pass), []byte(password)) != 1 {
				rw.Header().Set("WWW-Authenticate", `Basic realm="Finance Assistant"`)
				http.Error(rw, "unauthorized", http.StatusUnauthorized)
				return
			}
			next.ServeHTTP(rw, r)
		})
	}
}
