package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"time"

	"github.com/zenryokukun/oanda-bot/oanda"
	"github.com/zenryokukun/surfergopher/minmax"
)

// 稼働時の総利益。現在総歴-稼働時の総利益 = BOTの総利益　とするため。
var INITIAL_BALANCE = 500000.0
var BALANCE_FILE = "./balance.json"
var TRADE_FILE = "./trade.json"

// ロジックに使うパラメタ。コンパイル面倒だからファイルから読み取る。
type Param struct {
	Inst     string  // Instrument: "USD_JPY","EUR_USD"等
	Gran     string  // granularity："M5","H4",等。
	Seconds  int     // granularityを秒数で表したもの。"M5" -> 300
	Span     int     // Gran何個分で予測するか
	Thresh   float64 // レンジ判定の閾値
	ProfRate float64 // 利確ライン
	LossRate float64 // 損切ライン
	Spread   float64 // 許容スプレッド
	Units    int     // 取引量
}

// ファイルからパラメタを読みってParam structを返す
func loadParam(fpath string) *Param {
	b, err := os.ReadFile(fpath)
	if err != nil {
		panic(err)
	}
	p := &Param{}
	json.Unmarshal(b, p)
	return p
}

// パラメタをもとに保有ポジションを取得して返す。
// PositionDataにはLong,Shortそれぞれfieldがあるので留意。
func position(goq *oanda.Goquest, prm *Param) *oanda.PositionData {
	pos := oanda.NewPosition(goq, prm.Inst)
	data := pos.Extract()
	if data == nil {
		return nil
	}
	return data
}

// マーケットが空いているかチェック。パラメタは最後のロウソク足を想定
// 現在時刻と最後のロウソク足を比較し、prm.Granの3倍以上開いていたら閉じていると判断させる。
func isMarketOpen(cs oanda.CandleStick, prm *Param) bool {
	ct, err := time.Parse(layout(), cs.Time)
	if err != nil {
		fmt.Printf("marketopen:Could not parse time:%v\n", cs.Time)
		return false
	}
	now := time.Now().Unix()
	diff := now - ct.Unix()
	// 3倍を超えていたらマーケットが閉じていると判断
	if diff >= int64(prm.Seconds)*3 {
		fmt.Printf("Market might be closed...Last:%v,Now:%v,diff:%v\n", ct.Unix(), time.Now().UTC(), diff)
		return false
	}
	return true
}

// パラメタをもとにロウソク足取得
func candles(goq *oanda.Goquest, prm *Param) oanda.CandleSticks {
	span := prm.Span + 1 // ロウソク足が完成していないものが入っている可能性があるので＋1
	cd := oanda.NewCandles(goq, span, prm.Gran, prm.Inst, "", "", "")
	sticks := cd.ExtractMid()
	if sticks != nil {
		// 完成したロウソク足のみ抽出
		sticks = sticks.Complete()
		// 長さを超えている場合はslice
		st := len(sticks) - prm.Span
		sticks = sticks[st:]
		// 所定の長さに達していなかったらログ
		if len(sticks) != prm.Span {
			fmt.Printf("Stick length does not match Param. Stick.length:%v\n", len(sticks))
		}
	}
	return sticks
}

// 現在のPrice取得
func latestPrice(goq *oanda.Goquest, prm *Param) *oanda.Price {
	pr := oanda.NewPricing(goq, prm.Inst).Latest(prm.Inst)
	return pr
}

// botの総利益
func totalPL(goq *oanda.Goquest) float64 {
	acc := oanda.NewAccount(goq)
	data := acc.Extract()
	if data == nil {
		return 0.0
	}
	return data.Balance - INITIAL_BALANCE
}

// Longポジを持っていいれば"BUY"、Shortなら"SELL"を返す
// 両建て不可アカウントを想定しているため、両方は存在することは想定しない。
// ポジ無しの時は空文字を返す
func tradeSide(p *oanda.PositionData) string {
	if !p.Has() {
		return ""
	}
	if p.Long.Units > 0 {
		return "BUY"
	}
	if p.Short.Units < 0 {
		return "SELL"
	}
	return ""
}

