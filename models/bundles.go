package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

type Bundle struct {
	ID          int       `gorm:"primary_key;auto_increment"`
	Name        string    `gorm:"not null;"`
	Description string    `gorm:"not null;"`
	Price       float64   `gorm:"not null;"`
	Products    []Product `gorm:"many2many:product_bundles;association_autocreate:false;association_autoupdate:false;"`
	Images      []Image
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// BundleService defines the API for interacting with a bundle
type BundleService interface {
	BundleDB
}

// BundleDB defines operations for
type BundleDB interface {
	Create(*Bundle) error
	Update(*Bundle) error
	//	Delete(*Bundle) error
	GetByID(id int) (*Bundle, error)
	//	GetByFilter() ([]*Bundle, error)
	GetBundles() ([]*Bundle, error)
	GetProducts(products *[]Product) error
}
type bundleModel struct {
	BundleDB
}
type bundleValidator struct {
	BundleDB
}
type bundleDB struct {
	gorm *gorm.DB
}

func NewBundleService(db *gorm.DB) BundleService {
	return &bundleModel{
		BundleDB: bundleDB{
			gorm: db,
		},
	}
}
func (dbm bundleDB) GetByID(id int) (*Bundle, error) {
	var bundle Bundle
	var err error
	if err = dbm.gorm.Preload("Products").Find(&bundle, id).Error; err != nil {
		err = fmt.Errorf("Could not retreive bundle from db: %w", err)
	}
	return &bundle, err
}
func (dbm bundleDB) GetBundles() ([]*Bundle, error) {
	var bundles []*Bundle
	err := dbm.gorm.Preload("Products").Find(&bundles).Error
	return bundles, err
}

func (bv *bundleValidator) Create(b *Bundle) error {
	if err := Validate(b.Name, isRequired); err != nil {
		return err
	}
	if err := Validate(b.Description, isRequired); err != nil {
		return err
	}
	if err := Validate(b.Price, isRequired, isGreaterThan0); err != nil {
		return err
	}
	var prods []Product
	for _, prod := range b.Products {
		if prod.ID != 0 {
			prods = append(prods, prod)
		}
		b.Products = prods
	}
	return bv.BundleDB.Create(b)
}

func (dbm bundleDB) Create(b *Bundle) error {
	return dbm.gorm.Create(b).Error
}

func (bv *bundleValidator) Update(b *Bundle) error {
	if err := Validate(b.Name, isRequired); err != nil {
		return err
	}
	if err := Validate(b.Description, isRequired); err != nil {
		return err
	}
	if err := Validate(b.Price, isRequired, isGreaterThan0); err != nil {
		return err
	}
	var prods []Product
	for _, prod := range b.Products {
		if prod.ID != 0 {
			prods = append(prods, prod)
		}
		b.Products = prods
	}
	return bv.BundleDB.Update(b)
}

func (dbm bundleDB) Update(b *Bundle) error {
	return dbm.gorm.Save(b).Error
}

func (dbm bundleDB) GetProducts(products *[]Product) error {
	return dbm.gorm.Find(&products).Error
}
