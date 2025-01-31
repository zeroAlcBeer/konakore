package syncer

import (
	"fmt"

	myclient "konakore/pkg/client"
	"konakore/pkg/models"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"k8s.io/klog"
)

type Post struct {
	Id             int64  `gorm:"column:id" json:"id" form:"id"`
	Tags           string `gorm:"column:tags" json:"tags" form:"tags"`
	CreatedAt      int    `gorm:"column:created_at" json:"created_at" form:"created_at"`
	Source         string `gorm:"column:source" json:"source" form:"source"`
	Score          int    `gorm:"column:score" json:"score" form:"score"`
	Md5            string `gorm:"column:md5" json:"md5" form:"md5"`
	FileSize       int    `gorm:"column:file_size" json:"file_size" form:"file_size"`
	FileURL        string `gorm:"-" json:"file_url"`
	SampleURL      string `gorm:"-" json:"sample_url"`
	SampleWidth    int    `gorm:"column:sample_width" json:"sample_width" form:"sample_width"`
	SampleHeight   int    `gorm:"column:sample_height" json:"sample_height" form:"sample_height"`
	SampleFileSize int    `gorm:"column:sample_file_size" json:"sample_file_size" form:"sample_file_size"`
	JpegURL        string `gorm:"-" json:"jpeg_url"`
	JpegWidth      int    `gorm:"column:jpeg_width" json:"jpeg_width" form:"jpeg_width"`
	JpegHeight     int    `gorm:"column:jpeg_height" json:"jpeg_height" form:"jpeg_height"`
	JpegFileSize   int    `gorm:"column:jpeg_file_size" json:"jpeg_file_size" form:"jpeg_file_size"`
	Rating         string `gorm:"column:rating" json:"rating" form:"rating"`
	Status         string `gorm:"column:status" json:"status" form:"status"`
	Width          int    `gorm:"column:width" json:"width" form:"width"`
	Height         int    `gorm:"column:height" json:"height" form:"height"`
	ParentId       int64  `gorm:"column:parent_id" json:"parent_id" form:"parent_id"`
}

type Tag struct {
	ID    int64  `gorm:"column:id" json:"id" form:"id"`
	Name  string `gorm:"column:name" json:"name" form:"name"`
	Count int64  `gorm:"column:count" json:"count" form:"count"`
}

var (
	db *gorm.DB
)

func InitDB() {
	db = models.GetDb()
}

var reqclient = myclient.New()

func SetProxyUrl(proxy string) {
	reqclient.SetProxyUrl(proxy)
}

func getPosts(page int) ([]*Post, error) {
	var posts []*Post

	u := fmt.Sprintf("https://konachan.net/post.json?limit=%d&page=%d", 100, page)

	klog.Infof("request url: %s", u)
	err := reqclient.GetJSON(u, &posts)
	//str, err := getJson(u)
	//err = json.Unmarshal([]byte(str), &posts)

	return posts, err
}

func getTags(limit int) ([]*Tag, error) {
	var tags []*Tag

	u := fmt.Sprintf("https://konachan.net/tag.json?limit=%d&order=count", limit)

	klog.Infof("request url: %s", u)
	err := reqclient.GetJSON(u, &tags)
	//str, err := getJson(u)
	//err = json.Unmarshal([]byte(str), &tags)

	return tags, err
}

func currentPage() int {
	var count int64
	_ = db.Model(&Post{}).Count(&count)
	return int(count/100) + 1
}

func updatePosts(page int) error {
	klog.Infof("get posts... page: %d", page)

	posts, err := getPosts(page)
	if err != nil {
		klog.Errorf("get posts err: %s", err)
		return err
	}

	err = db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&posts).Error

	if err != nil {
		klog.Warningf("update posts to db err: %s", err)
	}
	return err
}
