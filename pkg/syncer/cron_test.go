package syncer

import (
	"testing"

	"github.com/CheerChen/konakore/pkg/models"
)

func TestUpdateTags(t *testing.T) {
	_, err := models.OpenDb("root:please_change@tcp(192.168.0.110:3307)/konakore?charset=utf8mb4&parseTime=True&loc=Local", "")
	if err != nil {
		t.Fatal(err)
	}
	InitDB()

	tws := models.NewTagWeightSystem()

	likedPosts := models.GetLikes()
	tws.Learn(likedPosts)

	// 取最后10个post
	lastPosts := likedPosts[len(likedPosts)-50:]

	for _, p := range lastPosts {
		tws.ScorePost(p)

		// 计算排名
		segmentStart := (p.Id / 300) * 300
		segmentEnd := segmentStart + 300
		segmentPosts := models.GetPostsInRange(segmentStart, segmentEnd)
		for _, sp := range segmentPosts {
			tws.ScorePost(sp)
		}

		rank := 1
		for _, sp := range segmentPosts {
			if sp.MyScore > p.MyScore {
				rank++
			}
		}

		t.Logf("Post ID: %d, Score: %f, Rank: %d in segment [%d - %d]", p.Id, p.MyScore, rank, segmentStart, segmentEnd)
		// b, _ := json.MarshalIndent(p.Alg, "", " ")
		// t.Log(string(b))
	}
}
