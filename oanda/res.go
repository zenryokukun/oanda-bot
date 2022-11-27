/*
 * Oanda API Responseの型を定める
 */
package oanda

const (
	// 想定外の計算値
	CalcError = -1
	// StructのFieldがnil
	MissingError = -2
	// 配列が空
	EmptyError = -3
	// 一致無し
	NoMatch = -4
)

type (
	base struct {
		statusCode int
	}

	// high,low,open. and close prices
	Hloc struct {
		H float64 `json:"h,string"`
		L float64 `json:"l,string"`
		C float64 `json:"c,string"`
		O float64 `json:"o,string"`
	}

	// CandleDataの MId or Ask or BidとTimeをマージしたもの。
	CandleStick struct {
		Complete bool
		Time     string
		Prices   *Hloc
	}

	CandleSticks []CandleStick

	CandleData struct {
		Complete bool `json:"complete"`
		// ******************************
		// request パラメタで
		// price:"B" -> Bidが埋まる
		// price:"A" -> Askが埋まる
		// price:"M" -> Midが埋まる(default)
		Mid *Hloc `json:"mid"`
		Ask *Hloc `json:"ask"`
		Bid *Hloc `json:"bid"`
		// ******************************
		Time   string `json:"time"`
		Volume int32  `json:"volume"`
	}

	// Granularity: "M" "W" "D" "H4" "H1" "M15" "M5" "M1" 等
	// Instrument: "USD_JPY" "EUR_USD" 等
	Candles struct {
		base
		CandleData  []CandleData `json:"candles"`
		Granularity string       `json:"granularity"`
		Instrument  string       `json:"instrument"`
	}

	buckets struct {
		Price        float64 `json:"price,string"`
		LongPercent  float64 `json:"longCountPercent,string"`
		ShortPercent float64 `json:"shortCountPercent,string"`
	}

	Book struct {
		Price       float64   `json:"price,string"`
		BucketWidth float64   `json:"bucketWidth,string"`
		Buckets     []buckets `json:"buckets"`
		UnixTime    int64     `json:"unixTime,string"`
	}

	PositionBook struct {
		base
		Book *Book `json:"positionBook"`
	}

	OrderBook struct {
		base
		Book *Book `json:"orderBook"`
	}

	PositionDataSide struct {
		// 保有中のポジションの枚数。過去の取引量ではない。
		Units int `json:"units,string"`
		// 保有中ポジションの平均取得価格
		Average float64 `json:"averagePrice,string"`
		// 保有中ポジションの取引id
		TradeIDs []string `json:"tradeIDs"`
		// これは何故か全期間損益。保有中のものでな無いので注意
		PL float64 `json:"pl,string"`
		// 未実現損益
		UnrealizedPL float64 `json:"unrealizedPL,string"`
	}

	PositionData struct {
		Instrument string `json:"instrument"`
		// アカウントの累計損益なので注意
		PL float64 `json:"pl,string"`
		// 未実現損益。保有中のポジションの未実現損益
		UnrealizedPL float64 `json:"unrealizedPL,string"`
		// 謎。手数料？でも累計でも0になってる。
		Margin float64 `json:"marginUsed:string"`
		// 謎。手数料？でも累計でも0になってる。
		Commission float64 `json:"commission,string"`

		// position別の累計情報
		Long  *PositionDataSide `json:"long"`
		Short *PositionDataSide `json:"short"`
	}

	Positions struct {
		base
		PositionsData []*PositionData `json:"positions"`
		LastID        string          `json:"lastTransactionID"`
	}

	Position struct {
		base
		PositionData *PositionData `json:"position"`
		LastID       string        `json:"lastTransactionID"`
	}

	transaction struct {
		ID        string `json:"id"`
		Time      string `json:"time"`
		BatchID   string `json:"batchID"`
		RequestID string `json:"requestID"`
	}

	marketCreateTransaction struct {
		transaction
		Instrument string `json:"instrument"`
		Units      int    `json:"units,string"`
		Reason     string `json:"reason"`
	}

	fillTransaction struct {
		transaction
		Type           string     `json:"type"`
		OrderID        string     `json:"orderID"`
		Instrument     string     `json:"instrument"`
		Units          int        `json:"units,string"`
		Reason         string     `json:"reason"`
		PL             float64    `json:"pl,string"`
		Commission     float64    `json:"commission,string"`
		Opened         tradeOpen  `json:"tradeOpened"`
		Closed         tradeClose `json:"tradeClosed"`
		AccountBalance float64    `json:"accountBalance,string"`
	}

	cancelTransaction struct {
		transaction
		Type    string `json:"type"`
		OrderID string `json:"orderID"`
		Reason  string `json:"reason"`
	}

	tradeOpen struct {
	}
	tradeClose struct {
	}

	// 取引注文時のレスポンス
	// POST: v3/accounts/{accountID}/orders
	Orders struct {
		base
		Transaction       transaction       `json:"orderCreateTransaction"`
		FillTransaction   fillTransaction   `json:"orderFillTransaction"`
		CancelTransaction cancelTransaction `json:"orderCancelTransaction"`
		LastID            string            `json:"lastTransactionID"`
	}

	// クローズ処理時のレスポンス
	// CloseOrders struct {
	// 	base
	// 	LongCreateTransaction  marketCreateTransaction `json:"longOrderCreateTransaction"`
	// 	ShortCreateTransaction marketCreateTransaction `json:"shortOrderCreateTransaction"`
	// 	LongFillTransaction    fillTransaction         `json:"longOrderFillTransaction"`
	// 	ShortFillTransaction   fillTransaction         `json:"shortOrderFillTransaction"`
	// 	LongCancelTransaction  cancelTransaction       `json:"longOrderCancelTransaction "`
	// 	ShortCancelTransaction cancelTransaction       `json:"shortOrderCancelTransaction"`
	// 	LastID                 string                  `json:"lastTransactionID"`
	// }

	// Get: v3/accounts/{accountID}/order/{orderSpecifier}
	GetOrder struct {
		base
		Data   *OrderData `json:"order"`
		LastID string     `json:"lastTransactionID"`
	}
	// 注文データ

	OrderData struct {
		Id          string `json:"id"`
		CreatedTime string `json:"createdTime"`
		// PENDING,FILLED,TRIGGERED,CANCELLED
		State string `json:"state"`
	}

	AccountData struct {
		// 証拠金維持率と思われる
		MarginRate        float64 `json:"marginRate,string"`
		MarginUsed        float64 `json:"marginUsed,string"`
		OpenTradeCount    int     `json:"openTradeCount"`
		OpenPositionCount int     `json:"openPositionCount"`
		PendingOrderCount int     `json:"pendingOrderCount"`
		// 総未実現利益
		UnrealizedPL float64 `json:"unrealizedPL,string"`
		// 総利益
		PL float64 `json:"pl,string"`
		// 残高
		Balance    float64        `json:"balance,string"`
		Commission float64        `json:"commission,string"`
		Positions  []PositionData `json:"positions"`
		Orders     []OrderData    `json:"orders"`
	}

	Account struct {
		base
		Data   *AccountData `json:"account"`
		LastID string       `json:"lastTransactionID"`
	}

	Ticker struct {
		Price     float64 `json:"price,string"`
		Liquidity int64   `json:"liquidity"`
	}

	Price struct {
		Time       string   `json:"time"`
		Instrument string   `json:"instrument"`
		Bids       []Ticker `json:"bids"`
		Asks       []Ticker `json:"asks"`
	}

	Pricing struct {
		base
		Time string `json:"time"`
		// "USD_JPY,EUR_USD"のように複数通貨指定できる。
		// 通貨単位でPriceが埋まる。
		Prices []Price `json:"prices"`
	}

	TradeData struct {
		ID           string  `json:"id"`
		Instrument   string  `json:"instrument"`
		Price        float64 `json:"price,string"`
		OpenTime     string  `json:"openTime"`
		CloseTime    string  `json:"closeTime"`
		State        string  `json:"state"`
		InitialUnits int     `json:"initialUnits,string"`
		CurrentUnits int     `json:"currentUnits,string"`
		UnrealizedPL float64 `json:"unrealizedPL,string"`
		RealizedPL   float64 `json:"realizedPL,string"`
	}

	Trade struct {
		base
		LastID    string     `json:"lastTransactionID"`
		TradeData *TradeData `json:"trade"`
	}

	Trades struct {
		base
		LastID    string       `json:"lastTransactionID"`
		TradeData []*TradeData `json:"trades"`
	}
)

