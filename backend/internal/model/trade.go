package model

type Trade struct {
	TradeID     int64
	BuyOrderID  int64
	SellOrderID int64
	BuyerID     string
	SellerID    string
	Price       int64
	Quantity    int64
	CreatedAt   int64
}

type Execution struct {
	Trade         Trade
	BuyLimitPrice int64
	AggressorSide Side
}
