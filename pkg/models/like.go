package models

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Like struct {
	Id int64 `gorm:"column:id" json:"id" form:"id"`
}

func (p *Post) Unlike(id int64) (err error) {
	return db.Delete(&Like{Id: id}).Error
}

func (p *Post) Like(id int64) (err error) {
	return db.Create(&Like{Id: id}).Error
}

func (p *Post) Done(id int64) (err error) {
	return db.Save(&Like{Id: id}).Error
}

func LikeAll(in []int64) (err error) {
	var likes []*Like
	for _, id := range in {
		likes = append(likes, &Like{Id: id})
	}

	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&likes).Error
}

func GetLikesStmt(query string) *gorm.DB {
	stmt := db.Model(&[]Post{}).Preload("Likes").Where("id in (select id from likes)")
	if query != "" {
		stmt = stmt.Where("tags LIKE ?", "%"+query+"%")
	}
	return stmt
}

func GetLikes() []*Post {
	var pts []*Post
	_ = GetLikesStmt("").
		Find(&pts).Error
	return pts
}

func GetLikeTags() []string {
	var pts []string
	_ = GetLikesStmt("").
		Pluck("tags", &pts).Error
	return pts
}
