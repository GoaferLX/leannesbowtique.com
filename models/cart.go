package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

type LineItem struct {
	Product  Product
	Quantity int
}
type Cart struct {
	ID        int `gorm:"primary_key"`
	Items     []LineItem
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

func (cart Cart) Subtotal() float64 {
	var total float64
	for _, lineItem := range cart.Items {
		total += lineItem.Product.Price * float64(lineItem.Quantity)
	}
	return total
}

func (cart Cart) Total() float64 {
	var total float64
	subtotal := cart.Subtotal()
	delivery := cart.DeliveryCharge()
	total = subtotal + delivery

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
}
type CartDB interface {
	New() (*Cart, error)
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

func (db cartDB) New() (*Cart, error) {
	cart := Cart{}
	if err := db.gorm.Create(&cart).Error; err != nil {
		return nil, err
	}
	return &cart, nil
}
func (db cartDB) GetCart(id int) (*Cart, error) {
	var cart Cart
	if err := db.gorm.First(&cart, id).Error; err != nil {
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
