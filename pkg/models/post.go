package models

import (
	"fmt"
	"net/url"

	"gorm.io/gorm"
)

type Post struct {
	Id             int64  `gorm:"column:id" json:"id" form:"id"`
	Tags           string `gorm:"column:tags" json:"tags" form:"tags"`
	CreatedAt      int    `gorm:"column:created_at" json:"created_at" form:"created_at"`
	Source         string `gorm:"column:source" json:"source" form:"source"`
	Score          int    `gorm:"column:score" json:"score" form:"score"`
	Md5            string `gorm:"column:md5" json:"md5" form:"md5"`
	FileSize       int    `gorm:"column:file_size" json:"file_size" form:"file_size"`
	SampleWidth    int    `gorm:"column:sample_width" json:"sample_width" form:"sample_width"`
	SampleHeight   int    `gorm:"column:sample_height" json:"sample_height" form:"sample_height"`
	SampleFileSize int    `gorm:"column:sample_file_size" json:"sample_file_size" form:"sample_file_size"`
	JpegWidth      int    `gorm:"column:jpeg_width" json:"jpeg_width" form:"jpeg_width"`
	JpegHeight     int    `gorm:"column:jpeg_height" json:"jpeg_height" form:"jpeg_height"`
	JpegFileSize   int    `gorm:"column:jpeg_file_size" json:"jpeg_file_size" form:"jpeg_file_size"`
	Rating         string `gorm:"column:rating" json:"rating" form:"rating"`
	Status         string `gorm:"column:status" json:"status" form:"status"`
	Width          int    `gorm:"column:width" json:"width" form:"width"`
	Height         int    `gorm:"column:height" json:"height" form:"height"`
	ParentId       int64  `gorm:"column:parent_id" json:"parent_id" form:"parent_id"`

	Likes      *Like  `gorm:"foreignKey:id" json:"likes"`
	FileURL    string `gorm:"-" json:"file_url"`
	SampleURL  string `gorm:"-" json:"sample_url"`
	JpegURL    string `gorm:"-" json:"jpeg_url"`
	PreviewURL string `gorm:"-" json:"preview_url"`
	//
	MyScore     float64 `gorm:"-" json:"my_score"`
	WaifuPillow bool    `gorm:"-" json:"waifu_pillow"`

	Alg map[string]float64 `gorm:"-" json:"Alg"`
}

func (p *Post) Save() (err error) {
	return db.Save(p).Error
}

func (p *Post) Last() (err error) {
	return db.Last(p).Error
}

func (p *Post) First(id int64) (err error) {
	return db.Where("id = ?", id).First(p).Error
}

func GetPostsStmt(query string) *gorm.DB {
	stmt := db.Model(&[]Post{}).Preload("Likes")
	if query != "" {
		stmt = stmt.Where("MATCH (`tags`) AGAINST (?)", query)
	}
	return stmt
}

func GetPostsInRange(start, end int64) []*Post {
	var posts []*Post
	db.Where("id >= ? AND id < ?", start, end).Find(&posts)
	return posts
}

func BuildURL(p *Post) {
	prefix := "https://konachan.net"

	p.SampleURL, _ = urlEncoded(fmt.Sprintf("%s/sample/%s/Konachan.com - %d sample.jpg", prefix, p.Md5, p.Id))
	p.JpegURL, _ = urlEncoded(fmt.Sprintf("%s/jpeg/%s/Konachan.com - %d %s.jpg", prefix, p.Md5, p.Id, p.Tags))
	p.FileURL, _ = urlEncoded(fmt.Sprintf("%s/image/%s/1.jpg", prefix, p.Md5))

	if p.SampleFileSize == 0 && p.JpegFileSize == 0 {
		p.FileURL, _ = urlEncoded(fmt.Sprintf("%s/image/%s/1.jpg", prefix, p.Md5))

		// can be gif
	}
}

func urlEncoded(str string) (string, error) {
	u, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
