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

// Defines API for interacting with products
type ProductService interface {
	ProductDB
	ProductsDB
}

// Defines all database interactions for a single product
// Expected return ono query is *Product (*sql.Row)
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
	//GetProducts(opts *ProductOpts) ([]*Product, error)
	GetBundles() ([]*Bundle, error)
}

// Defines all database interactions for a slice of products
// Expected return ono query is []*Product (*sql.Rows)
type ProductsDB interface {
	GetProducts(opts *SearchOpts) ([]*Product, error)
}

type productModel struct {
	ProductDB
	ProductsDB
}
type productValidator struct {
	ProductDB
}
type productDB struct {
	gorm *gorm.DB
}

type SearchOpts struct {
	Limit      int
	Search     string
	Sort       int
	Offset     int
	CategoryID int
	Total      int
}

type searchValidator struct {
	searchRegex *regexp.Regexp
	ProductsDB
}

type searchDB struct {
	gorm *gorm.DB
}

func NewProductService(db *gorm.DB) ProductService {
	return &productModel{
		ProductDB: &productValidator{
			ProductDB: &productDB{
				gorm: db,
			},
		},
		ProductsDB: &searchValidator{
			searchRegex: regexp.MustCompile(`^[a-z0-9.]`),
			ProductsDB: &searchDB{
				gorm: db,
			},
		},
	}

}

func (sv *searchValidator) GetProducts(opts *SearchOpts) ([]*Product, error) {
	if opts.CategoryID < 1 {
		opts.CategoryID = 0
	}
	searchError := errors.New("Not valid search criteria...")
	if !sv.searchRegex.MatchString(opts.Search) && opts.Search != "" {
		return nil, searchError
	}
	if opts.Limit == 0 || opts.Limit < -1 {
		opts.Limit = -1
	}
	if opts.Sort < 1 || opts.Sort > 4 {
		opts.Sort = 0
	}

	return sv.ProductsDB.GetProducts(opts)
}
func (sdb *searchDB) GetProducts(opts *SearchOpts) ([]*Product, error) {
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
		err = sdb.gorm.Joins("INNER JOIN product_categories on product_categories.product_id = products.id").Preload("Categories").Where("name LIKE ? AND category_id = ?", "%"+opts.Search+"%", opts.CategoryID).Or("description LIKE ? AND category_id = ?", "%"+opts.Search+"%", opts.CategoryID).Order(order).Offset(opts.Offset).Limit(opts.Limit).Find(&products).Error
		sdb.gorm.Joins("INNER JOIN product_categories on product_categories.product_id = products.id").Preload("Categories").Where("category_id = ? AND name LIKE ? AND products.deleted_at IS NULL", opts.CategoryID, "%"+opts.Search+"%").Or("category_id = ? AND description LIKE ? AND products.deleted_at IS NULL", opts.CategoryID, "%"+opts.Search+"%").Order(order).Table("products").Count(&opts.Total)
	} else {
		err = sdb.gorm.Preload("Categories").Where("name LIKE ?", "%"+opts.Search+"%").Or("description LIKE ?", "%"+opts.Search+"%").Order(order).Limit(opts.Limit).Offset(opts.Offset).Find(&products).Error
		sdb.gorm.Where("deleted_at IS NULL AND name LIKE ?", "%"+opts.Search+"%").Or("deleted_at IS NULL AND name LIKE ?", "%"+opts.Search+"%").Table("products").Count(&opts.Total)
	}
	if err != nil {
		return nil, err
	}
	return products, nil
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

func (dbm productDB) GetBundles() ([]*Bundle, error) {
	var bundles []*Bundle
	err := dbm.gorm.Preload("Products").Find(&bundles).Error
	return bundles, err
}
