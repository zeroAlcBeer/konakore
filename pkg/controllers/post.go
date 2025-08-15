package controllers

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/julienschmidt/httprouter"
	"github.com/morkid/paginate"

	"github.com/CheerChen/konakore/pkg/models"
	"github.com/CheerChen/konakore/pkg/ranker"
	"github.com/CheerChen/konakore/pkg/ranker/tfidf_hybrid"
	"github.com/CheerChen/konakore/pkg/syncer"
)

var (
	// Global ranker instance, initialized once.
	defaultRanker ranker.Ranker
)

// init initializes the default ranker for the application.
// It learns from all liked posts at startup.
func init() {
	defaultRanker = tfidf_hybrid.NewTfidfHybridRanker()
	defaultRanker.Learn(models.GetLikes())
}

// rankAndRespond handles the common logic of scoring, sorting, building URLs, and responding with JSON.
func rankAndRespond(w http.ResponseWriter, page *paginate.Page, posts []*models.Post) {
	// Score all posts using the default ranker.
	defaultRanker.ScoreAll(posts)

	// Sort posts by their new score in descending order.
	sort.Slice(posts, func(i, j int) bool {
		return posts[i].MyScore > posts[j].MyScore
	})

	// Build URLs for each post.
	for i := range posts {
		models.BuildURL(posts[i])
	}

	// Respond with JSON.
	cJson(w, posts, map[string]int64{
		"total": page.Total,
		"page":  page.Page,
		"size":  page.Size,
	})
}

// GetPosts fetches, ranks, and serves posts.
func GetPosts(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	query := r.URL.Query().Get("query")
	var posts []*models.Post
	page := paginate.New().With(models.GetPostsStmt(query)).Request(r).Response(&posts)

	rankAndRespond(w, &page, posts)
}

// GetLikes fetches, ranks, and serves liked posts.
func GetLikes(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	query := r.URL.Query().Get("query")
	var posts []*models.Post
	page := paginate.New().With(models.GetLikesStmt(query)).Request(r).Response(&posts)

	rankAndRespond(w, &page, posts)
}

// Like handles liking a post.
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
}

// Unlike handles unliking a post.
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
}

// Force triggers a force update of posts.
func Force(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	p, err := strconv.Atoi(r.URL.Query().Get("p"))

	if err != nil {
		http.Error(w, err.Error(), http.StatusNotAcceptable)
		return
	}
	syncer.ForceUpdatePosts(p)
	cJson(w, "OK", nil)
}
