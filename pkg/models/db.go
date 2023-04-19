package models

import (
	"strings"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var db *gorm.DB

func OpenDb(dsn, env string) (*gorm.DB, error) {
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		// NamingStrategy: schema.NamingStrategy{
		// 	SingularTable: true,
		// },
	})
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(env) == "dev" {
		db = db.Debug()
	}
	err = db.AutoMigrate(&Post{}, &Like{}, Tag{})
	if err != nil {
		return nil, err
	}
	return db, err
}

func GetDb() *gorm.DB {
	return db
}
