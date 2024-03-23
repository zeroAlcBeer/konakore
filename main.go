package main

import (
	"net/http"
	"os"

	"github.com/NYTimes/gziphandler"
	"github.com/julienschmidt/httprouter"
	log "github.com/kataras/golog"
	"github.com/rs/cors"

	"konakore/pkg/controllers"
	"konakore/pkg/models"
	"konakore/pkg/syncer"
)

func main() {
	// init db
	var err error
	dsn := os.Getenv("dsn")
	env := os.Getenv("env")
	_, err = models.OpenDb(dsn, env)
	if err != nil {
		log.Fatal(err)
	}

	// init syncer
	syncer.InitDB()
	proxy := os.Getenv("proxy")
	syncer.SetProxyUrl(proxy)
	spec := os.Getenv("sync_spec")
	syncer.AddCron(spec)

	// init local files
	models.CheckPath()
	models.AddLocalPosts()
	models.AddRemotePosts()
	models.UpdateTfIdf()

	router := httprouter.New()

	// assets
	router.GET("/", serveFile("/assets/index.html"))
	router.GET("/favicon.ico", serveFile("/assets/favicon.ico"))
	router.GET("/likes", serveFile("/assets/likes.html"))

	// post
	router.GET("/api/posts", controllers.GetPosts)
	router.GET("/api/likes", controllers.GetLikes)

	router.POST("/like/:id", controllers.Like)
	router.POST("/unlike/:id", controllers.Unlike)
	router.GET("/sample/:id", controllers.Sample)

	router.GET("/force", controllers.Force)

	handler := cors.Default().Handler(router)
	withGz := gziphandler.GzipHandler(handler)

	log.Infof("HTTP listening at: %s", ":80")
	log.Fatalf("%s", http.ListenAndServe(":80", withGz))
}

func serveFile(filename string) httprouter.Handle {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		http.ServeFile(w, r, filename)
	}
}
