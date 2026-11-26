package cli

import (
	"bytes"
	"errors"
	"flag"
	"io"
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

func TestReadCostsIntegralFloat(t *testing.T) {
	for in, want := range map[string][][]int64{
		`[[5.0, 1e3]]`:                {{5, 1000}},
		`[[-2.0E2, 0.0]]`:             {{-200, 0}},
		`[[9007199254740992.0]]`:      {{1 << 53}},
		`[[-9.007199254740992e15]]`:   {{-(1 << 53)}},
		`{"n":1,"edges":[[0,0,7e0]]}`: nil,
	} {
		c, err := ReadCosts(strings.NewReader(in), 0)
		switch {
		case err != nil:
			t.Errorf("%s: %v", in, err)
		case want != nil && !reflect.DeepEqual(c.Dense, want):
			t.Errorf("%s: got %v; want %v", in, c.Dense, want)
		case want == nil && (len(c.Edges) != 1 || c.Edges[0].Cost != 7):
			t.Errorf("%s: got %v; want cost 7", in, c.Edges)
		}
	}
	// Integral, but too large to be sure of without -scale.
	for _, in := range []string{`[[1e300]]`, `[[1.8014398509481984e16]]`, `[[9007199254740993.0]]`, `[[1e400]]`} {
		if _, err := ReadCosts(strings.NewReader(in), 0); err == nil || !strings.Contains(err.Error(), "out of range") {
			t.Errorf("%s: got %v; want out of range error", in, err)
		}
	}
	// Not integral, though some round to an integer as a float64.
	for _, in := range []string{`[[1e-3]]`, `[[1.0000000000000000001]]`, `[[4503599627370497.5]]`, `[[1e-400]]`} {
		if _, err := ReadCosts(strings.NewReader(in), 0); err == nil || !strings.Contains(err.Error(), "-scale") {
			t.Errorf("%s: got %v; want -scale hint", in, err)
		}
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

func TestParseMessages(t *testing.T) {
	mk := func() (*flag.FlagSet, *bytes.Buffer, *int) {
		var buf bytes.Buffer
		calls := 0
		fs := flag.NewFlagSet("name", flag.ContinueOnError)
		fs.SetOutput(&buf)
		ScaleFlag(fs)
		fs.Usage = func() {
			calls++
			_, _ = io.WriteString(fs.Output(), "USAGE\n")
		}
		return fs, &buf, &calls
	}

	fs, buf, calls := mk()
	if c, e := Parse(fs, []string{"-bogus"}); c != 1 || !e {
		t.Errorf("-bogus: %d %v", c, e)
	}
	if want := "name: flag provided but not defined: -bogus\nUSAGE\n"; buf.String() != want || *calls != 1 {
		t.Errorf("-bogus: stderr %q, %d usage calls; want %q, 1 call", buf.String(), *calls, want)
	}

	fs, buf, calls = mk()
	if c, e := Parse(fs, []string{"-scale", "x"}); c != 1 || !e {
		t.Errorf("-scale x: %d %v", c, e)
	}
	if !strings.HasPrefix(buf.String(), "name: invalid value \"x\" for flag -scale") || *calls != 1 {
		t.Errorf("-scale x: stderr %q, %d usage calls", buf.String(), *calls)
	}

	fs, buf, calls = mk()
	if c, e := Parse(fs, []string{"-h"}); c != 0 || !e {
		t.Errorf("-h: %d %v", c, e)
	}
	if buf.String() != "USAGE\n" || *calls != 1 {
		t.Errorf("-h: stderr %q, %d usage calls; want usage once", buf.String(), *calls)
	}

	fs, buf, calls = mk()
	if c, e := Parse(fs, []string{"foo"}); c != 1 || !e {
		t.Errorf("foo: %d %v", c, e)
	}
	if want := "name: unexpected argument foo\nUSAGE\n"; buf.String() != want || *calls != 1 {
		t.Errorf("foo: stderr %q, %d usage calls; want %q, 1 call", buf.String(), *calls, want)
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
