package generator

import (
	"errors"
	"math/rand"
	"sync"
	"time"

	"corematch/backend/internal/model"
)

const (
	ModeRandom  = "RANDOM"
	ModeNormal  = "NORMAL"
	ModeBullish = "BULLISH"
	ModeBearish = "BEARISH"

	BasePairPrice          int64 = model.DefaultPairPrice
	PriceRangeCash         int64 = 2 * model.CashScale
	MinPairQuantity        int64 = model.QuantityTick
	MaxPairQuantity        int64 = 100 * model.PairQuantityScale
	StressRateFloor        int   = 100_000
	MaxOrderRate           int   = 100_000
	StressRateBoostPercent int   = 5
	StressMakerLevels      int   = 16
	StressPriceLevels      int64 = 60
)

type SubmitFunc func(side model.Side, price int64, quantity int64)

type Service struct {
	mu      sync.Mutex
	running bool
	rate    int
	stop    chan struct{}
	submit  SubmitFunc
}

func NewService(submit SubmitFunc) *Service {
	return &Service{
		rate:   25000,
		submit: submit,
	}
}

func (s *Service) Start() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return false
	}

	s.stop = make(chan struct{})
	s.running = true
	go s.loop(s.stop)
	return true
}

func (s *Service) Stop() bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.running {
		return false
	}

	close(s.stop)
	s.running = false
	return true
}

func (s *Service) SetMode(mode string) error {
	switch mode {
	case ModeRandom, ModeNormal, ModeBullish, ModeBearish:
		return nil
	default:
		return errors.New("unsupported generator mode")
	}
}

func (s *Service) SetRate(rate int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if rate <= 0 || rate > MaxOrderRate {
		return errors.New("rate must be between 1 and 100000")
	}
	s.rate = rate
	return nil
}

func (s *Service) Status() string {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.running {
		return "RUNNING"
	}
	return "STOPPED"
}

func (s *Service) Mode() string {
	return ModeRandom
}

func (s *Service) Rate() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.rate
}

func (s *Service) loop(stop <-chan struct{}) {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	var carry float64

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			rate := s.config()
			mode := randomMode(rng)
			carry += float64(effectiveRate(rate)) / 100.0
			batchSize := int(carry)
			carry -= float64(batchSize)

			s.submitBatch(rng, mode, rate, batchSize)
		}
	}
}

func (s *Service) submitBatch(rng *rand.Rand, mode string, rate int, batchSize int) {
	if rate >= StressRateFloor {
		s.submitStressBatch(rng, mode, rate, batchSize)
		return
	}

	for i := 0; i < batchSize; i++ {
		side := pickSide(rng, mode)
		price := randomPrice(rng)
		quantity := randomQuantity(rng)
		s.submit(side, price, quantity)
	}
}

func (s *Service) submitStressBatch(rng *rand.Rand, mode string, selectedRate int, batchSize int) {
	if batchSize <= 0 {
		return
	}

	targetTrades := selectedRate / 100
	if targetTrades <= 0 {
		targetTrades = batchSize
	}

	makerLevels := minInt(StressMakerLevels, maxInt(1, batchSize/20))
	tradeOrders := minInt(targetTrades, batchSize-(makerLevels*2))
	if tradeOrders <= 0 {
		return
	}

	buyTakers, sellTakers := splitStressTakers(rng, mode, tradeOrders)
	s.submitStressMakers(rng, model.Sell, makerLevels, stressMakerUnits(rng, buyTakers))
	s.submitStressMakers(rng, model.Buy, makerLevels, stressMakerUnits(rng, sellTakers))

	buyCrossPrice := BasePairPrice + StressPriceLevels*model.PriceTick
	sellCrossPrice := BasePairPrice - StressPriceLevels*model.PriceTick
	for buyTakers > 0 || sellTakers > 0 {
		submitBuy := sellTakers == 0 || (buyTakers > 0 && rng.Intn(buyTakers+sellTakers) < buyTakers)
		if submitBuy {
			s.submit(model.Buy, buyCrossPrice, model.QuantityTick)
			buyTakers--
		} else {
			s.submit(model.Sell, sellCrossPrice, model.QuantityTick)
			sellTakers--
		}
	}
}