// 保有ポジと逆サイドを返す。決済の向きを指定するために使う
func closingSide(side string) string {
	if side == "BUY" {
		return "SELL"
	}
	if side == "SELL" {
		return "BUY"
	}
	return ""
}

// 売買判定ロジック
func breakThrough(v float64, inf *minmax.Inf) string {
	if v > inf.Maxv {
		return "BUY"
	}
	if v < inf.Minv {
		return "SELL"
	}
	return ""
}

// v: 現在価格。p:取得価格 side:"BUY"or"SELL",prm: Param
func isProfFilled(v float64, p float64, side string, prm *Param) bool {
	if side == "BUY" {
		return (v-p)/p >= prm.ProfRate
	} else {
		return (p-v)/p >= prm.ProfRate
	}
}

// v: 現在価格。p:取得価格 side:"BUY"or"SELL",prm: Param
func isLossFilled(v float64, p float64, side string, prm *Param) bool {
	if side == "BUY" {
		return (v-p)/p <= prm.LossRate
	} else {
		return (p-v)/p <= prm.LossRate
	}
}

// spreadが許容値になるまで待つ
func waitSpread(goq *oanda.Goquest, price *oanda.Price, prm *Param, secs int) *oanda.Price {
	if price.Spread() <= prm.Spread {
		return price
	}
	for i := 0; i < secs; i++ {
		time.Sleep(time.Second * 1)
		p := latestPrice(goq, prm)
		if p == nil {
			continue
		}
		if p.Spread() <= prm.Spread {
			return p
		}
	}
	return nil
}

// orderIDの注文がFILLEDになるまで待つ。sec秒待ってもFILLしない場合、falseを返す
func waitOrderFill(goq *oanda.Goquest, orderID string, sec int) bool {
	// 300ミリ秒ごとに実行
	for i := 0.3; i < float64(sec); i += 0.3 {
		order := oanda.NewOrderData(goq, orderID)
		if order == nil {
			fmt.Printf("Could not get orderID:%v \n", orderID)
			continue
		}
		status := order.OrderStatus()
		if status == "FILLED" {
			return true
		}
		time.Sleep(300 * time.Millisecond)
	}
	return false
}

// 成行き注文。両建て不可アカウントなので、openもcloseもこれで完結
// go で呼ぶこと。
func marketOrder(goq *oanda.Goquest, inst, side string, units int, ch chan string) {
	// 売りの場合はunitをマイナスで指定する仕様

	if side == "SELL" {
		units *= -1
	}
	// 注文してorder IDを抽出
	order := oanda.NewMarketOrder(goq, inst, units)
	id := order.Id()
	// IDが取得できない場合はリターン
	if id == "" {
		fmt.Printf("marketOrder: id was empty:%v", id)
		ch <- ""
		return
	}
	// orderが完了するまで待つ
	isFilled := waitOrderFill(goq, id, 6)

	if isFilled {
		ch <- id
		return
	}
	ch <- ""
}

// 保有ポジションをcloseする処理。ヘルパー。orderがFILLEDになるまで待つ。
func closeOrder(goq *oanda.Goquest, pos *oanda.PositionData, prm *Param, ch chan string) {
	posSide := tradeSide(pos)
	closeSide := closingSide(posSide)
	units := pos.Units()
	// marketOrderでSELL時はunit *= -1にする処理があるので、ここでは絶対値にしておく
	units = int(math.Abs(float64(units)))
	marketOrder(goq, prm.Inst, closeSide, units, ch)
}

// 実現損益をtweetメッセージに設定
func addClosingMsg(goq *oanda.Goquest, prm *Param, ids string, m *Message) {
	trades := oanda.NewTrades(goq, ids, "CLOSED", prm.Inst, "", "")
	realized, _ := trades.PL()
	m.realizedProf = realized
}

