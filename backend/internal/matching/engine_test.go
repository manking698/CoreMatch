package matching

import (
	"testing"

	"corematch/backend/internal/model"
)

func TestEngineMatchesFIFOWithinSamePrice(t *testing.T) {
	engine := NewEngine()

	firstSell := testOrder(1, "seller-1", model.Sell, model.DefaultPairPrice, 2*model.QuantityTick, 1)
	secondSell := testOrder(2, "seller-2", model.Sell, model.DefaultPairPrice, 2*model.QuantityTick, 2)
	buy := testOrder(3, "buyer", model.Buy, model.DefaultPairPrice, 3*model.QuantityTick, 3)

	if trades := engine.ProcessOrder(firstSell); len(trades) != 0 {
		t.Fatalf("expected first maker order to rest, got %d trades", len(trades))
	}
	if trades := engine.ProcessOrder(secondSell); len(trades) != 0 {
		t.Fatalf("expected second maker order to rest, got %d trades", len(trades))
	}

	executions := engine.ProcessOrder(buy)
	if len(executions) != 2 {
		t.Fatalf("expected 2 executions, got %d", len(executions))
	}
	if executions[0].Trade.SellOrderID != firstSell.OrderID {
		t.Fatalf("expected first FIFO order to match first, got sell order %d", executions[0].Trade.SellOrderID)
	}
	if executions[1].Trade.SellOrderID != secondSell.OrderID {
		t.Fatalf("expected second FIFO order to match second, got sell order %d", executions[1].Trade.SellOrderID)
	}
	if executions[1].Trade.Quantity != model.QuantityTick {
		t.Fatalf("expected partial fill on second order, got quantity %d", executions[1].Trade.Quantity)
	}
}

func TestEngineMatchesBestPriceBeforeWorsePrice(t *testing.T) {
	engine := NewEngine()

	worseAsk := testOrder(1, "seller-1", model.Sell, model.DefaultPairPrice+100*model.PriceTick, model.QuantityTick, 1)
	bestAsk := testOrder(2, "seller-2", model.Sell, model.DefaultPairPrice-100*model.PriceTick, model.QuantityTick, 2)
	buy := testOrder(3, "buyer", model.Buy, model.DefaultPairPrice+100*model.PriceTick, model.QuantityTick, 3)

	engine.ProcessOrder(worseAsk)
	engine.ProcessOrder(bestAsk)

	executions := engine.ProcessOrder(buy)
	if len(executions) != 1 {
		t.Fatalf("expected 1 execution, got %d", len(executions))
	}
	if executions[0].Trade.SellOrderID != bestAsk.OrderID {
		t.Fatalf("expected best ask to match first, got sell order %d", executions[0].Trade.SellOrderID)
	}
	if executions[0].Trade.Price != bestAsk.Price {
		t.Fatalf("expected maker price %d, got %d", bestAsk.Price, executions[0].Trade.Price)
	}
}

func testOrder(orderID int64, userID string, side model.Side, price int64, quantity int64, createdAt int64) *model.Order {
	return &model.Order{
		OrderID:   orderID,
		UserID:    userID,
		Symbol:    model.PairSymbol,
		Side:      side,
		Price:     price,
		Quantity:  quantity,
		Remaining: quantity,
		CreatedAt: createdAt,
	}
}
