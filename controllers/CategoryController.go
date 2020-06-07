package controllers

import (
	"goafweb/models"
	"goafweb/views"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type CategoryController struct {
	categoryService models.CategoryService
	indexView       *views.View
}

func NewCategories(cs models.CategoryService) *CategoryController {
	return &CategoryController{
		categoryService: cs,
		indexView:       views.NewView("index.gohtml", "views/category/catindex.gohtml"),
	}
}

type categoryForm struct {
	ID   int    `schema:"id"`
	Name string `schema:"name"`
}

func (cc *CategoryController) View(w http.ResponseWriter, r *http.Request) {
	var yield views.Page
	cats, err := cc.categoryService.GetCategories()
	if err != nil {
		yield.SetAlert(err)
		cc.indexView.RenderTemplate(w, r, yield)
	}
	yield.PageData = cats
	cc.indexView.RenderTemplate(w, r, yield)
}

func (cc *CategoryController) Create(w http.ResponseWriter, r *http.Request) {
	var yield views.Page
	var form categoryForm
	if err := parseGetForm(r, &form); err != nil {
		yield.SetAlert(err)
		cc.indexView.RenderTemplate(w, r, yield)
	}
	var cat models.Category = models.Category{
		Name: form.Name,
	}
	if err := cc.categoryService.Create(&cat); err != nil {
		yield.SetAlert(err)
		cc.indexView.RenderTemplate(w, r, yield)
	}
	cats, err := cc.categoryService.GetCategories()
	if err != nil {
		yield.SetAlert(err)
		cc.indexView.RenderTemplate(w, r, yield)
	}
	yield.PageData = cats
	yield.Alert = &views.Alert{
		Level:   "Success",
		Message: "Category added!",
	}
	cc.indexView.RenderTemplate(w, r, yield)
}

func (cc *CategoryController) Update(w http.ResponseWriter, r *http.Request) {
	var yield views.Page
	var form categoryForm
	if err := parseGetForm(r, &form); err != nil {
		yield.SetAlert(err)
		cc.indexView.RenderTemplate(w, r, yield)
	}
	var cat models.Category = models.Category{
		ID:   form.ID,
		Name: form.Name,
	}
	if err := cc.categoryService.Update(&cat); err != nil {
		yield.SetAlert(err)
		cc.indexView.RenderTemplate(w, r, yield)
	}
	cats, err := cc.categoryService.GetCategories()
	if err != nil {
		yield.SetAlert(err)
		cc.indexView.RenderTemplate(w, r, yield)
	}
	yield.PageData = cats
	yield.Alert = &views.Alert{
		Level:   "Success",
		Message: "Category Updated!",
	}
	cc.indexView.RenderTemplate(w, r, yield)
}
func (cc *CategoryController) Delete(w http.ResponseWriter, r *http.Request) {
	var yield views.Page
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		yield.SetAlert(err)
		cc.indexView.RenderTemplate(w, r, yield)
	}
	var cat models.Category = models.Category{
		ID: id,
	}
	if err := cc.categoryService.Delete(&cat); err != nil {
		yield.SetAlert(err)
		cc.indexView.RenderTemplate(w, r, yield)
	}

	cats, err := cc.categoryService.GetCategories()
	if err != nil {
		yield.SetAlert(err)
		cc.indexView.RenderTemplate(w, r, yield)
	}
	yield.PageData = cats
	yield.Alert = &views.Alert{
		Level:   "Success",
		Message: "Category Deleted!",
	}
	cc.indexView.RenderTemplate(w, r, yield)
}
