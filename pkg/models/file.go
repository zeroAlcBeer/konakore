package models

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/gorilla/websocket"
	log "github.com/kataras/golog"
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

// DownloadFile ...
func DownloadFile(file *KFile, u string) {
	file.BuildName(u)
	log.Infof("built name %s...", file.Name)

	// check
	log.Infof("downloading %v...", u)

	// path
	idxStr := fmt.Sprintf("%02d", file.Id/10000)
	err := ensureDir(path.Join(wpath, idxStr))
	if err != nil {
		log.Errorf("ensureDir err:", err)
		return
	}
	dst := path.Join(wpath, idxStr, file.Name)

	err = downloadFile(idxStr, u, file.Name, "p3terx")
	if err != nil {
		fmt.Printf("Error downloading file: %v\n", err)
	}

	log.Infof("save to ./%s", dst)
}

// 调用aria2的JSON-RPC接口下载文件（通过WebSocket）
func downloadFile(folderPath, u, fileName, secret string) error {
	// 构造WebSocket连接
	dialer := websocket.Dialer{}
	conn, _, err := dialer.Dial("ws://192.168.151.62:6800/jsonrpc", nil)
	if err != nil {
		return err
	}
	defer conn.Close()

	//re := regexp.MustCompile(`\.(mp4|jpg)(\?[^?]*|:[^:]*|$)$`)
	//extMatches := re.FindStringSubmatch(u)
	//if len(extMatches) > 1 {
	//	ext := "." + extMatches[1]
	//	fileName += ext
	//}

	// 构造JSON-RPC请求体
	requestBody := map[string]interface{}{
		"jsonrpc": "2.0",
		"method":  "aria2.addUri",
		"id":      "qwer",
		"params": []interface{}{
			"token:" + secret, // 使用RPC密钥
			[]string{u},
			map[string]string{
				"dir":             wpath + "/" + folderPath,
				"out":             fileName,
				"allow-overwrite": "true",
			},
		},
	}

	// 序列化JSON请求体
	jsonData, err := json.Marshal(requestBody)
	if err != nil {
		return err
	}

	// 发送数据到WebSocket服务
	err = conn.WriteMessage(websocket.TextMessage, jsonData)
	if err != nil {
		return err
	}

	// 接收响应（可选）
	_, message, err := conn.ReadMessage()
	if err != nil {
		return err
	}
	fmt.Printf("Received: %s\n", message)

	return nil
}
