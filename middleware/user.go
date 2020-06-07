package middleware

import (
	"context"
	"goafweb/models"
	"net/http"
)

type User struct {
	UserModel models.UserService
}

func (mw *User) AllowFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rememberCookie, err := r.Cookie("rememberToken")
		if err != nil {
			next(w, r)
			return
		}
		user, err := mw.UserModel.GetUserByRemember(rememberCookie.Value)
		if err != nil {
			next(w, r)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, "user", user)
		r = r.WithContext(ctx)

		next(w, r)
	}
}
func (mw *User) Allow(next http.Handler) http.HandlerFunc {
	return mw.AllowFunc(next.ServeHTTP)
}
