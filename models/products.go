package models

import (
	"errors"
	"regexp"
	"time"

	"github.com/jinzhu/gorm"
)

// Product represents a single product stored in the database
type Product struct {
	ID          int
	Name        string `gorm:"not_null"`
	Description string
	Categories  []Category `gorm:"many2many:product_categories;association_autocreate:false;association_autoupdate:false;"`
	Price       float64
	Images      []Image `gorm:"-"`
	CreatedAt   time.Time
	UpdatedAt   time.Time
	DeletedAt   *time.Time
}

// Defines API for interacting with a product
type ProductService interface {
	ProductDB
}

// Defines all database interactions for a single product
type ProductDB interface {
	// Operations for updating database
	// Create/Update/Delete
	Create(product *Product) error
	Update(product *Product) error
	Delete(id int) error
	// Operations for querying Database
	// Read
	GetCategories() ([]Category, error)
	GetByID(id int) (*Product, error)
	GetProducts(category string) ([]*Product, error)
}
type productModel struct {
	ProductDB
}
type productValidator struct {
	ProductDB
}
type productDB struct {
	gorm *gorm.DB
}

func NewProductService(db *gorm.DB) ProductService {
	return &productModel{
		ProductDB: &productValidator{
			ProductDB: &productDB{
				gorm: db,
			},
		},
	}
}

func (pv *productValidator) Create(product *Product) error {
	if err := Validate(product.Name, isRequired); err != nil {
		return err
	}
	if err := Validate(product.Description, isRequired); err != nil {
		return err
	}
	if err := Validate(product.Price, isRequired, isGreaterThan0); err != nil {
		return err
	}
	var ret []Category
	for _, cat := range product.Categories {
		if cat.ID != 0 {
			ret = append(ret, cat)
		}
		product.Categories = ret
	}
	return pv.ProductDB.Create(product)
}

func (dbm *productDB) Create(product *Product) error {
	return dbm.gorm.Create(product).Error
}

func (pv *productValidator) GetByID(id int) (*Product, error) {
	if err := Validate(id, isGreaterThan0); err != nil {
		return nil, err
	}
	return pv.ProductDB.GetByID(id)
}

func (dbm *productDB) GetByID(id int) (*Product, error) {
	var product Product
	if err := dbm.gorm.Preload("Categories").First(&product, id).Error; err != nil {
		return nil, err
	}
	return &product, nil
}

func (pv *productValidator) GetProducts(category string) ([]*Product, error) {
	catError := errors.New("Not a valid category")
	catRegex := regexp.MustCompile(`^[0-9]+`)
	if !catRegex.MatchString(category) && category != "" {
		return nil, catError
	}
	return pv.ProductDB.GetProducts(category)
}
func (dbm *productDB) GetProducts(category string) ([]*Product, error) {
	var products []*Product
	var err error
	if category != "" {
		err = dbm.gorm.Joins("INNER JOIN product_categories on product_categories.product_id = products.id").Preload("Categories").Where("category_id=?", category).Order("price desc").Find(&products).Error
	} else {
		err = dbm.gorm.Preload("Categories").Order("price desc").Find(&products).Error
	}
	if err != nil {
		return nil, err
	}
	return products, nil
}
func (pv *productValidator) Update(product *Product) error {
	var ret []Category
	for _, cat := range product.Categories {
		if cat.ID != 0 {
			ret = append(ret, cat)
		}
		product.Categories = ret
	}
	return pv.ProductDB.Update(product)
}
func (dbm *productDB) Update(product *Product) error {
	dbm.gorm.Set("gorm:association_autoupdate", false).Set("gorm:association_autocreate", false).Model(&product).Association("categories").Replace(product.Categories)
	return dbm.gorm.Save(product).Error
}
func (pv *productValidator) Delete(id int) error {
	if err := Validate(id, isRequired, isGreaterThan0); err != nil {
		return errors.New("Invalid ID, cannot delete")
	}
	return pv.ProductDB.Delete(id)
}
func (dbm *productDB) Delete(id int) error {
	product := &Product{ID: id}
	return dbm.gorm.Delete(product).Error
}
