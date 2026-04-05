package ledger

import (
	"fmt"
	"sync"
	"time"

	"github.com/bhupeshkumar/upi-apps/pkg/models"
)

type Store struct {
	mu      sync.RWMutex
	entries []models.LedgerEntry
	txns    map[string]*models.Transaction
}

func NewStore() *Store {
	return &Store{
		entries: make([]models.LedgerEntry, 0),
		txns:    make(map[string]*models.Transaction),
	}
}

func (s *Store) RecordDebit(txnID, account string, amount int64) {
	fmt.Println("ledger: DEBIT", account, amount, "txn", txnID)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, models.LedgerEntry{
		ID:        txnID + "_debit",
		TxnID:     txnID,
		Account:   account,
		EntryType: "DEBIT",
		Amount:    amount,
		CreatedAt: time.Now(),
	})
}

func (s *Store) RecordCredit(txnID, account string, amount int64) {
	fmt.Println("ledger: CREDIT", account, amount, "txn", txnID)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, models.LedgerEntry{
		ID:        txnID + "_credit",
		TxnID:     txnID,
		Account:   account,
		EntryType: "CREDIT",
		Amount:    amount,
		CreatedAt: time.Now(),
	})
}

func (s *Store) RecordRefund(txnID, account string, amount int64) {
	fmt.Println("ledger: REFUND", account, amount, "txn", txnID)
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries = append(s.entries, models.LedgerEntry{
		ID:        txnID + "_refund",
		TxnID:     txnID,
		Account:   account,
		EntryType: "REFUND",
		Amount:    amount,
		CreatedAt: time.Now(),
	})
}

func (s *Store) SaveTxn(txn *models.Transaction) {
	s.mu.Lock()
	defer s.mu.Unlock()
	txn.UpdatedAt = time.Now()
	s.txns[txn.ID] = txn
	fmt.Println("ledger: txn saved", txn.ID, "status", string(txn.Status))
}

func (s *Store) GetTxn(id string) (*models.Transaction, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	txn, ok := s.txns[id]
	return txn, ok
}

func (s *Store) GetEntries(txnID string) []models.LedgerEntry {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]models.LedgerEntry, 0)
	for _, e := range s.entries {
		if e.TxnID == txnID {
			result = append(result, e)
		}
	}
	return result
}
