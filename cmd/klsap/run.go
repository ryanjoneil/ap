package main

import (
	"errors"
	"flag"
	"fmt"
	"io"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/internal/cli"
	"github.com/ryanjoneil/ap/klsap"
)

const name = "klsap"

const usageText = `klsap ranks the k best assignments of a linear sum assignment problem, given
a square cost matrix, in nondecreasing order of cost.

Input is a dense cost matrix, a JSON array of arrays, or a sparse problem,
a JSON object {"n": N, "edges": [[u, v, cost], ...]}. Costs must be integers
unless -scale s is given, which multiplies each cost by s and rounds it to an
integer. Reported costs are divided by s. The flag -k is required.
The output has a "status" of "optimal" or "infeasible", and a list of
"assignments". If there are fewer than k assignments, all are listed.

Usage:

klsap -k 3 < input.json -cycles > output.json
echo '[[1.5,2],[2,1.5]]' | klsap -k 2 -scale 10

cat <<EOF | klsap -k 3 | jq
[
	[  90,  76,  75,  70 ],
	[  35,  85,  55,  65 ],
	[ 125,  95,  90, 105 ],
	[  45, 110,  95, 115 ]
]
EOF

Flags:
`

type assignment struct {
	Cost        any            `json:"cost"`
	Permutation ap.Permutation `json:"permutation"`
	Cycles      ap.Cycles      `json:"cycles,omitempty"`
	Matrix      ap.Matrix      `json:"matrix,omitempty"`
}

type output struct {
	Status      string       `json:"status"`
	Assignments []assignment `json:"assignments"`
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	k := fs.Int("k", 0, "number of assignments to rank (required, at least 1)")
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
	if *k < 1 {
		return cli.Fail(stderr, name, errors.New("-k must be at least 1"))
	}

	costs, err := cli.ReadCosts(stdin, *scale)
	if err != nil {
		return cli.Fail(stderr, name, err)
	}

	var r *klsap.KLSAP[int64]
	if costs.Dense != nil {
		r, err = klsap.New(costs.Dense, *k)
	} else {
		r, err = klsap.NewSparse(costs.N, costs.Edges, *k)
	}
	if err != nil {
		return cli.Fail(stderr, name, err)
	}

	out := output{Assignments: []assignment{}}
	for {
		p, err := r.Next()
		if errors.Is(err, ap.ErrInfeasible) {
			break // No further assignments.
		}
		if errors.Is(err, ap.ErrOverflow) {
			return cli.Fail(stderr, name, fmt.Errorf("%w; use a smaller -scale or smaller costs", err))
		}
		if err != nil {
			return cli.Fail(stderr, name, err)
		}
		a := assignment{Cost: cli.Unscale(r.Cost(), *scale), Permutation: p}
		if *cycles {
			a.Cycles = p.Cycles()
		}
		if *matrix {
			a.Matrix = p.Matrix()
		}
		out.Assignments = append(out.Assignments, a)
	}

	out.Status = cli.StatusOptimal
	if len(out.Assignments) == 0 {
		out.Status = cli.StatusInfeasible
	}
	if err := cli.Write(stdout, out); err != nil {
		return cli.Fail(stderr, name, err)
	}
	return 0
}
