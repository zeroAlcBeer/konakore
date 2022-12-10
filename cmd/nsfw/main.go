package main

import (
	log "github.com/kataras/golog"
	"konakore/pkg/models"
	"os"
)

func main() {

	models.OpenDb()
	wpath := os.Getenv("wpath")
	pics := models.LoadFiles(wpath)

	fileMap := make(map[int64]string)

	for _, pic := range pics {
		fileMap[pic.Id] = pic.Name
	}
	var likes []models.Post
	models.GetLikesStmt("").Find(&likes)

	var count int64
	for _, like := range likes {
		if like.Rating != "s" {
			log.Infof("remove %s", fileMap[like.Id])
			count++
			_ = os.Remove(fileMap[like.Id])
		}
	}
	log.Infof("removed %v", count)
}
