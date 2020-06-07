package views

import (
	"bytes"
	"errors"
	"html/template"
	"io"
	"log"
	"net/http"
	"path/filepath"

	"leannesbowtique.com/models"

	"github.com/gorilla/csrf"
)

var templateDir string = "views/bowstemplates/"
var templateExt string = ".gohtml"

type View struct {
	Templates  *template.Template
	LayoutFile string
}

//  Which layout file to load and which page to diplay within it
func NewView(layoutFile string, page string) *View {
	files := getTemplates()
	files = append(files, page)
	templates, err := template.New("").Funcs(template.FuncMap{
		"csrfField": func() (template.HTML, error) {
			return "", errors.New("csrfField not implemented")
		},
	}).ParseFiles(files...)
	if err != nil {
		log.Fatal("The template(s) could not be parsed")
	}

	return &View{
		Templates:  templates,
		LayoutFile: layoutFile,
	}
}

func getTemplates() []string {
	// views/templates/*gohtml
	files, err := filepath.Glob(templateDir + "*" + templateExt)
	if err != nil {
		log.Fatal("Could not get template files")
	}
	return files
}

func (v *View) RenderTemplate(w http.ResponseWriter, r *http.Request, data interface{}) error {
	var buffer bytes.Buffer
	csrfField := csrf.TemplateField(r)
	tpl := v.Templates.Funcs(template.FuncMap{
		"csrfField": func() template.HTML {
			return csrfField
		},
	})
	var yield Page
	switch d := data.(type) {
	case Page:
		yield = d
	default:
		yield = Page{
			PageData: d,
		}
	}
	if alert := getAlert(r); alert != nil {
		yield.Alert = alert
		clearAlert(w)
	}

	userctx := r.Context().Value("user")
	user, ok := userctx.(*models.User)
	if ok {
		yield.User = user
	}

	err := tpl.ExecuteTemplate(&buffer, v.LayoutFile, yield)
	if err != nil {
		log.Println(err)
		http.Error(w, "Something has gone horribly wrong:", http.StatusInternalServerError)
		return err
	}
	w.Header().Set("Content-Type", "text/html")
	io.Copy(w, &buffer)
	return nil
}

func (v *View) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if err := v.RenderTemplate(w, r, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
