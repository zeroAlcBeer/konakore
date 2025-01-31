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

	tagScores := make([]float64, 0, len(tags))
	totalScore := 0.0
	validTags := 0

	// 计算基础分数时增加最低频率限制
	for _, tag := range tags {
		if weight, exists := tws.weights[tag]; exists {
			globalFreq := float64(tws.likedTags[tag])/float64(len(tws.likedTags)) + 1e-6 // 防止除零
			tfidf := weight * math.Log(1/globalFreq)

			// 过滤掉过低权重的标签
			if tfidf > 0.1 {
				tagScores = append(tagScores, tfidf)
				totalScore += tfidf
				validTags++
			}
		}
	}

	// 根据分数排序（降序）
	sort.Sort(sort.Reverse(sort.Float64Slice(tagScores)))

	// 计算归一化加权平均
	sumWeight := 0.0
	weightedSum := 0.0
	for i, score := range tagScores {
		weight := math.Pow(0.7, float64(i))
		sumWeight += weight
		weightedSum += score * weight
	}
	averageScore := weightedSum / (sumWeight + 1e-6)

	// 标签数量惩罚（对数平衡）
	tagCount := float64(validTags)
	countPenalty := 1.0 / math.Log1p(math.E-1+tagCount*0.5) // 柔性惩罚

	// 最终得分组合
	mainScore := averageScore * countPenalty
	verificationScore := math.Log1p(totalScore) * 0.2 // 使用totalScore作为验证项

	p.MyScore = mainScore +
		verificationScore +
		math.Log1p(float64(p.Score))*0.1

	// 保留原始用户分数
	p.UserScore = float64(p.Score)
	p.WaifuPillow = p.Width > p.Height*2
}
