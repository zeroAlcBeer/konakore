package tfidf_hybrid

import (
	"math"
	"strings"

	"github.com/zeroAlcBeer/konakore/pkg/models"

	log "github.com/kataras/golog"
)

const RankerName = "tfidf_hybrid"

type Config struct {
	ProfileWeight, QualityWeight, CurationWeight float64 // Final score weights
	CurationMap                                  map[string]float64
	TypeWeights                                  map[int]float64
}

type TfidfHybridRanker struct {
	config Config
	tfIdf  map[string]float64
}

func NewTfidfHybridRanker() *TfidfHybridRanker {
	return &TfidfHybridRanker{
		config: Config{
			ProfileWeight:  0.8,
			QualityWeight:  0.15,
			CurationWeight: 0.05,
			CurationMap: map[string]float64{
				"s": 1.0, // Safe
				"q": 0.9, // Questionable
				"e": 0.7, // Explicit
			},
			TypeWeights: map[int]float64{
				0: 0.4, // General
				1: 3.0, // Artist
				3: 2.5, // Copyright/series
				4: 2.0, // Character
				5: 0.1, // Meta
				6: 2.0, // Brand/studio
			},
		},
		tfIdf: make(map[string]float64),
	}
}

func (r *TfidfHybridRanker) Name() string {
	return RankerName
}

func (r *TfidfHybridRanker) Learn(likedPosts []*models.Post) {
	post := &models.Post{}
	_ = post.Last()
	totalPosts := post.Id

	var likedTags []string
	for _, p := range likedPosts {
		likedTags = append(likedTags, p.Tags)
	}

	tf1 := make(map[string]int)
	tf2 := make(map[string]int)

	for _, pt := range likedTags {
		tags := strings.Split(pt, " ")
		for _, tag := range tags {
			tf1[tag]++
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
		tf := float64(tf1[tag]) / float64(tf2[tag])
		idf := math.Log(float64(totalPosts) / (float64(countMap[tag] + 1)))
		r.tfIdf[tag] = tf * idf * r.config.TypeWeights[typeMap[tag]]
	}
	log.Infof("[%s] learned TF-IDF for %d tags", RankerName, len(r.tfIdf))
}

func (r *TfidfHybridRanker) ScoreAll(posts []*models.Post) {
	if len(posts) == 0 {
		return
	}

	type scores struct {
		profile  float64
		quality  float64
		curation float64
	}

	rawScores := make([]scores, len(posts))
	maxProfile, maxQuality, maxCuration := 0.0, 0.0, 0.0

	// 1. First pass: calculate raw scores and find max values
	for i, p := range posts {
		// Profile Score (TF-IDF)
		var profileScore float64
		tags := strings.Split(p.Tags, " ")
		if p.Alg == nil {
			p.Alg = make(map[string]float64)
		}
		if len(tags) > 0 {
			var weight float64
			for _, tag := range tags {
				if (r.tfIdf[tag]) > 0 {
					weight += r.tfIdf[tag]
					p.Alg[tag] = r.tfIdf[tag]
				}
			}
			profileScore = weight / float64(len(tags))
		}
		rawScores[i].profile = profileScore

		// Quality Score
		qualityScore := math.Log1p(float64(p.Score))
		rawScores[i].quality = qualityScore

		// Curation Score
		curationScore := r.config.CurationMap[p.Rating]
		rawScores[i].curation = curationScore

		if profileScore > maxProfile {
			maxProfile = profileScore
		}
		if qualityScore > maxQuality {
			maxQuality = qualityScore
		}
		if curationScore > maxCuration {
			maxCuration = curationScore
		}
	}

	// 2. Second pass: normalize and calculate final score
	for i, p := range posts {
		normProfile := 0.0
		if maxProfile > 0 {
			normProfile = rawScores[i].profile / maxProfile
		}

		normQuality := 0.0
		if maxQuality > 0 {
			normQuality = rawScores[i].quality / maxQuality
		}

		normCuration := 0.0
		if maxCuration > 0 {
			normCuration = rawScores[i].curation / maxCuration
		}

		p.MyScore = r.config.ProfileWeight*normProfile +
			r.config.QualityWeight*normQuality +
			r.config.CurationWeight*normCuration

		if p.Alg == nil {
			p.Alg = make(map[string]float64)
		}
		p.Alg["profile_score"] = normProfile
		p.Alg["quality_score"] = normQuality
		p.Alg["curation_score"] = normCuration
	}
}
