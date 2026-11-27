package lsap_test

import (
	"errors"
	"math/rand"
	"testing"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/internal/bruteforce"
	"github.com/ryanjoneil/ap/lsap"
)

// namedKernel is a solver built by one of the two kernels.
type namedKernel struct {
	name string
	a    *lsap.LSAP[int64]
}

// fullEdges returns every entry of the square matrix c as an edge.
func fullEdges(c [][]int64) []ap.Edge[int64] {
	var edges []ap.Edge[int64]
	for i, row := range c {
		for j, cij := range row {
			edges = append(edges, ap.Edge[int64]{U: i, V: j, Cost: cij})
		}
	}
	return edges
}

// bothKernels builds the dense kernel with New and the sparse kernel with
// NewSparse on the full edge list of c.
func bothKernels(t *testing.T, c [][]int64) []namedKernel {
	t.Helper()

	d, err := lsap.New(c)
	if err != nil {
		t.Fatalf("%v: New: %v", c, err)
	}
	s, err := lsap.NewSparse(len(c), fullEdges(c))
	if err != nil {
		t.Fatalf("%v: NewSparse: %v", c, err)
	}
	return []namedKernel{{"dense", d}, {"sparse", s}}
}

// fullMask returns an n×n mask allowing every edge.
func fullMask(n int) [][]bool {
	ok := make([][]bool, n)
	for i := range ok {
		ok[i] = make([]bool, n)
		for j := range ok[i] {
			ok[i][j] = true
		}
	}
	return ok
}

// checkAssignMask calls Assign on a and compares the result with brute force
// on c restricted to mask. It returns the assignment, or nil if brute force
// reports the problem infeasible.
func checkAssignMask(t *testing.T, name string, a *lsap.LSAP[int64], c [][]int64, mask [][]bool) ap.Permutation {
	t.Helper()

	want, feasible := bruteforce.MinSum(c, mask)
	p, err := a.Assign()
	if !feasible {
		if !errors.Is(err, ap.ErrInfeasible) {
			t.Errorf("%s: %v mask %v: got (%v, %v); want ErrInfeasible", name, c, mask, p, err)
		}
		return nil
	}
	if err != nil {
		t.Fatalf("%s: %v mask %v: Assign: %v", name, c, mask, err)
	}
	for u, v := range p {
		if !mask[u][v] {
			t.Errorf("%s: %v mask %v: %v uses removed edge (%v, %v)", name, c, mask, p, u, v)
		}
	}
	if got := permutationCost(c, p); got != want {
		t.Errorf("%s: %v mask %v: permutation %v costs %v; want %v", name, c, mask, p, got, want)
	}
	if got := a.Cost(); got != want {
		t.Errorf("%s: %v mask %v: Cost() = %v; want %v", name, c, mask, got, want)
	}
	return p
}

func TestKernelsAgree(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for trial := 0; trial < 500; trial++ {
		n := 1 + rng.Intn(7)
		c := randomCosts(rng, n)
		want, _ := bruteforce.MinSum(c, nil)
		var costs []int64
		for _, k := range bothKernels(t, c) {
			if _, err := k.a.Assign(); err != nil {
				t.Fatalf("%s: %v: Assign: %v", k.name, c, err)
			}
			costs = append(costs, k.a.Cost())
		}
		if costs[0] != costs[1] || costs[0] != want {
			t.Errorf("%v: dense %v, sparse %v; brute force %v", c, costs[0], costs[1], want)
		}
	}
}

// checkDuals checks that every live edge has a nonnegative reduced cost, that
// every assigned edge has a zero reduced cost, and that ΣU + ΣV == Cost().
func checkDuals(t *testing.T, name string, a *lsap.LSAP[int64], c [][]int64, mask [][]bool, p ap.Permutation) {
	t.Helper()

	for u := range c {
		for v := range c[u] {
			rc, ok := a.ReducedCost(u, v)
			switch {
			case ok != mask[u][v]:
				t.Errorf("%s: %v: ReducedCost(%v, %v) ok = %v; want %v", name, c, u, v, ok, mask[u][v])
			case ok && rc < 0:
				t.Errorf("%s: %v: ReducedCost(%v, %v) = %v < 0", name, c, u, v, rc)
			case ok && p[u] == v && rc != 0:
				t.Errorf("%s: %v: assigned edge (%v, %v) has reduced cost %v", name, c, u, v, rc)
			}
		}
	}
	d := a.DualPrices()
	var sum int64
	for i := range d.U {
		sum += d.U[i] + d.V[i]
	}
	if sum != a.Cost() {
		t.Errorf("%s: %v: ΣU + ΣV = %v; Cost() = %v", name, c, sum, a.Cost())
	}
}

func TestDuals(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for trial := 0; trial < 500; trial++ {
		n := 1 + rng.Intn(7)
		c := randomCosts(rng, n)
		mask := fullMask(n)
		for u := range mask {
			for v := range mask[u] {
				mask[u][v] = rng.Float64() >= 0.3
			}
		}
		for _, k := range bothKernels(t, c) {
			if trial%2 == 1 {
				// Solve first, so that the removals re-solve incrementally.
				_, _ = k.a.Assign()
			}
			for u := range mask {
				for v := range mask[u] {
					if !mask[u][v] {
						k.a.Fix(u, v, false)
					}
				}
			}
			if p := checkAssignMask(t, k.name, k.a, c, mask); p != nil {
				checkDuals(t, k.name, k.a, c, mask, p)
			}
		}
	}
}

// runIncremental interleaves Assign and Fix(u, v, false) on a until the problem becomes
// infeasible, checking every Assign against brute force on the current mask.
// If alternate is true, every other step removes a random live edge rather
// than an assigned edge.
func runIncremental(t *testing.T, name string, a *lsap.LSAP[int64], c [][]int64, rng *rand.Rand, alternate bool) {
	t.Helper()

	n := len(c)
	mask := fullMask(n)
	for step := 0; ; step++ {
		// Every step removes a live edge, so a correct solver becomes
		// infeasible within n² steps.
		if step > n*n {
			t.Fatalf("%s: %v: still feasible after %v removals", name, c, step)
		}
		p := checkAssignMask(t, name, a, c, mask)
		if p == nil {
			return
		}
		checkDuals(t, name, a, c, mask, p)

		var u, v int
		if alternate && step%2 == 1 {
			var live [][2]int
			for i := range mask {
				for j := range mask[i] {
					if mask[i][j] {
						live = append(live, [2]int{i, j})
					}
				}
			}
			e := live[rng.Intn(len(live))]
			u, v = e[0], e[1]
		} else {
			u = rng.Intn(n)
			v = p[u]
		}
		mask[u][v] = false
		a.Fix(u, v, false)
	}
}

func TestIncrementalBothKernels(t *testing.T) {
	for trial := 0; trial < 300; trial++ {
		rng := rand.New(rand.NewSource(int64(trial)))
		n := 2 + rng.Intn(5)
		c := randomCosts(rng, n)
		for _, k := range bothKernels(t, c) {
			// Each kernel sees the same sequence of removals.
			runIncremental(t, k.name, k.a, c, rand.New(rand.NewSource(int64(trial))), true)
		}
	}
}
