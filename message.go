package main

import (
	"strconv"
	"time"
)

var (
	BOT_NAME = "Wavenauts"
	BOT_VER  = "v1.0"
)

// tweet用
type Message struct {
	units          int     // 保有量
	side           string  // 売玉か買玉か
	realizedProf   float64 // 実現損益
	unrealizedProf float64 //　評価額
	totalProf      float64 //　累計損益
	tags           string  // #ハッシュタグ

	didOpen  bool // このフレームでOPEN取引した
	didClose bool // このフレームでClose取引した

}

func NewMessage() *Message {
	m := &Message{tags: "#FX #USD_JPY"}
	return m
}

func (m *Message) open() {
	m.didOpen = true
}

func (m *Message) close() {
	m.didClose = true
}

func (m *Message) String() string {
	rp := strconv.Itoa(int(m.realizedProf))
	up := strconv.Itoa(int(m.unrealizedProf))
	tp := strconv.Itoa(int(m.totalProf))
	un := strconv.Itoa(m.units)

	var txt string
	if m.didClose && m.didOpen {
		txt = "『我OPENし、CLOSEす』\n"
	}
	if m.didClose && !m.didOpen {
		txt = "『我CLOSEす』\n"
	}
	if !m.didClose && m.didOpen {
		txt = "『我OPENす』\n"
	}

	msg := "[" + time.Now().Format("2006-01-02 15:04") + "]" + "\n"
	msg += "🐋" + BOT_NAME + "@" + BOT_VER + "🐋" + "\n"
	msg += "⚽" + "実現損益 :" + rp + "⚽" + "\n"
	msg += "💰" + "未実現損益:" + up + "💰" + "\n"
	msg += "🥎" + "保有量 :" + un + "🥎" + "\n"
	msg += "🗾" + "総利益  :" + tp + "🗾" + "\n"
	if len(txt) > 0 {
		msg += "\n"
		msg += txt + "\n"
	}
	msg += m.tags
	return msg
}
