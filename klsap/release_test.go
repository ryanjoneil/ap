package klsap

import (
	"errors"
	"testing"

	"github.com/ryanjoneil/ap"
)

// Once Next returns an error, the waiting subproblems are released.
func TestErrorReleasesQueue(t *testing.T) {
	// This instance's ranking overflows int8 while subproblems are waiting.
	a, err := New([][]int8{{0, 21, 14}, {1, 11, 13}, {0, 15, 13}}, 6)
	if err != nil {
		t.Fatal(err)
	}
	waiting := 0
	for {
		_, err := a.Next()
		if err != nil {
			if !errors.Is(err, ap.ErrOverflow) {
				t.Fatalf("Next = %v; want ErrOverflow", err)
			}
			break
		}
		waiting = a.queue.len()
	}
	if waiting == 0 {
		t.Fatal("no subproblem was waiting before the overflow")
	}
	if a.queue.len() != 0 || a.queue.min != nil || a.queue.max != nil || a.pending != nil || a.root != nil {
		t.Errorf("after ErrOverflow: queue %d, pending %v, root %v; want all released",
			a.queue.len(), a.pending, a.root)
	}
}
