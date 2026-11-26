// Command geninstance writes one benchmark instance to standard output in the
// input format of cmd/lsap: a JSON array of arrays for a dense instance, or
// {"n": N, "edges": [[u, v, cost], ...]} for a sparse one. Instances come from
// the classes of book Section 4.10.1, built by package internal/instances.
//
// Usage:
//
//	geninstance -class uniform|geometric|twocost|macholwien -n N
//	            [-degree D] [-k K] [-seed S]
//
// A degree of 0 (the default) writes a dense instance.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand/v2"
	"os"
	"strconv"

	"github.com/ryanjoneil/ap/internal/instances"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "geninstance:", err)
		os.Exit(1)
	}
}

func run() error {
	class := flag.String("class", "uniform", "uniform|geometric|twocost|macholwien")
	n := flag.Int("n", 0, "number of rows and columns")
	degree := flag.Int("degree", 0, "edges per row; 0 means dense")
	k := flag.Int64("k", 1000000, "cost range K (uniform, geometric)")
	seed := flag.Uint64("seed", 1, "random seed")
	flag.Parse()

	if *n <= 0 {
		return fmt.Errorf("-n must be positive")
	}
	rng := rand.New(rand.NewPCG(*seed, *seed))
	var c instances.Class
	switch *class {
	case "uniform":
		c = instances.Uniform(rng, *k)
	case "geometric":
		c = instances.Geometric(rng, *n, *k)
	case "twocost":
		c = instances.TwoCost(rng)
	case "macholwien":
		c = instances.MacholWien()
	default:
		return fmt.Errorf("unknown -class %q", *class)
	}

	// bufio.Writer errors are sticky and reported by Flush below.
	w := bufio.NewWriterSize(os.Stdout, 1<<20)
	var buf []byte
	if *degree == 0 {
		_ = w.WriteByte('[')
		for i, row := range instances.Dense(*n, c) {
			if i > 0 {
				_ = w.WriteByte(',')
			}
			_ = w.WriteByte('[')
			for j, x := range row {
				if j > 0 {
					_ = w.WriteByte(',')
				}
				buf = strconv.AppendInt(buf[:0], x, 10)
				_, _ = w.Write(buf)
			}
			_ = w.WriteByte(']')
		}
		_, _ = w.WriteString("]\n")
	} else {
		_, _ = fmt.Fprintf(w, `{"n":%d,"edges":[`, *n)
		for i, e := range instances.Sparse(rng, *n, *degree, c) {
			if i > 0 {
				_ = w.WriteByte(',')
			}
			_, _ = fmt.Fprintf(w, "[%d,%d,%d]", e.U, e.V, e.Cost)
		}
		_, _ = w.WriteString("]}\n")
	}
	return w.Flush()
}
