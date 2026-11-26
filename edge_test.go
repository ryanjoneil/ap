package ap_test

import (
	"errors"
	"fmt"
	"testing"

	"github.com/ryanjoneil/ap"
)

func TestErrInfeasible(t *testing.T) {
	err := fmt.Errorf("lsap: %w", ap.ErrInfeasible)
	if !errors.Is(err, ap.ErrInfeasible) || ap.ErrInfeasible.Error() != "ap: no feasible assignment" {
		t.Fail()
	}
	_ = ap.Edge[int64]{U: 0, V: 1, Cost: 5}
	_ = ap.Pair{U: 0, V: 1}
}
