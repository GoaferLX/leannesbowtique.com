package models

import (
	"errors"
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
	GetProducts(opts *ProductOpts) ([]*Product, error)
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

type ProductOpts struct {
	Limit      int
	Search     string
	CategoryID int
	Sort       int
}

func (pv *productValidator) GetProducts(opts *ProductOpts) ([]*Product, error) {
	/*
		catError := errors.New("Not a valid category")
		catRegex := regexp.MustCompile(`^[0-9]+`)
		if !catRegex.MatchString(opts.CategoryID) && opts.CategoryID != "" {
			return nil, catError
		}
	*/
	if opts.Limit == 0 || opts.Limit < -1 {
		opts.Limit = -1
	}
	if opts.Sort < 1 || opts.Sort > 4 {
		opts.Sort = 0
	}

	return pv.ProductDB.GetProducts(opts)
}

func (dbm *productDB) GetProducts(opts *ProductOpts) ([]*Product, error) {
	var products []*Product
	var err error
	var order string
	switch opts.Sort {
	case 1:
		order = "created_at desc"
	case 2:
		order = "created_at asc"
	case 3:
		order = "price desc"
	case 4:
		order = "price asc"
	default:
		order = "id asc"
	}
	if opts.CategoryID != 0 {
		err = dbm.gorm.Joins("INNER JOIN product_categories on product_categories.product_id = products.id").Preload("Categories").Where("category_id =?", opts.CategoryID).Limit(opts.Limit).Order(order).Find(&products).Error
	} else {
		err = dbm.gorm.Preload("Categories").Limit(opts.Limit).Order(order).Find(&products).Error
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
