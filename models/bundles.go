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
	Products    []Product `gorm:"many2many:product_bundles;"`
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
}
type bundleModel struct {
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
func (dbm bundleDB) Create(b *Bundle) error {
	return dbm.gorm.Create(b).Error
}
func (dbm bundleDB) Update(b *Bundle) error {
	return dbm.gorm.Save(b).Error
}
