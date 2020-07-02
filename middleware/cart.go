package middleware

import (
	"context"
	"fmt"
	"net/http"

	"leannesbowtique.com/models"
)

type CartMW struct {
	models.CartService
}

func (cmw CartMW) CheckCart(next http.Handler) http.HandlerFunc {
	return cmw.CheckCartFunc(next.ServeHTTP)
}

func (cmw CartMW) CheckCartFunc(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		/*
			cartcookie, err := r.Cookie("cartid")
			if err != nil {
				fmt.Println(err)
				next(w, r)
				return
			}
			cartid, err := strconv.Atoi(cartcookie.Value)
			if err != nil {
				fmt.Println(err)
				next(w, r)
				return
			}
			_ = cartid
		*/
		cart, err := cmw.CartService.GetCart(1)
		if err != nil {
			fmt.Println(err)
			next(w, r)
			return
		}
		ctx := r.Context()
		ctx = context.WithValue(ctx, "cart", cart)
		r = r.WithContext(ctx)
		next(w, r)
	}
}