type Checker interface {
	Check() bool
	Status(int)
}

func (b *base) Status(code int) {
	b.statusCode = code
}

func (b *base) Check() bool {
	return b.statusCode >= 200 && b.statusCode <= 299
}

func (acc *Account) Extract() *AccountData {
	if !acc.Check() {
		return nil
	}
	return acc.Data
}

func (c *Candles) Extract() []CandleData {
	if !c.Check() {
		return nil
	}
	return c.CandleData
}

func (c *Candles) ExtractMid() CandleSticks {
	data := c.Extract()
	if data == nil || len(data) == 0 {
		return nil
	}
	sticks := []CandleStick{}
	for _, d := range data {
		if d.Mid == nil {
			continue
		}
		stick := CandleStick{
			Complete: d.Complete,
			Prices:   d.Mid,
			Time:     d.Time,
		}
		sticks = append(sticks, stick)
	}
	return sticks
}

func (c CandleSticks) Complete() CandleSticks {
	newSticks := CandleSticks{}
	for _, s := range c {
		if s.Complete == true {
			newSticks = append(newSticks, s)
		}
	}
	return newSticks
}

func (s CandleSticks) Extract(hloc string) []float64 {
	vals := []float64{}
	for _, s := range s {
		val := 0.0
		switch hloc {
		case "L":
			val = s.Prices.L
		case "H":
			val = s.Prices.H
		case "O":
			val = s.Prices.O
		case "C":
			val = s.Prices.C
		default:
			val = 0.0
		}
		vals = append(vals, val)
	}
	return vals
}

