package idempotency

import (
	"fmt"
	"sync"

	"github.com/bhupeshkumar/upi-apps/pkg/models"
)

type Store struct {
	mu    sync.RWMutex
	cache map[string]*models.Transaction
}

func NewStore() *Store {
	return &Store{
		cache: make(map[string]*models.Transaction),
	}
}

func (s *Store) Get(key string) (*models.Transaction, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	txn, ok := s.cache[key]
	if ok {
		fmt.Println("idempotency: cache hit for key", key)
	}
	return txn, ok
}

func (s *Store) Set(key string, txn *models.Transaction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fmt.Println("idempotency: caching result for key", key)
	s.cache[key] = txn
}
