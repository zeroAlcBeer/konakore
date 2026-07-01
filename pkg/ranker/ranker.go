package ranker

import "github.com/zeroAlcBeer/konakore/pkg/models"

// Ranker defines the interface for a scoring algorithm.
type Ranker interface {
	// Learn trains the ranker model based on a set of posts (e.g., liked posts).
	Learn(posts []*models.Post)

	// ScoreAll scores a slice of posts.
	ScoreAll(posts []*models.Post)

	// Name returns the name of the ranker.
	Name() string
}
