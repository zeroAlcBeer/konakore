package models

import log "github.com/kataras/golog"

type Tag struct {
	ID    int64  `gorm:"column:id" json:"id" form:"id"`
	Name  string `gorm:"column:name" json:"name" form:"name"`
	Count int64  `gorm:"column:count" json:"count" form:"count"`
	Type  int    `gorm:"column:type" json:"type" form:"type"`
}

func GetTagCount() map[string]*Tag {
	itemMap := make(map[string]*Tag)
	var tags []*Tag
	err := db.Model(&[]Tag{}).Find(&tags).Error
	if err != nil {
		log.Error(err)
		return itemMap
	}
	for _, tag := range tags {
		itemMap[tag.Name] = tag
	}
	return itemMap
}
