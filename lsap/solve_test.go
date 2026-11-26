package lsap_test

import (
	"errors"
	"math/rand"
	"reflect"
	"testing"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/internal/bruteforce"
	"github.com/ryanjoneil/ap/lsap"
)

func permutationCost(c [][]int64, p ap.Permutation) int64 {
	var z int64
	for u, v := range p {
		z += c[u][v]
	}
	return z
}

func randomCosts(rng *rand.Rand, n int) [][]int64 {
	c := make([][]int64, n)
	for i := range c {
		c[i] = make([]int64, n)
		for j := range c[i] {
			c[i][j] = rng.Int63n(21) - 10 // [-10, 10]
		}
	}
	return c
}

func checkOptimal(t *testing.T, c [][]int64) {
	t.Helper()

	want, _ := bruteforce.MinSum(c, nil)
	a, err := lsap.New(c)
	if err != nil {
		t.Fatalf("%v: New: %v", c, err)
	}
	p, err := a.Assign()
	if err != nil {
		t.Fatalf("%v: Assign: %v", c, err)
	}

	if got := permutationCost(c, p); got != want {
		t.Errorf("%v: permutation %v costs %v; want %v", c, p, got, want)
	}
	if got := a.Cost(); got != want {
		t.Errorf("%v: Cost() = %v; want %v", c, got, want)
	}
}

func TestNegativeCosts(t *testing.T) {
	checkOptimal(t, [][]int64{
		{8, 2},
		{6, -7},
	})
	checkOptimal(t, [][]int64{
		{-7, 1, -2, -1},
		{-10, 8, -2, 9},
		{5, 8, -8, 9},
		{-3, 7, -1, 6},
	})
}

func TestMatchesBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for trial := 0; trial < 2000; trial++ {
		checkOptimal(t, randomCosts(rng, 2+rng.Intn(5)))
	}
}

func TestDeterministicWithTies(t *testing.T) {
	c := [][]int64{
		{1, 2, 1, 0},
		{1, 2, 2, 0},
		{0, 0, 0, 2},
		{1, 1, 1, 0},
	}

	assign := func() ap.Permutation {
		a, err := lsap.New(c)
		if err != nil {
			t.Fatal(err)
		}
		p, err := a.Assign()
		if err != nil {
			t.Fatal(err)
		}
		return p
	}

	first := assign()
	for run := 0; run < 100; run++ {
		if p := assign(); !reflect.DeepEqual(first, p) {
			t.Fatalf("run %v: got %v; first run got %v", run, p, first)
		}
	}
}

func TestNewErrors(t *testing.T) {
	for name, c := range map[string][][]int64{
		"empty": {}, "ragged": {{1, 2}, {3}}, "not square": {{1, 2}},
	} {
		if _, err := lsap.New(c); err == nil {
			t.Errorf("%s: want error", name)
		}
	}
	// M = 2·(1+1) + 1 + 1 = 6, and 4·(2+1)·6 = 72 ≤ 127.
	if _, err := lsap.New([][]int8{{-1, 1}, {1, -1}}); err != nil {
		t.Errorf("int8 within bound: %v", err)
	}
	// M = 2·(2+2) + 2 + 1 = 11, and 4·(2+1)·11 = 132 > 127.
	if _, err := lsap.New([][]int8{{-2, 2}, {2, -2}}); err == nil {
		t.Error("int8 past bound: want error")
	}
}

func TestNewCopiesInput(t *testing.T) {
	c := [][]int64{{1, 2}, {3, 4}}
	a, err := lsap.New(c)
	if err != nil {
		t.Fatal(err)
	}
	a.Remove(0, 0)
	_, _ = a.Assign()
	if !reflect.DeepEqual(c, [][]int64{{1, 2}, {3, 4}}) {
		t.Errorf("input modified: %v", c)
	}
}

func TestRemoveInfeasible(t *testing.T) {
	a, err := lsap.New([][]int64{{1, 2}, {3, 4}})
	if err != nil {
		t.Fatal(err)
	}
	a.Remove(0, 0)
	a.Remove(0, 1)
	if _, err := a.Assign(); !errors.Is(err, ap.ErrInfeasible) {
		t.Errorf("got %v; want ErrInfeasible", err)
	}
}

