package models

import (
	log "github.com/kataras/golog"
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

func LikeAll(in []int64) (err error) {
	var likes []*Like
	for _, id := range in {
		likes = append(likes, &Like{Id: id})
	}

	return db.Clauses(clause.OnConflict{DoNothing: true}).Create(&likes).Error
}

func Likes() []int64 {
	var likes []int64
	err := db.Model(&Like{}).Pluck("id", &likes).Error
	if err != nil {
		log.Errorf("get favorites err, %s", err.Error())
	}

	return likes
}

func GetLikesStmt(query string) *gorm.DB {
	stmt := db.Model(&[]Post{}).Where("id in (?)", Likes())
	if query != "" {
		stmt = stmt.Where("MATCH (`tags`) AGAINST (?)", query)
	}
	return stmt
}

func GetLikeTags() []string {
	var pts []string
	_ = db.Model(&[]Post{}).Where("id in (?)", Likes()).Pluck("tags", &pts).Error
	return pts
}

func IsLike(p *Post, likes []int64) {
	for _, id := range likes {
		if p.Id == id {
			p.IsLike = true
			break
		}
	}
}
