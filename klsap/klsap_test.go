package klsap_test

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"testing"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/internal/bruteforce"
	"github.com/ryanjoneil/ap/internal/instances"
	"github.com/ryanjoneil/ap/klsap"
	"github.com/ryanjoneil/ap/lsap"
)

// instance is a cost matrix, the mask of edges that exist, and a constructor
// for a solver of it that ranks up to k assignments.
type instance struct {
	name  string
	c     [][]int64
	mask  [][]bool
	build func(k int) (*klsap.KLSAP[int64], error)
}

func fullMask(n int) [][]bool {
	mask := make([][]bool, n)
	for i := range mask {
		mask[i] = make([]bool, n)
		for j := range mask[i] {
			mask[i][j] = true
		}
	}
	return mask
}

func denseInstance(name string, c [][]int64) instance {
	return instance{name, c, fullMask(len(c)), func(k int) (*klsap.KLSAP[int64], error) {
		return klsap.New(c, k)
	}}
}

// sparseInstance keeps each edge of c with probability density.
func sparseInstance(rng *rand.Rand, c [][]int64, density float64) instance {
	n := len(c)
	mask := make([][]bool, n)
	var edges []ap.Edge[int64]
	for i, row := range c {
		mask[i] = make([]bool, n)
		for j, cij := range row {
			if rng.Float64() < density {
				mask[i][j] = true
				edges = append(edges, ap.Edge[int64]{U: i, V: j, Cost: cij})
			}
		}
	}
	name := fmt.Sprintf("sparse %.2f", density)
	return instance{name, c, mask, func(k int) (*klsap.KLSAP[int64], error) {
		return klsap.NewSparse(n, edges, k)
	}}
}

// randomInstances returns dense uniform, two-cost and Machol–Wien instances
// of size n, and a sparse instance with density in [0.4, 0.8].
func randomInstances(rng *rand.Rand, n int) []instance {
	uniform := instances.Dense(n, instances.Uniform(rng, 20))
	return []instance{
		denseInstance("uniform", uniform),
		denseInstance("two-cost", instances.Dense(n, instances.TwoCost(rng))),
		denseInstance("Machol–Wien", instances.Dense(n, instances.MacholWien())),
		sparseInstance(rng, uniform, 0.4+0.4*rng.Float64()),
	}
}

func factorial(n int) int {
	f := 1
	for i := 2; i <= n; i++ {
		f *= i
	}
	return f
}

// drain calls Next until it returns an error, and returns every assignment
// and cost before it, and the error. It checks that the error repeats.
func drain[T ap.Integer](t *testing.T, a *klsap.KLSAP[T]) ([]ap.Permutation, []T, error) {
	t.Helper()

	var perms []ap.Permutation
	var costs []T
	for {
		p, err := a.Next()
		if err != nil {
			for range 2 {
				if q, again := a.Next(); again != err {
					t.Fatalf("Next after %v = (%v, %v); want the same error", err, q, again)
				}
			}
			return perms, costs, err
		}
		perms = append(perms, p)
		costs = append(costs, a.Cost())
	}
}

// checkRanking checks that the assignments are distinct permutations using
// only edges in mask, that each one's cost is costs[r], and that costs is
// want.
func checkRanking(
	t *testing.T,
	name string,
	c [][]int64,
	mask [][]bool,
	perms []ap.Permutation,
	costs []int64,
	want []int64,
) {
	t.Helper()

	if !slices.Equal(costs, want) {
		t.Fatalf("%s %v: costs = %v; want %v", name, c, costs, want)
	}
	seen := map[string]bool{}
	for r, p := range perms {
		if len(p) != len(c) {
			t.Fatalf("%s %v: assignment %d = %v; want length %d", name, c, r, p, len(c))
		}
		used := make([]bool, len(c))
		var cost int64
		for u, v := range p {
			if v < 0 || v >= len(c) || used[v] || !mask[u][v] {
				t.Fatalf("%s %v: assignment %d = %v uses an invalid edge (%d, %d)", name, c, r, p, u, v)
			}
			used[v] = true
			cost += c[u][v]
		}
		if cost != costs[r] {
			t.Fatalf("%s %v: assignment %d = %v costs %d; Cost = %d", name, c, r, p, cost, costs[r])
		}
		key := fmt.Sprint(p)
		if seen[key] {
			t.Fatalf("%s %v: assignment %d = %v repeats", name, c, r, p)
		}
		seen[key] = true
	}
}

