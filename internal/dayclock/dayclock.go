package dayclock

import (
	"strings"
	"time"
)

const MaxStartMinute = 1439

type Clock struct {
	Loc         *time.Location
	StartMinute int
}

func New(zone string, startMinute int, fallback *time.Location) Clock {
	loc := fallback
	if z := strings.TrimSpace(zone); z != "" {
		if l, err := time.LoadLocation(z); err == nil {
			loc = l
		}
	}
	if loc == nil {
		loc = time.UTC
	}
	if startMinute < 0 || startMinute > MaxStartMinute {
		startMinute = 0
	}
	return Clock{Loc: loc, StartMinute: startMinute}
}

func ValidZone(zone string) bool {
	z := strings.TrimSpace(zone)
	if z == "" {
		return true
	}
	_, err := time.LoadLocation(z)
	return err == nil
}

func (c Clock) loc() *time.Location {
	if c.Loc == nil {
		return time.UTC
	}
	return c.Loc
}

func (c Clock) Start(civil time.Time) time.Time {
	y, m, d := civil.Date()
	return time.Date(y, m, d, 0, c.StartMinute, 0, 0, c.loc())
}

func (c Clock) Civil(t time.Time) time.Time {
	local := t.In(c.loc())
	y, m, d := local.Date()
	civil := time.Date(y, m, d, 0, 0, 0, 0, time.UTC)
	if t.Before(c.Start(civil)) {
		civil = civil.AddDate(0, 0, -1)
	}
	return civil
}

func (c Clock) Bounds(t time.Time) (string, time.Time, time.Time) {
	civil := c.Civil(t)
	return civil.Format("2006-01-02"), c.Start(civil), c.Start(civil.AddDate(0, 0, 1))
}
