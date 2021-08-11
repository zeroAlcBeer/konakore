package models

import (
	"fmt"
	log "github.com/kataras/golog"
	"gorm.io/gorm"
	"math"
	"net/url"
	"strings"
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
	IsLike         bool   `gorm:"is_like" json:"is_like"`

	FileURL     string  `gorm:"-" json:"file_url"`
	SampleURL   string  `gorm:"-" json:"sample_url"`
	JpegURL     string  `gorm:"-" json:"jpeg_url"`
	PreviewURL  string  `gorm:"-" json:"preview_url"`
	TfIDf       float64 `gorm:"-" json:"tf_idf"`
	MyScore     float64 `gorm:"-" json:"my_score"`
	WaifuPillow bool    `gorm:"-" json:"waifu_pillow"`
}

type Tag struct {
	ID    int64  `gorm:"column:id" json:"id" form:"id"`
	Name  string `gorm:"column:name" json:"name" form:"name"`
	Count int64  `gorm:"column:count" json:"count" form:"count"`
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

func (p *Post) Unlike(id int64) (err error) {
	return db.Model(p).Update("is_like", 0).Error
}

func (p *Post) Like(id int64) (err error) {
	return db.Model(p).Update("is_like", 1).Error
}

func LikeAll(in []int64) (err error) {
	return db.Model(&Post{}).Where("id in (?)", in).Update("is_like", 1).Error
}

func Favorites() *gorm.DB {
	return db.Model(&[]Post{}).Where("is_like = ?", 1)
}

func GetPosts() *gorm.DB {
	return db.Model(&[]Post{})
}

func GetTags() map[string]int64 {
	tagCountMap := make(map[string]int64)
	var tags []*Tag
	err := db.Model(&[]Tag{}).Find(&tags).Error
	if err != nil {
		log.Error(err)
		return tagCountMap
	}
	for _, tag := range tags {
		tagCountMap[tag.Name] = tag.Count
	}
	return tagCountMap
}

func (p *Post) GetLikePosts() []string {
	var pts []string
	_ = db.Model(p).Where("is_like = ?", 1).Pluck("tags", &pts).Error
	return pts
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

	pts := post.GetLikePosts()
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

	countMap := GetTags()
	for tag, tf1 := range tf1 {
		if _, ok := countMap[tag]; !ok {
			countMap[tag] = 1
		}
		idf := math.Log(float64(lastId) / (float64(countMap[tag] + 1)))
		tf := float64(tf1) / float64(tf2[tag])
		tfIdf[tag] = tf * idf
		//idfMap[tag] = idf
	}
	log.Infof("available tags: %d", len(tfIdf))
}

func BuildURL(p *Post) {
	p.SampleURL, _ = urlEncoded(fmt.Sprintf("Konachan.com - %d sample.jpg", p.Id))
	p.SampleURL = fmt.Sprintf("https://konachan.com/sample/%s/%s", p.Md5, p.SampleURL)
	if p.IsLike {
		p.SampleURL = fmt.Sprintf("/sample/%d", p.Id)
	}

	p.JpegURL, _ = urlEncoded(fmt.Sprintf("Konachan.com - %d %s.jpg", p.Id, p.Tags))
	p.JpegURL = fmt.Sprintf("https://konachan.com/jpeg/%s/%s", p.Md5, p.JpegURL)

	if p.SampleFileSize == 0 && p.JpegFileSize == 0 {
		p.FileURL, _ = urlEncoded(fmt.Sprintf("Konachan.com - %d %s.jpg", p.Id, p.Tags))
		// can be gif
	} else {
		p.FileURL, _ = urlEncoded(fmt.Sprintf("Konachan.com - %d %s.png", p.Id, p.Tags))
		// can be jpg
	}
	p.FileURL = fmt.Sprintf("https://konachan.com/image/%s/%s", p.Md5, p.FileURL)
}

func urlEncoded(str string) (string, error) {
	u, err := url.Parse(str)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
