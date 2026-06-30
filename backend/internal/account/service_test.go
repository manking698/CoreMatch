package account

import (
	"testing"

	"corematch/backend/internal/model"
)

func TestReserveRejectsNonTickAlignedOrders(t *testing.T) {
	service := NewService(3)
	order := &model.Order{
		OrderID:   1,
		UserID:    "U0001",
		Symbol:    model.PairSymbol,
		Side:      model.Buy,
		Price:     model.DefaultPairPrice + 1,
		Quantity:  model.QuantityTick,
		Remaining: model.QuantityTick,
	}

	if service.Reserve(order) {
		t.Fatal("expected non-tick-aligned price to be rejected")
	}
}

func TestSettleExecutionRefundsBuyLimitDifference(t *testing.T) {
	service := NewService(3)

	buyOrder := &model.Order{
		OrderID:   1,
		UserID:    "U0003",
		Symbol:    model.PairSymbol,
		Side:      model.Buy,
		Price:     model.DefaultPairPrice + 100*model.PriceTick,
		Quantity:  model.QuantityTick,
		Remaining: model.QuantityTick,
	}
	sellOrder := &model.Order{
		OrderID:   2,
		UserID:    "U0001",
		Symbol:    model.PairSymbol,
		Side:      model.Sell,
		Price:     model.DefaultPairPrice,
		Quantity:  model.QuantityTick,
		Remaining: model.QuantityTick,
	}

	if !service.Reserve(buyOrder) {
		t.Fatal("expected buy reserve to pass")
	}
	if !service.Reserve(sellOrder) {
		t.Fatal("expected sell reserve to pass")
	}

	execution := model.Execution{
		BuyLimitPrice: buyOrder.Price,
		AggressorSide: model.Buy,
		Trade: model.Trade{
			TradeID:     1,
			BuyOrderID:  buyOrder.OrderID,
			SellOrderID: sellOrder.OrderID,
			BuyerID:     buyOrder.UserID,
			SellerID:    sellOrder.UserID,
			Price:       sellOrder.Price,
			Quantity:    model.QuantityTick,
		},
	}

	if err := service.SettleExecution(execution); err != nil {
		t.Fatalf("settlement failed: %v", err)
	}

	buyer := service.accounts[buyOrder.UserID]
	seller := service.accounts[sellOrder.UserID]
	if buyer.LockedCash != 0 {
		t.Fatalf("expected buyer locked cash to be zero, got %d", buyer.LockedCash)
	}
	if buyer.AvailablePair < model.QuantityTick {
		t.Fatalf("expected buyer pair quantity to increase by fill quantity")
	}
	if seller.LockedPair != 0 {
		t.Fatalf("expected seller locked pair to be zero, got %d", seller.LockedPair)
	}
	if seller.AvailableCash == 0 {
		t.Fatal("expected seller cash to increase")
	}
}
