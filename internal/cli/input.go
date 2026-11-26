package cli

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math"
	"strconv"
	"strings"

	"github.com/ryanjoneil/ap"
)

// Costs is a parsed cost matrix, dense or sparse.
type Costs struct {
	N     int              // number of rows (and columns)
	Dense [][]int64        // set for dense input (N = len(Dense)); not checked for squareness
	Edges []ap.Edge[int64] // set for sparse input
}

// Pairs is a parsed bipartite graph, dense or sparse.
type Pairs struct {
	NU, NV int
	Dense  [][]bool  // set for dense input; not checked for rectangularity
	Edges  []ap.Pair // set for sparse input
}

// number is a JSON number kept as its literal text. Unlike json.Number it
// rejects quoted strings.
type number string

func (n *number) UnmarshalJSON(b []byte) error {
	if len(b) == 0 || b[0] == '"' || b[0] == 'n' { // string or null
		return errors.New("expected a number")
	}
	*n = number(b)
	return nil
}

type sparseCosts struct {
	N     number     `json:"n"`
	Edges [][]number `json:"edges"`
}

type sparsePairs struct {
	NU    number     `json:"nu"`
	NV    number     `json:"nv"`
	Edges [][]number `json:"edges"`
}

// ReadCosts reads one JSON cost matrix from r: a top-level array is dense, a
// top-level object {"n": N, "edges": [[u, v, cost], ...]} is sparse. If scale
// is 0, costs must be integers: either integer literals in int64 range, or
// literals with a fraction or exponent, such as 5.0 or 1e3, whose value is an
// integer of magnitude at most 2^53. Otherwise each cost becomes
// math.Round(cost*scale). Scale must not be negative, NaN or infinite.
func ReadCosts(r io.Reader, scale float64) (Costs, error) {
	if math.IsNaN(scale) || math.IsInf(scale, 0) || scale < 0 {
		return Costs{}, fmt.Errorf("invalid -scale %v: must be a positive finite number", scale)
	}
	dec, first, err := open(r)
	if err != nil {
		return Costs{}, err
	}
	switch first {
	case '[':
		var rows [][]number
		if err := decode(dec, &rows, false); err != nil {
			return Costs{}, err
		}
		dense := make([][]int64, len(rows))
		for i, row := range rows {
			out := make([]int64, len(row))
			for j, s := range row {
				if out[j], err = cost(s, scale); err != nil {
					return Costs{}, fmt.Errorf("row %d, column %d: %w", i, j, err)
				}
			}
			dense[i] = out
		}
		return Costs{N: len(dense), Dense: dense}, nil
	case '{':
		var sp sparseCosts
		if err := decode(dec, &sp, true); err != nil {
			return Costs{}, err
		}
		n, err := integer(sp.N)
		if err != nil {
			return Costs{}, fmt.Errorf("n: %w", err)
		}
		edges := make([]ap.Edge[int64], len(sp.Edges))
		for i, e := range sp.Edges {
			if len(e) != 3 {
				return Costs{}, fmt.Errorf("edge %d: want 3 entries [u, v, cost], got %d", i, len(e))
			}
			u, err := integer(e[0])
			if err != nil {
				return Costs{}, fmt.Errorf("edge %d: u: %w", i, err)
			}
			v, err := integer(e[1])
			if err != nil {
				return Costs{}, fmt.Errorf("edge %d: v: %w", i, err)
			}
			c, err := cost(e[2], scale)
			if err != nil {
				return Costs{}, fmt.Errorf("edge %d: cost: %w", i, err)
			}
			edges[i] = ap.Edge[int64]{U: u, V: v, Cost: c}
		}
		return Costs{N: n, Edges: edges}, nil
	}
	return Costs{}, errors.New("input must be a JSON array (dense) or object (sparse)")
}

// ReadPairs reads one JSON bipartite graph from r: a top-level array of
// boolean rows is dense, a top-level object
// {"nu": NU, "nv": NV, "edges": [[u, v], ...]} is sparse.
func ReadPairs(r io.Reader) (Pairs, error) {
	dec, first, err := open(r)
	if err != nil {
		return Pairs{}, err
	}
	switch first {
	case '[':
		var rows [][]bool
		if err := decode(dec, &rows, false); err != nil {
			return Pairs{}, err
		}
		nv := 0
		if len(rows) > 0 {
			nv = len(rows[0])
		}
		return Pairs{NU: len(rows), NV: nv, Dense: rows}, nil
	case '{':
		var sp sparsePairs
		if err := decode(dec, &sp, true); err != nil {
			return Pairs{}, err
		}
		nu, err := integer(sp.NU)
		if err != nil {
			return Pairs{}, fmt.Errorf("nu: %w", err)
		}
		nv, err := integer(sp.NV)
		if err != nil {
			return Pairs{}, fmt.Errorf("nv: %w", err)
		}
		edges := make([]ap.Pair, len(sp.Edges))
		for i, e := range sp.Edges {
			if len(e) != 2 {
				return Pairs{}, fmt.Errorf("edge %d: want 2 entries [u, v], got %d", i, len(e))
			}
			u, err := integer(e[0])
			if err != nil {
				return Pairs{}, fmt.Errorf("edge %d: u: %w", i, err)
			}
			v, err := integer(e[1])
			if err != nil {
				return Pairs{}, fmt.Errorf("edge %d: v: %w", i, err)
			}
			edges[i] = ap.Pair{U: u, V: v}
		}
		return Pairs{NU: nu, NV: nv, Edges: edges}, nil
	}
	return Pairs{}, errors.New("input must be a JSON array (dense) or object (sparse)")
}