func TestRankingMatchesBruteForce(t *testing.T) {
	rng := rand.New(rand.NewPCG(1, 2))
	for trial := 0; trial < 100; trial++ {
		n := 1 + trial%6
		for _, inst := range randomInstances(rng, n) {
			all := bruteforce.AllSums(inst.c, inst.mask)

			// Every assignment.
			a, err := inst.build(factorial(n) + 5)
			if err != nil {
				t.Fatalf("%s %v: %v", inst.name, inst.c, err)
			}
			perms, costs, err := drain(t, a)
			if !errors.Is(err, ap.ErrInfeasible) {
				t.Fatalf("%s %v: Next = %v; want ErrInfeasible", inst.name, inst.c, err)
			}
			checkRanking(t, inst.name, inst.c, inst.mask, perms, costs, all)

			// The first k, with pruning.
			k := 1 + rng.IntN(factorial(n))
			a, err = inst.build(k)
			if err != nil {
				t.Fatalf("%s %v: %v", inst.name, inst.c, err)
			}
			perms, costs, err = drain(t, a)
			if !errors.Is(err, ap.ErrInfeasible) {
				t.Fatalf("%s %v k=%d: Next = %v; want ErrInfeasible", inst.name, inst.c, k, err)
			}
			checkRanking(t, inst.name, inst.c, inst.mask, perms, costs, all[:min(k, len(all))])
		}
	}
}

func TestSmallK(t *testing.T) {
	rng := rand.New(rand.NewPCG(3, 4))
	c := instances.Dense(5, instances.Uniform(rng, 100))

	ref, err := lsap.New(c)
	if err != nil {
		t.Fatal(err)
	}
	opt, err := ref.Assign()
	if err != nil {
		t.Fatal(err)
	}

	a, err := klsap.New(c, 1)
	if err != nil {
		t.Fatal(err)
	}
	perms, costs, err := drain(t, a)
	if !errors.Is(err, ap.ErrInfeasible) {
		t.Fatalf("k=1: Next = %v; want ErrInfeasible", err)
	}
	if len(perms) != 1 || !slices.Equal(perms[0], opt) || costs[0] != ref.Cost() {
		t.Fatalf("k=1: got %v costing %v; want [%v] costing [%v]", perms, costs, opt, ref.Cost())
	}

	a, err = klsap.New(c, 3)
	if err != nil {
		t.Fatal(err)
	}
	perms, costs, err = drain(t, a)
	if !errors.Is(err, ap.ErrInfeasible) {
		t.Fatalf("k=3: Next = %v; want ErrInfeasible", err)
	}
	checkRanking(t, "k=3", c, fullMask(5), perms, costs, bruteforce.AllSums(c, nil)[:3])
}

func TestOneByOne(t *testing.T) {
	a, err := klsap.New([][]int{{7}}, 3)
	if err != nil {
		t.Fatal(err)
	}
	if p, err := a.Next(); err != nil || !slices.Equal(p, ap.Permutation{0}) || a.Cost() != 7 {
		t.Fatalf("Next = (%v, %v), cost %v; want ([0], nil), cost 7", p, err, a.Cost())
	}
	for range 3 {
		if p, err := a.Next(); !errors.Is(err, ap.ErrInfeasible) {
			t.Fatalf("Next = (%v, %v); want ErrInfeasible", p, err)
		}
	}
}

func TestInfeasible(t *testing.T) {
	a, err := klsap.NewSparse(2, []ap.Edge[int]{{U: 0, V: 0, Cost: 1}, {U: 1, V: 0, Cost: 2}}, 5)
	if err != nil {
		t.Fatal(err)
	}
	for range 3 {
		if p, err := a.Next(); !errors.Is(err, ap.ErrInfeasible) {
			t.Fatalf("Next = (%v, %v); want ErrInfeasible", p, err)
		}
	}
}

