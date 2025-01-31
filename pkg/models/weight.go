package models

import (
	"math"
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

	var score float64
	var tagMatches int

	for _, tag := range tags {
		if weight, exists := tws.weights[tag]; exists {
			score += weight
			tagMatches++

			if coTags, exists := tws.cooccurrence[tag]; exists {
				for otherTag, coCount := range coTags {
					if _, liked := tws.likedTags[otherTag]; liked {
						score += float64(coCount) * 0.1 / float64(len(tags))
					}
				}
			}
		}
	}

	matchRatio := float64(tagMatches) / float64(len(tags))
	qualityFactor := 0.3 + 0.7*math.Log1p(matchRatio)

	normalizedScore := math.Log1p(float64(p.Score))

	p.MyScore = score*qualityFactor + normalizedScore*0.1
}
