package main

import "testing"

func TestVersionLineUsesInjectedVersion(t *testing.T) {
	old := version
	version = "9.9.9-test"
	t.Cleanup(func() { version = old })

	got := versionLine()
	want := "git-commit-later 9.9.9-test"
	if got != want {
		t.Fatalf("versionLine() = %q, want %q", got, want)
	}
}
