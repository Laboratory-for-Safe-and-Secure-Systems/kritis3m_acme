package types

import "time"

type Time struct {
	time.Time
}

func NewTime(t time.Time) *Time {
	return &Time{Time: t}
}

// Convert *time.Time to *Time, handling nil values
func NewTimeP(t *time.Time) *Time {
	if t == nil {
		return nil
	}
	return NewTime(*t)
}
