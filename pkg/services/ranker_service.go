package services

import (
	"sync"

	"github.com/CheerChen/konakore/pkg/models"
	"github.com/CheerChen/konakore/pkg/ranker"
	"github.com/CheerChen/konakore/pkg/ranker/tfidf_hybrid"

	log "github.com/kataras/golog"
)

// RankerService manages the lifecycle of the ranking model.
// It holds the current ranker and provides thread-safe access and retraining.
type RankerService struct {
	mu     sync.RWMutex
	ranker ranker.Ranker
}

// NewRankerService creates and initializes a new RankerService.
func NewRankerService() *RankerService {
	s := &RankerService{}
	s.Retrain()
	return s
}

// GetRanker returns the current, thread-safe ranker instance.
func (s *RankerService) GetRanker() ranker.Ranker {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.ranker
}

// Retrain creates a new ranker, trains it on the latest data,
// and replaces the old one.
func (s *RankerService) Retrain() {
	log.Info("Starting ranker retraining...")
	newRanker := tfidf_hybrid.NewTfidfHybridRanker()
	newRanker.Learn(models.GetLikes())

	s.mu.Lock()
	s.ranker = newRanker
	s.mu.Unlock()
	log.Info("Ranker retraining completed.")
}
