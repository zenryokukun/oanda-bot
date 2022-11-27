package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Slicer interface {
	Slice(int)
}

type (
	XY struct {
		X []int64   // openTime Unix
		Y []float64 // 価格
	}

	BalanceData struct {
		XY                // unixTime,価格
		TotalPL []float64 // 残高
	}

	// 取引履歴
	TradeData struct {
		XY              // unixTime,価格
		Side   []string // "BUY" | "SELL"
		Action []string // "OPEN" | "STRING"
	}
)

// ***************************************************
// Public
// ***************************************************
func writeBalance(fpath string, mlen int, x int64, y, balance float64) {
	bl := NewBalanceHistory()
	load(fpath, bl)
	bl.Add(x, y, balance)
	bl.Slice(mlen)
	dump(fpath, bl)
}

func writeTrade(fpath string, mlen int, x int64, y float64, side, action string) {
	td := NewTradeHistory()
	load(fpath, td)
	td.Add(x, y, side, action)
	td.Slice(mlen)
	dump(fpath, td)
}

// ***************************************************
// utility functions
// ***************************************************
func dump(fpath string, data Slicer) {
	f, err := os.Create(fpath)
	if err != nil {
		fmt.Println(err)
	}
	defer f.Close()
	b, err := json.MarshalIndent(data, "", " ")
	if err != nil {
		fmt.Println(err)
	}
	f.Write(b)
}

func load(fpath string, data Slicer) {
	_, err := os.Stat(fpath)
	if os.IsNotExist(err) {
		return
	}
	b, err := os.ReadFile(fpath)
	if err != nil {
		fmt.Println(err)
	}
	err = json.Unmarshal(b, data)
	if err != nil {
		fmt.Println(err)
	}
}

// ***************************************************
//  XY
// ***************************************************
func (xy *XY) Slice(mlen int) {
	lx := len(xy.X)
	ly := len(xy.Y)
	if lx != ly {
		panic("XY:mistached length.")
	}
	if lx <= mlen {
		return
	}
	st := lx - mlen
	xy.X = xy.X[st:]
	xy.Y = xy.Y[st:]
}

// ***************************************************
// Balance
// ***************************************************
func NewBalanceHistory() *BalanceData {
	return &BalanceData{}
}

func (b *BalanceData) Slice(mlen int) {
	lx := len(b.X)
	if lx != len(b.TotalPL) {
		panic("balanceData:mismatched length.")
	}
	if len(b.X) <= mlen {
		return
	}
	b.XY.Slice(mlen)
	st := lx - mlen
	b.TotalPL = b.TotalPL[st:]
}

func (b *BalanceData) Add(x int64, y, balance float64) {
	b.X = append(b.X, x)
	b.Y = append(b.Y, y)
	b.TotalPL = append(b.TotalPL, balance)
}

// ***************************************************
// Trade
// ***************************************************
func NewTradeHistory() *TradeData {
	return &TradeData{}
}

func (t *TradeData) Slice(mlen int) {
	lx := len(t.X)
	ls := len(t.Side)
	la := len(t.Action)
	if !(lx == ls && lx == la) {
		panic("tradeData:mistmatched length.")
	}
	if lx <= mlen {
		return
	}
	t.XY.Slice(mlen)
	st := lx - mlen
	t.Action = t.Action[st:]
	t.Side = t.Side[st:]
}

func (t *TradeData) Add(x int64, y float64, side, action string) {
	t.X = append(t.X, x)
	t.Y = append(t.Y, y)
	t.Side = append(t.Side, side)
	t.Action = append(t.Action, action)
}

// import (
// 	"encoding/json"
// 	"fmt"
// 	"os"
// )

// type (
// 	XY struct {
// 		X []int     // openTime Unix
// 		Y []float64 // 価格
// 	}

// 	// 取引履歴
// 	TradeHistoryData struct {
// 		XY
// 		Side   []string // "BUY" | "SELL"
// 		Action []string // "OPEN" | "STRING"
// 	}

// 	History struct {
// 		mlen  int    // データの最大保有量
// 		fpath string // 書き込み先ファイルパス
// 		Data  Slicer
// 	}

// 	TradeFactory struct {}

// 	BalanceFactory struct {}
// )

// type Slicer interface {
// 	Slice(mlen int) Slicer
// }

// type Initializer interface {
// 	Init() Slicer
// }

// func (tf *TradeFactory) Init(){

// }

// func (th *TradeHistoryData) Init() Slicer {
// 	return &TradeHistoryData{}
// }

// func (th *TradeHistoryData) Slice(mlen int) Slicer {
// 	ls := len(th.Side)
// 	la := len(th.Action)
// 	lx := len(th.X)
// 	ly := len(th.Y)
// 	if !(ls == la && ls == lx && ls == ly) {
// 		panic("TradeHistoryData length mismatch.")
// 	}
// 	if ls <= mlen {
// 		return th
// 	}
// 	st := ls - mlen

// 	th.X = th.X[st:]
// 	th.Y = th.Y[st:]
// 	th.Action = th.Action[st:]
// 	th.Side = th.Side[st:]
// 	return th
// }

// func NewHistory(fpath string, mlen int) *History {
// 	// fpathが存在しない場合作成。最終的には上書きされるが、
// 	// load時に存在しないとエラーになるので。
// 	_, err := os.Stat(fpath)
// 	if os.IsNotExist(err) {
// 		os.Create(fpath)
// 	}

// 	return &History{
// 		fpath: fpath,
// 		mlen:  mlen,
// 		Data:  &TradeHistoryData{},
// 	}
// }

// // ファイルからTradeHistoryDataをロード
// func (h *History) load() *TradeHistoryData {
// 	data := &TradeHistoryData{}
// 	b, err := os.ReadFile(h.fpath)
// 	_ = b
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	json.Unmarshal(b, data)
// 	return data
// }

// // ファイルに書き込む
// func (h *History) dump(data *TradeHistoryData) {
// 	// dataで上書きするためCreate。
// 	f, err := os.Create(h.fpath)
// 	if err != nil {
// 		fmt.Println(err)
// 	}
// 	defer f.Close()
// 	b, err := json.MarshalIndent(data, "", " ")
// 	f.Write(b)
// }

// // データを最大値を超えないようにスライスする
// func (h *History) slice(data *TradeHistoryData) *TradeHistoryData {
// 	ls := len(data.Side)
// 	la := len(data.Action)
// 	lx := len(data.X)
// 	ly := len(data.Y)
// 	if !(ls == la && ls == lx && ls == ly) {
// 		panic("TradeHistoryData length mismatch.")
// 	}
// 	if ls <= h.mlen {
// 		return data
// 	}
// 	st := ls - h.mlen

// 	data.X = data.X[st:]
// 	data.Y = data.Y[st:]
// 	data.Action = data.Action[st:]
// 	data.Side = data.Side[st:]
// 	return data
// }

// // fpathのデータをロードし、パラメタのデータを追加してファイルを上書き
// func (h *History) Merge(x int, y float64, side, action string) {
// 	// ファイルから読み取り
// 	hs := h.load()
// 	if hs == nil {
// 		return
// 	}
// 	hs.X = append(hs.X, x)
// 	hs.Y = append(hs.Y, y)
// 	hs.Side = append(hs.Side, side)
// 	hs.Action = append(hs.Action, action)
// 	h.dump(hs)
// }
