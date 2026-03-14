package models

import "time"

var defaultLoc *time.Location

func init() {
	loc, err := time.LoadLocation("America/Los_Angeles")
	if err != nil {
		defaultLoc = time.UTC
	} else {
		defaultLoc = loc
	}
}

type DateTimeInput struct {
	Year            int
	Month           int
	Day             int
	Hour            int
	Minute          int
	Second          int
	Location        *time.Location
	HasExplicitDate bool
}

func (dt DateTimeInput) Resolve(now time.Time) time.Time {
	loc := dt.Location
	if loc == nil {
		loc = defaultLoc
	}
	now = now.In(loc)

	var base time.Time
	if dt.HasExplicitDate {
		base = time.Date(dt.Year, time.Month(dt.Month), dt.Day, 0, 0, 0, 0, loc)
	} else {
		base = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, loc)
	}

	t := time.Date(base.Year(), base.Month(), base.Day(), dt.Hour, dt.Minute, dt.Second, 0, loc)
	if !dt.HasExplicitDate && t.After(now) {
		t = t.AddDate(0, 0, -1)
	}

	return t
}

func (dt DateTimeInput) Unix(now time.Time) int64 {
	return dt.Resolve(now).Unix()
}

func SplitYYYYMMDD(v int) (year, month, day int) {
	year = v / 10000
	md := v % 10000
	month = md / 100
	day = md % 100
	return
}

func SplitHHMMSS(v int) (hour, minute, second int) {
	hour = v / 10000
	ms := v % 10000
	minute = ms / 100
	second = ms % 100
	return
}
