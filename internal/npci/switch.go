package npci

import (
	"errors"
	"fmt"
	"strings"

	"github.com/bhupeshkumar/upi-apps/internal/bank"
)

type Switch struct {
	registry *PSPRegistry
	resolver *BankResolver
}

func NewSwitch(registry *PSPRegistry, resolver *BankResolver) *Switch {
	return &Switch{
		registry: registry,
		resolver: resolver,
	}
}

func (s *Switch) ResolveVPA(vpa string) (bank.Adapter, error) {
	fmt.Println("npci: resolving VPA", vpa)

	parts := strings.Split(vpa, "@")
	if len(parts) != 2 {
		return nil, errors.New("npci: malformed VPA " + vpa)
	}
	handle := parts[1]

	fmt.Println("npci: extracted handle", handle)

	pspName, err := s.registry.Resolve(handle)
	if err != nil {
		return nil, err
	}
	fmt.Println("npci: resolved handle", handle, "to PSP", pspName)

	adapter, err := s.resolver.GetBank(pspName)
	if err != nil {
		return nil, err
	}
	fmt.Println("npci: resolved PSP", pspName, "to bank", adapter.Name())

	return adapter, nil
}
