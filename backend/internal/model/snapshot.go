package model

type BookLevel struct {
	Price    int64 `json:"price"`
	Quantity int64 `json:"quantity"`
	Total    int64 `json:"total"`
}

type TradeView struct {
	TradeID  int64  `json:"tradeId"`
	Price    int64  `json:"price"`
	Quantity int64  `json:"quantity"`
	Side     string `json:"side"`
	Time     int64  `json:"time"`
}

type EngineStats struct {
	OrdersGenerated int64 `json:"ordersGenerated"`
	OrdersAccepted  int64 `json:"ordersAccepted"`
	OrdersRejected  int64 `json:"ordersRejected"`
	TradesExecuted  int64 `json:"tradesExecuted"`

	OrdersPerSecond int64 `json:"ordersPerSecond"`
	TradesPerSecond int64 `json:"tradesPerSecond"`
	MaxTradesPerSecond int64 `json:"maxTradesPerSecond"`

	OpenBidOrders int64 `json:"openBidOrders"`
	OpenAskOrders int64 `json:"openAskOrders"`

	BestBid int64 `json:"bestBid"`
	BestAsk int64 `json:"bestAsk"`

	EngineStatus        string  `json:"engineStatus"`
	GeneratorMode       string  `json:"generatorMode"`
	GeneratorRate       int     `json:"generatorRate"`
	GeneratorTargetRate int64   `json:"generatorTargetRate"`
	StartedAt           int64   `json:"startedAt"`
	UptimeSeconds       int64   `json:"uptimeSeconds"`
	AvgLatencyMs        float64 `json:"avgLatencyMs"`
}

type MarketSnapshot struct {
	PairSymbol       string      `json:"pairSymbol"`
	ReferencePrice   int64       `json:"referencePrice"`
	Last24hTradeCash int64       `json:"last24hTradeCash"`
	Last24hTradeQty  int64       `json:"last24hTradeQty"`
	Bids             []BookLevel `json:"bids"`
	Asks             []BookLevel `json:"asks"`
	RecentTrades     []TradeView `json:"recentTrades"`
	Stats            EngineStats `json:"stats"`
}
