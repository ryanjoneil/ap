package lsap_test

import (
	"errors"
	"math/rand/v2"
	"testing"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/internal/bruteforce"
	"github.com/ryanjoneil/ap/internal/instances"
	"github.com/ryanjoneil/ap/internal/intmath"
	"github.com/ryanjoneil/ap/lsap"
)

// toT converts c to T. The caller ensures every cost fits.
func toT[T ap.Integer](c [][]int64) [][]T {
	ct := make([][]T, len(c))
	for i, row := range c {
		ct[i] = make([]T, len(row))
		for j, cij := range row {
			ct[i][j] = T(cij)
		}
	}
	return ct
}

// buildDense builds a dense solver for c. Every edge exists, so it leaves
// mask all true.
func buildDense[T ap.Integer](c [][]int64, _ [][]bool) (*lsap.LSAP[T], error) {
	return lsap.New(toT[T](c))
}

// denseWantErr reports whether New must reject c for T: some cost is the
// smallest value of T, or M = n·(maxC − minC) + maxC + 1 doesn't fit below
// the largest value of T. The costs are small enough that M fits in int64.
func denseWantErr[T ap.Integer](c [][]int64) bool {
	minC, maxC := c[0][0], c[0][0]
	for _, row := range c {
		for _, cij := range row {
			minC = min(minC, cij)
			maxC = max(maxC, cij)
		}
	}
	m := int64(len(c))*(maxC-minC) + maxC + 1
	return minC == int64(intmath.MinOf[T]()) || m >= int64(intmath.MaxOf[T]())
}

// A single cost c gives A = c and v = c, so the dense margin
// B = (127 − c − 1)/2 holds v exactly when c ≤ 42.
func TestDenseMarginInt8(t *testing.T) {
	a, err := lsap.New([][]int8{{42}})
	if err != nil {
		t.Fatalf("New(42): %v", err)
	}
	if p, err := a.Assign(); err != nil || a.Cost() != 42 {
		t.Errorf("Assign(42) = (%v, %v), cost %v; want cost 42", p, err, a.Cost())
	}

	a, err = lsap.New([][]int8{{43}})
	if err != nil {
		t.Fatalf("New(43): %v", err)
	}
	if p, err := a.Assign(); !errors.Is(err, ap.ErrOverflow) {
		t.Errorf("Assign(43) = (%v, %v); want ErrOverflow", p, err)
	}
}

// Each of these passes New, then Assign returns ErrOverflow from a different
// check.
func TestDenseOverflowInt8(t *testing.T) {
	for name, c := range map[string][][]int8{
		// A = 127 leaves no margin at all.
		"no margin": {{-127}},
		// A = 42 and B = 42. Phase 1 sets v = [−42, −42], and phase 2 sets
		// u[1] = 1 + 42 = 43 > B.
		"phase 2": {{-42, -42}, {1, 1}},
		// A = 40 and B = 43, so the duals fit, but the cost 160 doesn't.
		"cost": {{40, 40, 40, 40}, {40, 40, 40, 40}, {40, 40, 40, 40}, {40, 40, 40, 40}},
	} {
		a, err := lsap.New(c)
		if err != nil {
			t.Fatalf("%s: New: %v", name, err)
		}
		if p, err := a.Assign(); !errors.Is(err, ap.ErrOverflow) {
			t.Errorf("%s: Assign = (%v, %v); want ErrOverflow", name, p, err)
		}
	}
}

func TestDenseNeverWrongInt8(t *testing.T) {
	checkNeverWrong[int8](t, buildDense[int8], denseWantErr[int8])
}

func TestDenseNeverWrongInt16(t *testing.T) {
	checkNeverWrong[int16](t, buildDense[int16], denseWantErr[int16])
}

