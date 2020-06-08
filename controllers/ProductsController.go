package controllers

import (
	"fmt"
	"math"
	"net/http"
	"strconv"

	"leannesbowtique.com/models"
	"leannesbowtique.com/views"

	"github.com/gorilla/mux"
)

type ProductsController struct {
	productService models.ProductService
	imageService   models.ImageService
	editView       *views.View
	productView    *views.View
	CreateView     *views.View
	productsView   *views.View
	indexView      *views.View
}
type ProductsForm struct {
	Name        string            `schema:"name"`
	Description string            `schema:"description"`
	Price       float64           `schema:"price"`
	Categories  []models.Category `schema:"categories"`
}

func NewProductsController(ps models.ProductService, is models.ImageService) *ProductsController {
	return &ProductsController{
		productService: ps,
		imageService:   is,
		editView:       views.NewView("index.gohtml", "views/product/editproduct.gohtml"),
		CreateView:     views.NewView("index.gohtml", "views/product/newproduct.gohtml"),
		productView:    views.NewView("index.gohtml", "views/product/product.gohtml"),
		productsView:   views.NewView("index.gohtml", "views/product/productsindex.gohtml"),
		indexView:      views.NewView("index.gohtml", "views/product/productsadminview.gohtml"),
	}
}

// GET /new
// NewProduct renders the page for adding a new product
// It retreives the available categories from storage to pass to the form
func (pc ProductsController) NewProduct(w http.ResponseWriter, r *http.Request) {
	var yield views.Page
	data := struct {
		Product    models.Product
		Categories []models.Category
	}{}
	yield.PageData = &data
	cats, err := pc.productService.GetCategories()
	if err != nil {
		yield.SetAlert(err)
	}
	data.Categories = cats
	pc.CreateView.RenderTemplate(w, r, yield)
}

// POST /new
// Create processes the request to create new product and adds to database
// Redirects to product page on success, redisplays page on failure
func (pc ProductsController) Create(w http.ResponseWriter, r *http.Request) {
	var yield views.Page
	var form ProductsForm
	data := struct {
		Product    *ProductsForm
		Categories []models.Category
	}{}
	yield.PageData = &data
	data.Product = &form
	cats, err := pc.productService.GetCategories()
	if err != nil {
		yield.SetAlert(err)
	}
	data.Categories = cats

	fmt.Println(yield)
	if err := parsePostForm(r, &form); err != nil {
		yield.SetAlert(err)
		pc.CreateView.RenderTemplate(w, r, yield)
		return
	}
	product := models.Product{
		Name:        form.Name,
		Description: form.Description,
		Price:       form.Price,
		Categories:  form.Categories,
	}
	if err := pc.productService.Create(&product); err != nil {
		yield.SetAlert(err)
		pc.CreateView.RenderTemplate(w, r, yield)
		return
	}
	url := fmt.Sprintf("/product/%d/edit", product.ID)
	http.Redirect(w, r, url, http.StatusFound)
}

func (pc *ProductsController) ViewProductsIndex(w http.ResponseWriter, r *http.Request) {
	yield := views.Page{}
	opts := &models.ProductOpts{}
	products, err := pc.productService.GetProducts(opts)
	if err != nil {
		yield.SetAlert(err)
		pc.indexView.RenderTemplate(w, r, yield)
		return
	}
	for _, product := range products {
		images, _ := pc.imageService.ByID(product.ID)
		product.Images = images
	}
	yield.PageData = products
	pc.indexView.RenderTemplate(w, r, yield)
}

// ProductByID is a helper function, it takes care of checking the requested
// product is valid (mux.Vars) and then retreives the product from the database
// along with all assoicated Images
// It handles errors and re-renders where needed.
func (pc *ProductsController) productByID(w http.ResponseWriter, r *http.Request) *models.Product {
	vars := mux.Vars(r)
	yield := views.Page{}
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		yield.SetAlert(err)
		pc.productView.RenderTemplate(w, r, yield)
		return nil
	}
	product, err := pc.productService.GetByID(id)
	if err != nil {
		yield.SetAlert(err)
		pc.productView.RenderTemplate(w, r, yield)
		return nil
	}
	images, _ := pc.imageService.ByID(product.ID)
	product.Images = images
	return product

}

// GET /product:id
// ViewProduct retreives a product and renders is on the view page
func (pc *ProductsController) ViewProduct(w http.ResponseWriter, r *http.Request) {
	yield := views.Page{}
	product := pc.productByID(w, r)

	yield.PageData = product
	pc.productView.RenderTemplate(w, r, yield)
}

// splitN splits a []Product into smaller slices for rendering in a grid
// TODO: Relocate this view logic to Views package
func splitN(n int, products []*models.Product) [][]*models.Product {
	ret := make([][]*models.Product, n)
	for i := 0; i < n; i++ {
		ret[i] = make([]*models.Product, 0)
	}
	for i, product := range products {
		bucket := i % n
		ret[bucket] = append(ret[bucket], product)
	}
	return ret
}

type ProductOptsForm struct {
	CategoryID int    `schema:"category"`
	Limit      int    `schema:"limit"`
	Search     string `schema:"search"`
	Sort       int    `schema:"sort"`
	PageNum    int    `schema:"pagenum"`
	Total      int
}

func (p ProductOptsForm) PageUp() string {
	num := p.PageNum + 1
	return strconv.Itoa(num)

}
func (p ProductOptsForm) PageDown() string {
	num := p.PageNum - 1
	return strconv.Itoa(num)
}

