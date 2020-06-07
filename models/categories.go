package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

// Category represents a Product category type stored in the database
type Category struct {
	ID        int
	Name      string `gorm:"unique;"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// Defines all operations for category service
type CategoryService interface {
	CategoryDB
}

// Defines all database interactions for the Categories table
type CategoryDB interface {
	Create(*Category) error
	Update(*Category) error
	Delete(*Category) error

	GetCategory(id int) (*Category, error)
	GetCategories() ([]*Category, error)
}

type categoryModel struct {
	CategoryDB
}
type categoryValidator struct {
	CategoryDB
}

type categoryDB struct {
	gorm *gorm.DB
}

func NewCategoryService(db *gorm.DB) CategoryService {
	return &categoryModel{
		CategoryDB: &categoryValidator{
			CategoryDB: &categoryDB{
				gorm: db,
			},
		},
	}
}

func (cv *categoryValidator) Create(cat *Category) error {
	if err := Validate(cat.Name, isRequired); err != nil {
		return err
	}
	return cv.CategoryDB.Create(cat)
}
func (dbm *categoryDB) Create(cat *Category) error {
	var err error
	if err = dbm.gorm.Create(cat).Error; err != nil {
		err = fmt.Errorf("Could not create the record: %w", err)
	}
	return err
}
func (cv *categoryValidator) Update(cat *Category) error {
	if err := Validate(cat.Name, isRequired); err != nil {
		return err
	}
	return cv.CategoryDB.Update(cat)
}
func (dbm *categoryDB) Update(cat *Category) error {
	var err error
	if err = dbm.gorm.Save(cat).Error; err != nil {
		err = fmt.Errorf("Could not save the record: %w", err)
	}
	return err
}

func (cv *categoryValidator) Delete(cat *Category) error {
	if err := Validate(cat.ID, isRequired, isGreaterThan0); err != nil {
		return err
	}
	return cv.CategoryDB.Delete(cat)
}

func (dbm *categoryDB) Delete(cat *Category) error {
	var err error
	if err = dbm.gorm.Delete(cat).Error; err != nil {
		err = fmt.Errorf("Could not delete the record: %w", err)
	}
	return err
}
func (dbm *categoryDB) GetCategory(id int) (*Category, error) {
	var cat *Category
	var err error
	if err = dbm.gorm.First(&cat, id).Error; err != nil {
		err = fmt.Errorf("Could not retreive from database: %w", err)
	}
	return cat, err
}

func (dbm *categoryDB) GetCategories() ([]*Category, error) {
	var categories []*Category
	if err := dbm.gorm.Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}

func (dbm *productDB) GetCategories() ([]Category, error) {
	var categories []Category
	if err := dbm.gorm.Order("name asc").Find(&categories).Error; err != nil {
		return nil, err
	}
	return categories, nil
}
