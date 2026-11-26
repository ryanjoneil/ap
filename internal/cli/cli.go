// Package cli holds the plumbing shared by the command-line binaries: JSON
// input formats, the -scale flag, flag parsing and exit codes, and JSON
// output.
//
// It validates JSON shape, integer-ness, int64 range and scale. It does not
// check squareness or whether edge endpoints are in range; the solvers'
// constructors do that.
package cli

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"strings"
)

// Status values reported in the "status" field of CLI output.
const (
	// StatusOptimal means an optimal assignment was found.
	StatusOptimal = "optimal"
	// StatusInfeasible means no feasible assignment exists.
	StatusInfeasible = "infeasible"
)

// Unscale converts a scaled integer cost back to its reported value. It
// returns x as an int64 if scale is 0 (no -scale given), and float64(x)/scale
// otherwise.
func Unscale(x int64, scale float64) any {
	if scale == 0 {
		return x
	}
	return float64(x) / scale
}

// ScaleFlag registers the -scale flag on fs and returns a pointer to its
// value. A value of 0, the default, means no scaling.
func ScaleFlag(fs *flag.FlagSet) *float64 {
	return fs.Float64("scale", 0, "multiply costs by `k` and round to integers (accepts non-integer costs)")
}

// Parse parses args with fs. The caller creates fs with flag.ContinueOnError
// and sets its output; Parse does not change it. It returns (0, true) if help
// was requested, (1, true) on a bad flag or an unexpected positional
// argument, and (0, false) on success, in which case the program continues.
func Parse(fs *flag.FlagSet, args []string) (code int, exit bool) {
	if err := fs.Parse(args); err != nil {
		if err == flag.ErrHelp {
			return 0, true
		}
		return 1, true
	}
	if fs.NArg() > 0 {
		_, _ = fmt.Fprintf(fs.Output(), "%s: unexpected argument %s\n", fs.Name(), fs.Arg(0))
		return 1, true
	}
	return 0, false
}

// Fail prints "name: err" and a newline to stderr and returns exit code 1.
// If err's message already starts with "name: " the prefix is not repeated.
func Fail(stderr io.Writer, name string, err error) int {
	msg := err.Error()
	if !strings.HasPrefix(msg, name+": ") {
		msg = name + ": " + msg
	}
	_, _ = io.WriteString(stderr, msg+"\n")
	return 1
}

// Write encodes v to stdout as one JSON value followed by a newline.
func Write(stdout io.Writer, v any) error {
	enc := json.NewEncoder(stdout)
	enc.SetEscapeHTML(false)
	return enc.Encode(v)
}