// overflowInstances searches small int8 instances with costs near the dense
// margin for ones whose root solves but whose ranking overflows.
func overflowInstances(t *testing.T, want int) [][][]int8 {
	t.Helper()

	rng := rand.New(rand.NewPCG(5, 6))
	var found [][][]int8
	for tries := 0; tries < 100_000 && len(found) < want; tries++ {
		n := 2 + rng.IntN(3)
		hi := 20 + rng.IntN(25)
		c := make([][]int8, n)
		for i := range c {
			c[i] = make([]int8, n)
			for j := range c[i] {
				c[i][j] = int8(rng.IntN(hi + 1))
			}
		}
		a, err := klsap.New(c, factorial(n))
		if err != nil {
			continue
		}
		if perms, _, err := drain(t, a); len(perms) > 0 && errors.Is(err, ap.ErrOverflow) {
			found = append(found, c)
		}
	}
	if len(found) < want {
		t.Fatalf("found %d overflowing instances; want %d", len(found), want)
	}
	return found
}

func TestOverflow(t *testing.T) {
	for _, c := range overflowInstances(t, 20) {
		c64 := make([][]int64, len(c))
		for i, row := range c {
			c64[i] = make([]int64, len(row))
			for j, cij := range row {
				c64[i][j] = int64(cij)
			}
		}

		a, err := klsap.New(c, factorial(len(c)))
		if err != nil {
			t.Fatal(err)
		}
		perms, costs8, err := drain(t, a)
		if !errors.Is(err, ap.ErrOverflow) {
			t.Fatalf("%v: Next = %v; want ErrOverflow", c, err)
		}
		costs := make([]int64, len(costs8))
		for r, cost := range costs8 {
			costs[r] = int64(cost)
		}
		all := bruteforce.AllSums(c64, nil)
		checkRanking(t, "int8", c64, fullMask(len(c)), perms, costs, all[:len(costs)])
	}
}

func TestNewErrors(t *testing.T) {
	c := [][]int{{1, 2}, {3, 4}}
	edges := []ap.Edge[int]{{U: 0, V: 0, Cost: 1}}
	for name, build := range map[string]func() (*klsap.KLSAP[int], error){
		"k = 0":        func() (*klsap.KLSAP[int], error) { return klsap.New(c, 0) },
		"sparse k = 0": func() (*klsap.KLSAP[int], error) { return klsap.NewSparse(1, edges, 0) },
		"empty":        func() (*klsap.KLSAP[int], error) { return klsap.New([][]int{}, 1) },
		"ragged":       func() (*klsap.KLSAP[int], error) { return klsap.New([][]int{{1, 2}, {3}}, 1) },
		"sparse n = 0": func() (*klsap.KLSAP[int], error) { return klsap.NewSparse(0, edges, 1) },
		"sparse range": func() (*klsap.KLSAP[int], error) { return klsap.NewSparse(1, []ap.Edge[int]{{U: 1}}, 1) },
	} {
		a, err := build()
		if err == nil || a != nil {
			t.Errorf("%s: got (%v, %v); want an error", name, a, err)
			continue
		}
		if !strings.HasPrefix(err.Error(), "klsap: ") {
			t.Errorf("%s: error %q; want prefix %q", name, err, "klsap: ")
		}
	}
}

func TestCallerOwnsPermutation(t *testing.T) {
	rng := rand.New(rand.NewPCG(7, 8))
	c := instances.Dense(4, instances.Uniform(rng, 10))

	ref, err := klsap.New(c, 24)
	if err != nil {
		t.Fatal(err)
	}
	want, _, _ := drain(t, ref)

	a, err := klsap.New(c, 24)
	if err != nil {
		t.Fatal(err)
	}
	for r := range want {
		p, err := a.Next()
		if err != nil {
			t.Fatalf("Next %d: %v", r, err)
		}
		if !slices.Equal(p, want[r]) {
			t.Fatalf("Next %d = %v; want %v", r, p, want[r])
		}
		for u := range p {
			p[u] = 0
		}
	}
}
