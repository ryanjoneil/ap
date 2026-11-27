package main

import (
	"bytes"
	"strings"
	"testing"
)

func runString(args []string, in string) (code int, out, errOut string) {
	var o, e bytes.Buffer
	code = run(args, strings.NewReader(in), &o, &e)
	return code, o.String(), e.String()
}

func TestDense(t *testing.T) {
	code, out, _ := runString(nil, "[[90,76,75,70],[35,85,55,65],[125,95,90,105],[45,110,95,115]]")
	if code != 0 || out != `{"status":"optimal","cost":265,"permutation":[3,2,1,0]}`+"\n" {
		t.Errorf("%d %q", code, out)
	}
}

func TestSparseInfeasible(t *testing.T) {
	code, out, _ := runString([]string{"-dual", "-rc"}, `{"n":2,"edges":[[0,0,1],[1,0,1]]}`)
	if code != 0 || out != `{"status":"infeasible"}`+"\n" {
		t.Errorf("%d %q", code, out)
	}
}

func TestScale(t *testing.T) {
	_, out, _ := runString([]string{"-scale", "10"}, `[[1.5,2],[2,1.5]]`)
	if !strings.Contains(out, `"cost":3,`) {
		t.Errorf("%q", out)
	}
}

func TestSparseRC(t *testing.T) {
	// Edges are given out of order to check the CSR (row, column) ordering.
	code, out, _ := runString([]string{"-rc"}, `{"n":2,"edges":[[1,1,1],[0,1,2],[0,0,1]]}`)
	if code != 0 || !strings.Contains(out, `"rc":[[0,0,0],[0,1,1],[1,1,0]]`) {
		t.Errorf("%d %q", code, out)
	}
}

func TestDual(t *testing.T) {
	code, out, _ := runString([]string{"-dual"}, `[[10,15,12],[51,75,23],[11,91,10]]`)
	if code != 0 || !strings.Contains(out, `"dual":{"u":[0,14,1],"v":[10,15,9]}`) {
		t.Errorf("%d %q", code, out)
	}
}

func TestInvalid(t *testing.T) {
	for _, in := range []string{`[[1,2]]`, `[[1],[2,3]]`, `{"n":0,"edges":[]}`, `nope`} {
		code, out, errOut := runString(nil, in)
		if code != 1 || out != "" || !strings.HasPrefix(errOut, "lsap: ") {
			t.Errorf("%q: %d %q %q", in, code, out, errOut)
		}
	}
}

func TestOverflow(t *testing.T) {
	code, out, errOut := runString(nil, `[[4000000000000000000]]`)
	want := "lsap: ap: arithmetic overflow; use a smaller -scale or smaller costs\n"
	if code != 1 || out != "" || errOut != want {
		t.Errorf("%d %q %q", code, out, errOut)
	}
}

func TestHelp(t *testing.T) {
	code, _, errOut := runString([]string{"-h"}, "")
	if code != 0 || !strings.Contains(errOut, "Usage") {
		t.Errorf("%d %q", code, errOut)
	}
}

func TestBadFlag(t *testing.T) {
	code, out, errOut := runString([]string{"-bogus"}, "")
	if code != 1 || out != "" || !strings.HasPrefix(errOut, "lsap: flag provided but not defined: -bogus\n") ||
		strings.Count(errOut, "Usage") != 1 {
		t.Errorf("%d %q %q", code, out, errOut)
	}
}
