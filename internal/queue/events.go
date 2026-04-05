package queue

import "github.com/bhupeshkumar/upi-apps/pkg/models"

type EventType string

const (
	EventDebitRequested  EventType = "DEBIT_REQUESTED"
	EventDebitDone       EventType = "DEBIT_DONE"
	EventCreditRequested EventType = "CREDIT_REQUESTED"
	EventCreditDone      EventType = "CREDIT_DONE"
	EventRefundRequested EventType = "REFUND_REQUESTED"
	EventCompleted       EventType = "COMPLETED"
	EventFailed          EventType = "FAILED"
)

type Event struct {
	Type    EventType
	TraceID string
	TxnID   string
	Payload *models.Transaction
}
