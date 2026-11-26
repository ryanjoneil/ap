package main

import (
	"flag"
	"fmt"
	"io"

	"github.com/ryanjoneil/ap"
	"github.com/ryanjoneil/ap/internal/cli"
	"github.com/ryanjoneil/ap/mcm"
)

const name = "mcm"

const usageText = `mcm finds a maximum cardinality matching (MCM) in a bipartite graph.

Input is a dense boolean adjacency matrix, a JSON array of arrays with one row
per vertex in U and one column per vertex in V, or a sparse graph, a JSON
object {"nu": NU, "nv": NV, "edges": [[u, v], ...]}. The two sides may differ
in size. The output has a "status" of "optimal", the "size" of the matching,
whether it is "perfect" (both sides the same size and fully matched), and
"matching", where entry u is the vertex in V matched to u, or -1.
-cycles and -matrix add the permutation forms, only for a perfect matching.

Usage:

mcm < input.json -cycles > output.json

cat <<EOF | mcm | jq
[
	[ true,  false, true  ],
	[ true,  true,  false ],
	[ false, true,  false ]
]
EOF

Flags:
`

type output struct {
	Status   string    `json:"status"`
	Size     int       `json:"size"`
	Perfect  bool      `json:"perfect"`
	Matching []int     `json:"matching"`
	Cycles   ap.Cycles `json:"cycles,omitempty"`
	Matrix   ap.Matrix `json:"matrix,omitempty"`
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(stderr)
	cycles := fs.Bool("cycles", false, "output cyclic assignment form (perfect matchings only)")
	matrix := fs.Bool("matrix", false, "output matrix assignment form (perfect matchings only)")
	fs.Usage = func() {
		_, _ = fmt.Fprint(fs.Output(), usageText)
		fs.PrintDefaults()
	}
	if code, exit := cli.Parse(fs, args); exit {
		return code
	}

	pairs, err := cli.ReadPairs(stdin)
	if err != nil {
		return cli.Fail(stderr, name, err)
	}

	var m *mcm.MCM
	if pairs.Dense != nil {
		m, err = mcm.New(pairs.Dense)
	} else {
		m, err = mcm.NewSparse(pairs.NU, pairs.NV, pairs.Edges)
	}
	if err != nil {
		return cli.Fail(stderr, name, err)
	}

	match := m.Match()
	out := output{
		Status:   cli.StatusOptimal,
		Size:     m.Size(),
		Perfect:  pairs.NU == pairs.NV && m.Size() == pairs.NU,
		Matching: match,
	}
	if out.Perfect {
		p := ap.Permutation(match)
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
