package version

import "testing"

func TestStringReturnsBuildTimeVersion(t *testing.T) {
	original := Version
	t.Cleanup(func() { Version = original })

	Version = "v1.2.3"
	if got, want := String(), "v1.2.3"; got != want {
		t.Fatalf("String() = %q, want %q", got, want)
	}
}

func TestStringFallsBackWhenUnset(t *testing.T) {
	original := Version
	t.Cleanup(func() { Version = original })

	Version = ""
	if got := String(); got == "" {
		t.Fatal("String() returned empty version")
	}
}
