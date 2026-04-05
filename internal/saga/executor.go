package saga

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/bhupeshkumar/upi-apps/internal/bank"
	"github.com/bhupeshkumar/upi-apps/internal/ledger"
	"github.com/bhupeshkumar/upi-apps/internal/queue"
	"github.com/bhupeshkumar/upi-apps/pkg/models"
)

type Executor struct {
	ledger *ledger.Store
	broker *queue.Broker
}

func NewExecutor(l *ledger.Store, b *queue.Broker) *Executor {
	return &Executor{ledger: l, broker: b}
}

func (e *Executor) Run(txn *models.Transaction, senderBank, receiverBank bank.Adapter) {
	fmt.Println("saga: starting txn", txn.ID, "trace", txn.TraceID)

	ctx := context.Background()

	e.broker.Publish(queue.Event{Type: queue.EventDebitRequested, TxnID: txn.ID, TraceID: txn.TraceID, Payload: txn})

	err := e.retryWithBackoff(3, func() error {
		return senderBank.Debit(ctx, txn.ID, txn.SenderVPA, txn.AmountPaise)
	})

	if err != nil {
		fmt.Println("saga: debit failed after all retries txn", txn.ID)
		txn.Status = models.StatusFailed
		e.ledger.SaveTxn(txn)
		e.broker.Publish(queue.Event{Type: queue.EventFailed, TxnID: txn.ID, TraceID: txn.TraceID, Payload: txn})
		return
	}

	e.ledger.RecordDebit(txn.ID, txn.SenderVPA, txn.AmountPaise)
	txn.Status = models.StatusDebited
	e.ledger.SaveTxn(txn)
	e.broker.Publish(queue.Event{Type: queue.EventDebitDone, TxnID: txn.ID, TraceID: txn.TraceID, Payload: txn})

	e.broker.Publish(queue.Event{Type: queue.EventCreditRequested, TxnID: txn.ID, TraceID: txn.TraceID, Payload: txn})

	err = e.retryWithBackoff(3, func() error {
		return receiverBank.Credit(ctx, txn.ID, txn.ReceiverVPA, txn.AmountPaise)
	})

	if err != nil {
		fmt.Println("saga: credit failed after all retries, initiating refund txn", txn.ID)
		e.broker.Publish(queue.Event{Type: queue.EventRefundRequested, TxnID: txn.ID, TraceID: txn.TraceID, Payload: txn})

		_ = senderBank.Refund(ctx, txn.ID, txn.SenderVPA, txn.AmountPaise)
		e.ledger.RecordRefund(txn.ID, txn.SenderVPA, txn.AmountPaise)

		txn.Status = models.StatusRefunded
		e.ledger.SaveTxn(txn)
		e.broker.Publish(queue.Event{Type: queue.EventFailed, TxnID: txn.ID, TraceID: txn.TraceID, Payload: txn})
		return
	}

	e.ledger.RecordCredit(txn.ID, txn.ReceiverVPA, txn.AmountPaise)
	txn.Status = models.StatusSuccess
	e.ledger.SaveTxn(txn)
	e.broker.Publish(queue.Event{Type: queue.EventCompleted, TxnID: txn.ID, TraceID: txn.TraceID, Payload: txn})
	fmt.Println("saga: completed txn", txn.ID)
}

func (e *Executor) retryWithBackoff(attempts int, fn func() error) error {
	var err error
	for i := 0; i < attempts; i++ {
		err = fn()
		if err == nil {
			return nil
		}
		fmt.Println("saga: retry attempt", i+1, "of", attempts, "err:", err.Error())
		time.Sleep(time.Duration(i+1) * 100 * time.Millisecond)
	}
	return errors.New("all retries exhausted: " + err.Error())
}
