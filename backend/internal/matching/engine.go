package matching

import (
	"sync"
	"sync/atomic"
	"time"

	"corematch/backend/internal/model"
)

type BookSnapshot struct {
	Bids          []model.BookLevel
	Asks          []model.BookLevel
	OpenBidOrders int64
	OpenAskOrders int64
	BestBid       int64
	BestAsk       int64
}

type Engine struct {
	mu            sync.RWMutex
	book          *OrderBook
	nextTradeID   atomic.Int64
	openBidOrders int64
	openAskOrders int64
}

func NewEngine() *Engine {
	return &Engine{book: NewOrderBook()}
}

func (e *Engine) ProcessOrder(order *model.Order) []model.Execution {
	e.mu.Lock()
	defer e.mu.Unlock()

	if order.Side == model.Buy {
		return e.processBuy(order)
	}
	return e.processSell(order)
}

func (e *Engine) processBuy(incoming *model.Order) []model.Execution {
	executions := make([]model.Execution, 0, 4)

	for incoming.Remaining > 0 && len(e.book.AskPrices) > 0 {
		bestAsk := e.book.AskPrices[0]
		if incoming.Price < bestAsk {
			break
		}

		maker := e.book.Asks[bestAsk][0]
		tradeQty := minInt64(incoming.Remaining, maker.Remaining)
		trade := model.Trade{
			TradeID:     e.nextTradeID.Add(1),
			BuyOrderID:  incoming.OrderID,
			SellOrderID: maker.OrderID,
			BuyerID:     incoming.UserID,
			SellerID:    maker.UserID,
			Price:       maker.Price,
			Quantity:    tradeQty,
			CreatedAt:   time.Now().UnixMilli(),
		}

		executions = append(executions, model.Execution{
			Trade:         trade,
			BuyLimitPrice: incoming.Price,
			AggressorSide: model.Buy,
		})

		incoming.Remaining -= tradeQty
		maker.Remaining -= tradeQty
		if maker.Remaining == 0 {
			e.book.popBestAsk()
			e.openAskOrders--
		}
	}

	if incoming.Remaining > 0 {
		e.book.add(incoming)
		e.openBidOrders++
	}

	return executions
}

func (e *Engine) processSell(incoming *model.Order) []model.Execution {
	executions := make([]model.Execution, 0, 4)

	for incoming.Remaining > 0 && len(e.book.BidPrices) > 0 {
		bestBid := e.book.BidPrices[0]
		if incoming.Price > bestBid {
			break
		}

		maker := e.book.Bids[bestBid][0]
		tradeQty := minInt64(incoming.Remaining, maker.Remaining)
		trade := model.Trade{
			TradeID:     e.nextTradeID.Add(1),
			BuyOrderID:  maker.OrderID,
			SellOrderID: incoming.OrderID,
			BuyerID:     maker.UserID,
			SellerID:    incoming.UserID,
			Price:       maker.Price,
			Quantity:    tradeQty,
			CreatedAt:   time.Now().UnixMilli(),
		}

		executions = append(executions, model.Execution{
			Trade:         trade,
			BuyLimitPrice: maker.Price,
			AggressorSide: model.Sell,
		})

		incoming.Remaining -= tradeQty
		maker.Remaining -= tradeQty
		if maker.Remaining == 0 {
			e.book.popBestBid()
			e.openBidOrders--
		}
	}

	if incoming.Remaining > 0 {
		e.book.add(incoming)
		e.openAskOrders++
	}

	return executions
}

func (e *Engine) Snapshot(limit int) BookSnapshot {
	e.mu.RLock()
	defer e.mu.RUnlock()

	snapshot := BookSnapshot{
		Bids:          buildLevels(e.book.BidPrices, e.book.Bids, limit),
		Asks:          buildLevels(e.book.AskPrices, e.book.Asks, limit),
		OpenBidOrders: e.openBidOrders,
		OpenAskOrders: e.openAskOrders,
	}

	if len(e.book.BidPrices) > 0 {
		snapshot.BestBid = e.book.BidPrices[0]
	}
	if len(e.book.AskPrices) > 0 {
		snapshot.BestAsk = e.book.AskPrices[0]
	}

	return snapshot
}

func buildLevels(prices []int64, levels map[int64][]*model.Order, limit int) []model.BookLevel {
	if len(prices) < limit {
		limit = len(prices)
	}

	result := make([]model.BookLevel, 0, limit)
	var cumulative int64
	for i := 0; i < limit; i++ {
		price := prices[i]
		var quantity int64
		for _, order := range levels[price] {
			quantity += order.Remaining
		}
		cumulative += quantity
		result = append(result, model.BookLevel{
			Price:    price,
			Quantity: quantity,
			Total:    cumulative,
		})
	}
	return result
}

func minInt64(a int64, b int64) int64 {
	if a < b {
		return a
	}
	return b
}