// checkMasked solves c after removing every edge with ok[i][j] == false, and
// checks the result against brute force with the same mask. A feasible result
// must never use a removed edge.
func checkMasked[T ap.Integer](t *testing.T, c [][]T, ok [][]bool) {
	t.Helper()

	c64 := make([][]int64, len(c))
	for i, row := range c {
		c64[i] = make([]int64, len(row))
		for j, cij := range row {
			c64[i][j] = int64(cij)
		}
	}
	want, feasible := bruteforce.MinSum(c64, ok)

	a, err := lsap.New(c)
	if err != nil {
		t.Fatalf("%v: New: %v", c, err)
	}
	for i := range ok {
		for j := range ok[i] {
			if !ok[i][j] {
				a.Remove(i, j)
			}
		}
	}

	p, err := a.Assign()
	if !feasible {
		if !errors.Is(err, ap.ErrInfeasible) {
			t.Errorf("%v mask %v: got (%v, %v); want ErrInfeasible", c, ok, p, err)
		}
		return
	}
	if err != nil {
		t.Fatalf("%v mask %v: Assign: %v", c, ok, err)
	}
	for u, v := range p {
		if !ok[u][v] {
			t.Errorf("%v mask %v: %v uses removed edge (%v, %v)", c, ok, p, u, v)
		}
	}
	if got := permutationCost(c64, p); got != want {
		t.Errorf("%v mask %v: permutation %v costs %v; want %v", c, ok, p, got, want)
	}
	if got := int64(a.Cost()); got != want {
		t.Errorf("%v mask %v: Cost() = %v; want %v", c, ok, got, want)
	}
}

// Random removals: the result matches brute force with the same mask, and
// never uses a removed edge.
func TestRemoveMatchesBruteForce(t *testing.T) {
	rng := rand.New(rand.NewSource(1))
	for trial := 0; trial < 1000; trial++ {
		n := 1 + rng.Intn(6)
		c := randomCosts(rng, n)
		ok := make([][]bool, n)
		for i := range ok {
			ok[i] = make([]bool, n)
			for j := range ok[i] {
				ok[i][j] = rng.Float64() >= 0.3
			}
		}
		checkMasked(t, c, ok)
	}
}

// Review Focus 1.
func TestCopyIndependent(t *testing.T) {
	c := [][]int64{{4, 1, 3}, {2, 0, 5}, {3, 2, 2}}
	a, err := lsap.New(c)
	if err != nil {
		t.Fatal(err)
	}
	_, _ = a.Assign()
	b := a.Copy()
	b.Remove(0, 1)
	b.Remove(1, 1)
	pa, _ := a.Assign()
	if a.Cost() != 5 || !reflect.DeepEqual(pa, ap.Permutation{1, 0, 2}) {
		t.Errorf("original changed: %v %v", pa, a.Cost())
	}
	a.Remove(2, 2)
	if _, ok := b.ReducedCost(2, 2); !ok {
		t.Error("Remove on the original removed the edge from the copy")
	}
	pb, _ := b.Assign()
	want, _ := bruteforce.MinSum(c, [][]bool{{true, false, true}, {true, false, true}, {true, true, true}})
	if b.Cost() != want {
		t.Errorf("copy: %v cost %v, want %v", pb, b.Cost(), want)
	}
}

func TestReducedCostOutOfRange(t *testing.T) {
	a, err := lsap.New([][]int64{{1}})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := a.ReducedCost(1, 0); ok {
		t.Error("want ok == false")
	}
}

// Every 2×2 int8 matrix with entries in [-1,1], with every removal mask,
// matches brute force.
func TestOverflowBoundInt8(t *testing.T) {
	vals := []int8{-1, 0, 1}
	for m := 0; m < 81; m++ {
		c := [][]int8{
			{vals[m%3], vals[m/3%3]},
			{vals[m/9%3], vals[m/27%3]},
		}
		for mask := 0; mask < 16; mask++ {
			ok := [][]bool{
				{mask&1 == 0, mask&2 == 0},
				{mask&4 == 0, mask&8 == 0},
			}
			checkMasked(t, c, ok)
		}
	}
}
