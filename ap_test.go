package ap

import (
	"reflect"
	"testing"
)

func check(t *testing.T, ap *AP, Z int, U, V []int, RC, X [][]int) {
	if Z != ap.Z {
		t.Errorf("expected Z = %v; got %v", Z, ap.Z)
	}
	if !reflect.DeepEqual(U, ap.U) {
		t.Errorf("expected U = %v; got %v", U, ap.U)
	}
	if !reflect.DeepEqual(V, ap.V) {
		t.Errorf("expected V = %v; got %v", V, ap.V)
	}
	for i := 0; i < len(RC); i++ {
		for j := 0; j < len(RC); j++ {
			if got := ap.RC(i, j); got != RC[i][j] {
				t.Errorf("expected RC(%d,%d) = %v; got %v", i, j, RC[i][j], got)
			}
			exp := X[i][j] == 1
			if got := ap.X(i, j); got != exp {
				t.Errorf("expected X(%d,%d) = %v; got %v", i, j, exp, got)
			}
			if got := (ap.Col(i) == j); got != exp {
				t.Errorf("expected (Col(%v) == %v) = %v; got %v", i, j, exp, got)
			}
			if got := (ap.Row(j) == i); got != exp {
				t.Errorf("expected (Row(%v) == %v) = %v; got %v", j, i, exp, got)
			}
		}
	}
}

func TestSolve(t *testing.T) {
	// Input.
	A := [][]int{
		{25, 40, 9, 15},
		{12, 33, 22, 31},
		{18, 37, 18, 30},
		{32, 11, 12, 26},
	}

	// Expected outputs.
	Z := 56
	U := []int{0, 3, 9, 0}
	V := []int{9, 11, 9, 15}
	RC := [][]int{
		{16, 29, 0, 0},
		{0, 19, 10, 13},
		{0, 17, 0, 6},
		{23, 0, 3, 11},
	}
	X := [][]int{
		{0, 0, 0, 1},
		{1, 0, 0, 0},
		{0, 0, 1, 0},
		{0, 1, 0, 0},
	}

	// Optimize and check the solution.
	ap := Create(len(A))
	for i, row := range A {
		for j, val := range row {
			ap.A[i][j] = val
		}
	}
	ap.Solve()
	check(t, ap, Z, U, V, RC, X)

	// Remove arcs, re-optimize, and check.
	M := 1000
	ap.Remove(2, 2, M)
	ap.Remove(1, 0, M)
	ap.Remove(0, 3, M)
	ap.Remove(3, 1, M)
	ap.Remove(2, 0, M)
	ap.Remove(3, 3, M)

	Z2 := 100
	U2 := []int{16, 22, 21, 19}
	V2 := []int{9, 11, -7, 9}
	RC2 := [][]int{
		{0, 13, 0, 975},
		{969, 0, 7, 0},
		{970, 5, 986, 0},
		{4, 970, 0, 972},
	}
	X2 := [][]int{
		{1, 0, 0, 0},
		{0, 1, 0, 0},
		{0, 0, 0, 1},
		{0, 0, 1, 0},
	}

	ap.Solve()
	check(t, ap, Z2, U2, V2, RC2, X2)
}
