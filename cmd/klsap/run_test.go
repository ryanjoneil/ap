package main

import (
	"bytes"
	"encoding/json"
	"os"
	"strings"
	"testing"
)

func runString(args []string, in string) (code int, out, errOut string) {
	var o, e bytes.Buffer
	code = run(args, strings.NewReader(in), &o, &e)
	return code, o.String(), e.String()
}

func TestRank(t *testing.T) {
	in, err := os.ReadFile("../lsap/data/4x4.json")
	if err != nil {
		t.Fatal(err)
	}
	code, out, errOut := runString([]string{"-k", "3"}, string(in))
	if code != 0 {
		t.Fatalf("%d %q", code, errOut)
	}
	var got struct {
		Status      string `json:"status"`
		Assignments []struct {
			Cost        float64 `json:"cost"`
			Permutation []int   `json:"permutation"`
		} `json:"assignments"`
	}
	if err := json.Unmarshal([]byte(out), &got); err != nil {
		t.Fatal(err)
	}
	if got.Status != "optimal" || len(got.Assignments) != 3 {
		t.Fatalf("%q", out)
	}
	first := got.Assignments[0]
	if first.Cost != 265 || len(first.Permutation) != 4 || first.Permutation[0] != 3 ||
		first.Permutation[1] != 2 || first.Permutation[2] != 1 || first.Permutation[3] != 0 {
		t.Errorf("first: %+v", first)
	}
	for i := 1; i < len(got.Assignments); i++ {
		if got.Assignments[i].Cost < got.Assignments[i-1].Cost {
			t.Errorf("costs decrease: %q", out)
		}
	}
}

func TestInfeasible(t *testing.T) {
	code, out, _ := runString([]string{"-k", "2"}, `{"n":2,"edges":[[0,0,1],[1,0,1]]}`)
	if code != 0 || out != `{"status":"infeasible","assignments":[]}`+"\n" {
		t.Errorf("%d %q", code, out)
	}
}

func TestFewer(t *testing.T) {
	code, out, _ := runString([]string{"-k", "5"}, `[[1]]`)
	want := `{"status":"optimal","assignments":[{"cost":1,"permutation":[0]}]}` + "\n"
	if code != 0 || out != want {
		t.Errorf("%d %q", code, out)
	}
}

func TestKRequired(t *testing.T) {
	for _, args := range [][]string{nil, {"-k", "0"}} {
		code, out, errOut := runString(args, `[[1]]`)
		if code != 1 || out != "" || errOut != "klsap: -k must be at least 1\n" {
			t.Errorf("%v: %d %q %q", args, code, out, errOut)
		}
	}
}

func TestScaleCyclesMatrix(t *testing.T) {
	code, out, _ := runString([]string{"-scale", "10", "-k", "2", "-cycles", "-matrix"}, `[[1.5,2],[2,1.5]]`)
	want := `{"status":"optimal","assignments":[` +
		`{"cost":3,"permutation":[0,1],"cycles":[[0],[1]],"matrix":[[true,false],[false,true]]},` +
		`{"cost":4,"permutation":[1,0],"cycles":[[0,1]],"matrix":[[false,true],[true,false]]}]}` + "\n"
	if code != 0 || out != want {
		t.Errorf("%d %q", code, out)
	}
}

func TestOverflow(t *testing.T) {
	code, out, errOut := runString([]string{"-k", "2"}, `[[4000000000000000000]]`)
	want := "klsap: ap: arithmetic overflow; use a smaller -scale or smaller costs\n"
	if code != 1 || out != "" || errOut != want {
		t.Errorf("%d %q %q", code, out, errOut)
	}
}

func TestBadInput(t *testing.T) {
	code, out, errOut := runString([]string{"-k", "1"}, `nope`)
	if code != 1 || out != "" || !strings.HasPrefix(errOut, "klsap: ") {
		t.Errorf("%d %q %q", code, out, errOut)
	}
}

func TestHelp(t *testing.T) {
	code, _, errOut := runString([]string{"-h"}, "")
	if code != 0 || !strings.Contains(errOut, "Usage") {
		t.Errorf("%d %q", code, errOut)
	}
}

// The root solves, but the child that forbids an edge raises the big-M past
// the overflow margin. The CLI discards the assignment already found: it
// fails with exit 1, the overflow message on stderr and nothing on stdout.
func TestOverflowAfterResults(t *testing.T) {
	const in = `[[0,3000000000000000000],[3000000000000000000,0]]`
	if code, _, errOut := runString([]string{"-k", "1"}, in); code != 0 {
		t.Fatalf("the root should solve: %d %q", code, errOut)
	}
	code, out, errOut := runString([]string{"-k", "2"}, in)
	want := "klsap: ap: arithmetic overflow; use a smaller -scale or smaller costs\n"
	if code != 1 || out != "" || errOut != want {
		t.Errorf("%d %q %q", code, out, errOut)
	}
}
