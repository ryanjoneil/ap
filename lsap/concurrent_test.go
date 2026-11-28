package lsap_test

import (
	"math/rand"
	"slices"
	"sync"
	"testing"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/internal/bruteforce"
)

// TestConcurrentCopy calls Copy concurrently on one solved parent. Each
// goroutine removes a different assigned edge from its copy and re-solves.
// Run it with -race.
func TestConcurrentCopy(t *testing.T) {
	const n, workers = 6, 8

	c := randomCosts(rand.New(rand.NewSource(1)), n)
	for _, k := range bothKernels(t, c) {
		parent := k.a
		p, err := parent.Assign()
		if err != nil {
			t.Fatalf("%s: Assign: %v", k.name, err)
		}
		cost, duals := parent.Cost(), parent.DualPrices()

		type result struct {
			p    ap.Permutation
			cost int64
			err  error
		}
		results := make([]result, workers)

		var wg sync.WaitGroup
		for w := range workers {
			wg.Add(1)
			go func() {
				defer wg.Done()
				a := parent.Copy()
				u := w % n
				a.Fix(u, p[u], false)
				q, err := a.Assign()
				results[w] = result{q, a.Cost(), err}
			}()
		}
		wg.Wait()

		for w, r := range results {
			u := w % n
			mask := fullMask(n)
			mask[u][p[u]] = false
			want, _ := bruteforce.MinSum(c, mask)
			switch {
			case r.err != nil:
				t.Errorf("%s: worker %d: Assign: %v", k.name, w, r.err)
			case r.p[u] == p[u]:
				t.Errorf("%s: worker %d: %v uses removed edge (%d, %d)", k.name, w, r.p, u, p[u])
			case permutationCost(c, r.p) != want || r.cost != want:
				t.Errorf("%s: worker %d: %v costs %d, Cost() = %d; want %d",
					k.name, w, r.p, permutationCost(c, r.p), r.cost, want)
			}
		}

		// The parent is unchanged.
		q, err := parent.Assign()
		if err != nil || !slices.Equal(q, p) || parent.Cost() != cost {
			t.Errorf("%s: parent after copies: (%v, %v) cost %d; want (%v, nil) cost %d",
				k.name, q, err, parent.Cost(), p, cost)
		}
		d := parent.DualPrices()
		if !slices.Equal(d.U, duals.U) || !slices.Equal(d.V, duals.V) {
			t.Errorf("%s: parent duals %v; want %v", k.name, d, duals)
		}
		for u := range c {
			for v := range c[u] {
				if _, ok := parent.ReducedCost(u, v); !ok {
					t.Errorf("%s: parent edge (%d, %d) removed", k.name, u, v)
				}
			}
		}
	}
}
