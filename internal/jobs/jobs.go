package jobs

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"
)

type Status string

const (
	Pending   Status = "pending"
	Completed Status = "completed"
	Cancelled Status = "cancelled"
	Failed    Status = "failed"
)

type Job struct {
	ID          string    `json:"id"`
	Repo        string    `json:"repo"`
	GitDir      string    `json:"gitDir"`
	Branch      string    `json:"branch"`
	BaseHEAD    string    `json:"baseHead"`
	Tree        string    `json:"tree"`
	Message     string    `json:"message"`
	AuthorName  string    `json:"authorName,omitempty"`
	AuthorEmail string    `json:"authorEmail,omitempty"`
	ScheduledAt time.Time `json:"scheduledAt"`
	CreatedAt   time.Time `json:"createdAt"`
	Status      Status    `json:"status"`
	Error       string    `json:"error,omitempty"`
	Commit      string    `json:"commit,omitempty"`
}

func dir() (string, error) {
	if v := os.Getenv("GIT_COMMIT_LATER_HOME"); v != "" {
		if err := os.MkdirAll(v, 0o700); err != nil {
			return "", err
		}
		return v, nil
	}
	cfg, err := os.UserConfigDir()
	if err != nil {
		return "", err
	}
	d := filepath.Join(cfg, "git-commit-later")
	if err := os.MkdirAll(d, 0o700); err != nil {
		return "", err
	}
	return d, nil
}

func path(id string) (string, error) {
	d, err := dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(d, id+".json"), nil
}

func Save(j Job) error {
	p, err := path(j.ID)
	if err != nil {
		return err
	}
	b, err := json.MarshalIndent(j, "", "  ")
	if err != nil {
		return err
	}
	tmp := p + ".tmp"
	if err := os.WriteFile(tmp, b, 0o600); err != nil {
		return err
	}
	return os.Rename(tmp, p)
}

func Load(id string) (Job, error) {
	p, err := path(id)
	if err != nil {
		return Job{}, err
	}
	b, err := os.ReadFile(p)
	if errors.Is(err, os.ErrNotExist) {
		return Job{}, fmt.Errorf("job %s not found", id)
	}
	if err != nil {
		return Job{}, err
	}
	var j Job
	if err := json.Unmarshal(b, &j); err != nil {
		return Job{}, err
	}
	return j, nil
}

func List() ([]Job, error) {
	d, err := dir()
	if err != nil {
		return nil, err
	}
	entries, err := os.ReadDir(d)
	if err != nil {
		return nil, err
	}
	out := make([]Job, 0, len(entries))
	for _, e := range entries {
		if e.IsDir() || filepath.Ext(e.Name()) != ".json" {
			continue
		}
		b, err := os.ReadFile(filepath.Join(d, e.Name()))
		if err != nil {
			continue
		}
		var j Job
		if json.Unmarshal(b, &j) == nil {
			out = append(out, j)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ScheduledAt.Before(out[j].ScheduledAt) })
	return out, nil
}
