package models

import log "github.com/kataras/golog"

type Tag struct {
	ID    int64  `gorm:"column:id" json:"id" form:"id"`
	Name  string `gorm:"column:name" json:"name" form:"name"`
	Count int64  `gorm:"column:count" json:"count" form:"count"`
}

func GetTagCount() map[string]int64 {
	tagCountMap := make(map[string]int64)
	var tags []*Tag
	err := db.Model(&[]Tag{}).Find(&tags).Error
	if err != nil {
		log.Error(err)
		return tagCountMap
	}
	for _, tag := range tags {
		tagCountMap[tag.Name] = tag.Count
	}
	return tagCountMap
}
