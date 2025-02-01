package controllers

import (
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/morkid/paginate"

	"konakore/pkg/models"
	"konakore/pkg/syncer"
)

// GetPosts ...
func GetPosts(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	query := r.URL.Query().Get("query")
	//?size=10&page=0&sort=-name
	var posts []models.Post
	page := paginate.New().With(models.GetPostsStmt(query)).Request(r).Response(&posts)

	weight := models.NewTagWeightSystem()
	weight.Learn(models.GetLikes())

	for index := range posts {
		weight.ScorePost(&posts[index])
		models.BuildURL(&posts[index])
	}

	cJson(w, posts, map[string]int64{
		"total": page.Total,
		"page":  page.Page,
		"size":  page.Size,
	})
	return
}

// GetLikes ...
func GetLikes(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	query := r.URL.Query().Get("query")
	//?size=10&page=0&sort=-name
	var posts []models.Post
	page := paginate.New().With(models.GetLikesStmt(query)).Request(r).Response(&posts)

	weight := models.NewTagWeightSystem()
	weight.Learn(models.GetLikes())
	for index := range posts {
		weight.ScorePost(&posts[index])
		models.BuildURL(&posts[index])
	}

	cJson(w, page.Items, map[string]int64{
		"total": page.Total,
		"page":  page.Page,
		"size":  page.Size,
	})
	return

}

// Like ...
func Like(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}
	post := &models.Post{}
	err = post.First(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	go post.Like(id)

	models.BuildURL(post)
	var target string
	if post.JpegFileSize != 0 {
		target = post.JpegURL
	} else {
		target = post.FileURL
	}
	models.DownloadFile(&models.KFile{Id: post.Id, Tags: post.Tags}, target)
	cJson(w, "OK", nil)
	return
}

// Unlike ...
func Unlike(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}
	post := &models.Post{}
	err = post.First(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	err = post.Unlike(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	err = models.DeleteFile(id)

	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	cJson(w, "OK", nil)
	return
}

// Force ...
func Force(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	p, err := strconv.Atoi(r.URL.Query().Get("p"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}
	syncer.ForceUpdatePosts(p)
	cJson(w, "OK", nil)
	return
}
