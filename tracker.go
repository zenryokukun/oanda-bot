package main

import "time"

type Tracker struct {
	Next int64
	// インターバルを秒数で。5分->300,4時間->14400
	Seconds int64
}

func (t *Tracker) next() {
	t.Next += t.Seconds
}

func (t *Tracker) IsPassed() bool {
	nu := time.Now().Unix()
	if nu >= t.Next {
		t.next()
		return true
	}
	return false
}

// インターバルを秒数で。5分->300,4時間->14400
func NewTracker(itv int64) *Tracker {
	nu := time.Now().Unix()
	prev := nu - nu%itv
	next := prev + itv
	return &Tracker{
		Seconds: itv,
		Next:    next,
	}
}
