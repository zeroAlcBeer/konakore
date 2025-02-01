package models

import (
	"math"
	"sort"
	"strings"
)

type TagWeightSystem struct {
	tagWeights          map[string]float64
	likedTagOccurrences map[string]int
	globalTagCount      map[string]int64
	totalLikedPosts     int
	totalPosts          int64
}

type weightedTag struct {
	tag         string
	weight      float64
	rarityScore float64 // Add rarityScore for sorting
}

func NewTagWeightSystem() *TagWeightSystem {
	return &TagWeightSystem{
		tagWeights:          make(map[string]float64),
		likedTagOccurrences: make(map[string]int),
		globalTagCount:      make(map[string]int64),
		totalLikedPosts:     0,
		totalPosts:          0,
	}
}

func (tws *TagWeightSystem) Learn(likedPosts []*Post) {
	tws.tagWeights = make(map[string]float64)
	tws.likedTagOccurrences = make(map[string]int)
	tws.totalLikedPosts = len(likedPosts)

	tws.globalTagCount = GetTagCount()

	for _, count := range tws.globalTagCount {
		if count > tws.totalPosts {
			tws.totalPosts = count
		}
	}

	for _, post := range likedPosts {
		tags := strings.Split(post.Tags, " ")
		for _, tag := range tags {
			if tag == "" {
				continue
			}
			tws.likedTagOccurrences[tag]++
		}
	}

	for tag, likedCount := range tws.likedTagOccurrences {
		tf := float64(likedCount) / float64(tws.totalLikedPosts)

		globalCount := float64(tws.globalTagCount[tag])
		if globalCount == 0 {
			continue
		}
		idf := math.Log(float64(tws.totalPosts) / globalCount)

		weight := tf * idf
		weight = math.Log1p(weight)

		tws.tagWeights[tag] = weight
	}
}

func (tws *TagWeightSystem) calculateRarityScore(tag string) float64 {
	globalCount := float64(tws.globalTagCount[tag])
	if globalCount == 0 {
		return 0
	}
	// Calculate rarity score as inverse of frequency
	return math.Log(float64(tws.totalPosts) / globalCount)
}

func (tws *TagWeightSystem) ScorePost(p *Post) {
	tags := strings.Split(p.Tags, " ")
	if len(tags) == 0 {
		return
	}

	p.Alg = make(map[string]float64)
	var weightedTags []weightedTag

	// Calculate weight and rarity for each tag
	for _, tag := range tags {
		if tag == "" {
			continue
		}

		if weight, exists := tws.tagWeights[tag]; exists {
			rarityScore := tws.calculateRarityScore(tag)
			weightedTags = append(weightedTags, weightedTag{
				tag:         tag,
				weight:      weight,
				rarityScore: rarityScore,
			})
		}
	}

	// Sort tags primarily by rarity score, then by weight if rarity scores are equal
	sort.Slice(weightedTags, func(i, j int) bool {
		if weightedTags[i].rarityScore != weightedTags[j].rarityScore {
			return weightedTags[i].rarityScore > weightedTags[j].rarityScore
		}
		return weightedTags[i].weight > weightedTags[j].weight
	})

	// Take top 10 tags
	maxTags := 10
	if len(weightedTags) < maxTags {
		maxTags = len(weightedTags)
	}

	// Calculate final score using selected tags
	totalScore := 0.0
	for i := 0; i < maxTags; i++ {
		tag := weightedTags[i].tag
		// Combine weight with rarity for final tag contribution
		// adjustedWeight := weightedTags[i].weight * weightedTags[i].rarityScore
		p.Alg[tag] = weightedTags[i].rarityScore
		totalScore += weightedTags[i].rarityScore
	}

	p.MyScore = totalScore
	p.WaifuPillow = p.Width > p.Height*2
}

func (tws *TagWeightSystem) GetTagWeight(tag string) float64 {
	return tws.tagWeights[tag]
}
