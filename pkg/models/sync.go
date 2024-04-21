package models

import (
	"os"

	log "github.com/kataras/golog"
)

const wpath = "/wallpaper"

func CheckPath() {
	if err := ensureDir(wpath); err != nil {
		log.Fatalf("Error reading path, %s", err)
	}
}

func ensureDir(dirName string) error {
	err := os.Mkdir(dirName, os.ModeDir)
	if err == nil || os.IsExist(err) {
		return nil
	}
	return err
}

func AddLocalPosts() {
	pics := LoadFiles(wpath)
	if len(pics) == 0 {
		log.Warnf("Wallpaper path empty!")
		return
	}

	var ids []int64
	for _, pic := range pics {
		ids = append(ids, pic.Id)
	}

	err := LikeAll(ids)
	if err != nil {
		log.Errorf("sync like post err: %s", err)
	}

	log.Infof("synced: %d", len(ids))
}

func AddRemotePosts() {
	pics := LoadFiles(wpath)
	if len(pics) == 0 {
		log.Warnf("Wallpaper path empty!")
		return
	}

	bMap := make(map[int64]bool)
	for _, pic := range pics {
		bMap[pic.Id] = true
	}

	pts := GetLikes()

	for _, post := range pts {
		if !bMap[post.Id] {
			log.Infof("found lost post: %d", post.Id)
			//BuildURL(post)
			//log.Infof("name built: %s", post.Tags)
			//if post.JpegFileSize != 0 && post.FileSize > (post.JpegFileSize*10) {
			//	go DownloadFile(&KFile{Id: post.Id, Tags: post.Tags}, post.JpegURL)
			//} else {
			//	go DownloadFile(&KFile{Id: post.Id, Tags: post.Tags}, post.FileURL)
			//}
		}
	}
}
