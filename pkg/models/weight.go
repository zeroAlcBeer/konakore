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

func (tws *TagWeightSystem) ScorePostV2(p *Post) {
	tags := strings.Split(p.Tags, " ")
	if len(tags) == 0 {
		return
	}

	// 新参数定义
	const (
		coreTagThreshold   = 0.3  // 核心标签权重阈值
		decayFactor        = 0.85 // 协同衰减系数
		maxEffectiveTags   = 8    // 有效标签上限
		baselineTagPenalty = 0.92 // 基础惩罚系数
	)

	// 阶段1：计算标签协同效应
	processedTags := make(map[string]struct{})
	uniqueTagScores := make([]float64, 0, len(tags))
	tagSynergy := 1.0 // 协同效应乘数

	for _, tag := range tags {
		if _, exists := processedTags[tag]; exists {
			continue
		}
		processedTags[tag] = struct{}{}

		if weight, exists := tws.weights[tag]; exists {
			// 计算标签协同衰减
			if coTags, ok := tws.cooccurrence[tag]; ok {
				synergyCount := 0
				for relatedTag := range coTags {
					if _, exists := processedTags[relatedTag]; exists {
						synergyCount++
					}
				}
				// 每有一个协同标签，衰减系数生效
				currentDecay := math.Pow(decayFactor, float64(synergyCount))
				tagSynergy *= currentDecay
			}

			// 增强核心标签权重
			adjustedWeight := weight
			if weight > coreTagThreshold {
				adjustedWeight += 0.5 * (weight - coreTagThreshold)
			}

			globalFreq := math.Max(float64(tws.likedTags[tag])/float64(len(tws.likedTags)), 1e-6)
			tfidf := adjustedWeight * math.Log(1/globalFreq)

			if tfidf > 0.1 {
				uniqueTagScores = append(uniqueTagScores, tfidf)
			}
		}
	}

	// 阶段2：动态标签数量处理
	sort.Sort(sort.Reverse(sort.Float64Slice(uniqueTagScores)))

	// 计算有效标签数量（考虑协同效应）
	effectiveTags := math.Min(float64(len(uniqueTagScores)), maxEffectiveTags)
	tagCountFactor := baselineTagPenalty * math.Pow(0.96, effectiveTags)

	// 阶段3：动态衰减加权
	sumWeights := 0.0
	weightedSum := 0.0
	remainingImpact := 1.0 // 剩余影响力

	for i, score := range uniqueTagScores {
		// 动态衰减系数：前25%标签保持全权重，之后指数衰减
		decay := 1.0
		if float64(i) > 0.25*effectiveTags {
			decay = math.Pow(0.8, float64(i)-0.25*effectiveTags)
		}

		currentWeight := remainingImpact * decay
		weightedSum += score * currentWeight
		sumWeights += currentWeight
		remainingImpact *= 0.92 // 每个标签减少8%剩余影响力
	}

	// 阶段4：最终得分合成
	normalizedScore := weightedSum / (sumWeights + 1e-6)
	rarityScore := math.Sqrt(sumWeights) * 0.7 // 使用平方根压缩稀有性分数

	p.MyScore = (normalizedScore * tagCountFactor * tagSynergy) +
		rarityScore +
		math.Log1p(float64(p.Score))*0.1

	// 调试信息
	p.Alg = make(map[string]float64)
	p.Alg["normalizedScore"] = normalizedScore
	p.Alg["rarityScore"] = rarityScore
	p.Alg["tagSynergy"] = tagSynergy
	p.Alg["tagCountFactor"] = tagCountFactor

	p.WaifuPillow = p.Width > p.Height*2
}
