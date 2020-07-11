package views

import (
	"net/http"
	"time"

	"leannesbowtique.com/models"
)

// Data structure to be passed to all views
type Page struct {
	Alert    *Alert
	PageData interface{}
	User     *models.User
	Cart     *models.Cart
}

type Alert struct {
	Level   string // warning/error/success etc
	Message string
}

func (p *Page) SetAlert(err error) {
	p.Alert = &Alert{
		Level:   "warning",
		Message: err.Error(),
	}
}
func (alert *Alert) PersistAlert(w http.ResponseWriter) {
	alertLvl := http.Cookie{
		Name:     "alertLvl",
		Value:    alert.Level,
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Now().Add(5 * time.Minute),
		MaxAge:   120,
	}
	alertMsg := http.Cookie{
		Name:     "alertMsg",
		Value:    alert.Message,
		HttpOnly: true,
		Path:     "/",
		Expires:  time.Now().Add(60 * time.Second),
		MaxAge:   120,
	}
	http.SetCookie(w, &alertLvl)
	http.SetCookie(w, &alertMsg)
}

func clearAlert(w http.ResponseWriter) {
	alertLvl := http.Cookie{
		Name:     "alertLvl",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now(),
	}
	alertMsg := http.Cookie{
		Name:     "alertMsg",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Expires:  time.Now(),
	}
	http.SetCookie(w, &alertLvl)
	http.SetCookie(w, &alertMsg)
}
func getAlert(r *http.Request) *Alert {
	alertLvl, err := r.Cookie("alertLvl")
	if err != nil {
		return nil
	}
	alertMsg, err := r.Cookie("alertMsg")
	if err != nil {
		return nil
	}
	return &Alert{
		Level:   alertLvl.Value,
		Message: alertMsg.Value,
	}
}
