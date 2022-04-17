package main

import (
	"fmt"
	log "github.com/kataras/golog"
	"github.com/robfig/cron/v3"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	myclient "konakore/pkg/client"
	"os"
)

var db *gorm.DB

func main() {
	var err error
	dsn := os.Getenv("dsn")
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	debug := os.Getenv("debug")
	if debug != "" {
		db = db.Debug()
	}

	// cron
	c := cron.New()
	c.Start()

	_, err = c.AddFunc(os.Getenv("spec_update_tag"), UpdateTags)
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.AddFunc(os.Getenv("spec_newest_post"), NewestPosts)
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.AddFunc(os.Getenv("spec_oldest_post"), OldestPosts)
	if err != nil {
		log.Fatal(err)
	}

	select {}
}

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

func getPosts(page int) ([]*Post, error) {
	var posts []*Post
	client := myclient.New()
	u := fmt.Sprintf("https://konachan.com/post.json?limit=%d&page=%d", 100, page)

	log.Infof("req: %s", u)
	err := client.GetJSON(u, &posts)

	return posts, err
}

func getTags(limit int) ([]*Tag, error) {
	var tags []*Tag
	client := myclient.New()

	u := fmt.Sprintf("https://konachan.com/tag.json?limit=%d&order=count", limit)

	log.Infof("req: %s", u)
	err := client.GetJSON(u, &tags)

	return tags, err
}

func currentPage() int {
	var count int64
	_ = db.Model(&Post{}).Count(&count)
	return int(count/100) + 1
}

func updatePosts(page int) {
	log.Infof("update posts in page %d...", page)
	posts, err := getPosts(page)
	if err != nil {
		log.Errorf("get posts err: %s", err)
		return
	}

	err = db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(&posts).Error

	if err != nil {
		log.Errorf("UpdateAll posts err: %s", err)
	}
}

func NewestPosts() {
	updatePosts(1)
}

func OldestPosts() {
	updatePosts(currentPage())
}

func UpdateTags() {
	log.Infof("update tags ...")
	tags, err := getTags(30000)
	if err != nil {
		log.Errorf("get tags err: %s", err)
		return
	}

	err = db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(tags[0:10000]).Error
	if err != nil {
		log.Errorf("UpdateAll tags err: %s", err)
	}

	err = db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(tags[10000:20000]).Error
	if err != nil {
		log.Errorf("UpdateAll tags err: %s", err)
	}

	err = db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(tags[20000:30000]).Error
	if err != nil {
		log.Errorf("UpdateAll tags err: %s", err)
	}
}
