package cli

import (
	"bytes"
	"errors"
	"flag"
	"math"
	"reflect"
	"strings"
	"testing"

	"github.com/ryanjoneil/ap"
)

func TestReadCosts(t *testing.T) {
	c, err := ReadCosts(strings.NewReader(`[[1,2],[3,4]]`), 0)
	if err != nil || c.N != 2 || !reflect.DeepEqual(c.Dense, [][]int64{{1, 2}, {3, 4}}) {
		t.Errorf("dense: %+v %v", c, err)
	}
	c, err = ReadCosts(strings.NewReader(" \n{\"n\":3,\"edges\":[[0,1,15],[2,0,11]]} \n"), 0)
	if err != nil || c.N != 3 || !reflect.DeepEqual(c.Edges, []ap.Edge[int64]{{U: 0, V: 1, Cost: 15}, {U: 2, V: 0, Cost: 11}}) {
		t.Errorf("sparse: %+v %v", c, err)
	}
	c, err = ReadCosts(strings.NewReader(`[[1.25, 0.005]]`), 100)
	if err != nil || !reflect.DeepEqual(c.Dense, [][]int64{{125, 1}}) {
		t.Errorf("scale: %v %v", c.Dense, err)
	}
}

func TestReadCostsFloatNeedsScale(t *testing.T) {
	_, err := ReadCosts(strings.NewReader(`[[1.5]]`), 0)
	if err == nil || !strings.Contains(err.Error(), "-scale") {
		t.Errorf("got %v", err)
	}
}

func TestReadCostsErrors(t *testing.T) {
	for _, tc := range []struct {
		in    string
		scale float64
	}{
		{``, 0}, {`  `, 0}, {`5`, 0}, {`"x"`, 0}, {`[[1]] [[2]]`, 0}, {`[[1]] x`, 0},
		{`{"n":1,"edges":[[0,0]]}`, 0}, {`{"n":1,"edges":[[0,0,1,2]]}`, 0},
		{`{"n":1,"edges":[],"x":1}`, 0}, {`{"n":1.5,"edges":[]}`, 0}, {`{"n":1,"edges":[[0.5,0,1]]}`, 0},
		{`[[99999999999999999999]]`, 0}, {`[[1e30]]`, 1}, {`[[1]]`, -1}, {`[[1]]`, math.NaN()}, {`[[1]]`, math.Inf(1)},
		{`[["a"]]`, 0}, {`[1]`, 0}, {`[[1]`, 0}, {`{"n":1,"edges":[[0,0,"x"]]}`, 0},
		{`{"n":99999999999999999999,"edges":[]}`, 0}, {`{"n":1,"edges":[[0,0,1]]} {}`, 0},
	} {
		if _, err := ReadCosts(strings.NewReader(tc.in), tc.scale); err == nil {
			t.Errorf("%q scale %v: want error", tc.in, tc.scale)
		}
	}
}

func TestReadCostsTrailingMessage(t *testing.T) {
	_, err := ReadCosts(strings.NewReader(`[[1]] [[2]]`), 0)
	if err == nil || !strings.Contains(err.Error(), "trailing data") {
		t.Errorf("got %v", err)
	}
}

func TestReadPairs(t *testing.T) {
	p, err := ReadPairs(strings.NewReader(`[[true,false],[true,true]]`))
	if err != nil || p.NU != 2 || p.NV != 2 || !reflect.DeepEqual(p.Dense, [][]bool{{true, false}, {true, true}}) {
		t.Errorf("dense: %+v %v", p, err)
	}
	p, err = ReadPairs(strings.NewReader(`{"nu":2,"nv":3,"edges":[[0,2],[1,0]]}`))
	if err != nil || p.NU != 2 || p.NV != 3 || !reflect.DeepEqual(p.Edges, []ap.Pair{{U: 0, V: 2}, {U: 1, V: 0}}) {
		t.Errorf("sparse: %+v %v", p, err)
	}
}

func TestReadPairsErrors(t *testing.T) {
	for _, in := range []string{``, `[[1,0]]`, `{"nu":1,"nv":1,"edges":[[0]]}`, `{"nu":1,"nv":1,"edges":[[0,0,1]]}`,
		`[[true]] [[true]]`, `{"n":1,"edges":[]}`, `7`} {
		if _, err := ReadPairs(strings.NewReader(in)); err == nil {
			t.Errorf("%q: want error", in)
		}
	}
}

func TestUnscale(t *testing.T) {
	if v := Unscale(125, 0); v != int64(125) {
		t.Errorf("got %#v", v)
	}
	if v := Unscale(125, 100); v != 1.25 {
		t.Errorf("got %#v", v)
	}
}

func TestParse(t *testing.T) {
	mk := func() (*flag.FlagSet, *float64, *bytes.Buffer) {
		var buf bytes.Buffer
		fs := flag.NewFlagSet("x", flag.ContinueOnError)
		fs.SetOutput(&buf)
		return fs, ScaleFlag(fs), &buf
	}
	fs, _, _ := mk()
	if c, e := Parse(fs, []string{"-h"}); c != 0 || !e {
		t.Errorf("-h: %d %v", c, e)
	}
	fs, _, _ = mk()
	if c, e := Parse(fs, []string{"-bogus"}); c != 1 || !e {
		t.Errorf("-bogus: %d %v", c, e)
	}
	fs, s, _ := mk()
	if c, e := Parse(fs, []string{"-scale", "10"}); c != 0 || e || *s != 10 {
		t.Errorf("scale: %d %v %v", c, e, *s)
	}
	fs, _, buf := mk()
	if c, e := Parse(fs, []string{"foo"}); c != 1 || !e || !strings.Contains(buf.String(), "x: unexpected argument foo") {
		t.Errorf("positional: %d %v %q", c, e, buf.String())
	}
}

func TestFail(t *testing.T) {
	var b bytes.Buffer
	if c := Fail(&b, "lsap", errors.New("bad")); c != 1 || b.String() != "lsap: bad\n" {
		t.Errorf("%d %q", c, b.String())
	}
	b.Reset()
	Fail(&b, "lsap", errors.New("lsap: bad"))
	if b.String() != "lsap: bad\n" {
		t.Errorf("%q", b.String())
	}
}

func TestWrite(t *testing.T) {
	var b bytes.Buffer
	if err := Write(&b, map[string]any{"status": StatusOptimal, "s": "<&>"}); err != nil || b.String() != "{\"s\":\"<&>\",\"status\":\"optimal\"}\n" {
		t.Errorf("%v %q", err, b.String())
	}
}
