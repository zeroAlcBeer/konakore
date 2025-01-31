package models

import (
	"math"
	"sort"
	"strings"
)

type TagWeightSystem struct {
	weights      map[string]float64
	cooccurrence map[string]map[string]int
	likedTags    map[string]int
}

func NewTagWeightSystem() *TagWeightSystem {
	return &TagWeightSystem{
		weights:      make(map[string]float64),
		cooccurrence: make(map[string]map[string]int),
		likedTags:    make(map[string]int),
	}
}

func (tws *TagWeightSystem) Learn(likedPosts []*Post) {
	tws.weights = make(map[string]float64)
	tws.cooccurrence = make(map[string]map[string]int)

	for _, post := range likedPosts {
		tags := strings.Split(post.Tags, " ")

		for _, tag := range tags {
			tws.likedTags[tag]++

			if _, exists := tws.cooccurrence[tag]; !exists {
				tws.cooccurrence[tag] = make(map[string]int)
			}

			for _, otherTag := range tags {
				if tag != otherTag {
					tws.cooccurrence[tag][otherTag]++
				}
			}
		}
	}

	totalPosts := float64(len(likedPosts))
	for tag, count := range tws.likedTags {
		frequency := float64(count) / totalPosts

		coWeight := 0.0
		if coTags, exists := tws.cooccurrence[tag]; exists {
			for otherTag, coCount := range coTags {
				if otherCount, ok := tws.likedTags[otherTag]; ok {
					coWeight += float64(coCount) * float64(otherCount)
				}
			}
		}

		tws.weights[tag] = (frequency*0.7 + (coWeight/totalPosts)*0.3) *
			math.Log1p(float64(count))
	}
}

func (tws *TagWeightSystem) ScorePost(p *Post) {
	tags := strings.Split(p.Tags, " ")
	if len(tags) == 0 {
		p.MyScore = 0.0
		return
	}

	var tagScores []float64
	var totalScore float64

	for _, tag := range tags {
		if weight, exists := tws.weights[tag]; exists {
			// weight := tws.weights[tag]

			// TF-IDF计算
			// localFreq := 1.0 / float64(len(tags))
			globalFreq := float64(tws.likedTags[tag]) / float64(len(tws.likedTags))
			tfidf := weight * math.Log(1/globalFreq)

			tagScores = append(tagScores, tfidf)
			totalScore += tfidf
		}
	}
	tagComplexityPenalty := math.Pow(0.8, float64(len(tags)-1))

	averageScore := 0.0
	sort.Sort(sort.Reverse(sort.Float64Slice(tagScores)))

	for i, score := range tagScores {
		weight := math.Pow(0.7, float64(i))
		averageScore += score * weight
	}

	p.MyScore = averageScore*tagComplexityPenalty +
		math.Log1p(float64(p.Score))*0.1

	p.UserScore = float64(p.Score)
	p.WaifuPillow = p.Width > p.Height*2
}
