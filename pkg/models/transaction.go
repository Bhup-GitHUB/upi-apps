package models

import "time"

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusDebited   Status = "DEBITED"
	StatusSuccess   Status = "SUCCESS"
	StatusFailed    Status = "FAILED"
	StatusRefunded  Status = "REFUNDED"
)

type Transaction struct {
	ID             string
	TraceID        string
	IdempotencyKey string
	SenderVPA      string
	ReceiverVPA    string
	AmountPaise    int64
	Status         Status
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type LedgerEntry struct {
	ID        string
	TxnID     string
	Account   string
	EntryType string
	Amount    int64
	CreatedAt time.Time
}
