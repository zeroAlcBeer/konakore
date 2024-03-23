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

	_, err := c.AddFunc("0 0 * * *", UpdateTags)
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

var currentParam int = 1

func NewestPosts() {
	updatePosts(currentParam)

	// 更新 currentParam 的值，如果是，则重置为1，否则加1
	if currentParam == 12 {
		currentParam = 1
	} else {
		currentParam++
	}
}

func OldestPosts() {
	p := currentPage()
	if p > 2000 {
		return
	}
	updatePosts(p)
}

func ForceUpdatePosts(p int) {
	if p > 2000 {
		return
	}
	updatePosts(p)
}

func UpdateTags() {
	klog.Infof("update tags...")
	tags, err := getTags(10000)
	if err != nil {
		klog.Errorf("get tags err: %s", err)
		return
	}

	err = db.Clauses(clause.OnConflict{
		UpdateAll: true,
	}).Create(tags).Error
	if err != nil {
		klog.Warningf("update tags err: %s", err)
	}
}
