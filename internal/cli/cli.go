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
	return fs.Float64("scale", 0, "multiply costs by `s` and round them to integers, to accept non-integer costs (0, the default, means off)")
}

// Parse parses args with fs. The caller creates fs with flag.ContinueOnError,
// sets its output and usage function. Parse returns (0, false) on success, in
// which case the program continues. If help was requested it prints the usage
// once and returns (0, true). On a bad flag or an unexpected positional
// argument it prints "name: message" and the usage, and returns (1, true).
func Parse(fs *flag.FlagSet, args []string) (code int, exit bool) {
	// Silence the flag package's own message and usage, so that the error
	// gets the "name: " prefix and the usage is printed exactly once.
	out, usage := fs.Output(), fs.Usage
	fs.SetOutput(io.Discard)
	fs.Usage = func() {}
	err := fs.Parse(args)
	fs.SetOutput(out)
	fs.Usage = usage

	switch {
	case err == flag.ErrHelp:
		printUsage(fs)
		return 0, true
	case err != nil:
		_, _ = fmt.Fprintf(out, "%s: %v\n", fs.Name(), err)
		printUsage(fs)
		return 1, true
	case fs.NArg() > 0:
		_, _ = fmt.Fprintf(out, "%s: unexpected argument %s\n", fs.Name(), fs.Arg(0))
		printUsage(fs)
		return 1, true
	}
	return 0, false
}

// printUsage calls fs.Usage, or prints the flag package's default usage if
// fs.Usage is nil.
func printUsage(fs *flag.FlagSet) {
	if fs.Usage != nil {
		fs.Usage()
		return
	}
	_, _ = fmt.Fprintf(fs.Output(), "Usage of %s:\n", fs.Name())
	fs.PrintDefaults()
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
