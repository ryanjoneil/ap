package lsap_test

import (
	"math/rand"
	"reflect"
	"testing"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/lsap"
)

// bruteForce returns the minimum cost of any assignment for a small matrix.
func bruteForce(c [][]int64) int64 {
	n := len(c)
	best := ap.MaxOf[int64]()
	used := make([]bool, n)

	var rec func(i int, sum int64)
	rec = func(i int, sum int64) {
		if i == n {
			if sum < best {
				best = sum
			}
			return
		}
		for j := 0; j < n; j++ {
			if !used[j] {
				used[j] = true
				rec(i+1, sum+c[i][j])
				used[j] = false
			}
		}
	}

	rec(0, 0)
	return best
}

func copyMatrix(c [][]int64) [][]int64 {
	d := make([][]int64, len(c))
	for i := range c {
		d[i] = append([]int64{}, c[i]...)
	}
	return d
}

func permutationCost(c [][]int64, p ap.Permutation) int64 {
	var z int64
	for u, v := range p {
		z += c[u][v]
	}
	return z
}

func checkOptimal(t *testing.T, c [][]int64) {
	t.Helper()

	want := bruteForce(c)
	a := lsap.New(copyMatrix(c))
	p := a.Assign()

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
		n := 2 + rng.Intn(5)
		c := make([][]int64, n)
		for i := range c {
			c[i] = make([]int64, n)
			for j := range c[i] {
				c[i][j] = rng.Int63n(21) - 10 // [-10, 10]
			}
		}
		checkOptimal(t, c)
	}
}

func TestDeterministicWithTies(t *testing.T) {
	c := [][]int64{
		{1, 2, 1, 0},
		{1, 2, 2, 0},
		{0, 0, 0, 2},
		{1, 1, 1, 0},
	}

	first := lsap.New(copyMatrix(c)).Assign()
	for run := 0; run < 100; run++ {
		if p := lsap.New(copyMatrix(c)).Assign(); !reflect.DeepEqual(first, p) {
			t.Fatalf("run %v: got %v; first run got %v", run, p, first)
		}
	}
}
