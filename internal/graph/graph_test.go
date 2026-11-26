package graph

import (
	"slices"
	"strings"
	"testing"

	"github.com/ryanjoneil/ap"
)

func TestFromEdges(t *testing.T) {
	g, c, err := FromEdges(3, []ap.Edge[int64]{{U: 2, V: 0, Cost: 11}, {U: 0, V: 2, Cost: 12}, {U: 0, V: 1, Cost: 15}})
	if err != nil {
		t.Fatal(err)
	}
	if !slices.Equal(g.Start, []int{0, 2, 2, 3}) || !slices.Equal(g.Adj, []int{1, 2, 0}) || !slices.Equal(c, []int64{15, 12, 11}) {
		t.Errorf("got %v %v %v", g.Start, g.Adj, c)
	}
	if g.Find(0, 2) != 1 || g.Find(1, 0) != -1 || g.Find(5, 0) != -1 || g.Find(0, -1) != -1 {
		t.Error("Find")
	}
}

func TestFromPairs(t *testing.T) {
	in := []ap.Pair{{U: 1, V: 2}, {U: 0, V: 1}, {U: 1, V: 0}}
	g, err := FromPairs(2, 3, in)
	if err != nil {
		t.Fatal(err)
	}
	if g.NU != 2 || g.NV != 3 || !slices.Equal(g.Start, []int{0, 1, 3}) || !slices.Equal(g.Adj, []int{1, 0, 2}) {
		t.Errorf("got %+v", g)
	}
}

func TestErrors(t *testing.T) {
	cases := map[string]error{}
	_, err := FromPairs(0, 1, nil)
	cases["empty"] = err
	_, err = FromPairs(2, 2, []ap.Pair{{U: 0, V: 2}})
	cases["out of range"] = err
	_, err = FromPairs(2, 2, []ap.Pair{{U: 0, V: 1}, {U: 0, V: 1}})
	cases["duplicate"] = err
	_, err = FromAdjacency([][]bool{{true}, {true, false}})
	cases["ragged"] = err
	_, _, err = FromMatrix([][]int64{{1, 2}})
	cases["not square"] = err
	for name, err := range cases {
		if err == nil || !strings.Contains(err.Error(), name) {
			t.Errorf("%s: got %v", name, err)
		}
	}
}

func TestEmptyInputs(t *testing.T) {
	_, err1 := FromAdjacency([][]bool{})
	_, err2 := FromAdjacency([][]bool{{}})
	_, _, err3 := FromEdges[int](0, nil)
	_, _, err4 := FromMatrix[int](nil)
	_, _, err5 := FromMatrix([][]int{{}})
	for i, err := range []error{err1, err2, err3, err4, err5} {
		if err == nil || !strings.Contains(err.Error(), "empty") {
			t.Errorf("case %d: got %v", i, err)
		}
	}
}

func TestFromEdgesErrors(t *testing.T) {
	_, _, err := FromEdges(2, []ap.Edge[int]{{U: 2, V: 0}})
	if err == nil || !strings.Contains(err.Error(), "out of range") {
		t.Errorf("got %v", err)
	}
	_, _, err = FromEdges(2, []ap.Edge[int]{{U: 1, V: 1, Cost: 1}, {U: 1, V: 1, Cost: 2}})
	if err == nil || !strings.Contains(err.Error(), "duplicate") {
		t.Errorf("got %v", err)
	}
}

func TestFromAdjacencyRectangular(t *testing.T) {
	in := [][]bool{{true, false, true}, {false, false, false}}
	g, err := FromAdjacency(in)
	if err != nil {
		t.Fatal(err)
	}
	if g.NU != 2 || g.NV != 3 || !slices.Equal(g.Start, []int{0, 2, 2}) || !slices.Equal(g.Adj, []int{0, 2}) {
		t.Errorf("got %+v", g)
	}
	in[0][0] = false
	if !slices.Equal(g.Adj, []int{0, 2}) {
		t.Error("aliases caller memory")
	}
}

func TestFromMatrix(t *testing.T) {
	in := [][]int{{4, 5}, {6, 7}}
	g, c, err := FromMatrix(in)
	if err != nil {
		t.Fatal(err)
	}
	if g.NU != 2 || g.NV != 2 || !slices.Equal(g.Start, []int{0, 2, 4}) || !slices.Equal(g.Adj, []int{0, 1, 0, 1}) || !slices.Equal(c, []int{4, 5, 6, 7}) {
		t.Errorf("got %+v %v", g, c)
	}
	in[0][0] = 99
	if c[0] != 4 {
		t.Error("costs alias caller memory")
	}
	if _, _, err := FromMatrix([][]int{{1, 2}, {3}}); err == nil || !strings.Contains(err.Error(), "not square") {
		t.Errorf("got %v", err)
	}
}