func (p *Position) Extract() *PositionData {
	if !p.Check() {
		return nil
	}
	return p.PositionData
}

// Long,Shortいずれかを保有しているか
func (p *PositionData) Has() bool {
	if p.Long == nil || p.Short == nil {
		return false
	}
	if p.Long.Units > 0 || p.Short.Units > 0 {
		return true
	}
	return false
}

// Long,Shortの保有ポジションを返す。
// 両建て出来ないアカウントが前提（両建て不可）
// LongがあればLongを、ShortがあればShortを返す。
func (p *PositionData) Side() *PositionDataSide {
	if p.Long == nil || p.Short == nil {
		return nil
	}
	if p.Long.Units > 0 {
		return p.Long
	}
	if p.Short.Units > 0 {
		return p.Short
	}
	return nil
}

//　保有ポジションを返す。決済の取引量を決めるため。
func (p *PositionData) Units() int {
	if !p.Has() {
		return 0
	}
	if p.Long.Units > 0 {
		return p.Long.Units
	}
	return p.Short.Units
}

// trade idsを取得
func (p *PositionData) Ids() string {
	side := p.Side()
	if side.TradeIDs == nil {
		return ""
	}
	ids := ""
	for _, id := range side.TradeIDs {
		ids += id + ","
	}
	return ids[:len(ids)-1]
}

// Spread計算。ask-bid。片道の取引は(ask-bid)/2になるはずだが、考慮しない。
func (p *Price) Spread() float64 {
	ask, bid := p.Latest()
	if ask == EmptyError || bid == EmptyError {
		return EmptyError
	}
	return ask - bid
}

// askとbidの中央値を返す
func (p *Price) Mid() float64 {
	ask, bid := p.Latest()
	if ask == EmptyError || bid == EmptyError {
		return EmptyError
	}
	return (ask + bid) / 2
}

// 直近のaskとbidを、この順番で返す
func (p *Price) Latest() (float64, float64) {
	if len(p.Asks) == 0 || len(p.Bids) == 0 {
		return EmptyError, EmptyError
	}
	lastAsk := p.Asks[len(p.Asks)-1]
	lastBid := p.Bids[len(p.Bids)-1]
	return lastAsk.Price, lastBid.Price
}

// Pricingは複数の通貨の指定が出来きる。instrumentでfilterする。
// マッチしない場合はnilを返す
func (p *Pricing) filter(instrument string) *Price {
	for _, p := range p.Prices {
		if p.Instrument == instrument {
			return &p
		}
	}
	return nil
}

// Pricingからspreadを計算して返す
// 計算できないときはMissingErrorかEmptyErrorを返す
func (p *Pricing) Spread(instrument string) float64 {
	if !p.Check() {
		return MissingError
	}
	if len(p.Prices) == 0 {
		return EmptyError
	}
	price := p.filter(instrument)
	if price == nil {
		return NoMatch
	}
	return price.Spread()
}

// 直近の*Priceを返す
func (p *Pricing) Latest(instrument string) *Price {
	if !p.Check() {
		return nil
	}
	if len(p.Prices) == 0 {
		return nil
	}
	price := p.filter(instrument)
	if price == nil {
		return nil
	}
	return price
}

func (o *Orders) Id() string {
	if !o.Check() {
		return ""
	}
	// o.Transactionはvalueなのでnilにはならない。なのでチェックしない
	return o.Transaction.ID
}

func (o *OrderData) OrderStatus() string {
	return o.State
}

func (g *GetOrder) OrderStatus() string {
	if !g.Check() || g.Data == nil {
		return ""
	}
	return g.Data.OrderStatus()
}

func (t *Trades) Extract() []*TradeData {
	if !t.Check() {
		return nil
	}
	return t.TradeData
}

// 実現損益、評価損益を返す
func (t *Trades) PL() (float64, float64) {
	td := t.Extract()
	realized := 0.0
	unrealized := 0.0
	if td == nil {
		return realized, unrealized
	}
	for _, d := range td {
		realized += d.RealizedPL
		unrealized += d.UnrealizedPL
	}
	return realized, unrealized
}

// 保有量を返す
func (t *Trades) Units() int {
	td := t.Extract()
	units := 0
	if td == nil {
		return units
	}
	for _, d := range td {
		units += d.CurrentUnits
	}
	return units
}

// 直近のTradeを返す
// TradesをAPIで取得する際に、stateやinstrumentで絞っておくこと
// 直近取引の結果としてみる場合、注文がFILLEDになっていることを確かめておくこと
func (t *Trades) Latest() *TradeData {
	trades := t.Extract()
	if trades == nil {
		return nil
	}
	return trades[0]
}
