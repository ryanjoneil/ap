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

const (
	denseIn = `[[true,false],[true,true]]`
	rectIn  = `{"nu":2,"nv":3,"edges":[[0,2],[1,2]]}`
)

func TestDense(t *testing.T) {
	code, out, _ := runString(nil, denseIn)
	if code != 0 || out != `{"status":"optimal","size":2,"perfect":true,"matching":[0,1]}`+"\n" {
		t.Errorf("%d %q", code, out)
	}
}

func TestRectangular(t *testing.T) {
	code, out, _ := runString(nil, rectIn)
	if code != 0 || out != `{"status":"optimal","size":1,"perfect":false,"matching":[2,-1]}`+"\n" {
		t.Errorf("%d %q", code, out)
	}
}

func TestCyclesOnlyWhenPerfect(t *testing.T) {
	code, out, _ := runString([]string{"-cycles"}, denseIn)
	if code != 0 || !strings.Contains(out, `"cycles":[[0],[1]]`) {
		t.Errorf("%d %q", code, out)
	}
	code, out, _ = runString([]string{"-cycles", "-matrix"}, rectIn)
	if code != 0 || strings.Contains(out, `"cycles"`) || strings.Contains(out, `"matrix"`) {
		t.Errorf("%d %q", code, out)
	}
}

func TestMatrixWhenPerfect(t *testing.T) {
	code, out, _ := runString([]string{"-matrix"}, denseIn)
	if code != 0 || !strings.Contains(out, `"matrix":[[true,false],[false,true]]`) {
		t.Errorf("%d %q", code, out)
	}
}

func TestSquareNotPerfect(t *testing.T) {
	code, out, _ := runString([]string{"-cycles"}, `[[true,false],[true,false]]`)
	if code != 0 || !strings.Contains(out, `"size":1,"perfect":false`) || strings.Contains(out, `"cycles"`) {
		t.Errorf("%d %q", code, out)
	}
}

func TestInvalid(t *testing.T) {
	code, out, errOut := runString(nil, `[[1,0]]`)
	if code != 1 || out != "" || !strings.HasPrefix(errOut, "matching: ") {
		t.Errorf("%d %q %q", code, out, errOut)
	}
}

func TestNoScaleFlag(t *testing.T) {
	code, out, _ := runString([]string{"-scale", "10"}, denseIn)
	if code != 1 || out != "" {
		t.Errorf("%d %q", code, out)
	}
}

func TestHelp(t *testing.T) {
	code, _, errOut := runString([]string{"-h"}, "")
	if code != 0 || !strings.Contains(errOut, "Usage") {
		t.Errorf("%d %q", code, errOut)
	}
}
