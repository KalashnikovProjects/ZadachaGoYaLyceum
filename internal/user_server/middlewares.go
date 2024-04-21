package user_server

import (
	"context"
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/auth"
	"net/http"
)

func CheckTokenCookie(next http.Handler, exclude []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		logged := true
		cookie, err := r.Cookie("token")
		if err != nil {
			logged = false
		} else if _, err := auth.LoadUserIdFromToken(cookie.Value); err != nil {
			logged = false
		}

		r = r.WithContext(context.WithValue(r.Context(), "logged", logged))
		for _, ex := range exclude {
			if ex == r.URL.Path {
				next.ServeHTTP(w, r)
				return
			}
		}

		if !logged {
			http.Redirect(w, r, "/login?redirect="+r.URL.Path, http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
