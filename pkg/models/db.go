package models

import (
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	myclient "konakore/pkg/client"
	"log"
	"os"
)

var (
	db     *gorm.DB
	client myclient.Client
)

func OpenDb() {
	var err error
	dsn := os.Getenv("dsn")
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}
	debug := os.Getenv("debug")
	if debug != "" {
		db = db.Debug()
	}
	//err = db.AutoMigrate(&Post{}, &Like{}, &Tag{})
	//if err != nil {
	//	log.Fatal(err)
	//}
}
