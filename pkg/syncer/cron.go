package syncer

import (
	"log"

	"github.com/robfig/cron/v3"
	"gorm.io/gorm/clause"
	"k8s.io/klog"
)

func AddCron(spec string) {
	c := cron.New()
	c.Start()

	_, err := c.AddFunc(spec, UpdateTags)
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.AddFunc(spec, NewestPosts)
	if err != nil {
		log.Fatal(err)
	}

	_, err = c.AddFunc(spec, OldestPosts)
	if err != nil {
		log.Fatal(err)
	}
}

func NewestPosts() {
	updatePosts(1)
}

func OldestPosts() {
	p := currentPage()
	if p > 2000 {
		return
	}
	updatePosts(p)
}

func UpdateTags() {
	klog.Infof("update tags...")
	tags, err := getTags(20000)
	if err != nil {
		klog.Errorf("get tags err: %s", err)
		return
	}

	err = db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(tags[0:10000]).Error
	if err != nil {
		klog.Warningf("update tags err: %s", err)
	}

	err = db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(tags[10000:20000]).Error
	if err != nil {
		klog.Warningf("update tags err: %s", err)
	}
}
