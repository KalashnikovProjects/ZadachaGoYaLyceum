package user_server

import (
	"github.com/KalashnikovProjects/ZadachaGoYaLyceum/internal/auth"
	"net/http"
)

func CheckTokenCookie(next http.Handler, exclude []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, ex := range exclude {
			if ex == r.URL.Path {
				next.ServeHTTP(w, r)
				return
			}
		}
		cookie, err := r.Cookie("token")
		if err != nil {
			http.Redirect(w, r, "/login?redirect="+r.URL.Path, http.StatusSeeOther)
			return
		}
		if _, err := auth.LoadUserIdFromToken(cookie.Value); err != nil {
			http.Redirect(w, r, "/login?redirect="+r.URL.Path, http.StatusSeeOther)
			return
		}
		next.ServeHTTP(w, r)
	})
}
