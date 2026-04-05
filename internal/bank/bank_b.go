package bank

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"
)

type BankB struct{}

func NewBankB() *BankB {
	return &BankB{}
}

func (b *BankB) Name() string {
	return "BankB (HDFC)"
}

func (b *BankB) Debit(ctx context.Context, txnID, vpa string, amount int64) error {
	fmt.Println("bank_b: debit started vpa", vpa, "amount", amount, "txn", txnID)

	ctx, cancel := context.WithTimeout(ctx, 700*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		time.Sleep(randomDelay(400, 700))
		if rand.Intn(10) < 2 {
			done <- errors.New("bank_b: core system error on debit")
			return
		}
		done <- nil
	}()

	select {
	case err := <-done:
		if err != nil {
			fmt.Println("bank_b: debit failed", err.Error(), "txn", txnID)
		} else {
			fmt.Println("bank_b: debit success txn", txnID)
		}
		return err
	case <-ctx.Done():
		fmt.Println("bank_b: debit timeout txn", txnID)
		return errors.New("bank_b: debit timed out")
	}
}

func (b *BankB) Credit(ctx context.Context, txnID, vpa string, amount int64) error {
	fmt.Println("bank_b: credit started vpa", vpa, "amount", amount, "txn", txnID)

	ctx, cancel := context.WithTimeout(ctx, 700*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		time.Sleep(randomDelay(400, 700))
		if rand.Intn(10) < 2 {
			done <- errors.New("bank_b: core system error on credit")
			return
		}
		done <- nil
	}()

	select {
	case err := <-done:
		if err != nil {
			fmt.Println("bank_b: credit failed", err.Error(), "txn", txnID)
		} else {
			fmt.Println("bank_b: credit success txn", txnID)
		}
		return err
	case <-ctx.Done():
		fmt.Println("bank_b: credit timeout txn", txnID)
		return errors.New("bank_b: credit timed out")
	}
}

func (b *BankB) Refund(ctx context.Context, txnID, vpa string, amount int64) error {
	fmt.Println("bank_b: refund started vpa", vpa, "amount", amount, "txn", txnID)
	time.Sleep(randomDelay(200, 400))
	fmt.Println("bank_b: refund done txn", txnID)
	return nil
}
