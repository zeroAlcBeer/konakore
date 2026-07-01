package ranker_test

import (
	"sort"
	"testing"

	"github.com/zeroAlcBeer/konakore/pkg/models"
	"github.com/zeroAlcBeer/konakore/pkg/ranker"
	"github.com/zeroAlcBeer/konakore/pkg/ranker/tfidf"
	"github.com/zeroAlcBeer/konakore/pkg/ranker/tfidf_hybrid"

	log "github.com/kataras/golog"
)

const (
	// IMPORTANT: Replace with your actual test database credentials
	TestDbDsn        = "root:please_change@tcp(192.168.0.110:3307)/konakore?charset=utf8mb4&parseTime=True&loc=Local"
	TotalPostsToTest = 1500 // Corresponds to 5 pages of 300 posts each
)

func Test_Ranker_Integration(t *testing.T) {
	// 1. Setup database connection
	_, err := models.OpenDb(TestDbDsn, "dev")
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	log.Info("Successfully connected to test database.")

	// 2. Get all liked posts for training and evaluation
	allLikedPosts := models.GetLikes()
	if len(allLikedPosts) == 0 {
		log.Warn("No liked posts found in the database. The evaluation might not be meaningful.")
	}
	likedIDMap := make(map[int64]struct{}, len(allLikedPosts))
	for _, p := range allLikedPosts {
		likedIDMap[p.Id] = struct{}{}
	}

	// 3. Define the rankers to be tested
	rankers := []ranker.Ranker{
		tfidf.NewTfidfRanker(),
		tfidf_hybrid.NewTfidfHybridRanker(),
	}

	// 4. Fetch the posts to be tested
	var postsToTest []models.Post
	stmt := models.GetPostsStmt("").Order("id desc").Limit(TotalPostsToTest)
	if err := stmt.Find(&postsToTest).Error; err != nil {
		t.Fatalf("Failed to fetch posts for testing: %v", err)
	}
	log.Infof("Fetched %d posts for evaluation.", len(postsToTest))

	// 5. Evaluate each ranker
	for _, r := range rankers {
		t.Run(r.Name(), func(t *testing.T) {
			// Create a slice of pointers for the ranker
			postsToScore := make([]*models.Post, len(postsToTest))
			for i := range postsToTest {
				// Make a copy to prevent rankers from interfering with each other
				postCopy := postsToTest[i]
				postsToScore[i] = &postCopy
			}

			// Train and score all posts at once
			r.Learn(allLikedPosts)
			r.ScoreAll(postsToScore)

			// Sort posts by score in descending order
			sort.Slice(postsToScore, func(i, j int) bool {
				return postsToScore[i].MyScore > postsToScore[j].MyScore
			})

			// Calculate the average rank of liked posts
			var totalRank, likedCount int
			for i, post := range postsToScore {
				if _, isLiked := likedIDMap[post.Id]; isLiked {
					likedCount++
					totalRank += (i + 1) // Rank is 1-based index
				}
			}

			var avgRank float64
			if likedCount > 0 {
				avgRank = float64(totalRank) / float64(likedCount)
			}

			// Log score breakdown for debugging if it's the tfidf_hybrid ranker
			if r.Name() == tfidf_hybrid.RankerName {
				t.Logf("--- Score breakdown for [%s] ---", r.Name())
				likedPostsLogged := 0
				unlikedPostsLogged := 0
				for i, post := range postsToScore {
					_, isLiked := likedIDMap[post.Id]
					if isLiked && likedPostsLogged < 5 {
						t.Logf("  Liked Post ID %d (Rank %d): Final=%.4f | Profile(w*s=%.4f, s=%.4f) | Quality(w*s=%.4f, s=%.4f) | Curation(w*s=%.4f, s=%.4f)",
							post.Id, i+1, post.MyScore,
							post.Alg["profile_score"]*0.8, post.Alg["profile_score"],
							post.Alg["quality_score"]*0.15, post.Alg["quality_score"],
							post.Alg["curation_score"]*0.05, post.Alg["curation_score"])
						likedPostsLogged++
					}
					if !isLiked && unlikedPostsLogged < 5 {
						t.Logf("Unliked Post ID %d (Rank %d): Final=%.4f | Profile(w*s=%.4f, s=%.4f) | Quality(w*s=%.4f, s=%.4f) | Curation(w*s=%.4f, s=%.4f)",
							post.Id, i+1, post.MyScore,
							post.Alg["profile_score"]*0.8, post.Alg["profile_score"],
							post.Alg["quality_score"]*0.15, post.Alg["quality_score"],
							post.Alg["curation_score"]*0.05, post.Alg["curation_score"])
						unlikedPostsLogged++
					}
					if likedPostsLogged >= 5 && unlikedPostsLogged >= 5 {
						break
					}
				}
			}

			// Log results
			t.Logf("--- Results for [%s] ---", r.Name())
			t.Logf("Found %d liked posts out of %d total posts.", likedCount, len(postsToScore))
			t.Logf("Average rank of liked posts: %.2f", avgRank)
			if avgRank == 0 {
				t.Log("Warning: Average rank is 0. This could mean no liked posts were found in the test set.")
			}
		})
	}
}
