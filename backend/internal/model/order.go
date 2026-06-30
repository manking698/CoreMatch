package model

type Side string

const (
	Buy  Side = "BUY"
	Sell Side = "SELL"
)

type Order struct {
	OrderID   int64
	UserID    string
	Symbol    string
	Side      Side
	Price     int64
	Quantity  int64
	Remaining int64
	CreatedAt int64
}
