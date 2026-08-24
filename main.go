package main

import (
	"crypto/rand"
	"encoding/hex"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/milan/git-commit-later/internal/gitrepo"
	"github.com/milan/git-commit-later/internal/jobs"
	sched "github.com/milan/git-commit-later/internal/schedule"
)

const version = "0.1.0"

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "list":
			must(listJobs())
		case "cancel":
			must(cancelJob(os.Args[2:]))
		case "run":
			must(runJob(os.Args[2:]))
		case "version", "--version", "-v":
			fmt.Println("git-commit-later", version)
		case "help", "--help", "-h":
			usage()
		default:
			must(scheduleJob(os.Args[1:]))
		}
	} else {
		usage()
	}
}

func usage() {
	fmt.Print(`git commit-later - schedule a staged Git snapshot for later

Usage:
  git commit-later "message" --in 2h
  git commit-later "message" --at "2026-08-25 09:00"
  git commit-later list
  git commit-later cancel <job-id>
  git commit-later run <job-id> [--wait]

The scheduled commit captures the staged index now. The commit only runs if
that branch still points to the same HEAD at execution time.
`)
}

func id() string { b := make([]byte, 4); _, _ = rand.Read(b); return hex.EncodeToString(b) }

func scheduleJob(args []string) error {
	fs := flag.NewFlagSet("schedule", flag.ContinueOnError)
	at := fs.String("at", "", "absolute/local time")
	in := fs.String("in", "", "duration from now, e.g. 2h or 30m")
	// Accept flags after message by manually splitting message from options.
	message := ""
	var flagArgs []string
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "-") {
			flagArgs = append(flagArgs, args[i:]...)
			break
		}
		if message != "" {
			message += " "
		}
		message += args[i]
	}
	if err := fs.Parse(flagArgs); err != nil {
		return err
	}
	if strings.TrimSpace(message) == "" {
		return fmt.Errorf("commit message is required")
	}
	when, err := sched.Parse(*at, *in, time.Now())
	if err != nil {
		return err
	}
	snap, err := gitrepo.Capture()
	if err != nil {
		return err
	}
	j := jobs.Job{ID: id(), Repo: snap.Repo, GitDir: snap.GitDir, Branch: snap.Branch, BaseHEAD: snap.HEAD, Tree: snap.Tree, Message: message, AuthorName: snap.AuthorName, AuthorEmail: snap.AuthorEmail, ScheduledAt: when, CreatedAt: time.Now(), Status: jobs.Pending}
	if err := jobs.Save(j); err != nil {
		return err
	}
	if err := startWorker(j.ID); err != nil {
		return fmt.Errorf("job saved as %s but worker could not start: %w", j.ID, err)
	}
	fmt.Printf("Scheduled %s on %s for %s (job %s)\n", short(snap.Tree), snap.Branch, when.Format(time.RFC1123), j.ID)
	return nil
}

func listJobs() error {
	js, err := jobs.List()
	if err != nil {
		return err
	}
	if len(js) == 0 {
		fmt.Println("No scheduled commits.")
		return nil
	}
	fmt.Printf("%-10s %-10s %-20s %-16s %s\n", "ID", "STATUS", "WHEN", "BRANCH", "MESSAGE")
	for _, j := range js {
		fmt.Printf("%-10s %-10s %-20s %-16s %s\n", j.ID, j.Status, j.ScheduledAt.Format("2006-01-02 15:04"), j.Branch, j.Message)
	}
	return nil
}

func cancelJob(args []string) error {
	if len(args) != 1 {
		return fmt.Errorf("usage: git commit-later cancel <job-id>")
	}
	j, err := jobs.Load(args[0])
	if err != nil {
		return err
	}
	if j.Status != jobs.Pending {
		return fmt.Errorf("job %s is %s", j.ID, j.Status)
	}
	j.Status = jobs.Cancelled
	if err := jobs.Save(j); err != nil {
		return err
	}
	fmt.Println("Cancelled", j.ID)
	return nil
}

func runJob(args []string) error {
	fs := flag.NewFlagSet("run", flag.ContinueOnError)
	wait := fs.Bool("wait", false, "wait until scheduled time")
	if len(args) == 0 {
		return fmt.Errorf("usage: git commit-later run <job-id> [--wait]")
	}
	jid := args[0]
	if err := fs.Parse(args[1:]); err != nil {
		return err
	}
	j, err := jobs.Load(jid)
	if err != nil {
		return err
	}
	if j.Status != jobs.Pending {
		return nil
	}
	if *wait && time.Now().Before(j.ScheduledAt) {
		time.Sleep(time.Until(j.ScheduledAt))
	}
	j, err = jobs.Load(jid)
	if err != nil {
		return err
	}
	if j.Status != jobs.Pending {
		return nil
	}
	if !*wait && time.Now().Before(j.ScheduledAt) {
		return fmt.Errorf("job is not due until %s", j.ScheduledAt.Format(time.RFC3339))
	}
	commit, err := gitrepo.CreateCommitAndAdvance(j.Repo, j.Branch, j.BaseHEAD, j.Tree, j.Message, j.AuthorName, j.AuthorEmail)
	if err != nil {
		j.Status = jobs.Failed
		j.Error = err.Error()
		_ = jobs.Save(j)
		return err
	}
	j.Status = jobs.Completed
	j.Commit = commit
	if err := jobs.Save(j); err != nil {
		return err
	}
	return nil
}

func startWorker(jobID string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	cmd := exec.Command(exe, "run", jobID, "--wait")
	null, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return err
	}
	cmd.Stdin = null
	cmd.Stdout = null
	cmd.Stderr = null
	detach(cmd)
	if err := cmd.Start(); err != nil {
		return err
	}
	if cmd.Process != nil {
		_ = cmd.Process.Release()
	}
	return nil
}

func short(s string) string {
	if len(s) > 10 {
		return s[:10]
	}
	return s
}
func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

var _ = strconv.Itoa