// 保有中ポジションの情報をメッセージにセット。
func addPositionMsg(goq *oanda.Goquest, prm *Param, ids string, m *Message) {
	trades := oanda.NewTrades(goq, ids, "OPEN", prm.Inst, "", "")
	_, unrealized := trades.PL() // 評価額
	units := trades.Units()
	if units > 0 {
		m.side = "LONG"
	} else {
		m.side = "SHORT"
	}
	m.unrealizedProf = unrealized
	m.units = units
}

func addTotalPLMsg(pl float64, m *Message) {
	m.totalProf = pl
}

// ロジック部分
func frame(goq *oanda.Goquest, prm *Param) *Message {
	pos := position(goq, prm)
	sticks := candles(goq, prm)
	price := latestPrice(goq, prm)

	// graph用データの最大個数
	mlen := 5000

	// このフレームでの新規取引フラグ
	willClose := false
	chOrder := make(chan string, 1)

	// chOrderの受け皿。open取引のorderID
	// close時には設定しないで。
	openOrderId := ""

	msg := NewMessage() // tweet用

	// apiで取得できないデータがあれば処理なし
	if pos == nil || sticks == nil || price == nil {
		fmt.Println("pos,sticks,or price is nil.")
		return msg
	}

	// マーケットが閉じているっぽければ処理なし
	if !isMarketOpen(sticks[len(sticks)-1], prm) {
		return msg
	}

	// 現在価格
	current := price.Mid()

	// 現在価格が取得できない場合は処理なし
	if current == oanda.EmptyError {
		fmt.Println("current is EmptyError")
		return msg
	}

	// 最後のロウソク足のopentime
	openTime := toUnix(sticks[len(sticks)-1].Time)

	// 最大値と最小値をセット。AddWrapしてるが今のところ使う予定なし
	highs, lows := sticks.Extract("H"), sticks.Extract("L")
	inf := minmax.NewInf(highs, lows).AddWrap(current)

	// 値幅
	vel := 1 - (inf.Minv / inf.Maxv)
	// 新規取引判定 "BUY","SELL",""
	dec := breakThrough(current, inf)
	// 保有ポジ。long->"BUY", short->"SELL", なし->""
	side := tradeSide(pos)

	// 逆向きポジを持っていて、かつ値幅が閾値を超えていれば決済。
	if len(dec) > 0 {
		if len(side) > 0 && side != dec && vel > prm.Thresh {
			willClose = true
		}
	}

	// ポジションがあり、上でcloseしていない場合、利確・損切処理を実施
	if len(side) > 0 && !willClose {
		// 平均取得価格
		posData := pos.Side()
		avg := posData.Average
		if isLossFilled(current, avg, side, prm) {
			willClose = true
		}
		if isProfFilled(current, avg, side, prm) {
			willClose = true
		}
	}

	// ****************************************************
	// 保有ポジを閉じる処理
	// ****************************************************
	if willClose {
		// spreadが許容値になるまで待つ。待っても収まらない場合は取引しない。
		price = waitSpread(goq, price, prm, 15)
		if price != nil {
			fmt.Println("closing!!!")
			go closeOrder(goq, pos, prm, chOrder)
			// 結局待つwww
			<-chOrder
			// tradeグラフ用データをファイルに出力
			writeTrade(TRADE_FILE, mlen, openTime, current, closingSide(side), "CLOSE")
		}
	}

	// ****************************************************
	// 新規購入処理。
	// 新規取引判定されている場合で、保有ポジションが無い場合、
	// もしくは本フレームでクローズしている場合、新規取引
	// ****************************************************
	if len(dec) > 0 {
		if len(side) == 0 || willClose {
			price = waitSpread(goq, price, prm, 15)
			if price != nil {
				fmt.Println("opening!!!")
				go marketOrder(goq, prm.Inst, dec, prm.Units, chOrder)
				<-chOrder
				// tradeグラフ用データをファイルに出力
				writeTrade(TRADE_FILE, mlen, openTime, current, closingSide(side), "OPEN")
			}
		}
	}

	// ****************************************************
	// tweet処理
	// ****************************************************
	tradeIDs := pos.Ids()
	if willClose {
		// closeした場合は確定損益を設定
		addClosingMsg(goq, prm, tradeIDs, msg)
		if len(openOrderId) > 0 {
			// 同じフレームで新規open取引をしていたら、その情報を設定
			// 新規取引なので新たにポジションをとりなおす。
			newPos := position(goq, prm)
			newTradeIds := newPos.Ids()
			addPositionMsg(goq, prm, newTradeIds, msg)
		}
	} else {
		// 決済されていない場合、保有ポジションの情報を設定。無い場合は全てzero-valueになる（はず）。
		addPositionMsg(goq, prm, tradeIDs, msg)
	}

	acc := oanda.NewAccount(goq)
	accData := acc.Extract()
	var tpl, upl float64 // 総利益,評価額込みの総利益
	if accData != nil {
		tpl = accData.Balance - INITIAL_BALANCE
		upl = tpl + accData.UnrealizedPL
	}
	// tpl := totalPL(goq)
	addTotalPLMsg(tpl, msg)

	// balance用データをファイルに出力
	writeBalance(BALANCE_FILE, mlen, openTime, current, upl)

	return msg
}

