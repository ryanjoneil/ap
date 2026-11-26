// Command timesolve times the lsap solver on one instance file, for
// comparison with external solvers by bench/compare.py.
//
// It reads the instance (cmd/lsap input format) from the path given as its
// only argument, then times lsap.New or lsap.NewSparse plus Assign, excluding
// JSON parsing. The timing includes the solver's internal copy of its input.
// It prints {"solver":"ap-lsap","seconds":S,"cost":C}.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/ryanjoneil/ap/internal/cli"
	"github.com/ryanjoneil/ap/lsap"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "timesolve:", err)
		os.Exit(1)
	}
}

func run() error {
	if len(os.Args) != 2 {
		return fmt.Errorf("usage: timesolve instance.json")
	}
	f, err := os.Open(os.Args[1])
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()
	costs, err := cli.ReadCosts(f, 1)
	if err != nil {
		return err
	}

	start := time.Now()
	var s *lsap.LSAP[int64]
	if costs.Dense != nil {
		s, err = lsap.New(costs.Dense)
	} else {
		s, err = lsap.NewSparse(costs.N, costs.Edges)
	}
	if err != nil {
		return err
	}
	if _, err = s.Assign(); err != nil {
		return err
	}
	elapsed := time.Since(start)

	return json.NewEncoder(os.Stdout).Encode(map[string]any{
		"solver": "ap-lsap", "seconds": elapsed.Seconds(), "cost": s.Cost(),
	})
}
