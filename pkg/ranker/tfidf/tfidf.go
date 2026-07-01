package tfidf

import (
	"math"
	"strings"

	"github.com/zeroAlcBeer/konakore/pkg/models"

	log "github.com/kataras/golog"
)

const RankerName = "tf-idf"

// TfidfRanker uses a TF-IDF algorithm to score posts based on tags.
type TfidfRanker struct {
	TypeWeights map[int]float64
	tfIdf       map[string]float64
}

// NewTfidfRanker creates a new TfidfRanker.
func NewTfidfRanker() *TfidfRanker {
	return &TfidfRanker{
		TypeWeights: map[int]float64{
			0: 0.4, // General tags
			1: 3.0, // Artist
			3: 2.5, // Copyright/series
			4: 2.0, // Character
			5: 0.1, // Meta tags
			6: 2.0, // Brand/studio
		},
		tfIdf: make(map[string]float64),
	}
}

// Name returns the name of the ranker.
func (r *TfidfRanker) Name() string {
	return RankerName
}

// Learn trains the TF-IDF model on a set of liked posts.
func (r *TfidfRanker) Learn(likedPosts []*models.Post) {
	lastId := int64(40 * 10000)
	post := &models.Post{}
	err := post.Last()
	if err != nil {
		log.Warnf("get last id err: %s", err)
	} else {
		lastId = post.Id
	}
	log.Infof("post last id: %d", lastId)

	var likedTags []string
	for _, p := range likedPosts {
		likedTags = append(likedTags, p.Tags)
	}

	tf1 := make(map[string]int)
	tf2 := make(map[string]int)

	for _, pt := range likedTags {
		tags := strings.Split(pt, " ")
		for _, tag := range tags {
			tf1[tag] += 1
			tf2[tag] += len(tags)
		}
	}

	countMap := models.GetTagCount()
	typeMap := models.GetTagType()

	for tag := range tf1 {
		if _, ok := countMap[tag]; !ok {
			countMap[tag] = 1
		}
		if _, ok := typeMap[tag]; !ok {
			typeMap[tag] = 0
		}
		// Term Frequency = (Number of times tag appears in liked posts) / (Total number of tags in liked posts)
		tf := float64(tf1[tag]) / float64(tf2[tag])
		// Inverse Document Frequency = log(Total number of posts / Number of posts with this tag + 1)
		idf := math.Log(float64(lastId) / (float64(countMap[tag] + 1)))
		r.tfIdf[tag] = tf * idf * r.TypeWeights[typeMap[tag]]
	}
	log.Infof("[%s] learned %d tags", RankerName, len(r.tfIdf))
}

// ScoreAll scores a slice of posts.
func (r *TfidfRanker) ScoreAll(posts []*models.Post) {
	for _, p := range posts {
		r.scorePost(p)
	}
}

// scorePost scores a single post based on the learned TF-IDF weights.
func (r *TfidfRanker) scorePost(p *models.Post) {
	tags := strings.Split(p.Tags, " ")
	if len(tags) == 0 {
		return
	}

	if p.Alg == nil {
		p.Alg = make(map[string]float64)
	}

	var weight float64
	for _, tag := range tags {
		if t, ok := r.tfIdf[tag]; ok {
			weight += t
			p.Alg[tag] = t
		}
	}

	if len(tags) > 0 {
		weight = weight / float64(len(tags))
	}

	p.MyScore = weight
	p.WaifuPillow = p.Width > p.Height*2
}
