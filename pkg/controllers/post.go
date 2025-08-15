package controllers

import (
	"net/http"
	"sort"
	"strconv"

	"github.com/CheerChen/konakore/pkg/models"
	"github.com/CheerChen/konakore/pkg/services"
	"github.com/CheerChen/konakore/pkg/syncer"

	"github.com/julienschmidt/httprouter"
	"github.com/morkid/paginate"
)

// GetPosts returns a handler that serves ranked posts.
func GetPosts(rs *services.RankerService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		query := r.URL.Query().Get("query")
		var posts []*models.Post
		page := paginate.New().With(models.GetPostsStmt(query)).Request(r).Response(&posts)

		// Get the current ranker, then score, sort, and build URLs
		ranker := rs.GetRanker()
		ranker.ScoreAll(posts)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].MyScore > posts[j].MyScore
		})
		for i := range posts {
			models.BuildURL(posts[i])
		}

		cJson(w, posts, map[string]int64{
			"total": page.Total,
			"page":  page.Page,
			"size":  page.Size,
		})
	}
}

// GetLikes returns a handler that serves ranked liked posts.
func GetLikes(rs *services.RankerService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
		query := r.URL.Query().Get("query")
		var posts []*models.Post
		page := paginate.New().With(models.GetLikesStmt(query)).Request(r).Response(&posts)

		// Get the current ranker, then score, sort, and build URLs
		ranker := rs.GetRanker()
		ranker.ScoreAll(posts)
		sort.Slice(posts, func(i, j int) bool {
			return posts[i].MyScore > posts[j].MyScore
		})
		for i := range posts {
			models.BuildURL(posts[i])
		}

		cJson(w, posts, map[string]int64{
			"total": page.Total,
			"page":  page.Page,
			"size":  page.Size,
		})
	}
}

// Like returns a handler for liking a post and triggers a ranker retrain.
func Like(rs *services.RankerService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

		// Trigger a retrain in the background
		go rs.Retrain()

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
}

// Unlike returns a handler for unliking a post and triggers a ranker retrain.
func Unlike(rs *services.RankerService) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

		// Trigger a retrain in the background
		go rs.Retrain()

		cJson(w, "OK", nil)
	}
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
