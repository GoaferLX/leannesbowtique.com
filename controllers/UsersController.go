package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"leannesbowtique.com/models"
	"leannesbowtique.com/rand"
	"leannesbowtique.com/views"
)

type UsersController struct {
	SignUpView   *views.View
	LoginView    *views.View
	ForgotPWView *views.View
	ResetPWView  *views.View
	UserService  models.UserService
	EmailService *MailController
}

func NewUsers(us models.UserService, emailer *MailController) *UsersController {
	return &UsersController{
		SignUpView:   views.NewView("index.gohtml", "views/user/signup.gohtml"),
		LoginView:    views.NewView("index.gohtml", "views/user/login.gohtml"),
		ForgotPWView: views.NewView("index.gohtml", "views/user/forgotpw.gohtml"),
		ResetPWView:  views.NewView("index.gohtml", "views/user/resetpw.gohtml"),
		UserService:  us,
		EmailService: emailer,
	}
}

type resetPWForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
	Token    string `schema:"token"`
}

type signupForm struct {
	Name     string `schema:"name"`
	Email    string `schema:"email"`
	Password string `schema:"password"`
}
type loginForm struct {
	Email    string `schema:"email"`
	Password string `schema:"password"`
}

// POST /signup
// Create processes a new user and adds to database if okay
// Reloads signup page and returns errors if now
func (uc *UsersController) Create(w http.ResponseWriter, r *http.Request) {
	var form signupForm
	var yield views.Page
	yield.PageData = &form
	if err := parsePostForm(r, &form); err != nil {
		yield.SetAlert(err)
		uc.SignUpView.RenderTemplate(w, r, yield)
		return
	}

	var user models.User = models.User{
		Name:     form.Name,
		Email:    form.Email,
		Password: form.Password,
	}

	if err := uc.UserService.CreateUser(&user); err != nil {
		yield.SetAlert(err)
		uc.SignUpView.RenderTemplate(w, r, yield)
		return
	}

	if err := uc.UserService.Login(w, &user); err != nil {
		yield.SetAlert(fmt.Errorf("Your account has been created but you could not be logged in: %w", err))
		yield.Alert.PersistAlert(w)
		http.Redirect(w, r, "/login", http.StatusFound)
		return
	}
	alert := &views.Alert{
		Level:   "success",
		Message: "You have been signed up and logged in!",
	}
	alert.PersistAlert(w)
	http.Redirect(w, r, "/products", http.StatusFound)

}

// /logout
// Logs a user out and returns them to login page
func (uc *UsersController) Logout(w http.ResponseWriter, r *http.Request) {
	// Expire old cookie
	cookie := &http.Cookie{
		Name:     "rememberToken",
		Value:    "",
		Path:     "/",
		Expires:  time.Now(),
		HttpOnly: true,
	}
	// Set expired cookie to log user out
	http.SetCookie(w, cookie)

	// Set new rememberToken and update db
	// If for any reason old cookie persists, it will not match database RememberToken

	token, _ := rand.RememberToken()
	user := User(r.Context())
	user.RememberToken = token
	if err := uc.UserService.UpdateUser(user); err != nil {
		log.Print(err)
	}
	// Redirect user to login page
	http.Redirect(w, r, "/login", http.StatusFound)
}

// /login POST
// Processes a users login attempt and logs them in or returns errors
func (uc *UsersController) Login(w http.ResponseWriter, r *http.Request) {
	var yield views.Page
	var form loginForm
	yield.PageData = &form
	if err := parsePostForm(r, &form); err != nil {
		yield.SetAlert(err)
		uc.LoginView.RenderTemplate(w, r, yield)
		return
	}
	user, err := uc.UserService.Authenticate(form.Email, form.Password)
	if err != nil {
		yield.SetAlert(err)
		uc.LoginView.RenderTemplate(w, r, yield)
		return
	}
	if err := uc.UserService.Login(w, user); err != nil {
		yield.SetAlert(err)
		uc.LoginView.RenderTemplate(w, r, yield)
		return
	}

	url := r.URL.Query().Get("dest")
	http.Redirect(w, r, url, http.StatusFound)

}

func User(ctx context.Context) *models.User {
	if temp := ctx.Value("user"); temp != nil {
		if user, ok := temp.(*models.User); ok {
			return user
		}
	}
	return nil
}

// POST /forgot
// forgot will display the form for initiating a password reset, process input
// and initiate the reset
func (uc *UsersController) Forgot(w http.ResponseWriter, r *http.Request) {
	var yield views.Page
	var form resetPWForm
	yield.PageData = &form
	if err := parsePostForm(r, &form); err != nil {
		yield.SetAlert(err)
		uc.ForgotPWView.RenderTemplate(w, r, yield)
		return
	}
	token, err := uc.UserService.InitiatePWReset(form.Email)
	if err != nil {
		yield.SetAlert(err)
		uc.ForgotPWView.RenderTemplate(w, r, yield)
		return
	}
	if err := uc.EmailService.ResetPw(form.Email, token); err != nil {
		yield.SetAlert(err)
		uc.ForgotPWView.RenderTemplate(w, r, yield)
		return
	}
	alert := views.Alert{Level: "Success", Message: "A reset token has been sent to the email you provided"}
	alert.PersistAlert(w)
	http.Redirect(w, r, "/reset", http.StatusFound)
}

// GET /reset
// ResetPW will display the page for resetting the password and prefill a Token
// if one exists
func (uc *UsersController) ResetPW(w http.ResponseWriter, r *http.Request) {
	var yield views.Page
	var form resetPWForm
	yield.PageData = &form
	if err := parseGetForm(r, &form); err != nil {
		yield.SetAlert(err)
	}
	uc.ResetPWView.RenderTemplate(w, r, yield)
}

// POST /reset
// Reset will process a users password reset, processing the provided token and
// return any errors if they exist
func (uc *UsersController) Reset(w http.ResponseWriter, r *http.Request) {
	var yield views.Page
	var form resetPWForm
	yield.PageData = &form
	if err := parsePostForm(r, &form); err != nil {
		yield.SetAlert(err)
		uc.ResetPWView.RenderTemplate(w, r, yield)
		return
	}
	user, err := uc.UserService.CompletePWReset(form.Token, form.Password)
	if err != nil {
		yield.SetAlert(err)
		uc.ResetPWView.RenderTemplate(w, r, yield)
		return
	}
	uc.UserService.Login(w, user)
	alert := views.Alert{Level: "Success", Message: "Your password has been reset"}
	alert.PersistAlert(w)
	http.Redirect(w, r, "/", http.StatusFound)

}
