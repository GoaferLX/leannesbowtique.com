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
	OrderView   *views.View
	MailService *MailController
}

func NewCartController(cs models.CartService, ms *MailController) *CartController {
	return &CartController{
		CartService: cs,
		CartView:    views.NewView("index.gohtml", "views/cart/cart.gohtml"),
		OrderView:   views.NewView("index.gohtml", "views/cart/order.gohtml"),
		MailService: ms,
	}
}

func CartFromContext(r *http.Request) *models.Cart {
	var cart *models.Cart
	cartctx := r.Context().Value("cart")
	cart, ok := cartctx.(*models.Cart)
	if !ok {
		cart = &models.Cart{}
		fmt.Println("Created a new cart")
	}
	return cart
}

type basketForm struct {
	ProductID int `schema:"productid"`
}

type orderForm struct {
	Email string `schema:"email"`
}

func (cc *CartController) AddToCart(w http.ResponseWriter, r *http.Request) {
	cart := CartFromContext(r)

	form := basketForm{}
	if err := parseGetForm(r, &form); err != nil {
		fmt.Println(err)
	}

	product := models.Product{ID: form.ProductID}
	var alert views.Alert
	if err := cc.CartService.AddItem(cart, product); err != nil {
		alert.Level = "Error"
		alert.Message = err.Error()
	} else {
		alert.Level = "Success"
		alert.Message = "Added to basket"
	}

	alert.PersistAlert(w)
	cc.CartService.AssignCookie(w, cart)
	http.Redirect(w, r, "/products", http.StatusFound)
	return
}

func (cc *CartController) ViewCart(w http.ResponseWriter, r *http.Request) {
	cart := CartFromContext(r)
	var yield views.Page
	yield.PageData = cart
	cc.CartView.RenderTemplate(w, r, yield)
}

func (cc *CartController) DeleteItem(w http.ResponseWriter, r *http.Request) {
	cart := CartFromContext(r)

	form := basketForm{}
	if err := parseGetForm(r, &form); err != nil {
		fmt.Println(err)
	}

	err := cc.CartService.DeleteItem(cart, form.ProductID)
	if err != nil {
		fmt.Println(err)
	}

	var yield views.Page
	yield.PageData = cart
	cc.CartView.RenderTemplate(w, r, yield)
}

func (cc *CartController) Empty(w http.ResponseWriter, r *http.Request) {
	cart := CartFromContext(r)
	cc.CartService.Empty(cart)

	var yield views.Page
	yield.PageData = cart
	cc.CartView.RenderTemplate(w, r, yield)
}

func (cc *CartController) Order(w http.ResponseWriter, r *http.Request) {
	cart := CartFromContext(r)

	var form orderForm
	var yield views.Page
	if err := parsePostForm(r, &form); err != nil {
		yield.SetAlert(err)
		cc.CartView.RenderTemplate(w, r, yield)
		return
	}
	if err := cc.CartService.PlaceOrder(cart, form.Email); err != nil {
		yield.SetAlert(fmt.Errorf("Something went wrong: %w", err))
		cc.CartView.RenderTemplate(w, r, yield)
		return
	}

	if err := cc.MailService.Order(form.Email, cart); err != nil {
		yield.SetAlert(err)
		cc.CartView.RenderTemplate(w, r, yield)
		return
	}

	if err := cc.MailService.OrderConfirm(form.Email, cart); err != nil {
		yield.SetAlert(err)
		cc.CartView.RenderTemplate(w, r, yield)
		return
	}

	alert := &views.Alert{Level: "Success", Message: "Order Placed: Please expect an email from us to confirm"}
	alert.PersistAlert(w)
	http.Redirect(w, r, "/products", http.StatusFound)

}