func main() {
	goq := oanda.NewGoquest("./key.json", "live")
	prm := loadParam("./param.json")
	// 4hに設定
	tracker := NewTracker(4 * 60 * 60)
	prm.Gran = "M1"
	prm.Inst = "EUR_USD"
	prm.Units = 100
	prm.LossRate = 0.001
	prm.ProfRate = 0.001
	prm.Seconds = 60
	_ = tracker

	// posSide := tradeSide(pos)
	// closeSide := closingSide(posSide)
	// units := pos.Units()
	// fmt.Println(posSide, closeSide, units, pos.Has())
	// fmt.Printf("%+v\n", pos.Short.Units)
	// ch := make(chan string, 1)
	// go closeOrder(goq, pos, prm, ch)
	// v := <-ch
	// fmt.Println(v)

	for {
		tick(int64(prm.Seconds))
		fmt.Println(time.Now())
		msg := frame(goq, prm)
		if tracker.IsPassed() {
			twitter := NewTwitter("./twitter.json")
			twitter.tweet(msg.String(), nil)
		}
	}
}

// Test Codes

// writeTrade("trade_test.json", 3, 4, 109, "BUY", "CLOSE")

// for {
// 	tick(60)
// }

// tweet テスト
// ch := make(chan string)
// tw := NewTwitter("./twitter.json")
// m := NewMessage()
// go marketOrder(goq, "EUR_USD", "BUY", 100, ch)
// <-ch
// p.Inst = "EUR_USD"
// newPos := position(goq, p)
// newTradeIds := newPos.Ids()
// addPositionMsg(goq, p, newTradeIds, m)
// // addPositionMsg(goq, p, d.Ids(), m)
// // addClosingMsg(goq, p, d.Ids(), m)
// tw.tweet(m.String(), nil)

// fmt.Println(d.Units())
// fmt.Println(d.Ids())
// t := oanda.NewTrades(goq, "779,777", "CLOSED", "EUR_USD", "", "")
// a, b := t.PL()
// fmt.Println(a)
// fmt.Println(b)
// prettyPrint(t)

// t := oanda.NewTrades(goq, "", "CLOSED", "EUR_USD", "5", "")
// prettyPrint(t.Latest())

// 成り行き注文してorderIDを取得
// p.Inst = "EUR_USD"
// p.Units = 100
// ch := make(chan string)
// go marketOrder(goq, p, "SELL", ch)
// id := <-ch
// fmt.Println(id)
// prettyPrint(oanda.NewOrderData(goq, id))

// 注文してorderIdを取得
// o := oanda.NewMarketOrder(goq, p.Inst, 1000)
// fmt.Println(o.Id())

// orderIdからorderのstatusを取得
// o := oanda.NewOrderData(goq, "750")
// fmt.Println(o.OrderStatus())

// frame(goq, p)

// routine_test()
// test(goq)

// msg := `(✿✪‿✪｡)ﾉｺﾝﾁｬ♡。私は開発中のオアンダ・ボットです。`
// tweetImage_test(msg, "./test.png")
// msg_test("./test.png")
