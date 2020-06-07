package models

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"leannesbowtique.com/hash"
	"leannesbowtique.com/rand"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// User reprensents a user as saved in the database
// Used to model a user single user throughout the app and mirror in database
type User struct {
	ID            int    `gorm:"primary_key;"`
	Name          string `gorm:"not_null;"`
	Email         string `gorm:"not_null;unique_index;"`
	Password      string `gorm:"-"`
	PasswordHash  string `gorm:"not_null;"`
	RememberToken string `gorm:"-"`
	RememberHash  string `gorm:"not_null;unique_index;"`
	CreatedAt     time.Time
	UpdatedAt     time.Time
	DeletedAt     *time.Time
}

// Defines API for interacting with a User
type UserService interface {
	Authenticate(email, password string) (*User, error)
	Login(w http.ResponseWriter, user *User) error
	UserDB
	InitiatePWReset(email string) (string, error)
	CompletePWReset(token, newPW string) (*User, error)
}

// Defines all database interactions for a single user
type UserDB interface {
	// Methods for querying db
	GetUser(id int) (*User, error)
	GetUserByEmail(email string) (*User, error)
	GetUserByRemember(token string) (*User, error)

	// Methods for editing database
	CreateUser(user *User) error
	UpdateUser(user *User) error
}

type userModel struct {
	UserDB
	pwResetService
	PwPepper string
}

type userValidator struct {
	UserDB
	hmac     hash.HMAC
	PwPepper string
}

type userDB struct {
	gorm *gorm.DB
}

func NewUserService(db *gorm.DB, userPwPepper, hmacSecretKey string) UserService {
	hmac := hash.NewHMAC(hmacSecretKey)
	return &userModel{
		UserDB: &userValidator{
			UserDB: &userDB{
				gorm: db,
			},
			hmac:     hmac,
			PwPepper: userPwPepper,
		},
		PwPepper:       userPwPepper,
		pwResetService: newPWResetService(db, hmac),
	}
}

func (val *userValidator) GetUser(id int) (*User, error) {
	err := isGreaterThan0(id)
	if err != nil {
		return nil, errors.New("ID Cannot be zero (0)")
	}
	return val.UserDB.GetUser(id)
}
func (dbm *userDB) GetUser(id int) (*User, error) {
	var user User
	err := dbm.gorm.First(&user, id).Error
	return &user, err
}

func (uv *userValidator) GetUserByEmail(email string) (*User, error) {
	if err := Validate(email, isRequired, isEmailFormat); err != nil {
		return nil, err
	}
	return uv.UserDB.GetUserByEmail(email)
}

func (dbm *userDB) GetUserByEmail(email string) (*User, error) {
	var user User
	err := dbm.gorm.Where("email = ?", email).First(&user).Error
	return &user, err
}
func (uv *userValidator) GetUserByRemember(token string) (*User, error) {
	rememberHash := uv.hmac.Hash(token)
	return uv.UserDB.GetUserByRemember(rememberHash)
}
func (dbm *userDB) GetUserByRemember(token string) (*User, error) {
	var user User
	err := dbm.gorm.Where("remember_hash = ?", token).First(&user).Error
	return &user, err
}
func (uv *userValidator) CreateUser(user *User) error {
	if err := Validate(user.Email, isRequired, isEmailFormat); err != nil {
		return err
	}
	if err := Validate(user.Password, isRequired, isMinLength(4)); err != nil {
		return err
	}
	pwhash, err := bcrypt.GenerateFromPassword([]byte(user.Password+uv.PwPepper), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(pwhash)
	user.Password = ""

	token, err := rand.RememberToken()
	if err != nil {
		return err
	}
	user.RememberToken = token
	user.RememberHash = uv.hmac.Hash(user.RememberToken)

	_, err = uv.GetUserByEmail(user.Email)
	if err == nil {
		return errors.New("That email address is already taken")
	}
	return uv.UserDB.CreateUser(user)

}
func (dbm *userDB) CreateUser(user *User) error {
	return dbm.gorm.Create(user).Error
}
func (uv *userValidator) UpdateUser(user *User) error {
	user.RememberHash = uv.hmac.Hash(user.RememberToken)
	if user.Password != "" {
		pwhash, err := bcrypt.GenerateFromPassword([]byte(user.Password+uv.PwPepper), bcrypt.DefaultCost)
		if err != nil {
			return err
		}
		user.PasswordHash = string(pwhash)
		user.Password = ""
	}
	return uv.UserDB.UpdateUser(user)
}
func (dbm *userDB) UpdateUser(user *User) error {
	return dbm.gorm.Save(user).Error
}

func (um *userModel) Authenticate(email, password string) (*User, error) {
	user, err := um.GetUserByEmail(email)
	if err != nil {
		return nil, err
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password+um.PwPepper))
	if err != nil {
		log.Print(err)
		return nil, errors.New("Invalid Password")
	}
	return user, nil
}

func (um *userModel) Login(w http.ResponseWriter, user *User) error {
	if user.RememberToken == "" {
		token, err := rand.RememberToken()
		if err != nil {
			return err
		}
		user.RememberToken = token
		err = um.UpdateUser(user)
		if err != nil {
			return err
		}
	}

	rememberCookie := http.Cookie{
		Name:     "rememberToken",
		Value:    user.RememberToken,
		Path:     "/",
		Expires:  time.Now().Add(24 * time.Hour),
		HttpOnly: true,
		//	Secure:   true,
	}
	http.SetCookie(w, &rememberCookie)
	return nil
}

func (um *userModel) InitiatePWReset(email string) (string, error) {
	user, err := um.GetUserByEmail(email)
	if err != nil {
		return "", err
	}
	pwr := pwReset{
		UserID: user.ID,
	}
	if err := um.pwResetService.Create(&pwr); err != nil {
		return "", err
	}
	return pwr.Token, nil
}

func (um *userModel) CompletePWReset(token, newPw string) (*User, error) {
	pwr, err := um.pwResetService.ByToken(token)
	if err != nil {
		return nil, fmt.Errorf("Token not valid: %w", err)
	}

	if time.Now().Sub(pwr.CreatedAt) > (12 * time.Hour) {
		return nil, errors.New("Token no longer valid")
	}
	user, err := um.GetUser(pwr.UserID)
	if err != nil {
		return nil, err
	}
	user.Password = newPw
	if err = um.UpdateUser(user); err != nil {
		return nil, err
	}
	um.pwResetService.Delete(pwr.ID)
	return user, nil
}
