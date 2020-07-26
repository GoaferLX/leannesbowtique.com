package controllers

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"text/template"
	"time"

	"leannesbowtique.com/models"
	"leannesbowtique.com/views"

	"github.com/mailgun/mailgun-go"
)

type MailController struct {
	ContactView *views.View
	mg          mailgun.Mailgun
}
type ContactForm struct {
	Email   string `schema:"email"`
	Subject string `schema:"subject"`
	Message string `schema:"message"`
}

func NewMail(mg mailgun.Mailgun) *MailController {
	return &MailController{
		ContactView: views.NewView("index.gohtml", "views/static/contact.gohtml"),
		mg:          mg,
	}
}
func (mc *MailController) Contact(w http.ResponseWriter, r *http.Request) {
	var form ContactForm
	var yield views.Page
	if err := parsePostForm(r, &form); err != nil {
		yield.SetAlert(err)
		mc.ContactView.RenderTemplate(w, r, yield)
		return
	}

	emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,16}$`)
	if !emailRegex.MatchString(form.Email) {
		yield.SetAlert(errors.New("Not a valid email"))
		mc.ContactView.RenderTemplate(w, r, yield)
		return
	}

	msg := mc.mg.NewMessage(form.Email, form.Subject, form.Message, "leanne@leannesbowtique.com")
	msg.AddBCC("support@leannesbowtique.com")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	_, _, err := mc.mg.Send(ctx, msg)

	if err != nil {
		log.Print(err)
		err = errors.New("There has been an error sending your message. Please try again. If the problem persists please let us know by emailing directly to support@leannesbowtique.com")
		yield.SetAlert(err)
		mc.ContactView.RenderTemplate(w, r, yield)
	}
	yield.Alert = &views.Alert{
		Level:   "Success",
		Message: "Your message has been sent. Thanks! We will reply ASAP.",
	}
	mc.ContactView.RenderTemplate(w, r, yield)
}

const (
	resetPWSubject = "Instructions for resetting your password."
)
const resetTextTmpl = `Hi there!

It appears that you have requested a password reset. If this was you, please follow the link below to update your password:<br/>

%s

If you are asked for a token, please use the following value:

%s

If you didn't request a password reset you can safely ignore this email and your account will not be changed.<br/>

All the best,
Leanne @ Leanne's Bowtique`
const resetHTMLTmpl = `Hi there!<br/>
<br/>
It appears that you have requested a password reset. If this was you, please follow the link below to update your password:<br/>
<br/>
<a href="%s">%s</a><br/>
<br/>
If you are asked for a token, please use the following value:<br/>
<br/>
%s<br/>
<br/>
If you didn't request a password reset you can safely ignore this email and your account will not be changed.<br/>
<br/>
All the best,<br />
Leanne @ Leanne's Bowtique`

func (mc *MailController) ResetPw(toEmail, token string) error {
	v := url.Values{}
	v.Set("token", token)
	resetURL := "localhost:3000/reset" + "?" + v.Encode()
	resetText := fmt.Sprintf(resetTextTmpl, resetURL, token)

	message := mc.mg.NewMessage("Leanne <support@leannesbowtique.com>", resetPWSubject, resetText, toEmail)
	resetHTML := fmt.Sprintf(resetHTMLTmpl, resetURL, resetURL, token)
	message.SetHtml(resetHTML)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	_, _, err := mc.mg.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("Mailgun Error, could not send: %w", err)
	}
	return nil
}

func (mc *MailController) OrderConfirm(email string, cart *models.Cart) error {
	orderConfirm := fmt.Sprintf(orderConfirmTMPL, cart.Total())
	message := mc.mg.NewMessage("Leannes Bowtique <leanne@leannesbowtique.com>", "Your Order", orderConfirm, email)
	message.SetHtml(orderConfirm)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	_, _, err := mc.mg.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("Mailgun Error, could not send: %w", err)
	}
	return nil
}

const orderConfirmTMPL = `<p>Hi there!</p>

<p>Thank you for ordering with Leannes Bowtique!<br />
Your order for £%.2f has been submitted!</p>

<p>We will be in touch to confirm the order and provide payment details.<br />
The order will be shipped upon successful payment.</p>

<p>If you don't hear from us within 48 hours, please get in touch at <a href="mailto:leanne@leannesbowtique.com">
leanne@leannesbowtique.com</a> to confirm the status of your order.</p>

<p>All the best,<br />
Leanne @ Leanne's Bowtique</p>

<p>If you think you have received this email in error, please ignore it.  No need to panic.</p>`

const orderTMPL = `You have a new order from {{.Email}}!

Order Details:
{{range .Cart.Items}}
ID: {{.Product.ID}}
Name: {{.Product.Name}}
Description: {{.Product.Description}}
Price: {{printf "%.2f"  .Product.Price}}
Quantity: {{.Quantity}}
{{end}}
Total: {{printf "%.2f" .Cart.Total}}

Proud of you babe xx`

func (mc *MailController) Order(email string, cart *models.Cart) error {

	temp, err := template.New("").Parse(orderTMPL)
	if err != nil {
		fmt.Println(err)
	}
	var buffer bytes.Buffer
	order := struct {
		Email string
		Cart  *models.Cart
	}{
		Email: email,
		Cart:  cart,
	}
	if err := temp.Execute(&buffer, order); err != nil {
		fmt.Println(err)
	}
	message := mc.mg.NewMessage(email, "New Order", buffer.String(), "leanne@leannesbowtique.com")
	message.AddBCC("support@leannesbowtique.com")
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	_, _, err = mc.mg.Send(ctx, message)
	if err != nil {
		return fmt.Errorf("Mailgun Error, could not send: %w", err)
	}
	return nil
}
