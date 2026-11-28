package instances

import (
	"os"
	"testing"
)

// LargeEnv is the environment variable that turns on the large benchmarks,
// whose names start with BenchmarkLarge. They take seconds to minutes per
// run, so go test -bench . skips them unless it is set:
//
//	AP_BENCH_LARGE=1 go test -run '^$' -bench Large ./lsap ./klsap
const LargeEnv = "AP_BENCH_LARGE"

// SkipUnlessLarge skips tb unless the environment variable LargeEnv is set
// to a non-empty value.
func SkipUnlessLarge(tb testing.TB) {
	tb.Helper()
	if os.Getenv(LargeEnv) == "" {
		tb.Skip("set " + LargeEnv + "=1 to run the large benchmarks")
	}
}