func splitStressTakers(rng *rand.Rand, mode string, total int) (int, int) {
	if total <= 0 {
		return 0, 0
	}

	buyTarget := (total * buyPercentForMode(mode)) / 100
	buyJitter := maxInt(1, total/20)
	buyTakers := buyTarget - buyJitter + rng.Intn(buyJitter*2+1)
	if buyTakers < 1 {
		buyTakers = 1
	}
	if buyTakers > total-1 {
		buyTakers = total - 1
	}
	return buyTakers, total - buyTakers
}

func stressMakerUnits(rng *rand.Rand, takerCount int) int {
	if takerCount <= 0 {
		return 1
	}

	minUnits := maxInt(1, (takerCount*88)/100)
	maxUnits := maxInt(minUnits, (takerCount*106)/100)
	return minUnits + rng.Intn(maxUnits-minUnits+1)
}

func (s *Service) submitStressMakers(rng *rand.Rand, side model.Side, levels int, totalUnits int) int {
	if levels <= 0 || totalUnits <= 0 {
		return 0
	}

	if levels > totalUnits {
		levels = totalUnits
	}

	offsets := rng.Perm(int(StressPriceLevels))
	remainingUnits := totalUnits
	submitted := 0

	for i := 0; i < levels; i++ {
		remainingLevels := levels - i
		qtyUnits := remainingUnits
		if remainingLevels > 1 {
			maxUnits := remainingUnits - (remainingLevels - 1)
			average := maxInt(1, maxUnits/remainingLevels)
			upper := minInt(maxUnits, maxInt(1, average*2))
			qtyUnits = 1 + rng.Intn(upper)
		}
		remainingUnits -= qtyUnits

		offsetTicks := int64(offsets[i%len(offsets)] + 1)
		price := BasePairPrice + offsetTicks*model.PriceTick
		if side == model.Buy {
			price = BasePairPrice - offsetTicks*model.PriceTick
		}

		s.submit(side, price, int64(qtyUnits)*model.QuantityTick)
		submitted++
	}

	return submitted
}

func (s *Service) config() int {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.rate
}

func effectiveRate(rate int) int {
	if rate < StressRateFloor {
		return rate
	}
	return rate + (rate*StressRateBoostPercent)/100
}

func minInt(a int, b int) int {
	if a < b {
		return a
	}
	return b
}

func maxInt(a int, b int) int {
	if a > b {
		return a
	}
	return b
}

func randomMode(rng *rand.Rand) string {
	switch rng.Intn(3) {
	case 0:
		return ModeNormal
	case 1:
		return ModeBullish
	default:
		return ModeBearish
	}
}

func pickSide(rng *rand.Rand, mode string) model.Side {
	if rng.Intn(100) < buyPercentForMode(mode) {
		return model.Buy
	}
	return model.Sell
}

func buyPercentForMode(mode string) int {
	buyPercent := 50
	switch mode {
	case ModeBullish:
		buyPercent = 65
	case ModeBearish:
		buyPercent = 35
	}
	return buyPercent
}

func randomPrice(rng *rand.Rand) int64 {
	rangeTicks := PriceRangeCash / model.PriceTick
	offsetTicks := rng.Int63n(rangeTicks*2+1) - rangeTicks
	return BasePairPrice + offsetTicks*model.PriceTick
}

func randomQuantity(rng *rand.Rand) int64 {
	rangeTicks := (MaxPairQuantity - MinPairQuantity) / model.QuantityTick
	return MinPairQuantity + rng.Int63n(rangeTicks+1)*model.QuantityTick
}
