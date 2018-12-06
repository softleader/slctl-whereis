package main

import (
	"github.com/araddon/dateparse"
	"time"
)

func parse(date string) (t time.Time) {
	if date == "" {
		return time.Now()
	}
	if "today" == date {
		return time.Now()
	}
	if "tomorrow" == date || "tmw" == date {
		return time.Now().AddDate(0, 0, 1)
	}
	if "yesterday" == date || "ytd" == date {
		return time.Now().AddDate(0, 0, -1)
	}
	t, err := dateparse.ParseLocal(date)
	if err != nil {
		t, _ = time.Parse(layout, date)
	}
	return
}
