package instances

import "testing"

func TestSkipUnlessLarge(t *testing.T) {
	t.Setenv(LargeEnv, "")
	if !t.Run("unset", func(t *testing.T) {
		SkipUnlessLarge(t)
		t.Error("SkipUnlessLarge did not skip with " + LargeEnv + " unset")
	}) {
		t.Fatal("unset subtest failed")
	}

	t.Setenv(LargeEnv, "1")
	ran := false
	t.Run("set", func(t *testing.T) {
		SkipUnlessLarge(t)
		ran = true
	})
	if !ran {
		t.Error("SkipUnlessLarge skipped with " + LargeEnv + "=1")
	}
}