// open peeks at the first non-whitespace byte of r and returns a decoder
// positioned at the start of the value.
func open(r io.Reader) (*json.Decoder, byte, error) {
	br := bufio.NewReader(r)
	for {
		b, err := br.ReadByte()
		if err == io.EOF {
			return nil, 0, errors.New("empty input")
		}
		if err != nil {
			return nil, 0, err
		}
		if b == ' ' || b == '\t' || b == '\r' || b == '\n' {
			continue
		}
		if err := br.UnreadByte(); err != nil {
			return nil, 0, err
		}
		dec := json.NewDecoder(br)
		dec.UseNumber()
		return dec, b, nil
	}
}

// decode decodes one value into v and requires that nothing but whitespace
// follows it.
func decode(dec *json.Decoder, v any, strict bool) error {
	if strict {
		dec.DisallowUnknownFields()
	}
	if err := dec.Decode(v); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}
	if _, err := dec.Token(); err != io.EOF {
		return errors.New("invalid JSON: trailing data after the input value")
	}
	return nil
}

// integer parses a JSON number that must be an integer fitting in an int.
func integer(s number) (int, error) {
	if s == "" {
		return 0, errors.New("missing")
	}
	v, err := strconv.ParseInt(string(s), 10, strconv.IntSize)
	if err != nil {
		return 0, fmt.Errorf("%q is not an integer in range", string(s))
	}
	return int(v), nil
}

// cost converts a JSON number to an int64 cost, applying scale if nonzero.
func cost(s number, scale float64) (int64, error) {
	if scale == 0 {
		v, err := strconv.ParseInt(string(s), 10, 64)
		if err == nil {
			return v, nil
		}
		if errors.Is(err, strconv.ErrRange) {
			return 0, fmt.Errorf("cost %s is out of int64 range", string(s))
		}
		if _, ferr := strconv.ParseFloat(string(s), 64); ferr != nil && !errors.Is(ferr, strconv.ErrRange) {
			return 0, fmt.Errorf("invalid cost %s", string(s))
		}
		// A fraction or exponent, as in 5.0 or 1e3.
		switch v, integral, ok := integerLiteral(string(s)); {
		case ok:
			return v, nil
		case integral:
			return 0, fmt.Errorf("cost %s is out of range: a cost with a fraction or exponent must be at most 2^53 in magnitude", string(s))
		}
		return 0, fmt.Errorf("cost %s is not an integer: use -scale to accept non-integer costs", string(s))
	}
	f, err := strconv.ParseFloat(string(s), 64)
	if err != nil {
		return 0, fmt.Errorf("invalid cost %s", string(s))
	}
	x := math.Round(f * scale)
	// -2^63 and 2^63 are exact in float64; the upper bound is exclusive.
	if !(x >= -(1<<63) && x < 1<<63) {
		return 0, fmt.Errorf("cost %s scaled by %v is out of int64 range", string(s), scale)
	}
	return int64(x), nil
}

// integerLiteral decides from its decimal digits whether the JSON number
// literal s, which may have a fraction or exponent, is an integer. It works
// on the text rather than on a float64, which would round a literal such as
// 1.0000000000000000001 to an integer. It returns the integer's value and
// ok = true if |value| ≤ 2^53, and integral = true for any integer.
func integerLiteral(s string) (v int64, integral, ok bool) {
	neg := strings.HasPrefix(s, "-")
	mant, exp := strings.TrimPrefix(s, "-"), 0
	if i := strings.IndexAny(mant, "eE"); i >= 0 {
		e, err := strconv.Atoi(mant[i+1:])
		switch {
		case err == nil:
			exp = e
		case strings.HasPrefix(mant[i+1:], "-"):
			exp = -(1 << 30) // Too small for any digits to make an integer.
		default:
			exp = 1 << 30 // Too large for any nonzero digits to fit.
		}
		mant = mant[:i]
	}
	digits := mant
	if i := strings.IndexByte(mant, '.'); i >= 0 {
		digits = mant[:i] + mant[i+1:]
		exp -= len(mant) - i - 1
	}

	// The value is digits·10^exp. Drop leading and trailing zeros.
	digits = strings.TrimLeft(digits, "0")
	if digits == "" {
		return 0, true, true
	}
	trimmed := strings.TrimRight(digits, "0")
	exp += len(digits) - len(trimmed)
	digits = trimmed

	switch {
	case exp < 0:
		return 0, false, false
	case len(digits)+exp > 16: // At least 10^16 > 2^53.
		return 0, true, false
	}
	v, err := strconv.ParseInt(digits+strings.Repeat("0", exp), 10, 64)
	if err != nil || v > 1<<53 {
		return 0, true, false
	}
	if neg {
		v = -v
	}
	return v, true, true
}