// checkNeverWrong runs 2000 random trials of at most 4×4 problems with costs
// in [−K, K], where K ranges from 1 to the largest value of T less one. Each
// trial builds a solver with build, then alternately removes an assigned edge
// and a random edge, solving after each removal, until brute force (in int64)
// says the problem is infeasible. Every Assign must either return
// ap.ErrOverflow, which ends the trial, or agree with brute force.
//
// checkNeverWrong passes build an all-true mask, and build sets mask[i][j] to
// false for each edge (i, j) it leaves out of the solver. wantErr reports
// whether build must fail for c; trials for which it does are skipped after
// checking that build failed.
//
// Over all trials, at least one Assign must solve and at least one must
// return ap.ErrOverflow.
func checkNeverWrong[T ap.Integer](
	t *testing.T,
	build func(c [][]int64, mask [][]bool) (*lsap.LSAP[T], error),
	wantErr func(c [][]int64) bool,
) {
	t.Helper()

	maxT := int64(intmath.MaxOf[T]())
	ks := []int64{1, 2, 5, 10, 20, 40, maxT / 4, maxT / 2, maxT - 1}
	rng := rand.New(rand.NewPCG(1, 2))
	solved, overflowed := 0, 0

	for trial := 0; trial < 2000; trial++ {
		n := 1 + rng.IntN(4)
		k := ks[rng.IntN(len(ks))]
		c := make([][]int64, n)
		mask := make([][]bool, n)
		for i := range c {
			c[i] = make([]int64, n)
			mask[i] = make([]bool, n)
			for j := range c[i] {
				c[i][j] = rng.Int64N(2*k+1) - k
				mask[i][j] = true
			}
		}

		a, err := build(c, mask)
		if want := wantErr(c); want != (err != nil) {
			t.Fatalf("%v: build error = %v; want error %v", c, err, want)
		}
		if err != nil {
			continue
		}

		for step := 0; ; step++ {
			// Every other step removes an assigned edge, so a correct solver
			// becomes infeasible within 2n² steps.
			if step > 2*n*n {
				t.Fatalf("%v: still feasible after %v steps", c, step)
			}
			want, feasible := bruteforce.MinSum(c, mask)
			p, err := a.Assign()
			if errors.Is(err, ap.ErrOverflow) {
				overflowed++
				// The overflow is sticky, and carried by Copy.
				if _, err := a.Assign(); !errors.Is(err, ap.ErrOverflow) {
					t.Fatalf("%v mask %v: second Assign = %v; want ErrOverflow", c, mask, err)
				}
				if _, err := a.Copy().Assign(); !errors.Is(err, ap.ErrOverflow) {
					t.Fatalf("%v mask %v: Copy().Assign() = %v; want ErrOverflow", c, mask, err)
				}
				break
			}
			if !feasible {
				if !errors.Is(err, ap.ErrInfeasible) {
					t.Fatalf("%v mask %v: got (%v, %v); want ErrInfeasible or ErrOverflow", c, mask, p, err)
				}
				break
			}
			if err != nil {
				t.Fatalf("%v mask %v: Assign: %v", c, mask, err)
			}
			solved++
			if got := int64(a.Cost()); got != want {
				t.Fatalf("%v mask %v: Cost() = %v; want %v", c, mask, got, want)
			}
			if got := permutationCost(c, p); got != want {
				t.Fatalf("%v mask %v: permutation %v costs %v; want %v", c, mask, p, got, want)
			}
			for u, v := range p {
				if !mask[u][v] {
					t.Fatalf("%v mask %v: %v uses removed edge (%v, %v)", c, mask, p, u, v)
				}
			}

			u := rng.IntN(n)
			v := p[u]
			if step%2 == 1 {
				v = rng.IntN(n)
			}
			mask[u][v] = false
			a.Remove(u, v)
		}
	}

	if solved == 0 || overflowed == 0 {
		t.Errorf("%v steps solved and %v overflowed; want both > 0", solved, overflowed)
	}
}

// The benchmark classes solve in int32 at their benchmark sizes, with the same
// cost as in int64.
func TestNoFalseAlarmsInt32(t *testing.T) {
	if testing.Short() {
		t.Skip("large instances")
	}

	const n = 1000
	for name, c := range map[string][][]int64{
		"uniform":     instances.Dense(n, instances.Uniform(rand.New(rand.NewPCG(1, 1)), 1_000_000)),
		"geometric":   instances.Dense(n, instances.Geometric(rand.New(rand.NewPCG(1, 2)), n, 1_000_000)),
		"two-cost":    instances.Dense(n, instances.TwoCost(rand.New(rand.NewPCG(1, 3)))),
		"machol-wien": instances.Dense(300, instances.MacholWien()),
	} {
		t.Run(name, func(t *testing.T) {
			a64, err := lsap.New(c)
			if err != nil {
				t.Fatalf("New int64: %v", err)
			}
			if _, err := a64.Assign(); err != nil {
				t.Fatalf("Assign int64: %v", err)
			}

			a32, err := lsap.New(toT[int32](c))
			if err != nil {
				t.Fatalf("New int32: %v", err)
			}
			if _, err := a32.Assign(); err != nil {
				t.Fatalf("Assign int32: %v", err)
			}
			if got, want := int64(a32.Cost()), a64.Cost(); got != want {
				t.Errorf("int32 cost %v; int64 cost %v", got, want)
			}
		})
	}
}
