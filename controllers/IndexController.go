package controllers

import "leannesbowtique.com/views"

func StaticPage() *Page {
	return &Page{
		About:    views.NewView("index.gohtml", "views/static/about.gohtml"),
		NotFound: views.NewView("index.gohtml", "views/static/404.gohtml"),
	}
}

type Page struct {
	About    *views.View
	NotFound *views.View
}
