package middleware

import (
	"net/http"
	"strings"

	"leannesbowtique.com/models"
	"leannesbowtique.com/views"
)

type RequireUser struct {
}

func (mw *RequireUser) AllowFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		userctx := ctx.Value("user")
		user, _ := userctx.(*models.User)

		if user == nil {
			alert := &views.Alert{
				Level:   "Warning",
				Message: "You must be signed in to access that page",
			}
			alert.PersistAlert(w)
			dest := strings.TrimLeft(r.RequestURI, "/")
			url := "/login?dest=" + dest
			http.Redirect(w, r, url, http.StatusFound)
			return
		}
		next(w, r)
	}
}
func (mw *RequireUser) AllowHandle(next http.Handler) http.HandlerFunc {
	return mw.AllowFunc(next.ServeHTTP)
}
