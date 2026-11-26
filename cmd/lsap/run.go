package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"slices"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/internal/cli"
	"github.com/ryanjoneil/ap/lsap"
)

const name = "lsap"

const usageText = `lsap solves linear sum assignment problems, given a square cost matrix.

Input is a dense cost matrix, a JSON array of arrays, or a sparse problem,
a JSON object {"n": N, "edges": [[u, v, cost], ...]}. Costs must be integers
unless -scale k is given, which multiplies each cost by k and rounds it to an
integer. Reported costs, dual prices and reduced costs are divided by k.
The output has a "status" of "optimal" or "infeasible".

Usage:

lsap < input.json -dual -rc > output.json
echo '[[1.5,2],[2,1.5]]' | lsap -scale 10

cat <<EOF | lsap | jq
[
	[  90,  76,  75,  70 ],
	[  35,  85,  55,  65 ],
	[ 125,  95,  90, 105 ],
	[  45, 110,  95, 115 ]
]
EOF

Flags:
`

type dualOut struct {
	U []any `json:"u"`
	V []any `json:"v"`
}

type output struct {
	Status      string         `json:"status"`
	Cost        any            `json:"cost,omitempty"`
	Permutation ap.Permutation `json:"permutation,omitempty"`
	Cycles      ap.Cycles      `json:"cycles,omitempty"`
	Matrix      ap.Matrix      `json:"matrix,omitempty"`
	Dual        *dualOut       `json:"dual,omitempty"`
	RC          any            `json:"rc,omitempty"`
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	cycles := fs.Bool("cycles", false, "output cyclic assignment form")
	dual := fs.Bool("dual", false, "output dual prices")
	matrix := fs.Bool("matrix", false, "output matrix assignment form")
	rc := fs.Bool("rc", false, "output reduced cost matrix")
	scale := cli.ScaleFlag(fs)
	fs.Usage = func() {
		_, _ = fmt.Fprint(fs.Output(), usageText)
		fs.PrintDefaults()
	}
	if code, exit := cli.Parse(fs, args); exit {
		return code
	}

	costs, err := cli.ReadCosts(stdin, *scale)
	if err != nil {
		return cli.Fail(stderr, name, err)
	}

	var a *lsap.LSAP[int64]
	if costs.Dense != nil {
		a, err = lsap.New(costs.Dense)
	} else {
		a, err = lsap.NewSparse(costs.N, costs.Edges)
	}
	if err != nil {
		return cli.Fail(stderr, name, err)
	}

	p, err := a.Assign()
	if errors.Is(err, ap.ErrInfeasible) {
		return write(stdout, stderr, output{Status: cli.StatusInfeasible})
	}
	if err != nil {
		return cli.Fail(stderr, name, err)
	}

	out := output{
		Status:      cli.StatusOptimal,
		Cost:        cli.Unscale(a.Cost(), *scale),
		Permutation: p,
	}
	if *cycles {
		out.Cycles = p.Cycles()
	}
	if *matrix {
		out.Matrix = p.Matrix()
	}
	if *dual {
		d := a.DualPrices()
		out.Dual = &dualOut{U: unscaleAll(d.U, *scale), V: unscaleAll(d.V, *scale)}
	}
	if *rc {
		if costs.Dense != nil {
			out.RC = denseRC(a, len(costs.Dense), *scale)
		} else {
			out.RC = sparseRC(a, costs.Edges, *scale)
		}
	}
	return write(stdout, stderr, out)
}

func write(stdout, stderr io.Writer, v any) int {
	if err := cli.Write(stdout, v); err != nil {
		return cli.Fail(stderr, name, err)
	}
	return 0
}

func unscaleAll(xs []int64, scale float64) []any {
	out := make([]any, len(xs))
	for i, x := range xs {
		out[i] = cli.Unscale(x, scale)
	}
	return out
}

func denseRC(a *lsap.LSAP[int64], n int, scale float64) [][]any {
	m := make([][]any, n)
	for u := range m {
		m[u] = make([]any, n)
		for v := range m[u] {
			r, _ := a.ReducedCost(u, v)
			m[u][v] = cli.Unscale(r, scale)
		}
	}
	return m
}

// sparseRC lists [u, v, reduced cost] for every edge in CSR order (row, then
// column). The solver's input order is untouched; a sorted copy is used.
func sparseRC(a *lsap.LSAP[int64], edges []ap.Edge[int64], scale float64) [][3]any {
	sorted := slices.Clone(edges)
	slices.SortStableFunc(sorted, func(x, y ap.Edge[int64]) int {
		if x.U != y.U {
			return x.U - y.U
		}
		return x.V - y.V
	})
	out := make([][3]any, len(sorted))
	for i, e := range sorted {
		r, _ := a.ReducedCost(e.U, e.V)
		out[i] = [3]any{e.U, e.V, cli.Unscale(r, scale)}
	}
	return out
}
