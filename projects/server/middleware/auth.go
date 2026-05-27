package middleware

import (
	"net/http"

	"issue_tracker/utils"
)

func RequireAPIKey(expectedKey string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("x-api-key") != expectedKey {
				utils.WriteError(w, utils.NewUnauthorized())
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}
