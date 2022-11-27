package main

import (
	"time"
)

// itv秒分スリープする関数。itv -> 秒。5分なら300。
func tick(itv int64) {
	micro := itv * 1000000              // seconds -> microseconds
	now := time.Now().UnixMicro()       // 現在時刻をmicro秒で。
	rem := now % micro                  // 前回時刻からの経過秒。itv:300,12:06 -> 1分
	prev := now - rem                   // 前回時刻
	next := prev + micro                // 次回時刻
	diff := time.Duration(next - now)   // 現在時刻から次回時刻までのmicro秒数。
	time.Sleep(diff * time.Microsecond) // 次回時刻まで待つ。
}
