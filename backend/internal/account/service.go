package account

import (
	"fmt"
	"sync"
	"sync/atomic"

	"corematch/backend/internal/model"
)

type Service struct {
	mu       sync.Mutex
	accounts map[string]*model.Account
	userIDs  []string
	nextUser atomic.Uint64
}

func NewService(userCount int) *Service {
	s := &Service{
		accounts: make(map[string]*model.Account, userCount),
		userIDs:  make([]string, 0, userCount),
	}

	for i := 1; i <= userCount; i++ {
		userID := fmt.Sprintf("U%04d", i)
		account := &model.Account{UserID: userID}

		switch i % 3 {
		case 0:
			account.AvailableCash = 10_000_000_000 * model.CashScale
			account.AvailablePair = 1_000 * model.PairQuantityScale
		case 1:
			account.AvailableCash = 1_000_000_000 * model.CashScale
			account.AvailablePair = 20_000 * model.PairQuantityScale
		default:
			account.AvailableCash = 5_000_000_000 * model.CashScale
			account.AvailablePair = 10_000 * model.PairQuantityScale
		}

		s.accounts[userID] = account
		s.userIDs = append(s.userIDs, userID)
	}

	return s
}

func (s *Service) RandomUserID() string {
	index := s.nextUser.Add(1)
	return s.userIDs[int(index%uint64(len(s.userIDs)))]
}

func (s *Service) Reserve(order *model.Order) bool {
	if order.Price <= 0 || order.Quantity <= 0 || order.Remaining <= 0 {
		return false
	}
	if order.Price%model.PriceTick != 0 || order.Quantity%model.QuantityTick != 0 {
		return false
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	account := s.accounts[order.UserID]
	if account == nil {
		return false
	}

	switch order.Side {
	case model.Buy:
		requiredCash := model.CostCash(order.Price, order.Quantity)
		if account.AvailableCash < requiredCash {
			return false
		}
		account.AvailableCash -= requiredCash
		account.LockedCash += requiredCash
		return true
	case model.Sell:
		if account.AvailablePair < order.Quantity {
			return false
		}
		account.AvailablePair -= order.Quantity
		account.LockedPair += order.Quantity
		return true
	default:
		return false
	}
}

func (s *Service) SettleExecution(exec model.Execution) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	trade := exec.Trade
	buyer := s.accounts[trade.BuyerID]
	seller := s.accounts[trade.SellerID]
	if buyer == nil || seller == nil {
		return fmt.Errorf("missing account for trade %d", trade.TradeID)
	}

	tradeCost := model.CostCash(trade.Price, trade.Quantity)
	reservedCost := model.CostCash(exec.BuyLimitPrice, trade.Quantity)
	if reservedCost < tradeCost {
		return fmt.Errorf("buy limit below trade price for trade %d", trade.TradeID)
	}
	if buyer.LockedCash < reservedCost {
		return fmt.Errorf("buyer locked cash underflow for trade %d", trade.TradeID)
	}
	if seller.LockedPair < trade.Quantity {
		return fmt.Errorf("seller locked pair underflow for trade %d", trade.TradeID)
	}

	refund := reservedCost - tradeCost
	buyer.LockedCash -= reservedCost
	buyer.AvailableCash += refund
	buyer.AvailablePair += trade.Quantity

	seller.LockedPair -= trade.Quantity
	seller.AvailableCash += tradeCost

	return nil
}
