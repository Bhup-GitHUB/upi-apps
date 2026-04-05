package gateway

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/bhupeshkumar/upi-apps/internal/ledger"
	"github.com/bhupeshkumar/upi-apps/internal/psp"
	"github.com/bhupeshkumar/upi-apps/pkg/models"
)

type Handler struct {
	pspService *psp.Service
	ledger     *ledger.Store
}

func NewHandler(p *psp.Service, l *ledger.Store) *Handler {
	return &Handler{pspService: p, ledger: l}
}

func (h *Handler) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/pay", h.handlePay)
	mux.HandleFunc("/status/", h.handleStatus)
	mux.HandleFunc("/health", h.handleHealth)
}

func (h *Handler) handlePay(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	var body struct {
		SenderVPA   string `json:"sender_vpa"`
		ReceiverVPA string `json:"receiver_vpa"`
		AmountPaise int64  `json:"amount_paise"`
	}

	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		fmt.Println("gateway: bad request body", err.Error())
		writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	idempotencyKey := r.Header.Get("Idempotency-Key")
	if idempotencyKey == "" {
		writeError(w, http.StatusBadRequest, "Idempotency-Key header is required")
		return
	}

	traceID := TraceIDFromContext(r.Context())

	txn, err := h.pspService.InitiatePayment(psp.PayRequest{
		SenderVPA:      body.SenderVPA,
		ReceiverVPA:    body.ReceiverVPA,
		AmountPaise:    body.AmountPaise,
		IdempotencyKey: idempotencyKey,
		TraceID:        traceID,
	})

	if err != nil {
		fmt.Println("gateway: payment initiation failed", err.Error(), "trace", traceID)
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}

	fmt.Println("gateway: payment accepted txn", txn.ID, "trace", traceID)

	writeJSON(w, http.StatusAccepted, map[string]interface{}{
		"txn_id":   txn.ID,
		"trace_id": traceID,
		"status":   txn.Status,
		"message":  "payment initiated",
	})
}

func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}

	txnID := strings.TrimPrefix(r.URL.Path, "/status/")
	if txnID == "" {
		writeError(w, http.StatusBadRequest, "txn_id is required")
		return
	}

	fmt.Println("gateway: status check for txn", txnID)

	txn, ok := h.ledger.GetTxn(txnID)
	if !ok {
		writeError(w, http.StatusNotFound, "transaction not found")
		return
	}

	entries := h.ledger.GetEntries(txnID)

	writeJSON(w, http.StatusOK, map[string]interface{}{
		"txn_id":         txn.ID,
		"trace_id":       txn.TraceID,
		"sender_vpa":     txn.SenderVPA,
		"receiver_vpa":   txn.ReceiverVPA,
		"amount_paise":   txn.AmountPaise,
		"status":         txn.Status,
		"created_at":     txn.CreatedAt,
		"updated_at":     txn.UpdatedAt,
		"ledger_entries": formatEntries(entries),
	})
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	fmt.Println("gateway: health check")
	writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func formatEntries(entries []models.LedgerEntry) []map[string]interface{} {
	result := make([]map[string]interface{}, 0, len(entries))
	for _, e := range entries {
		result = append(result, map[string]interface{}{
			"type":       e.EntryType,
			"account":    e.Account,
			"amount":     e.Amount,
			"created_at": e.CreatedAt,
		})
	}
	return result
}

func writeJSON(w http.ResponseWriter, status int, body interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(body)
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}
