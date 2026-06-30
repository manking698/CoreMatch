package matching

import (
	"sort"

	"corematch/backend/internal/model"
)

type OrderBook struct {
	Bids map[int64][]*model.Order
	Asks map[int64][]*model.Order

	BidPrices []int64
	AskPrices []int64
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		Bids: make(map[int64][]*model.Order),
		Asks: make(map[int64][]*model.Order),
	}
}

func (b *OrderBook) add(order *model.Order) {
	if order.Side == model.Buy {
		if _, ok := b.Bids[order.Price]; !ok {
			b.BidPrices = insertBidPrice(b.BidPrices, order.Price)
		}
		b.Bids[order.Price] = append(b.Bids[order.Price], order)
		return
	}

	if _, ok := b.Asks[order.Price]; !ok {
		b.AskPrices = insertAskPrice(b.AskPrices, order.Price)
	}
	b.Asks[order.Price] = append(b.Asks[order.Price], order)
}

func (b *OrderBook) popBestBid() {
	if len(b.BidPrices) == 0 {
		return
	}
	price := b.BidPrices[0]
	queue := b.Bids[price]
	if len(queue) <= 1 {
		delete(b.Bids, price)
		b.BidPrices = b.BidPrices[1:]
		return
	}
	b.Bids[price] = queue[1:]
}

func (b *OrderBook) popBestAsk() {
	if len(b.AskPrices) == 0 {
		return
	}
	price := b.AskPrices[0]
	queue := b.Asks[price]
	if len(queue) <= 1 {
		delete(b.Asks, price)
		b.AskPrices = b.AskPrices[1:]
		return
	}
	b.Asks[price] = queue[1:]
}

func insertBidPrice(prices []int64, price int64) []int64 {
	index := sort.Search(len(prices), func(i int) bool {
		return prices[i] <= price
	})
	return insertPrice(prices, index, price)
}

func insertAskPrice(prices []int64, price int64) []int64 {
	index := sort.Search(len(prices), func(i int) bool {
		return prices[i] >= price
	})
	return insertPrice(prices, index, price)
}

func insertPrice(prices []int64, index int, price int64) []int64 {
	prices = append(prices, 0)
	copy(prices[index+1:], prices[index:])
	prices[index] = price
	return prices
}
