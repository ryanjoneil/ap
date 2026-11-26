package main

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/internal/cli"
	"github.com/ryanjoneil/ap/lbap"
)

const name = "lbap"

const usageText = `lbap solves linear bottleneck assignment problems, given a square cost
matrix. It minimizes the largest cost used in a perfect assignment.

Input is a dense cost matrix, a JSON array of arrays, or a sparse problem,
a JSON object {"n": N, "edges": [[u, v, cost], ...]}. Costs must be integers
unless -scale k is given, which multiplies each cost by k and rounds it to an
integer. The reported cost, the bottleneck value, is divided by k.
The output has a "status" of "optimal" or "infeasible".

Usage:

lbap < input.json -cycles > output.json

cat <<EOF | lbap | jq
[
	[  90,  76,  75,  70 ],
	[  35,  85,  55,  65 ],
	[ 125,  95,  90, 105 ],
	[  45, 110,  95, 115 ]
]
EOF

Flags:
`

type output struct {
	Status      string         `json:"status"`
	Cost        any            `json:"cost,omitempty"`
	Permutation ap.Permutation `json:"permutation,omitempty"`
	Cycles      ap.Cycles      `json:"cycles,omitempty"`
	Matrix      ap.Matrix      `json:"matrix,omitempty"`
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	cycles := fs.Bool("cycles", false, "output cyclic assignment form")
	matrix := fs.Bool("matrix", false, "output matrix assignment form")
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

	var a *lbap.LBAP[int64]
	if costs.Dense != nil {
		a, err = lbap.New(costs.Dense)
	} else {
		a, err = lbap.NewSparse(costs.N, costs.Edges)
	}
	if err != nil {
		return cli.Fail(stderr, name, err)
	}

	var out output
	p, err := a.Assign()
	switch {
	case errors.Is(err, ap.ErrInfeasible):
		out = output{Status: cli.StatusInfeasible}
	case err != nil:
		return cli.Fail(stderr, name, err)
	default:
		out = output{
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
	}
	if err := cli.Write(stdout, out); err != nil {
		return cli.Fail(stderr, name, err)
	}
	return 0
}
