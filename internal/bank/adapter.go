package bank

import "context"

type Adapter interface {
	Debit(ctx context.Context, txnID, vpa string, amount int64) error
	Credit(ctx context.Context, txnID, vpa string, amount int64) error
	Refund(ctx context.Context, txnID, vpa string, amount int64) error
	Name() string
}
