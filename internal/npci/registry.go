package npci

import (
	"errors"
	"fmt"
)

type PSPRegistry struct {
	handles map[string]string
}

func NewPSPRegistry() *PSPRegistry {
	r := &PSPRegistry{
		handles: make(map[string]string),
	}

	r.handles["okaxis"] = "AXIS_PSP"
	r.handles["okhdfcbank"] = "HDFC_PSP"
	r.handles["ybl"] = "YES_PSP"
	r.handles["paytm"] = "PAYTM_PSP"

	fmt.Println("npci: PSP registry seeded with", len(r.handles), "handles")
	return r
}

func (r *PSPRegistry) Resolve(handle string) (string, error) {
	psp, ok := r.handles[handle]
	if !ok {
		return "", errors.New("npci: unknown handle " + handle)
	}
	return psp, nil
}
