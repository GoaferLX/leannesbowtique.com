package models

import (
	"errors"
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"time"

	"github.com/jinzhu/gorm"
)

type Cart struct {
	ID        int `gorm:"primary_key;"`
	Items     []CartItem
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
	Ordered   bool
}

type CartItem struct {
	CartID    int     `gorm:"primary_key;auto_increment:false"`
	ProductID int     `gorm:"primary_key;auto_increment:false"`
	Product   Product `gorm:"association_autoupdate:false;association_autocreate:false"`
	Quantity  int     `gorm:"not null;"`
}

func (cart Cart) Subtotal() float64 {
	var total float64
	for _, cartItem := range cart.Items {
		total += cartItem.Product.Price * float64(cartItem.Quantity)
	}
	return total
}

func (cart Cart) Total() float64 {
	total := cart.Subtotal() + cart.DeliveryCharge()
	return total
}

func (cart Cart) DeliveryCharge() float64 {
	var delivery float64
	if subtotal := cart.Subtotal(); subtotal < 15 {
		delivery = 1.50
	}
	return delivery
}

type CartService interface {
	CartDB
	OrderService
	NewCart() (*Cart, error)
	AddItem(*Cart, Product) error
	DeleteItem(*Cart, int) error
	Empty(*Cart) error
	AssignCookie(http.ResponseWriter, *Cart)
	//Order(*Cart) error
}

type cartService struct {
	CartDB
}

func NewCartService(db *gorm.DB) CartService {
	return &cartService{
		CartDB: &cartDB{
			gorm: db,
		},
	}
}
func (cs *cartService) AssignCookie(w http.ResponseWriter, cart *Cart) {
	cartid := strconv.Itoa(cart.ID)
	cartCookie := http.Cookie{
		Name:     "cartID",
		Value:    cartid,
		Expires:  time.Now().Add(7 * 24 * time.Hour),
		Path:     "/",
		HttpOnly: true,
		//	Secure:   true,
	}
	http.SetCookie(w, &cartCookie)
}
func (cs *cartService) NewCart() (*Cart, error) {
	cart := &Cart{}
	return cart, nil
}

func (cs *cartService) AddItem(cart *Cart, product Product) error {
	for i, cartItem := range cart.Items {
		if cartItem.Product.ID == product.ID {
			cart.Items[i].Quantity++
			return cs.CartDB.Update(cart)
		}
	}
	cart.Items = append(cart.Items, CartItem{Product: product, Quantity: 1})
	if cart.ID == 0 {
		return cs.CartDB.Create(cart)
	}
	return cs.CartDB.Update(cart)
}

func (cs *cartService) DeleteItem(cart *Cart, productid int) error {

	for i, cartItem := range cart.Items {
		if cartItem.Product.ID == productid {
			if cartItem.Quantity > 1 {
				cart.Items[i].Quantity--
				return cs.CartDB.Update(cart)
			}
			cart.Items = append(cart.Items[:i], cart.Items[i+1:]...)
		}
	}

	return cs.CartDB.Update(cart)
}
func (cs *cartService) Empty(cart *Cart) error {
	cart.Items = []CartItem{}
	return cs.CartDB.Update(cart)
}

type CartDB interface {
	Create(*Cart) error
	Update(*Cart) error
	Delete(*Cart) error
	GetCart(id int) (*Cart, error)
}

type cartDB struct {
	gorm *gorm.DB
}

func (db *cartDB) Create(cart *Cart) error {
	return db.gorm.Create(cart).Error
}

func (db *cartDB) GetCart(id int) (*Cart, error) {
	var cart Cart
	if err := db.gorm.Preload("Items.Product").First(&cart, id).Error; err != nil {
		return nil, err
	}
	return &cart, nil
}

func (db *cartDB) Update(cart *Cart) error {
	var ids []int
	for _, product := range cart.Items {
		ids = append(ids, product.ProductID)
	}
	if len(ids) == 0 {
		db.gorm.Exec("DELETE FROM cart_items WHERE cart_id IN (?)", cart.ID)
	}
	db.gorm.Exec("DELETE FROM cart_items WHERE cart_id IN (?) AND product_id NOT IN (?)", cart.ID, ids)
	if err := db.gorm.Save(cart).Error; err != nil {
		return err
	}
	return nil
}

func (db *cartDB) Delete(cart *Cart) error {
	if err := db.gorm.Delete(cart).Error; err != nil {
		return err
	}
	return nil
}

func (cs *cartService) PlaceOrder(cart *Cart, email string) error {
	if len(cart.Items) == 0 {
		return errors.New("Your cart is empty")
	}
	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("Not a valid email address")
	}
	cart.Ordered = true
	if err := cs.CartDB.Update(cart); err != nil {
		return fmt.Errorf("Could not place order: %w", err)
	}
	if err := cs.CartDB.Delete(cart); err != nil {
		return fmt.Errorf("Could not place order: %w", err)
	}
	return nil
}

type OrderService interface {
	PlaceOrder(*Cart, string) error
}
