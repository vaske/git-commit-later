package gitrepo

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Snapshot struct {
	Repo, GitDir, Branch, HEAD, Tree, AuthorName, AuthorEmail string
}

func git(repo string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	if repo != "" {
		cmd.Dir = repo
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		msg := strings.TrimSpace(stderr.String())
		if msg == "" {
			msg = err.Error()
		}
		return "", fmt.Errorf("git %s: %s", strings.Join(args, " "), msg)
	}
	return strings.TrimSpace(stdout.String()), nil
}

func Capture() (Snapshot, error) {
	root, err := git("", "rev-parse", "--show-toplevel")
	if err != nil {
		return Snapshot{}, fmt.Errorf("not inside a Git repository: %w", err)
	}
	root, _ = filepath.Abs(root)
	gd, err := git(root, "rev-parse", "--absolute-git-dir")
	if err != nil {
		return Snapshot{}, err
	}
	branch, err := git(root, "symbolic-ref", "--quiet", "--short", "HEAD")
	if err != nil {
		return Snapshot{}, fmt.Errorf("detached HEAD is not supported in v0.1")
	}
	head, err := git(root, "rev-parse", "HEAD")
	if err != nil {
		return Snapshot{}, fmt.Errorf("repository must have at least one commit: %w", err)
	}
	tree, err := git(root, "write-tree")
	if err != nil {
		return Snapshot{}, err
	}
	diff, err := git(root, "diff", "--cached", "--quiet")
	if err == nil && diff == "" {
		return Snapshot{}, fmt.Errorf("nothing staged; stage files first with git add")
	}
	// git diff --quiet exits 1 when there are changes; our helper turns that into an error.
	// Verify the tree actually differs from HEAD instead.
	headTree, err := git(root, "rev-parse", "HEAD^{tree}")
	if err != nil {
		return Snapshot{}, err
	}
	if tree == headTree {
		return Snapshot{}, fmt.Errorf("nothing staged; stage files first with git add")
	}
	name, _ := git(root, "config", "user.name")
	email, _ := git(root, "config", "user.email")
	return Snapshot{root, gd, branch, head, tree, name, email}, nil
}

func CreateCommitAndAdvance(repo, branch, baseHEAD, tree, message, authorName, authorEmail string) (string, error) {
	current, err := git(repo, "rev-parse", "refs/heads/"+branch)
	if err != nil {
		return "", err
	}
	if current != baseHEAD {
		return "", fmt.Errorf("branch %s moved: scheduled from %.10s, now %.10s", branch, baseHEAD, current)
	}

	cmd := exec.Command("git", "commit-tree", tree, "-p", baseHEAD, "-m", message)
	cmd.Dir = repo
	cmd.Env = os.Environ()
	if authorName != "" {
		cmd.Env = append(cmd.Env, "GIT_AUTHOR_NAME="+authorName, "GIT_COMMITTER_NAME="+authorName)
	}
	if authorEmail != "" {
		cmd.Env = append(cmd.Env, "GIT_AUTHOR_EMAIL="+authorEmail, "GIT_COMMITTER_EMAIL="+authorEmail)
	}
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("git commit-tree: %s", strings.TrimSpace(stderr.String()))
	}
	commit := strings.TrimSpace(stdout.String())
	if _, err := git(repo, "update-ref", "refs/heads/"+branch, commit, baseHEAD); err != nil {
		return "", err
	}
	return commit, nil
}
