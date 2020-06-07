package controllers

import (
	"errors"
	"fmt"
	"goafweb/models"
	"goafweb/views"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

type ArticlesController struct {
	ArticlesModel models.ArticleService
	New           *views.View
	articleView   *views.View
	editView      *views.View
	articlesView  *views.View
}
type ArticleForm struct {
	Title   string `schema:"title"`
	Content string `schema:"content"`
}

func NewArticlesController(model models.ArticleService) *ArticlesController {
	return &ArticlesController{
		ArticlesModel: model,
		New:           views.NewView("index.gohtml", "views/article/newarticle.gohtml"),
		articleView:   views.NewView("index.gohtml", "views/article/showarticle.gohtml"),
		editView:      views.NewView("index.gohtml", "views/article/editarticle.gohtml"),
		articlesView:  views.NewView("index.gohtml", "views/article/articlesindex.gohtml"),
	}
}

func (ac *ArticlesController) ViewArticle(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	id, _ := strconv.Atoi(vars["id"])
	page := views.Page{}
	article, err := ac.ArticlesModel.GetArticleByID(id)
	if err != nil {
		page.Alert = &views.Alert{
			Level:   "warning",
			Message: err.Error(),
		}

		ac.articleView.RenderTemplate(w, r, page)
		return
	}

	ac.articleView.RenderTemplate(w, r, article)
}

func (ac *ArticlesController) Edit(w http.ResponseWriter, r *http.Request) {
	yield := views.Page{}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		yield.SetAlert(err)
		ac.editView.RenderTemplate(w, r, yield)
	}
	article, err := ac.ArticlesModel.GetArticleByID(id)
	if err != nil {
		yield.SetAlert(err)
		ac.editView.RenderTemplate(w, r, yield)
		return
	}

	ctx := r.Context()
	user := User(ctx)

	if article.Author != user.ID {
		yield.SetAlert(errors.New("You don't have permission to view this page"))
		ac.editView.RenderTemplate(w, r, yield)
		return
	}
	yield.PageData = article
	ac.editView.RenderTemplate(w, r, yield)

}
func (ac ArticlesController) Create(w http.ResponseWriter, r *http.Request) {
	var page views.Page
	var form ArticleForm
	if err := parsePostForm(r, &form); err != nil {
		page.SetAlert(err)
		ac.New.RenderTemplate(w, r, page)
		return
	}
	ctx := r.Context()
	user := User(ctx)

	article := models.Article{
		Title:   form.Title,
		Content: form.Content,
		Author:  user.ID,
	}
	if err := ac.ArticlesModel.Create(&article); err != nil {
		page.SetAlert(err)
		ac.New.RenderTemplate(w, r, page)
		return
	}
	url := fmt.Sprintf("/article/%d", article.ID)
	http.Redirect(w, r, url, http.StatusFound)
}
func (ac *ArticlesController) Update(w http.ResponseWriter, r *http.Request) {
	yield := views.Page{}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		yield.SetAlert(err)
		ac.editView.RenderTemplate(w, r, yield)
	}
	article, err := ac.ArticlesModel.GetArticleByID(id)
	if err != nil {
		yield.SetAlert(err)
		ac.editView.RenderTemplate(w, r, yield)
		return
	}

	ctx := r.Context()
	userctx := ctx.Value("user")
	user := userctx.(*models.User)

	if article.Author != user.ID {
		http.Error(w, "permissions error", http.StatusForbidden)
		return
	}
	yield.PageData = article
	var form ArticleForm
	if err := parsePostForm(r, &form); err != nil {
		yield.SetAlert(err)
		ac.editView.RenderTemplate(w, r, yield)
		return
	}
	article.Title = form.Title
	article.Content = form.Content
	err = ac.ArticlesModel.Update(article)
	if err != nil {
		yield.SetAlert(err)
		ac.editView.RenderTemplate(w, r, yield)
		return
	}
	url := fmt.Sprintf("/article/%d", article.ID)
	http.Redirect(w, r, url, http.StatusFound)
}

func (ac *ArticlesController) Delete(w http.ResponseWriter, r *http.Request) {
	yield := views.Page{}
	vars := mux.Vars(r)
	id, err := strconv.Atoi(vars["id"])
	if err != nil {
		yield.SetAlert(err)
		ac.editView.RenderTemplate(w, r, yield)
	}
	article, err := ac.ArticlesModel.GetArticleByID(id)
	if err != nil {
		yield.SetAlert(err)
		ac.editView.RenderTemplate(w, r, yield)
		return
	}

	ctx := r.Context()
	userctx := ctx.Value("user")
	user := userctx.(*models.User)

	if article.Author != user.ID {
		http.Error(w, "You are not permitted to delete this!", http.StatusForbidden)
		return
	}

	err = ac.ArticlesModel.Delete(article.ID)
	if err != nil {
		yield.SetAlert(err)
		ac.editView.RenderTemplate(w, r, yield)
		return
	}
	fmt.Fprint(w, "article deleted")
}

// GET /galleries
func (ac *ArticlesController) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userctx := ctx.Value("user")
	user := userctx.(*models.User)

	articles, err := ac.ArticlesModel.GetArticlesByUser(user.ID)
	if err != nil {
		http.Error(w, "Something went wrong.", http.StatusInternalServerError)
		return
	}
	var yield views.Page
	yield.PageData = articles
	ac.articlesView.RenderTemplate(w, r, yield)
}
