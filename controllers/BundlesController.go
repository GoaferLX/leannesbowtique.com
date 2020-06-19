package controllers

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"leannesbowtique.com/models"
	"leannesbowtique.com/views"
)

type BundlesController struct {
	BundleView     *views.View
	BundlesView    *views.View
	NewBundleView  *views.View
	EditBundleView *views.View
	BundleService  models.BundleService
}
type BundlesForm struct {
	Name        string           `schema:"name"`
	Description string           `schema:"description"`
	Price       float64          `schema:"price"`
	Products    []models.Product `schema:"products"`
}

func NewBundlesController(bs models.BundleService) *BundlesController {
	return &BundlesController{
		BundleView:     views.NewView("index.gohtml", "views/bundles/bundle.gohtml"),
		BundlesView:    views.NewView("index.gohtml", "views/bundles/bundles.gohtml"),
		NewBundleView:  views.NewView("index.gohtml", "views/bundles/newbundle.gohtml"),
		EditBundleView: views.NewView("index.gohtml", "views/bundles/editbundle.gohtml"),
		BundleService:  bs,
	}
}

// GET /new
func (bc BundlesController) NewBundle(w http.ResponseWriter, r *http.Request) {
	var yield views.Page
	data := struct {
		Bundle   models.Bundle
		Products []models.Product
	}{}
	yield.PageData = &data
	var products []models.Product

	err := bc.BundleService.GetProducts(&products)
	if err != nil {
		yield.SetAlert(err)
		bc.NewBundleView.RenderTemplate(w, r, yield)
		return
	}
	data.Products = products
	bc.NewBundleView.RenderTemplate(w, r, yield)

}

// POST /new
// Create processes the request to create a new bundle and adds to database
// Redirects to bundle page on success, redisplays page on failure
func (bc BundlesController) Create(w http.ResponseWriter, r *http.Request) {
	var yield views.Page
	var form BundlesForm
	yield.PageData = &form

	if err := parsePostForm(r, &form); err != nil {
		yield.SetAlert(err)
		bc.NewBundleView.RenderTemplate(w, r, yield)
		return
	}
	bundle := models.Bundle{
		Name:        form.Name,
		Description: form.Description,
		Price:       form.Price,
		Products:    form.Products,
	}
	if err := bc.BundleService.Create(&bundle); err != nil {
		yield.SetAlert(err)
		bc.NewBundleView.RenderTemplate(w, r, yield)
		return
	}
	url := fmt.Sprintf("/bundle/%d/edit", bundle.ID)
	http.Redirect(w, r, url, http.StatusFound)
}
func (bc *BundlesController) GetByID(w http.ResponseWriter, r *http.Request) *models.Bundle {
	var yield *views.Page
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		yield.SetAlert(err)
		bc.EditBundleView.RenderTemplate(w, r, yield)
		return nil
	}
	bundle, err := bc.BundleService.GetByID(id)
	if err != nil {
		yield.SetAlert(err)
		bc.EditBundleView.RenderTemplate(w, r, yield)
		return nil
	}
	return bundle
}

// GET /edit
func (bc *BundlesController) Edit(w http.ResponseWriter, r *http.Request) {
	var yield views.Page
	data := struct {
		Bundle   *models.Bundle
		Products []models.Product
	}{}
	yield.PageData = &data
	var products []models.Product
	err := bc.BundleService.GetProducts(&products)
	if err != nil {
		yield.SetAlert(err)
		bc.NewBundleView.RenderTemplate(w, r, yield)
		return
	}
	data.Products = products
	bundle := bc.GetByID(w, r)
	data.Bundle = bundle
	bc.EditBundleView.RenderTemplate(w, r, yield)
}

// POST /edit
func (bc BundlesController) Update(w http.ResponseWriter, r *http.Request) {
	var yield views.Page
	var form BundlesForm
	yield.PageData = &form
	bundle := bc.GetByID(w, r)
	if err := parsePostForm(r, &form); err != nil {
		yield.SetAlert(err)
		bc.NewBundleView.RenderTemplate(w, r, yield)
		return
	}

	bundle.Name = form.Name
	bundle.Description = form.Description
	bundle.Price = form.Price
	bundle.Products = form.Products

	if err := bc.BundleService.Update(bundle); err != nil {
		yield.SetAlert(err)
		bc.NewBundleView.RenderTemplate(w, r, yield)
		return
	}
	url := fmt.Sprintf("/bundle/%d/edit", bundle.ID)
	http.Redirect(w, r, url, http.StatusFound)
}

func (bc *BundlesController) ImageUpload(w http.ResponseWriter, r *http.Request) {
	var yield views.Page
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		yield.SetAlert(err)
		bc.EditBundleView.RenderTemplate(w, r, yield)
		return
	}

	if err := r.ParseMultipartForm(1 << 20); err != nil {
		yield.SetAlert(err)
		bc.EditBundleView.RenderTemplate(w, r, yield)
		return
	}
	bundle, _ := bc.BundleService.GetByID(id)
	// Iterate over uploaded files to process them.
	fheaders := r.MultipartForm.File["images"]
	for _, f := range fheaders {
		// Open the uploaded file
		file, err := f.Open()
		if err != nil {
			yield.SetAlert(err)
			bc.EditBundleView.RenderTemplate(w, r, yield)
			return
		}
		defer file.Close()

		is := models.NewImageService("bundles")
		// Create the image
		if err := is.Create(bundle.ID, file, f.Filename); err != nil {
			yield.SetAlert(err)
			bc.EditBundleView.RenderTemplate(w, r, yield)
			return
		}
	}

	alert := &views.Alert{Level: "Success", Message: "Images uploaded successfully"}
	alert.PersistAlert(w)
	url := fmt.Sprintf("/bundle/%d/edit", bundle.ID)
	http.Redirect(w, r, url, http.StatusFound)
}

func (bc *BundlesController) ViewBundle(w http.ResponseWriter, r *http.Request) {
	var yield views.Page
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		yield.SetAlert(err)
		bc.BundleView.RenderTemplate(w, r, yield)
		return
	}

	bundle, err := bc.BundleService.GetByID(id)
	if err != nil {
		yield.SetAlert(err)
		bc.BundleView.RenderTemplate(w, r, yield)
		return
	}
	is := models.NewImageService("bundles")
	images, _ := is.GetByEntityID(bundle.ID)
	bundle.Images = images

	yield.PageData = bundle
	bc.BundleView.RenderTemplate(w, r, yield)
}

func (bc *BundlesController) ViewBundles(w http.ResponseWriter, r *http.Request) {
	var yield views.Page
	bundles, err := bc.BundleService.GetBundles()
	//[]*Bundle
	if err != nil {
		yield.SetAlert(err)
		bc.BundlesView.RenderTemplate(w, r, yield)
		return
	}
	is := models.NewImageService("bundles")
	for _, bundle := range bundles {
		images, _ := is.GetByEntityID(bundle.ID)
		bundle.Images = images
	}
	yield.PageData = bundles
	bc.BundlesView.RenderTemplate(w, r, yield)
}
