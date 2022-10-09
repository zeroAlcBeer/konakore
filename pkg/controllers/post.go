package controllers

import (
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/morkid/paginate"

	"konakore/pkg/models"
)

// GetPosts ...
func GetPosts(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	query := r.URL.Query().Get("query")
	//?size=10&page=0&sort=-name
	var posts []models.Post
	page := paginate.New().With(models.GetPostsStmt(query)).Request(r).Response(&posts)

	avg := models.AvgMap(posts)
	//likes := models.Likes()
	for index := range posts {
		models.Mark(&posts[index], avg)
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

	avg := models.AvgMap(posts)
	for index := range posts {
		models.Mark(&posts[index], avg)
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
	err = post.Like(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}

	models.BuildURL(post)
	if post.JpegFileSize != 0 && post.FileSize > (post.JpegFileSize*10) {
		go models.DownloadFile(&models.KFile{Id: post.Id, Tags: post.Tags}, post.JpegURL)
	} else {
		go models.DownloadFile(&models.KFile{Id: post.Id, Tags: post.Tags}, post.FileURL)
	}
	models.UpdateTfIdf()
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
	models.UpdateTfIdf()
	cJson(w, "OK", nil)
	return
}

// Sample ...
func Sample(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	id, err := strconv.ParseInt(ps.ByName("id"), 10, 64)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}

	pic, err := models.GetFileById(id)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	if pic.Header == "" {
		http.Error(w, "file format error", http.StatusNotFound)
		return
	}

	byte, err := ioutil.ReadFile(pic.Name)

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}
	w.Header().Set("Content-type", pic.Header)
	w.Header().Set("Cache-control", "max-age=315360000")
	w.Write(byte)
	byte = nil
}
