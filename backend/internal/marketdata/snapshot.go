package marketdata

import (
	"sync"
	"time"

	"corematch/backend/internal/model"
)

const volumeWindowSeconds int64 = 24 * 60 * 60

type volumeBucket struct {
	second int64
	cash   int64
	qty    int64
}

type VolumeSnapshot struct {
	Cash int64
	Qty  int64
}

type RecentTrades struct {
	mu      sync.RWMutex
	limit   int
	trades  []model.TradeView
	buckets []volumeBucket
}

func NewRecentTrades(limit int) *RecentTrades {
	return &RecentTrades{
		limit:   limit,
		trades:  make([]model.TradeView, 0, limit),
		buckets: make([]volumeBucket, volumeWindowSeconds),
	}
}

func (r *RecentTrades) AddExecutions(executions []model.Execution) {
	if len(executions) == 0 {
		return
	}

	start := 0
	if len(executions) > r.limit {
		start = len(executions) - r.limit
	}

	newTrades := make([]model.TradeView, 0, len(executions)-start)
	nowSecond := time.Now().Unix()
	bucketUpdates := make(map[int64]VolumeSnapshot, min(len(executions), 64))

	for _, execution := range executions {
		trade := execution.Trade
		second := trade.CreatedAt / 1000
		if second <= 0 {
			second = nowSecond
		}
		update := bucketUpdates[second]
		update.Cash += model.CostCash(trade.Price, trade.Quantity)
		update.Qty += trade.Quantity
		bucketUpdates[second] = update
	}

	for i := len(executions) - 1; i >= start; i-- {
		trade := executions[i].Trade
		newTrades = append(newTrades, model.TradeView{
			TradeID:  trade.TradeID,
			Price:    trade.Price,
			Quantity: trade.Quantity,
			Side:     string(executions[i].AggressorSide),
			Time:     trade.CreatedAt,
		})
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.trades = append(newTrades, r.trades...)
	if len(r.trades) > r.limit {
		r.trades = r.trades[:r.limit]
	}

	for second, update := range bucketUpdates {
		r.addVolumeLocked(second, update.Cash, update.Qty)
	}
}

func (r *RecentTrades) Snapshot() []model.TradeView {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]model.TradeView, len(r.trades))
	copy(result, r.trades)
	return result
}

func (r *RecentTrades) Volume24h() VolumeSnapshot {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cutoff := time.Now().Unix() - volumeWindowSeconds + 1
	var snapshot VolumeSnapshot
	for _, bucket := range r.buckets {
		if bucket.second >= cutoff {
			snapshot.Cash += bucket.cash
			snapshot.Qty += bucket.qty
		}
	}
	return snapshot
}

func (r *RecentTrades) addVolumeLocked(second int64, cash int64, qty int64) {
	index := second % volumeWindowSeconds
	bucket := &r.buckets[index]
	if bucket.second != second {
		bucket.second = second
		bucket.cash = 0
		bucket.qty = 0
	}
	bucket.cash += cash
	bucket.qty += qty
}
