package klsap

import (
	"errors"
	"math/rand/v2"
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

// After each call to Next, the queue holds at most k − r nodes, where r is the
// number of assignments returned so far: offer drops every node that can't
// reach the ranking. An 8×8 uniform instance has up to 7 children per
// assignment, far more than the 20 − r that fit.
func TestQueueBounded(t *testing.T) {
	rng := rand.New(rand.NewPCG(5, 5))
	const n, k = 8, 20
	costs := make([][]int64, n)
	for i := range costs {
		costs[i] = make([]int64, n)
		for j := range costs[i] {
			costs[i][j] = rng.Int64N(100)
		}
	}
	a, err := New(costs, k)
	if err != nil {
		t.Fatal(err)
	}
	full := false
	for r := 1; r <= k; r++ {
		if _, err := a.Next(); err != nil {
			t.Fatalf("Next %d: %v", r, err)
		}
		if got := a.queue.len(); got > k-r {
			t.Fatalf("after Next %d, queue holds %d nodes; want at most %d", r, got, k-r)
		} else if got == k-r {
			full = true
		}
	}
	if !full {
		t.Error("queue never reached its bound, so the test checks nothing")
	}
}
