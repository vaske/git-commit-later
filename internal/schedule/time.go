package schedule

import (
	"fmt"
	"strings"
	"time"
)

func Parse(at, in string, now time.Time) (time.Time, error) {
	if at != "" && in != "" {
		return time.Time{}, fmt.Errorf("use either --at or --in, not both")
	}
	if in != "" {
		d, err := time.ParseDuration(in)
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid --in duration: %w", err)
		}
		if d <= 0 {
			return time.Time{}, fmt.Errorf("--in must be in the future")
		}
		return now.Add(d), nil
	}
	if at == "" {
		return time.Time{}, fmt.Errorf("provide --at or --in")
	}
	layouts := []string{time.RFC3339, "2006-01-02 15:04", "2006-01-02 15:04:05", "15:04"}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, at, now.Location()); err == nil {
			if layout == "15:04" {
				t = time.Date(now.Year(), now.Month(), now.Day(), t.Hour(), t.Minute(), 0, 0, now.Location())
				if !t.After(now) {
					t = t.Add(24 * time.Hour)
				}
			}
			if !t.After(now) {
				return time.Time{}, fmt.Errorf("scheduled time must be in the future")
			}
			return t, nil
		}
	}
	if strings.EqualFold(at, "tomorrow") {
		t := now.AddDate(0, 0, 1)
		return time.Date(t.Year(), t.Month(), t.Day(), 9, 0, 0, 0, now.Location()), nil
	}
	return time.Time{}, fmt.Errorf("cannot parse --at %q (try RFC3339, '2006-01-02 15:04', or '15:04')", at)
}
