package main

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"
)

func runString(args []string, in string) (code int, out, errOut string) {
	var o, e bytes.Buffer
	code = run(args, strings.NewReader(in), &o, &e)
	return code, o.String(), e.String()
}

func TestDense(t *testing.T) {
	costs := [][]int{{90, 76, 75, 70}, {35, 85, 55, 65}, {125, 95, 90, 105}, {45, 110, 95, 115}}
	code, out, _ := runString(nil, "[[90,76,75,70],[35,85,55,65],[125,95,90,105],[45,110,95,115]]")
	if code != 0 {
		t.Fatalf("code %d", code)
	}
	var got struct {
		Status      string  `json:"status"`
		Cost        float64 `json:"cost"`
		Permutation []int   `json:"permutation"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got.Status != "optimal" || got.Cost != 90 {
		t.Fatalf("%q", out)
	}
	if len(got.Permutation) != 4 {
		t.Fatalf("%q", out)
	}
	seen := make([]bool, 4)
	worst := 0
	for u, v := range got.Permutation {
		if v < 0 || v >= 4 || seen[v] {
			t.Fatalf("invalid permutation %v", got.Permutation)
		}
		seen[v] = true
		worst = max(worst, costs[u][v])
	}
	if worst != 90 {
		t.Errorf("largest cost %d, want 90", worst)
	}
}

func TestSparseInfeasible(t *testing.T) {
	code, out, _ := runString([]string{"-cycles", "-matrix"}, `{"n":2,"edges":[[0,0,1],[1,0,1]]}`)
	if code != 0 || out != `{"status":"infeasible"}`+"\n" {
		t.Errorf("%d %q", code, out)
	}
}

func TestScale(t *testing.T) {
	_, out, _ := runString([]string{"-scale", "10"}, `[[1.5,2],[2,1.5]]`)
	if !strings.Contains(out, `"cost":1.5,`) {
		t.Errorf("%q", out)
	}
}

func TestInvalid(t *testing.T) {
	for _, in := range []string{`[[1,2]]`, `[[1],[2,3]]`, `{"n":0,"edges":[]}`, `nope`} {
		code, out, errOut := runString(nil, in)
		if code != 1 || out != "" || !strings.HasPrefix(errOut, "lbap: ") {
			t.Errorf("%q: %d %q %q", in, code, out, errOut)
		}
	}
}

func TestHelp(t *testing.T) {
	code, _, errOut := runString([]string{"-h"}, "")
	if code != 0 || !strings.Contains(errOut, "Usage") {
		t.Errorf("%d %q", code, errOut)
	}
}
