package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestArchiveNameForUnix(t *testing.T) {
	got := installLib(t, "archive_name darwin arm64")
	if got != "git-commit-later_darwin_arm64.tar.gz" {
		t.Fatalf("got %q", got)
	}
}

func TestArchiveNameForWindows(t *testing.T) {
	got := installLib(t, "archive_name windows amd64")
	if got != "git-commit-later_windows_amd64.zip" {
		t.Fatalf("got %q", got)
	}
}

func TestNormalizeOS(t *testing.T) {
	cases := map[string]string{
		"Darwin":  "darwin",
		"Linux":   "linux",
		"MINGW64": "windows",
	}
	for in, want := range cases {
		got := installLib(t, "normalize_os "+in)
		if got != want {
			t.Fatalf("normalize_os %s = %q, want %q", in, got, want)
		}
	}
}

func TestNormalizeArch(t *testing.T) {
	cases := map[string]string{
		"x86_64":  "amd64",
		"amd64":   "amd64",
		"aarch64": "arm64",
		"arm64":   "arm64",
	}
	for in, want := range cases {
		got := installLib(t, "normalize_arch "+in)
		if got != want {
			t.Fatalf("normalize_arch %s = %q, want %q", in, got, want)
		}
	}
}

func TestReleaseAssetURL(t *testing.T) {
	latest := installLib(t, "release_asset_url latest git-commit-later_darwin_arm64.tar.gz")
	wantLatest := "https://github.com/vaske/git-commit-later/releases/latest/download/git-commit-later_darwin_arm64.tar.gz"
	if latest != wantLatest {
		t.Fatalf("latest = %q, want %q", latest, wantLatest)
	}

	pinned := installLib(t, "release_asset_url v0.1.0 git-commit-later_linux_amd64.tar.gz")
	wantPinned := "https://github.com/vaske/git-commit-later/releases/download/v0.1.0/git-commit-later_linux_amd64.tar.gz"
	if pinned != wantPinned {
		t.Fatalf("pinned = %q, want %q", pinned, wantPinned)
	}
}

func TestReleaseTag(t *testing.T) {
	if got := installLib(t, "VERSION=latest; release_tag"); got != "latest" {
		t.Fatalf("latest tag = %q", got)
	}
	if got := installLib(t, "VERSION=v0.1.0; release_tag"); got != "v0.1.0" {
		t.Fatalf("v-prefixed tag = %q", got)
	}
	if got := installLib(t, "VERSION=0.1.0; release_tag"); got != "v0.1.0" {
		t.Fatalf("unprefixed tag = %q", got)
	}
}

func installLib(t *testing.T, expr string) string {
	t.Helper()
	cmd := exec.Command("sh", "-c", "GIT_COMMIT_LATER_LIB=1; . ./install.sh; "+expr)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("%s: %v\n%s", expr, err, out)
	}
	return strings.TrimSpace(string(out))
}
