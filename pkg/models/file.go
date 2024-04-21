package models

import (
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/imroc/req/v3"
	log "github.com/kataras/golog"

	myclient "konakore/pkg/client"
)

type KFile struct {
	Id   int64  `json:"id"`
	Name string `json:"name"`
	Ext  string `json:"ext"`
	Size int64
	Tags string `json:"tags"`

	Header string `json:"header"`
}

type KFiles []KFile

const FileNameLengthLimit = 200

func LoadFiles(path string) (pics KFiles) {
	err := filepath.Walk(path, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			log.Warnf("walk %v", path)
			return nil
		}

		id := regexp.MustCompile(`\d+`).FindString(info.Name())
		if id == "" {
			log.Warnf("walk id empty %v", info.Name())
			return nil
		}
		var pic KFile
		pic.Id, err = strconv.ParseInt(id, 10, 64)
		if err != nil {
			return nil
		}
		pic.Ext = filepath.Ext(info.Name())
		pic.Name = path
		pic.Size = info.Size()

		if strings.HasSuffix(pic.Name, ".png") {
			pic.Header = "image/png"
		} else if strings.HasSuffix(pic.Name, ".jpg") {
			pic.Header = "image/jpeg"
		} else if strings.HasSuffix(pic.Name, ".gif") {
			pic.Header = "image/gif"
		} else {
			return nil
		}

		pics = append(pics, pic)

		return nil
	})
	if err != nil {
		fmt.Printf("walk error [%v]\n", err)
	}

	return
}

func GetFileById(id int64) (pic KFile, err error) {
	pics := LoadFiles(wpath)

	for _, p := range pics {
		if p.Id == id {
			return p, nil
		}
	}

	return pic, errors.New("record not found")
}

func DeleteFile(id int64) (err error) {
	pic, err := GetFileById(id)

	if err != nil {
		return
	}
	return os.Remove(pic.Name)

}

func (pic *KFile) BuildName(u string) {
	// reduce tags
	tags := strings.Split(pic.Tags, " ")

	for len(pic.Tags) >= FileNameLengthLimit {
		tags = tags[:len(tags)-1]
		pic.Tags = strings.Join(tags, " ")
	}

	// replace special char
	var re = regexp.MustCompile(`[\\/:*?""<>|]`)
	pic.Tags = re.ReplaceAllString(pic.Tags, `$1`)

	if strings.Contains(u, "png") {
		pic.Ext = "png"
	}
	pic.Ext = "jpg"

	pic.Name = fmt.Sprintf("Konachan.com - %d %s.%s", pic.Id, pic.Tags, pic.Ext)
}

var reqclient = myclient.New()

// DownloadFile ...
func DownloadFile(file *KFile, u string) {
	file.BuildName(u)
	log.Infof("built name %s...", file.Name)

	// check
	log.Infof("downloading %v...", u)
	exist, err := reqclient.CheckDownloadUrl(u)
	if err != nil {
		log.Error(err)
	}
	if !exist {
		if strings.Contains(u, ".jpg") {
			u = strings.Replace(u, ".jpg", ".png", 1)
		} else if strings.Contains(u, ".png") {
			u = strings.Replace(u, ".png", ".gif", 1)
		} else if strings.Contains(u, ".gif") {
			log.Error("retry download limit")
			return
		}
		DownloadFile(file, u)
	}

	// path
	idxStr := fmt.Sprintf("%02d", file.Id/10000)
	err = ensureDir(path.Join(wpath, idxStr))
	if err != nil {
		log.Errorf("ensureDir err:", err)
		return
	}
	dst := path.Join(wpath, idxStr, file.Name)

	// done in callback
	callback := func(info req.DownloadInfo) {
		fmt.Printf("downloaded %.2f%%\n", float64(info.DownloadedSize)/float64(info.Response.ContentLength)*100.0)

		if info.Response.ContentLength != 0 && (info.DownloadedSize == info.Response.ContentLength) {
			err = (&Post{}).Done(file.Id)
		}

	}

	// real download
	err = reqclient.Download(u, dst, callback)
	if err != nil {
		log.Error(err)
	}
	log.Infof("save to ./%s", dst)
}
