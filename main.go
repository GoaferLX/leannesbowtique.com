package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"

	"leannesbowtique.com/controllers"
	"leannesbowtique.com/middleware"
	"leannesbowtique.com/models"
	"leannesbowtique.com/rand"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/csrf"
	"github.com/gorilla/mux"
	"github.com/mailgun/mailgun-go"
)

func main() {
	prod := flag.Bool("prod", false, "Provide this flag in production. This ensures that a .config file is provided before the application starts.")
	flag.Parse()

	cfg := LoadConfig(*prod)
	dbcfg := cfg.Database
	mgcfg := cfg.Mailgun
	mgclient := mailgun.NewMailgun(mgcfg.Domain, mgcfg.APIKey)
	mgclient.SetAPIBase(mailgun.APIBaseEU)

	services, err := models.NewServices(
		models.WithGorm(dbcfg.Dialect, dbcfg.dsn(), cfg.isProd()),
		models.WithUsers(cfg.PWPepper, cfg.HMACKey),
		models.WithProducts(),
		models.WithCategories(),
		//	models.WithImages(),
		models.WithBundles(),
	)
	if err != nil {
		log.Fatal(err)
	}
	if err := services.AutoMigrate(); err != nil {
		log.Fatal("Could not initiate database tables")
	}
	mailController := controllers.NewMail(mgclient)
	usersController := controllers.NewUsers(services.UserService, mailController)
	productsController := controllers.NewProductsController(services.ProductService, models.NewImageService("products"))
	categoryController := controllers.NewCategories(services.CategoryService)
	bundlesController := controllers.NewBundlesController(services.BundleService)

	// Inititate middlewares
	userMW := middleware.User{UserModel: services.UserService}
	authMW := middleware.RequireUser{}
	csrfbytes, err := rand.Bytes(32)
	if err != nil {
		log.Fatal(err)
	}
	csrfmw := csrf.Protect(csrfbytes, csrf.Secure(cfg.isProd()))

	r := mux.NewRouter()
	r.NotFoundHandler = controllers.StaticPage().NotFound
	// Asset routes
	assetsHandler := http.FileServer(http.Dir("./assets/"))
	r.PathPrefix("/css/").Handler(assetsHandler)
	r.PathPrefix("/js/").Handler(assetsHandler)
	r.PathPrefix("/imgs/").Handler(assetsHandler)
	// Image routes
	imageHandler := http.FileServer(http.Dir("./images/"))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", imageHandler))
	// Static pages
	r.Handle("/about", controllers.StaticPage().About).Methods("GET")
	// Contact rputes
	r.Handle("/contact", mailController.ContactView).Methods("GET")
	r.HandleFunc("/contact", mailController.Contact).Methods("POST")
	// User routes
	r.Handle("/signup", authMW.AllowHandle(usersController.SignUpView)).Methods("GET")
	r.HandleFunc("/signup", authMW.AllowFunc(usersController.Create)).Methods("POST")
	r.Handle("/login", usersController.LoginView).Methods("GET")
	r.HandleFunc("/login", usersController.Login).Methods("POST")
	r.HandleFunc("/logout", authMW.AllowFunc(usersController.Logout))

	r.Handle("/forgot", usersController.ForgotPWView).Methods("GET")
	r.HandleFunc("/forgot", usersController.Forgot).Methods("POST")
	r.HandleFunc("/reset", usersController.ResetPW).Methods("GET")
	r.HandleFunc("/reset", usersController.Reset).Methods("POST")
	// Product routes
	r.Handle("/product/new", authMW.AllowFunc(productsController.NewProduct)).Methods("GET")
	r.HandleFunc("/product/new", authMW.AllowFunc(productsController.Create)).Methods("POST")
	r.HandleFunc("/product/{id:[0-9]+}", productsController.ViewProduct).Methods("GET")
	r.HandleFunc("/product/{id:[0-9]+}/edit", authMW.AllowFunc(productsController.Edit)).Methods("GET")
	r.HandleFunc("/product/{id:[0-9]+}/edit", authMW.AllowFunc(productsController.Update)).Methods("POST")
	r.HandleFunc("/product/{id:[0-9]+}/delete", authMW.AllowFunc(productsController.Delete)).Methods("GET")
	// Product image handling
	r.HandleFunc("/product/{id:[0-9]+}/uploadimage", authMW.AllowFunc(productsController.ImageUpload)).Methods("POST")
	r.HandleFunc("/product/{id:[0-9]+}/deleteimage/{filename}", authMW.AllowFunc(productsController.DeleteImage)).Methods("POST")
	// Products collective
	r.HandleFunc("/", productsController.ViewProducts).Methods("GET")
	r.HandleFunc("/products", productsController.ViewProducts).Methods("GET")
	r.HandleFunc("/productsindex", authMW.AllowFunc(productsController.ViewProductsIndex)).Methods("GET")

	r.HandleFunc("/product/category/", authMW.AllowFunc(categoryController.View)).Methods("GET")
	r.HandleFunc("/product/category/", authMW.AllowFunc(categoryController.Create)).Methods("POST")
	r.HandleFunc("/product/category/{id:[0-9]+}", authMW.AllowFunc(categoryController.Update)).Methods("POST")
	r.HandleFunc("/product/category/{id:[0-9]+}", authMW.AllowFunc(categoryController.Delete)).Methods("GET")

	r.HandleFunc("/bundle/view", bundlesController.ViewBundle).Methods("GET")
	r.HandleFunc("/bundles/view", bundlesController.ViewBundles).Methods("GET")
	log.Printf("Server listening on port: %d", cfg.Port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", cfg.Port), csrfmw(userMW.Allow(r))))

}
