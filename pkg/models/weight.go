package models

import (
	"math"
	"strings"
)

// TagInfo stores information about a tag
type TagInfo struct {
	Weight float64
	Type   int
	Count  int64
}

type TagWeightSystem struct {
	// Store tag weights and their information
	TagWeights map[string]*TagInfo
	// Store tag type weights to give different importance to different tag types
	TypeWeights map[int]float64
}

func NewTagWeightSystem() *TagWeightSystem {
	return &TagWeightSystem{
		TagWeights: make(map[string]*TagInfo),
		TypeWeights: map[int]float64{
			0: 0.8, // General tags
			1: 1.5, // Artist tags
			3: 1.3, // IP tags
			4: 1.1, // Character tags
			5: 0.8, // logo, watermark
			6: 1.2, // brand
		},
	}
}

// Learn analyzes liked posts to build tag weights
func (tws *TagWeightSystem) Learn(likedPosts []*Post) {
	// Get global tag counts for normalization
	globalTagCount := GetTagCount()
	totalLikedPosts := float64(len(likedPosts))

	// First pass: count tag occurrences in liked posts
	likedTagCount := make(map[string]int64)
	for _, post := range likedPosts {
		tags := strings.Split(post.Tags, " ")
		for _, tag := range tags {
			likedTagCount[tag]++
		}
	}

	// Calculate weights for each tag
	for tag, likedCount := range likedTagCount {
		var globalCount int64
		item, ok := globalTagCount[tag]
		if ok {
			globalCount = item.Count
		} else {
			continue
		}

		// Calculate tag weight using TF-IDF inspired formula
		frequency := float64(likedCount) / totalLikedPosts
		inverse := math.Log(float64(globalCount) / float64(likedCount))
		weight := frequency * inverse

		// Store tag information
		tws.TagWeights[tag] = &TagInfo{
			Weight: weight,
			Count:  int64(likedCount),
			Type:   item.Type,
		}
	}
}

// ScorePost calculates a score for a new post based on learned weights
func (tws *TagWeightSystem) ScorePost(p *Post) {
	tags := strings.Split(p.Tags, " ")
	if len(tags) == 0 {
		return
	}

	p.Alg = make(map[string]float64)
	totalScore := 0.0
	totalWeight := 0.0

	// Calculate weighted score for each tag
	for _, tag := range tags {
		if tagInfo, exists := tws.TagWeights[tag]; exists {
			// Apply type weight if available
			typeWeight := tws.TypeWeights[tagInfo.Type]

			// Calculate tag contribution
			tagScore := tagInfo.Weight * typeWeight
			totalScore += tagScore
			totalWeight += typeWeight

			// Store important factors in Alg map
			p.Alg[tag] = tagScore
		}
	}

	// Normalize score by total weight to prevent bias from too many tags
	if totalWeight > 0 {
		p.MyScore = totalScore / math.Log(totalWeight+1)
	}

	// Additional features
	p.WaifuPillow = p.Width > p.Height*2
}

func (tws *TagWeightSystem) ScorePostV2(p *Post) {
	tags := strings.Split(p.Tags, " ")
	if len(tags) == 0 {
		return
	}

	p.Alg = make(map[string]float64)
	typeScores := make(map[int]float64)
	typeCounts := make(map[int]int)

	// 对每种类型分别追踪最高分
	typeMaxScores := make(map[int]float64)

	// 首先计算每个标签的分数
	for _, tag := range tags {
		if tagInfo, exists := tws.TagWeights[tag]; exists {
			tagType := tagInfo.Type
			typeWeight := tws.TypeWeights[tagType]

			tagScore := tagInfo.Weight * typeWeight

			// 记录每种类型的最高分
			if tagScore > typeMaxScores[tagType] {
				typeMaxScores[tagType] = tagScore
			}

			// 仍然累计总分（用于Alg map）
			typeScores[tagType] += tagScore
			typeCounts[tagType]++

			p.Alg[tag] = tagScore
		}
	}

	// 计算最终分数
	totalScore := 0.0
	for tagType, maxScore := range typeMaxScores {
		// 使用最高分作为基础
		baseScore := maxScore

		// 如果有多个标签，额外的标签只贡献较小的增益
		if typeCounts[tagType] > 1 {
			additionalScore := (typeScores[tagType] - maxScore) /
				math.Log(float64(typeCounts[tagType])+1)
			baseScore += additionalScore * 0.3 // 额外标签的影响度降低
		}

		totalScore += baseScore
	}

	// 最终归一化
	numTypes := float64(len(typeMaxScores))
	if numTypes > 0 {
		p.MyScore = totalScore / math.Log(numTypes+1)
	}

	p.WaifuPillow = p.Width > p.Height*2
}

// Helper function to normalize scores to 0-1 range
func (tws *TagWeightSystem) NormalizeScores(posts []*Post) {
	var maxScore, minScore float64
	maxScore = math.Inf(-1)
	minScore = math.Inf(1)

	// Find max and min scores
	for _, p := range posts {
		if p.MyScore > maxScore {
			maxScore = p.MyScore
		}
		if p.MyScore < minScore {
			minScore = p.MyScore
		}
	}

	// Normalize scores
	scoreRange := maxScore - minScore
	if scoreRange > 0 {
		for _, p := range posts {
			p.MyScore = (p.MyScore - minScore) / scoreRange
		}
	}
}
