package controllers

import (
	"fmt"
	"net/http"

	"leannesbowtique.com/models"
	"leannesbowtique.com/views"
)

type CartController struct {
	CartView *views.View
}

func NewCartController(cs models.CartService) *CartController {
	return &CartController{
		CartView: views.NewView("index.gohtml", "views/cart/cart.gohtml"),
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

	cart.Items = []models.LineItem{{models.Product{ID: 1, Name: "Bow 1", Price: 5.0}, 1}, {models.Product{ID: 2, Name: "Bow 2", Price: 10.0}, 1}, {models.Product{ID: 3, Name: "Bow 3", Price: 2.0}, 2}}
	//cart = &models.Cart{}
	yield.PageData = cart
	cc.CartView.RenderTemplate(w, r, yield)

}
