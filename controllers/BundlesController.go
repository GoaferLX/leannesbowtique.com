package controllers

import (
	"net/http"

	"leannesbowtique.com/models"
	"leannesbowtique.com/views"
)

type BundlesController struct {
	BundleView    *views.View
	BundlesView   *views.View
	BundleService models.BundleService
}

func NewBundlesController(bs models.BundleService) *BundlesController {
	return &BundlesController{
		BundleView:    views.NewView("index.gohtml", "views/bundles/bundle.gohtml"),
		BundlesView:   views.NewView("index.gohtml", "views/bundles/bundles.gohtml"),
		BundleService: bs,
	}
}

func (bc *BundlesController) ViewBundle(w http.ResponseWriter, r *http.Request) {
	var yield views.Page
	bundle, err := bc.BundleService.GetByID(1)
	if err != nil {
		yield.SetAlert(err)
		bc.BundleView.RenderTemplate(w, r, yield)
		return
	}
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
