package models

import (
	log "github.com/kataras/golog"
	"os"
)

var wpath = ""

func CheckPath() {
	wpath = os.Getenv("wpath")
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

func Sync() {
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

	return
}
