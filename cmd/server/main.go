package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/bhupeshkumar/upi-apps/internal/bank"
	"github.com/bhupeshkumar/upi-apps/internal/gateway"
	"github.com/bhupeshkumar/upi-apps/internal/idempotency"
	"github.com/bhupeshkumar/upi-apps/internal/ledger"
	"github.com/bhupeshkumar/upi-apps/internal/npci"
	"github.com/bhupeshkumar/upi-apps/internal/psp"
	"github.com/bhupeshkumar/upi-apps/internal/queue"
	"github.com/bhupeshkumar/upi-apps/internal/saga"
)

func main() {
	fmt.Println("server: starting UPI prototype")

	bankA := bank.NewBankA()
	bankB := bank.NewBankB()

	registry := npci.NewPSPRegistry()
	resolver := npci.NewBankResolver(bankA, bankB)
	npciSwitch := npci.NewSwitch(registry, resolver)

	broker := queue.NewBroker()
	ledgerStore := ledger.NewStore()
	sagaExecutor := saga.NewExecutor(ledgerStore, broker)

	idemStore := idempotency.NewStore()
	rateLimiter := psp.NewRateLimiter(5, time.Minute)

	pspService := psp.NewService(idemStore, npciSwitch, sagaExecutor, ledgerStore, broker, rateLimiter)

	handler := gateway.NewHandler(pspService, ledgerStore)

	mux := http.NewServeMux()
	handler.RegisterRoutes(mux)

	middlewareChain := gateway.LoggingMiddleware(gateway.TraceMiddleware(mux))

	go logEvents(broker)

	fmt.Println("server: listening on :8080")
	fmt.Println("server: seeded VPAs: alice@okaxis, bob@okhdfcbank, charlie@okaxis, diana@ybl")

	if err := http.ListenAndServe(":8080", middlewareChain); err != nil {
		fmt.Println("server: failed to start", err.Error())
	}
}

func logEvents(broker *queue.Broker) {
	for event := range broker.Subscribe() {
		fmt.Println("event:", string(event.Type), "txn", event.TxnID, "trace", event.TraceID)
	}
}
