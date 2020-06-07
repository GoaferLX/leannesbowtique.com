package models

import (
	"fmt"
	"time"

	"github.com/jinzhu/gorm"
)

// Defines a single Article as stored in the database
// Can be used to model a news article or blo
type Article struct {
	ID        int
	Title     string `gorm:"not_null"`
	Content   string `gorm:"not_null"`
	Author    int    `gorm:"not_null"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

// Defines API for interacting with an Article
type ArticleService interface {
	ArticleDB
}

// Defines all database interactions for a single article
type ArticleDB interface {
	// Database retreival
	GetArticleByID(id int) (*Article, error)
	GetArticlesByUser(id int) ([]Article, error)
	// Database modification
	Create(article *Article) error
	Update(article *Article) error
	Delete(id int) error
}

type articleModel struct {
	ArticleDB
}

type articleValidator struct {
	ArticleDB
}

type articleDB struct {
	gorm *gorm.DB
}

func NewArticleService(db *gorm.DB) ArticleService {
	return &articleModel{
		ArticleDB: &articleValidator{
			ArticleDB: &articleDB{
				gorm: db,
			},
		},
	}
}

func (av *articleValidator) Create(article *Article) error {
	if err := Validate(article.Title, isRequired); err != nil {
		return err
	}
	if err := Validate(article.Content, isRequired); err != nil {
		return err
	}
	if err := Validate(article.Author, isRequired, isGreaterThan0); err != nil {
		return err
	}
	return av.ArticleDB.Create(article)
}

func (dbm *articleDB) Create(article *Article) error {
	return dbm.gorm.Create(article).Error
}

func (av *articleValidator) GetArticleByID(id int) (*Article, error) {
	if err := Validate(id, isGreaterThan0); err != nil {
		return nil, err
	}
	return av.ArticleDB.GetArticleByID(id)
}

func (dbm *articleDB) GetArticleByID(id int) (*Article, error) {
	article := Article{}
	err := dbm.gorm.First(&article, id).Error
	if err != nil {
		return nil, err
	}
	return &article, nil
}

func (dbm *articleDB) GetArticlesByUser(authorID int) ([]Article, error) {
	var articles []Article

	results := dbm.gorm.Where("author = ?", authorID)
	if err := results.Find(&articles).Error; err != nil {
		return nil, err
	}
	return articles, nil
}

func (dbm *articleDB) Update(article *Article) error {
	return dbm.gorm.Save(article).Error
}

func (av *articleValidator) Delete(id int) error {
	if err := Validate(id, isGreaterThan0); err != nil {
		return err
	}
	return av.ArticleDB.Delete(id)
}

func (dbm *articleDB) Delete(id int) error {
	article := &Article{ID: id}
	var err error
	if err = dbm.gorm.Delete(article).Error; err != nil {
		err = fmt.Errorf("Could not delete: %w", err)
	}
	return err
}
