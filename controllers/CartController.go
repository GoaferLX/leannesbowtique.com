package controllers

import (
	"fmt"
	"net/http"

	"leannesbowtique.com/models"
	"leannesbowtique.com/views"
)

type CartController struct {
	CartService models.CartService
	CartView    *views.View
}

func NewCartController(cs models.CartService) *CartController {
	return &CartController{
		CartService: cs,
		CartView:    views.NewView("index.gohtml", "views/cart/cart.gohtml"),
	}
}

func (cc *CartController) ViewCart(w http.ResponseWriter, r *http.Request) {
	var cart *models.Cart
	ctx := r.Context()
	cartctx := ctx.Value("cart")
	cart, ok := cartctx.(*models.Cart)
	if !ok {
		cart = &models.Cart{}
		fmt.Println("Created a new cart")
	}
	var yield views.Page

	yield.PageData = cart
	cc.CartView.RenderTemplate(w, r, yield)

}
