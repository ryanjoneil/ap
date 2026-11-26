package instances

import (
	"math"
	"math/rand/v2"
	"reflect"
	"testing"

	"github.com/ryanjoneil/ap/internal/bruteforce"
)

func rng(seed uint64) *rand.Rand { return rand.New(rand.NewPCG(seed, seed)) }

func TestDeterministic(t *testing.T) {
	a := Dense(50, Uniform(rng(1), 1000))
	b := Dense(50, Uniform(rng(1), 1000))
	c := Dense(50, Uniform(rng(2), 1000))
	if !reflect.DeepEqual(a, b) {
		t.Fatal("same seed gave different matrices")
	}
	if reflect.DeepEqual(a, c) {
		t.Fatal("different seeds gave equal matrices")
	}
}

func TestRanges(t *testing.T) {
	const k = 1000
	for _, row := range Dense(40, Uniform(rng(3), k)) {
		for _, v := range row {
			if v < 0 || v > k {
				t.Fatalf("uniform entry %d out of [0,%d]", v, k)
			}
		}
	}
	for _, row := range Dense(40, TwoCost(rng(4))) {
		for _, v := range row {
			if v != 1 && v != 1000000 {
				t.Fatalf("two-cost entry %d", v)
			}
		}
	}
	max := int64(math.Floor(math.Sqrt2 * (k - 1)))
	for _, row := range Dense(40, Geometric(rng(5), 40, k)) {
		for _, v := range row {
			if v < 0 || v > max {
				t.Fatalf("geometric entry %d out of [0,%d]", v, max)
			}
		}
	}
}

func TestMacholWien(t *testing.T) {
	m := Dense(4, MacholWien())
	if m[2][3] != 6 || m[0][3] != 0 {
		t.Fatalf("got %v", m)
	}
}

func TestSparse(t *testing.T) {
	const n = 8
	for _, tc := range []struct{ degree, want int }{{3, 3}, {20, 8}, {5, 5}} {
		edges := Sparse(rng(7), n, tc.degree, MacholWien())
		ok := make([][]bool, n)
		for i := range ok {
			ok[i] = make([]bool, n)
		}
		count := make([]int, n)
		for _, e := range edges {
			if ok[e.U][e.V] {
				t.Fatalf("duplicate edge %d,%d", e.U, e.V)
			}
			ok[e.U][e.V] = true
			count[e.U]++
			if e.Cost != int64(e.U*e.V) {
				t.Fatalf("cost %d", e.Cost)
			}
		}
		for i, c := range count {
			if c != tc.want {
				t.Fatalf("degree %d: row %d has %d", tc.degree, i, c)
			}
		}
		if got := bruteforce.MaxMatching(ok); got != n {
			t.Fatalf("degree %d: max matching %d", tc.degree, got)
		}
	}
}

func TestSparseLarge(t *testing.T) {
	edges := Sparse(rng(9), 100000, 10, Uniform(rng(10), 100))
	if len(edges) != 1000000 {
		t.Fatalf("got %d edges", len(edges))
	}
}

func TestBookDegree(t *testing.T) {
	if BookDegree(1000) != 20 || BookDegree(1024) != 20 {
		t.Fatal(BookDegree(1000), BookDegree(1024))
	}
}
