package stats

import (
	"sync"
	"sync/atomic"
	"time"

	"corematch/backend/internal/model"
)

type Service struct {
	ordersGenerated atomic.Int64
	ordersAccepted  atomic.Int64
	ordersRejected  atomic.Int64
	tradesExecuted  atomic.Int64

	ordersPerSecond    atomic.Int64
	tradesPerSecond    atomic.Int64
	maxTradesPerSecond atomic.Int64

	latencyTotalNs atomic.Int64
	latencySamples atomic.Int64

	startedAt      time.Time
	startedAtMilli int64

	mu             sync.Mutex
	lastSampleAt   time.Time
	lastOrderCount int64
	lastTradeCount int64
}

func NewService() *Service {
	now := time.Now()
	return &Service{
		startedAt:      now,
		startedAtMilli: now.UnixMilli(),
		lastSampleAt:   now,
	}
}

func (s *Service) IncOrdersGenerated() {
	s.ordersGenerated.Add(1)
}

func (s *Service) IncOrdersAccepted() {
	s.ordersAccepted.Add(1)
}

func (s *Service) IncOrdersRejected() {
	s.ordersRejected.Add(1)
}

func (s *Service) AddTrades(count int64) {
	s.tradesExecuted.Add(count)
}

func (s *Service) ObserveLatency(duration time.Duration) {
	if duration <= 0 {
		return
	}
	s.latencyTotalNs.Add(duration.Nanoseconds())
	s.latencySamples.Add(1)
}

func (s *Service) ResetMaxTradesPerSecond() {
	s.maxTradesPerSecond.Store(0)
}

func (s *Service) BuildEngineStats(openBidOrders int64, openAskOrders int64, bestBid int64, bestAsk int64, engineStatus string, generatorMode string, generatorRate int) model.EngineStats {
	targetRate := int64(generatorRate)
	s.updateRates(targetRate, engineStatus == "RUNNING")

	latencySamples := s.latencySamples.Load()
	var avgLatencyMs float64
	if latencySamples > 0 {
		avgLatencyMs = float64(s.latencyTotalNs.Load()) / float64(latencySamples) / float64(time.Millisecond)
	}

	maxTradesPerSecond := s.maxTradesPerSecond.Load()
	if targetRate > 0 && maxTradesPerSecond > targetRate {
		maxTradesPerSecond = targetRate
	}

	return model.EngineStats{
		OrdersGenerated: s.ordersGenerated.Load(),
		OrdersAccepted:  s.ordersAccepted.Load(),
		OrdersRejected:  s.ordersRejected.Load(),
		TradesExecuted:  s.tradesExecuted.Load(),

		OrdersPerSecond:    s.ordersPerSecond.Load(),
		TradesPerSecond:    s.tradesPerSecond.Load(),
		MaxTradesPerSecond: maxTradesPerSecond,

		OpenBidOrders: openBidOrders,
		OpenAskOrders: openAskOrders,

		BestBid: bestBid,
		BestAsk: bestAsk,

		EngineStatus:        engineStatus,
		GeneratorMode:       generatorMode,
		GeneratorRate:       generatorRate,
		GeneratorTargetRate: int64(generatorRate),
		StartedAt:           s.startedAtMilli,
		UptimeSeconds:       int64(time.Since(s.startedAt).Seconds()),
		AvgLatencyMs:        avgLatencyMs,
	}
}

func (s *Service) CurrentRates() (int64, int64) {
	return s.ordersPerSecond.Load(), s.tradesPerSecond.Load()
}

func (s *Service) CurrentCounters() (int64, int64, int64, int64) {
	return s.ordersGenerated.Load(), s.ordersAccepted.Load(), s.ordersRejected.Load(), s.tradesExecuted.Load()
}

func (s *Service) updateRates(targetTradeCap int64, trackMax bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	elapsed := now.Sub(s.lastSampleAt).Seconds()
	if elapsed <= 0 {
		return
	}

	orderCount := s.ordersGenerated.Load()
	tradeCount := s.tradesExecuted.Load()
	currentOrdersPerSecond := int64(float64(orderCount-s.lastOrderCount) / elapsed)
	currentTradesPerSecond := int64(float64(tradeCount-s.lastTradeCount) / elapsed)
	s.ordersPerSecond.Store(currentOrdersPerSecond)
	s.tradesPerSecond.Store(currentTradesPerSecond)
	if trackMax {
		cappedTradesPerSecond := currentTradesPerSecond
		if targetTradeCap > 0 && cappedTradesPerSecond > targetTradeCap {
			cappedTradesPerSecond = targetTradeCap
		}
		if cappedTradesPerSecond > s.maxTradesPerSecond.Load() {
			s.maxTradesPerSecond.Store(cappedTradesPerSecond)
		}
	}

	s.lastSampleAt = now
	s.lastOrderCount = orderCount
	s.lastTradeCount = tradeCount
}
