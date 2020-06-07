package models

import (
	"fmt"
	"time"

	"leannesbowtique.com/hash"
	"leannesbowtique.com/rand"

	"github.com/jinzhu/gorm"
)

type pwReset struct {
	ID        int
	UserID    int    `gorm:"not null"`
	Token     string `gorm:"-"`
	TokenHash string `gorm:"not null;unique_index"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt *time.Time
}

type pwResetService interface {
	ByToken(token string) (*pwReset, error)
	Create(pwr *pwReset) error
	Delete(id int) error
}
type pwResetModel struct {
	pwResetService
}
type pwResetValidator struct {
	pwResetService
	hmac hash.HMAC
}

type pwResetDB struct {
	gorm *gorm.DB
}

func newPWResetService(db *gorm.DB, hmac hash.HMAC) pwResetService {
	return &pwResetValidator{
		pwResetService: &pwResetDB{
			gorm: db,
		},
		hmac: hmac,
	}
}
func (pwrreset *pwResetValidator) ByToken(token string) (*pwReset, error) {
	if err := Validate(token, isRequired); err != nil {
		return nil, err
	}
	tokenHash := pwrreset.hmac.Hash(token)
	return pwrreset.pwResetService.ByToken(tokenHash)

}
func (pwrdb *pwResetDB) ByToken(tokenHash string) (*pwReset, error) {
	var pwr pwReset
	err := pwrdb.gorm.Where("token_hash = ?", tokenHash).First(&pwr).Error
	if err != nil {
		return nil, err
	}
	return &pwr, nil
}

func (pwrreset *pwResetValidator) Create(pwr *pwReset) error {
	var err error
	if err = Validate(pwr.UserID, isRequired, isGreaterThan0); err != nil {
		return err
	}
	pwr.Token, err = rand.RememberToken()
	if err != nil {
		return fmt.Errorf("Could not create a Reset Token: %w", err)
	}
	pwr.TokenHash = pwrreset.hmac.Hash(pwr.Token)
	return pwrreset.pwResetService.Create(pwr)
}

func (pwrdb *pwResetDB) Create(pwr *pwReset) error {
	return pwrdb.gorm.Create(pwr).Error
}

func (pwrreset *pwResetValidator) Delete(id int) error {
	var err error
	if err = Validate(id, isRequired, isGreaterThan0); err != nil {
		return err
	}
	return pwrreset.pwResetService.Delete(id)
}
func (pwrdb *pwResetDB) Delete(id int) error {
	pwr := pwReset{ID: id}
	return pwrdb.gorm.Delete(&pwr).Error
}
