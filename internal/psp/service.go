package psp

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/bhupeshkumar/upi-apps/internal/idempotency"
	"github.com/bhupeshkumar/upi-apps/internal/ledger"
	"github.com/bhupeshkumar/upi-apps/internal/npci"
	"github.com/bhupeshkumar/upi-apps/internal/queue"
	"github.com/bhupeshkumar/upi-apps/internal/saga"
	"github.com/bhupeshkumar/upi-apps/pkg/models"
)

const fraudAmountThreshold int64 = 100000

type PayRequest struct {
	SenderVPA      string
	ReceiverVPA    string
	AmountPaise    int64
	IdempotencyKey string
	TraceID        string
}

type Service struct {
	idempotency *idempotency.Store
	npciSwitch  *npci.Switch
	saga        *saga.Executor
	ledger      *ledger.Store
	broker      *queue.Broker
	rateLimiter *RateLimiter
}

func NewService(
	idem *idempotency.Store,
	sw *npci.Switch,
	ex *saga.Executor,
	l *ledger.Store,
	b *queue.Broker,
	rl *RateLimiter,
) *Service {
	return &Service{
		idempotency: idem,
		npciSwitch:  sw,
		saga:        ex,
		ledger:      l,
		broker:      b,
		rateLimiter: rl,
	}
}

func (s *Service) InitiatePayment(req PayRequest) (*models.Transaction, error) {
	fmt.Println("psp: payment request sender", req.SenderVPA, "receiver", req.ReceiverVPA, "amount", req.AmountPaise, "trace", req.TraceID)

	if err := validateVPA(req.SenderVPA); err != nil {
		fmt.Println("psp: invalid sender VPA", req.SenderVPA)
		return nil, err
	}
	if err := validateVPA(req.ReceiverVPA); err != nil {
		fmt.Println("psp: invalid receiver VPA", req.ReceiverVPA)
		return nil, err
	}
	if req.AmountPaise <= 0 {
		fmt.Println("psp: invalid amount", req.AmountPaise)
		return nil, errors.New("amount must be greater than zero")
	}

	if err := s.rateLimiter.Allow(req.SenderVPA); err != nil {
		return nil, err
	}

	if req.AmountPaise > fraudAmountThreshold {
		fmt.Println("psp: fraud check flagged amount", req.AmountPaise, "from", req.SenderVPA)
		return nil, errors.New("transaction flagged by fraud check")
	}

	if existing, ok := s.idempotency.Get(req.IdempotencyKey); ok {
		fmt.Println("psp: duplicate request, returning cached txn", existing.ID)
		return existing, nil
	}

	txn := &models.Transaction{
		ID:             newID("txn"),
		TraceID:        req.TraceID,
		IdempotencyKey: req.IdempotencyKey,
		SenderVPA:      req.SenderVPA,
		ReceiverVPA:    req.ReceiverVPA,
		AmountPaise:    req.AmountPaise,
		Status:         models.StatusPending,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	s.ledger.SaveTxn(txn)
	s.idempotency.Set(req.IdempotencyKey, txn)

	fmt.Println("psp: txn created", txn.ID, "forwarding to NPCI")

	go s.process(txn)

	return txn, nil
}

func (s *Service) process(txn *models.Transaction) {
	senderBank, err := s.npciSwitch.ResolveVPA(txn.SenderVPA)
	if err != nil {
		fmt.Println("psp: could not resolve sender VPA", txn.SenderVPA, err.Error())
		txn.Status = models.StatusFailed
		s.ledger.SaveTxn(txn)
		return
	}

	receiverBank, err := s.npciSwitch.ResolveVPA(txn.ReceiverVPA)
	if err != nil {
		fmt.Println("psp: could not resolve receiver VPA", txn.ReceiverVPA, err.Error())
		txn.Status = models.StatusFailed
		s.ledger.SaveTxn(txn)
		return
	}

	s.saga.Run(txn, senderBank, receiverBank)
}

func validateVPA(vpa string) error {
	if !strings.Contains(vpa, "@") {
		return errors.New("invalid VPA: " + vpa)
	}
	parts := strings.Split(vpa, "@")
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return errors.New("invalid VPA format: " + vpa)
	}
	return nil
}

func newID(prefix string) string {
	return prefix + "_" + fmt.Sprintf("%d", time.Now().UnixNano())
}
