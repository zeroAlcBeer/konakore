package models

import (
	"math"
	"strings"
)

type TagWeightSystem struct {
	weights      map[string]float64
	cooccurrence map[string]map[string]int
	likedTags    map[string]int
	rarityBonus  map[string]float64
}

func NewTagWeightSystem() *TagWeightSystem {
	return &TagWeightSystem{
		weights:      make(map[string]float64),
		cooccurrence: make(map[string]map[string]int),
		likedTags:    make(map[string]int),
		rarityBonus:  make(map[string]float64),
	}
}

func (tws *TagWeightSystem) Learn(likedPosts []*Post) {
	tws.weights = make(map[string]float64)
	tws.cooccurrence = make(map[string]map[string]int)
	tws.rarityBonus = make(map[string]float64)

	// 统计个人喜好标签频率
	tagFrequency := make(map[string]int)
	totalLikedPosts := len(likedPosts)
	globalTagCount := GetTagCount()

	// 1. 统计标签频率和共现关系
	for _, post := range likedPosts {
		tags := strings.Split(post.Tags, " ")

		for _, tag := range tags {
			tagFrequency[tag]++

			// 初始化共现矩阵
			if _, exists := tws.cooccurrence[tag]; !exists {
				tws.cooccurrence[tag] = make(map[string]int)
			}

			// 更新共现关系
			for _, otherTag := range tags {
				if tag != otherTag {
					tws.cooccurrence[tag][otherTag]++
				}
			}
		}
	}

	// 2. 计算稀有标签加成
	for tag, count := range tagFrequency {
		// 稀有度计算：在全局图片池中出现次数越少，稀有度越高
		globalCount := globalTagCount[tag]

		// 对数衰减的稀有度计算
		rarityScore := math.Log(1.0 / (float64(globalCount) + 1))

		// 个人喜好频率
		personalFrequency := float64(count) / float64(totalLikedPosts)

		// 稀有标签加成：稀有 + 被个人喜欢
		tws.rarityBonus[tag] = rarityScore * personalFrequency * 2

		// 标签权重综合考虑共现和稀有性
		tws.weights[tag] = personalFrequency * (1 + tws.rarityBonus[tag])
	}
}

func (tws *TagWeightSystem) ScorePost(p *Post) {
	tags := strings.Split(p.Tags, " ")
	if len(tags) == 0 {
		return
	}

	var baseScore float64
	var rarityBonusScore float64
	var cooccurrenceScore float64

	// 1. 基础标签匹配分数
	for _, tag := range tags {
		// 直接匹配权重
		if weight, exists := tws.weights[tag]; exists {
			baseScore += weight

			// 稀有标签额外加成
			if bonus, exists := tws.rarityBonus[tag]; exists {
				rarityBonusScore += bonus
			}

			// 共现标签关联分数
			if coTags, exists := tws.cooccurrence[tag]; exists {
				for _, coCount := range coTags {
					cooccurrenceScore += float64(coCount) * 0.1
				}
			}
		}
	}

	// 2. 标签数量衰减因子
	// 让少标签图片有机会获得高分
	tagCountFactor := 1.0 / math.Log1p(float64(len(tags)))

	// 3. 综合评分
	p.MyScore = (baseScore + rarityBonusScore + cooccurrenceScore) * tagCountFactor
	p.Alg = make(map[string]float64)
	p.Alg["baseScore"] = baseScore
	p.Alg["rarityBonusScore"] = rarityBonusScore
	p.Alg["cooccurrenceScore"] = cooccurrenceScore
	p.Alg["tagCountFactor"] = tagCountFactor
	p.WaifuPillow = p.Width > p.Height*2
}
