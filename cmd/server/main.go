package main

import (
	"github.com/NYTimes/gziphandler"
	"github.com/julienschmidt/httprouter"
	log "github.com/kataras/golog"
	"github.com/rs/cors"
	"net/http"

	"konakore/pkg/controllers"
	"konakore/pkg/models"
)

func main() {

	models.OpenDb()
	models.CheckPath()
	models.Sync()
	models.UpdateTfIdf()

	router := httprouter.New()

	// assets
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		http.ServeFile(w, r, "/assets/index.html")
	})

	router.GET("/favicon.ico", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		http.ServeFile(w, r, "/assets/favicon.ico")
	})

	// post
	router.GET("/posts", controllers.GetPosts)
	router.GET("/likes", controllers.GetLikes)
	router.GET("/like/:id", controllers.Like)
	router.GET("/unlike/:id", controllers.Unlike)
	router.GET("/sample/:id", controllers.Sample)

	handler := cors.Default().Handler(router)
	withGz := gziphandler.GzipHandler(handler)

	log.Infof("HTTP listening at: %s", ":80")
	log.Fatalf("%s", http.ListenAndServe(":80", withGz))
}
