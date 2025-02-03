package models

import (
	"math"
	"strings"

	log "github.com/kataras/golog"
)

type TagWeightSystem struct {
	TypeWeights map[int]float64
	tfIdf       map[string]float64
}

func NewTagWeightSystem() *TagWeightSystem {
	return &TagWeightSystem{
		TypeWeights: map[int]float64{
			0: 0.4, // General tags降低
			1: 3.0, // Artist提升
			3: 2.5, // IP提升
			4: 2.0, // Character提升
			5: 0.1, // Meta tags保持低
			6: 2.0, // Brand提升
		},
		tfIdf: make(map[string]float64),
	}
}

func (tws *TagWeightSystem) Learn(likedPosts []*Post) {
	lastId := int64(38 * 10000)
	post := &Post{}
	err := post.Last()
	if err != nil {
		log.Warnf("get lastid err: %s", err)
	} else {
		lastId = post.Id
	}

	log.Infof("post last id: %d", lastId)

	pts := GetLikeTags()
	tf1 := make(map[string]int)
	tf2 := make(map[string]int)

	for _, pt := range pts {
		tags := strings.Split(pt, " ")
		for _, tag := range tags {
			tf1[tag] += 1
			tf2[tag] += len(tags)
		}
	}

	countMap := GetTagCount()
	typeMap := GetTagType()

	for tag := range tf1 {
		if _, ok := countMap[tag]; !ok {
			countMap[tag] = 1
		}
		if _, ok := typeMap[tag]; !ok {
			typeMap[tag] = 0
		}
		// 词频 = tag 在图片出现的次数 / 图片的 tag 总数
		tf := float64(tf1[tag]) / float64(tf2[tag])
		// 逆文档频率 = log( 图片总数 / 包含此 tag 的图片数 + 1）
		idf := math.Log(float64(lastId) / (float64(countMap[tag] + 1)))
		tws.tfIdf[tag] = tf * idf * tws.TypeWeights[typeMap[tag]]
		//idfMap[tag] = idf
	}
	log.Infof("available tags: %d", len(tws.tfIdf))

}

func (tws *TagWeightSystem) ScorePost(p *Post) {
	tags := strings.Split(p.Tags, " ")
	if len(tags) == 0 {
		return
	}

	p.Alg = make(map[string]float64)

	var weight float64
	// Calculate the TF-IDF value for each tag and track the maximum TF-IDF value
	for _, tag := range tags {
		if t, ok := tws.tfIdf[tag]; ok {
			weight += t
			p.Alg[tag] = t
		}
	}
	weight = weight / float64(len(tags))

	p.MyScore = weight

	p.WaifuPillow = p.Width > p.Height*2
}
