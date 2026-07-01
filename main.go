package main

import (
	"net/http"
	"os"

	"github.com/NYTimes/gziphandler"
	"github.com/julienschmidt/httprouter"
	log "github.com/kataras/golog"
	"github.com/rs/cors"

	"github.com/zeroAlcBeer/konakore/pkg/controllers"
	"github.com/zeroAlcBeer/konakore/pkg/models"
	"github.com/zeroAlcBeer/konakore/pkg/services"
	"github.com/zeroAlcBeer/konakore/pkg/syncer"
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

	// Init the ranker service, which trains the first model.
	rankerService := services.NewRankerService()

	router := httprouter.New()

	// assets
	router.GET("/", serveFile("/assets/index.html"))
	router.GET("/favicon.ico", serveFile("/assets/favicon.ico"))
	router.GET("/likes", serveFile("/assets/likes.html"))

	// post endpoints with ranker service injected
	router.GET("/api/posts", controllers.GetPosts(rankerService))
	router.GET("/api/likes", controllers.GetLikes(rankerService))

	router.POST("/like/:id", controllers.Like(rankerService))
	router.POST("/unlike/:id", controllers.Unlike(rankerService))

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