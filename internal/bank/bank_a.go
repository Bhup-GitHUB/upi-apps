package bank

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type BankA struct{}

func NewBankA() *BankA {
	return &BankA{}
}

func (b *BankA) Name() string {
	return "BankA (Axis)"
}

func (b *BankA) Debit(ctx context.Context, txnID, vpa string, amount int64) error {
	fmt.Println("bank_a: debit started vpa", vpa, "amount", amount, "txn", txnID)

	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		time.Sleep(randomDelay(300, 500))
		if rand.Intn(10) < 2 {
			done <- errors.New("bank_a: core system error on debit")
			return
		}
		done <- nil
	}()

	select {
	case err := <-done:
		if err != nil {
			fmt.Println("bank_a: debit failed", err.Error(), "txn", txnID)
		} else {
			fmt.Println("bank_a: debit success txn", txnID)
		}
		return err
	case <-ctx.Done():
		fmt.Println("bank_a: debit timeout txn", txnID)
		return errors.New("bank_a: debit timed out")
	}
}

func (b *BankA) Credit(ctx context.Context, txnID, vpa string, amount int64) error {
	fmt.Println("bank_a: credit started vpa", vpa, "amount", amount, "txn", txnID)

	ctx, cancel := context.WithTimeout(ctx, 500*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		time.Sleep(randomDelay(300, 500))
		if rand.Intn(10) < 2 {
			done <- errors.New("bank_a: core system error on credit")
			return
		}
		done <- nil
	}()

	select {
	case err := <-done:
		if err != nil {
			fmt.Println("bank_a: credit failed", err.Error(), "txn", txnID)
		} else {
			fmt.Println("bank_a: credit success txn", txnID)
		}
		return err
	case <-ctx.Done():
		fmt.Println("bank_a: credit timeout txn", txnID)
		return errors.New("bank_a: credit timed out")
	}
}

func (b *BankA) Refund(ctx context.Context, txnID, vpa string, amount int64) error {
	fmt.Println("bank_a: refund started vpa", vpa, "amount", amount, "txn", txnID)
	time.Sleep(randomDelay(200, 400))
	fmt.Println("bank_a: refund done txn", txnID)
	return nil
}

func randomDelay(minMs, maxMs int) time.Duration {
	n := minMs + rand.Intn(maxMs-minMs)
	return time.Duration(n) * time.Millisecond
}