// GET /Products
// Index of all products, searchable by category
func (pc *ProductsController) ViewProducts(w http.ResponseWriter, r *http.Request) {
	yield := views.Page{}
	data := struct {
		Products   []*models.Product
		Categories []models.Category
		Form       *ProductOptsForm
	}{}
	yield.PageData = &data
	var form ProductOptsForm = ProductOptsForm{PageNum: 1, Limit: 6, Sort: 3}
	data.Form = &form
	if err := parseGetForm(r, &form); err != nil {
		fmt.Println(err)
	}
	if form.Limit > -1 {
		form.Limit = 6
	}
	if form.Sort < 1 || form.Sort > 4 {
		form.Sort = 3
	}
	offset := ((form.PageNum) * form.Limit) - form.Limit

	opts := &models.ProductOpts{CategoryID: form.CategoryID, Limit: form.Limit, Sort: form.Sort, Offset: offset}
	products, err := pc.productService.GetProducts(opts)
	fmt.Println(opts.Total)
	form.Total = int(math.Ceil(float64(opts.Total) / float64(opts.Limit)))

	data.Products = products
	if err != nil {
		yield.SetAlert(err)
		pc.productView.RenderTemplate(w, r, yield)
		return
	}

	// Attach images to each product
	for _, product := range products {
		images, _ := pc.imageService.ByID(product.ID)
		product.Images = images
	}

	// Get categories list for search form
	cats, _ := pc.productService.GetCategories()
	data.Categories = cats
	pc.productsView.RenderTemplate(w, r, yield)
}

// GET /edit
// Edit renders the page to edit a products information and fills it with
// data retreived from the database for that ProductID
func (pc *ProductsController) Edit(w http.ResponseWriter, r *http.Request) {
	yield := views.Page{}
	data := struct {
		Categories []models.Category
		Product    *models.Product
	}{}
	yield.PageData = &data

	product := pc.productByID(w, r)
	data.Product = product
	data.Categories, _ = pc.productService.GetCategories()

	pc.editView.RenderTemplate(w, r, yield)
}

// POST /edit
// Update parses the information from the /edit page and updates the product in
// the database
// Redirects to the product page on success
func (pc *ProductsController) Update(w http.ResponseWriter, r *http.Request) {
	yield := views.Page{}
	data := struct {
		Categories []models.Category
		Product    *models.Product
	}{}
	yield.PageData = &data
	data.Categories, _ = pc.productService.GetCategories()
	product := pc.productByID(w, r)
	data.Product = product

	var form ProductsForm
	if err := parsePostForm(r, &form); err != nil {
		yield.SetAlert(err)
		pc.editView.RenderTemplate(w, r, yield)
		return
	}
	product.Name = form.Name
	product.Description = form.Description
	product.Price = form.Price
	product.Categories = form.Categories

	if err := pc.productService.Update(product); err != nil {
		yield.SetAlert(err)
		pc.editView.RenderTemplate(w, r, yield)
		return
	}

	url := fmt.Sprintf("/product/%d", product.ID)
	http.Redirect(w, r, url, http.StatusFound)
}

// GET /delete
// Deletes a product from the database
// TODO Remove image directory at same time
func (pc *ProductsController) Delete(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	var alert *views.Alert
	if err := pc.productService.Delete(id); err != nil {
		alert = &views.Alert{Level: "Warning", Message: err.Error()}
	} else {
		alert = &views.Alert{Level: "Success", Message: "Deleted!"}
	}
	alert.PersistAlert(w)
	http.Redirect(w, r, "/productsindex", http.StatusFound)
}

// TODO Seperate image processing to seperate Controller
// POST product/:id/uploadimage
func (pc *ProductsController) ImageUpload(w http.ResponseWriter, r *http.Request) {
	var yield views.Page
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		yield.SetAlert(err)
		pc.editView.RenderTemplate(w, r, yield)
		return
	}

	if err := r.ParseMultipartForm(1 << 20); err != nil {
		yield.SetAlert(err)
		pc.editView.RenderTemplate(w, r, yield)
		return
	}
	product, _ := pc.productService.GetByID(id)
	// Iterate over uploaded files to process them.
	fheaders := r.MultipartForm.File["images"]
	for _, f := range fheaders {
		// Open the uploaded file
		file, err := f.Open()
		if err != nil {
			yield.SetAlert(err)
			pc.editView.RenderTemplate(w, r, yield)
			return
		}
		defer file.Close()

		// Create the image
		if err := pc.imageService.Create(product.ID, file, f.Filename); err != nil {
			yield.SetAlert(err)
			pc.editView.RenderTemplate(w, r, yield)
			return
		}
	}

	alert := &views.Alert{Level: "Success", Message: "Images uploaded successfully"}
	alert.PersistAlert(w)
	url := fmt.Sprintf("/product/%d/edit", product.ID)
	http.Redirect(w, r, url, http.StatusFound)
}

// /product/:id/deleteimage/:filename
func (pc *ProductsController) DeleteImage(w http.ResponseWriter, r *http.Request) {
	yield := views.Page{}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		yield.SetAlert(err)
		pc.editView.RenderTemplate(w, r, yield)
	}

	product, err := pc.productService.GetByID(id)
	if err != nil {
		yield.SetAlert(err)
		pc.editView.RenderTemplate(w, r, yield)
		return
	}
	yield.PageData = product

	filename := vars["filename"]
	// Build the Image model
	i := models.Image{
		Filename: filename,
		EntityID: product.ID,
	}
	// Try to delete the image.
	err = pc.imageService.Delete(&i)
	if err != nil {
		// Render the edit page with any errors.
		yield.PageData = product
		yield.SetAlert(err)
		pc.editView.RenderTemplate(w, r, yield)
		return
	}
	url := fmt.Sprintf("/product/%d/edit", product.ID)
	http.Redirect(w, r, url, http.StatusFound)
}
