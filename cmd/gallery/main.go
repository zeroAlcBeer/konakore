package main

import (
	"embed"
	"net/http"

	"github.com/NYTimes/gziphandler"
	"github.com/julienschmidt/httprouter"
	log "github.com/kataras/golog"
	"github.com/rs/cors"

	"konakore/pkg/controllers"
	"konakore/pkg/models"
)

var (
	//go:embed assets/*
	f embed.FS
)

func main() {

	models.OpenDb()
	models.CheckPath()
	models.Sync()
	models.UpdateTfIdf()

	router := httprouter.New()

	// assets
	router.GET("/", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		bytes, err := f.ReadFile("assets/index.html")
		if err != nil {
			http.NotFound(w, r)
			return
		}
		w.Write(bytes)
	})
	router.ServeFiles("/web/*filepath", http.FS(f))

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
