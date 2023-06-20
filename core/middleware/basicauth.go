package middleware

import "net/http"

func BasicAuthAdmin(next http.HandlerFunc, username, password string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		givenUsername, givenPassword, ok := r.BasicAuth()

		if !ok || givenUsername != username || givenPassword != password {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			http.Error(w, "Unauthorized", http.StatusUnauthorized)

			return
		}

		next.ServeHTTP(w, r)
	}
}
