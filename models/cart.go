package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type CartItem struct {
	CartID    int `gorm:"primary_key;auto_increment:false"`
	ProductID int `gorm:"primary_key;auto_increment:false"`
	Product   Product
	Quantity  int `gorm:"not null;"`
}
type Cart struct {
	ID        int `gorm:"primary_key;"`
	Items     []CartItem
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
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

func (cart *Cart) AddItem(product Product) error {
	for _, cartItem := range cart.Items {
		if cartItem.Product.ID == product.ID {
			cartItem.Quantity++
		} else {
			cart.Items = append(cart.Items, CartItem{Product: product, Quantity: 1})
		}
	}
	return nil
}

func (cart *Cart) DeleteItem(product *Product) error {
	for i, cartItem := range cart.Items {
		if cartItem.Product.ID == product.ID {
			cart.Items = append(cart.Items[:i], cart.Items[i+1:]...)
		}
	}
	return nil
}
func (cart *Cart) Empty() error {
	cart.Items = []CartItem{}
	return nil
}

type CartService interface {
	CartDB
}
type CartDB interface {
	NewCart() (*Cart, error)
	Update(*Cart) error
	Delete(*Cart) error
	GetCart(id int) (*Cart, error)
}
type cartDB struct {
	gorm *gorm.DB
}

func NewCartService(db *gorm.DB) CartService {
	return &cartDB{
		gorm: db,
	}
}

func (db cartDB) NewCart() (*Cart, error) {
	cart := Cart{}
	if err := db.gorm.Create(&cart).Error; err != nil {
		return nil, err
	}
	return &cart, nil
}
func (db cartDB) GetCart(id int) (*Cart, error) {
	var cart Cart
	if err := db.gorm.Preload("Items.Product").First(&cart, id).Error; err != nil {
		return nil, err
	}
	return &cart, nil
}
func (db cartDB) Update(cart *Cart) error {
	if err := db.gorm.Save(cart).Error; err != nil {
		return err
	}
	return nil
}

func (db cartDB) Delete(cart *Cart) error {
	if err := db.gorm.Delete(cart).Error; err != nil {
		return err
	}
	return nil
}

func (cartItem *CartItem) EditQuantity(quantity int) error {
	if quantity != 0 {
		cartItem.Quantity = quantity
	}
	return nil
}
