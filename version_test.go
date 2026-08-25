package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

func TestInstallHookCreatesExecutableHook(t *testing.T) {
	path := filepath.Join(t.TempDir(), "pre-push")
	if err := installHook(path, "/usr/local/bin/git-commit-later"); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(b)
	if !strings.HasPrefix(content, "#!/bin/sh\n") {
		t.Fatalf("hook missing shell header:\n%s", content)
	}
	if !strings.Contains(content, "/usr/local/bin/git-commit-later' run-due --repo") {
		t.Fatalf("hook missing run-due command:\n%s", content)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if info.Mode().Perm() != 0o755 {
		t.Fatalf("mode = %v, want 0755", info.Mode().Perm())
	}
}

func TestInstallHookReplacesManagedBlock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "post-checkout")
	initial := "#!/bin/sh\nexisting\n\n# BEGIN git-commit-later\nold\n# END git-commit-later\ntrailer\n"
	if err := os.WriteFile(path, []byte(initial), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := installHook(path, "/new/path/git-commit-later"); err != nil {
		t.Fatal(err)
	}

	b, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	content := string(b)
	if strings.Contains(content, "\nold\n") {
		t.Fatalf("old managed block remained:\n%s", content)
	}
	if strings.Count(content, "# BEGIN git-commit-later") != 1 {
		t.Fatalf("managed block count mismatch:\n%s", content)
	}
	if !strings.Contains(content, "existing") || !strings.Contains(content, "trailer") {
		t.Fatalf("existing hook content was not preserved:\n%s", content)
	}
}

func TestShellQuoteEscapesSingleQuotes(t *testing.T) {
	got := shellQuote("/tmp/that's-it/git-commit-later")
	want := "'/tmp/that'\\''s-it/git-commit-later'"
	if got != want {
		t.Fatalf("shellQuote() = %q, want %q", got, want)
	}
}
