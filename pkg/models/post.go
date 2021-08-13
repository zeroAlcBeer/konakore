package models

import (
	"fmt"
	"math"
	"net/url"
	"strings"

	log "github.com/kataras/golog"
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

	IsLike      bool    `gorm:"-" json:"is_like"`
	FileURL     string  `gorm:"-" json:"file_url"`
	SampleURL   string  `gorm:"-" json:"sample_url"`
	JpegURL     string  `gorm:"-" json:"jpeg_url"`
	PreviewURL  string  `gorm:"-" json:"preview_url"`
	TfIDf       float64 `gorm:"-" json:"tf_idf"`
	MyScore     float64 `gorm:"-" json:"my_score"`
	WaifuPillow bool    `gorm:"-" json:"waifu_pillow"`
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
	stmt := db.Model(&[]Post{})
	if query != "" {
		stmt = stmt.Where("MATCH (`tags`) AGAINST (?)", query)
	}
	return stmt
}

func Mark(p *Post, avgMap map[string]float64) {
	tags := strings.Split(p.Tags, " ")

	// version 2
	for _, tag := range tags {
		if t, ok := tfIdf[tag]; ok {
			p.TfIDf += t
		}
	}
	p.TfIDf = p.TfIDf / float64(len(tags))
	p.MyScore = p.TfIDf + math.Log(float64(p.Score+1)/avgMap[p.Rating])/float64(len(tags))

	p.WaifuPillow = p.Width > p.Height*2
	//_ = p.SortTagsByTfIdf(tfIdf)
}

func AvgMap(ps []Post) map[string]float64 {
	avgMap := make(map[string]float64)
	sumMap := make(map[string]float64)
	lenMap := make(map[string]int)
	for _, post := range ps {
		if _, ok := sumMap[post.Rating]; !ok {
			sumMap[post.Rating] = float64(post.Score)
		} else {
			sumMap[post.Rating] += float64(post.Score)
		}

		if _, ok := lenMap[post.Rating]; !ok {
			lenMap[post.Rating] = 1
		} else {
			lenMap[post.Rating] += 1
		}
	}

	for ranting, sum := range sumMap {
		if l, ok := lenMap[ranting]; ok {
			avgMap[ranting] = float64(sum) / float64(l)
		}
	}
	log.Infof("created avgMap: %v", avgMap)

	return avgMap
}

var tfIdf map[string]float64

func UpdateTfIdf() {
	lastId := int64(30 * 10000)
	post := &Post{}
	err := post.Last()
	if err != nil {
		log.Warnf("get lastid err: %s", err)
	}
	lastId = post.Id
	log.Infof("post last id: %d", lastId)

	pts := GetLikeTags()
	tf1 := make(map[string]int)
	tf2 := make(map[string]int)

	for _, pt := range pts {
		tags := strings.Split(pt, " ")
		for _, tag := range tags {

			if _, ok := tf1[tag]; !ok {
				tf1[tag] = 1
			} else {
				tf1[tag] += 1
			}

			if _, ok := tf2[tag]; !ok {
				tf2[tag] = len(tags)
			} else {
				tf2[tag] += len(tags)
			}

		}
	}

	tfIdf = make(map[string]float64)
	//idfMap := make(map[string]float64)

	countMap := GetTagCount()
	for tag, _ := range tf1 {
		if _, ok := countMap[tag]; !ok {
			countMap[tag] = 1
		}
		// 词频 = tag 在图片出现的次数 / 图片的 tag 总数
		tf := float64(tf1[tag]) / float64(tf2[tag])
		// 逆文档频率 = log( 图片总数 / 包含此 tag 的图片数 + 1）
		idf := math.Log(float64(lastId) / (float64(countMap[tag] + 1)))
		tfIdf[tag] = tf * idf
		//idfMap[tag] = idf
	}
	log.Infof("available tags: %d", len(tfIdf))
}

func BuildURL(p *Post) {
	p.SampleURL, _ = urlEncoded(fmt.Sprintf("https://konachan.com/sample/%s/Konachan.com - %d sample.jpg", p.Md5, p.Id))
	if p.IsLike {
		p.SampleURL = fmt.Sprintf("/sample/%d", p.Id)
	}

	p.JpegURL, _ = urlEncoded(fmt.Sprintf("https://konachan.com/jpeg/%s/Konachan.com - %d %s.jpg", p.Md5, p.Id, p.Tags))
	p.FileURL, _ = urlEncoded(fmt.Sprintf("https://konachan.com/image/%s/1.png", p.Md5))

	if p.SampleFileSize == 0 && p.JpegFileSize == 0 {
		p.FileURL, _ = urlEncoded(fmt.Sprintf("https://konachan.com/image/%s/1.jpg", p.Md5))
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
