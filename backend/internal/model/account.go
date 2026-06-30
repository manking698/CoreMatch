package model

const (
	PairSymbol = "NVDA"

	// PairQuantityScale stores pair quantities in micro-units.
	PairQuantityScale int64 = 1_000_000

	// CashScale stores cash balances in micro currency units.
	CashScale int64 = 1_000_000

	// PriceTick is 0.001 cash units.
	PriceTick int64 = CashScale / 1000

	// QuantityTick is 1 whole pair unit.
	QuantityTick int64 = PairQuantityScale

	// DefaultPairPrice is 192.530.
	DefaultPairPrice int64 = 192_530_000
)

type Account struct {
	UserID string

	AvailableCash int64
	LockedCash    int64

	AvailablePair int64
	LockedPair    int64
}

func CostCash(price int64, quantity int64) int64 {
	if price <= 0 || quantity <= 0 {
		return 0
	}
	return (price * quantity) / PairQuantityScale
}
