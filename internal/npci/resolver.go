package npci

import (
	"errors"
	"fmt"

	"github.com/bhupeshkumar/upi-apps/internal/bank"
)

type BankResolver struct {
	pspToBank map[string]bank.Adapter
}

func NewBankResolver(bankA bank.Adapter, bankB bank.Adapter) *BankResolver {
	r := &BankResolver{
		pspToBank: make(map[string]bank.Adapter),
	}

	r.pspToBank["AXIS_PSP"] = bankA
	r.pspToBank["PAYTM_PSP"] = bankA
	r.pspToBank["HDFC_PSP"] = bankB
	r.pspToBank["YES_PSP"] = bankB

	fmt.Println("npci: bank resolver seeded with", len(r.pspToBank), "PSP mappings")
	return r
}

func (r *BankResolver) GetBank(pspName string) (bank.Adapter, error) {
	adapter, ok := r.pspToBank[pspName]
	if !ok {
		return nil, errors.New("npci: no bank found for PSP " + pspName)
	}
	return adapter, nil
}
