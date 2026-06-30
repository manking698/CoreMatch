package app

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"

	"corematch/backend/internal/account"
	"corematch/backend/internal/generator"
	"corematch/backend/internal/marketdata"
	"corematch/backend/internal/matching"
	"corematch/backend/internal/model"
	"corematch/backend/internal/stats"
)

type App struct {
	Accounts  *account.Service
	Engine    *matching.Engine
	Stats     *stats.Service
	Trades    *marketdata.RecentTrades
	Hub       *marketdata.Hub
	Generator *generator.Service

	nextOrderID atomic.Int64
}

const (
	orderBookSnapshotDepth = 30
	recentTradeLimit       = 30
)

func New() *App {
	a := &App{
		Accounts: account.NewService(1000),
		Engine:   matching.NewEngine(),
		Stats:    stats.NewService(),
		Trades:   marketdata.NewRecentTrades(recentTradeLimit),
		Hub:      marketdata.NewHub(),
	}
	a.Generator = generator.NewService(a.SubmitGeneratedOrder)
	return a
}

func (a *App) SubmitGeneratedOrder(side model.Side, price int64, quantity int64) {
	startedAt := time.Now()
	defer func() {
		a.Stats.ObserveLatency(time.Since(startedAt))
	}()

	a.Stats.IncOrdersGenerated()

	order := &model.Order{
		OrderID:   a.nextOrderID.Add(1),
		UserID:    a.Accounts.RandomUserID(),
		Symbol:    model.PairSymbol,
		Side:      side,
		Price:     price,
		Quantity:  quantity,
		Remaining: quantity,
		CreatedAt: time.Now().UnixMilli(),
	}

	if !a.Accounts.Reserve(order) {
		a.Stats.IncOrdersRejected()
		return
	}

	a.Stats.IncOrdersAccepted()
	executions := a.Engine.ProcessOrder(order)
	for _, execution := range executions {
		if err := a.Accounts.SettleExecution(execution); err != nil {
			log.Panicf("settlement invariant failed: %v", err)
		}
	}

	if len(executions) > 0 {
		a.Stats.AddTrades(int64(len(executions)))
		a.Trades.AddExecutions(executions)
	}
}

func (a *App) RegisterRoutes(mux *http.ServeMux) {
	mux.HandleFunc("/", a.handleRoot)
	mux.HandleFunc("/healthz", a.handleHealthz)
	mux.HandleFunc("/ws", a.handleWS)
	mux.HandleFunc("/api/snapshot", a.handleSnapshot)
	mux.HandleFunc("/api/generator/start", a.handleGeneratorStart)
	mux.HandleFunc("/api/generator/stop", a.handleGeneratorStop)
	mux.HandleFunc("/api/generator/mode", a.handleGeneratorMode)
	mux.HandleFunc("/api/generator/rate", a.handleGeneratorRate)
}

func (a *App) StartSnapshotLoop(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				payload, err := json.Marshal(a.BuildSnapshot())
				if err != nil {
					log.Printf("snapshot encode failed: %v", err)
					continue
				}
				a.Hub.Broadcast(payload)
			}
		}
	}()
}

func (a *App) StartConsoleLoop(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				ordersPerSecond, tradesPerSecond := a.Stats.CurrentRates()
				generated, accepted, rejected, _ := a.Stats.CurrentCounters()
				book := a.Engine.Snapshot(1)
				log.Printf("orders/sec=%d trades/sec=%d generated=%d accepted=%d rejected=%d openBids=%d openAsks=%d",
					ordersPerSecond,
					tradesPerSecond,
					generated,
					accepted,
					rejected,
					book.OpenBidOrders,
					book.OpenAskOrders,
				)
			}
		}
	}()
}

func (a *App) BuildSnapshot() model.MarketSnapshot {
	book := a.Engine.Snapshot(orderBookSnapshotDepth)
	volume := a.Trades.Volume24h()
	return model.MarketSnapshot{
		PairSymbol:       model.PairSymbol,
		ReferencePrice:   model.DefaultPairPrice,
		Last24hTradeCash: volume.Cash,
		Last24hTradeQty:  volume.Qty,
		Bids:             book.Bids,
		Asks:             book.Asks,
		RecentTrades:     a.Trades.Snapshot(),
		Stats: a.Stats.BuildEngineStats(
			book.OpenBidOrders,
			book.OpenAskOrders,
			book.BestBid,
			book.BestAsk,
			a.Generator.Status(),
			a.Generator.Mode(),
			a.Generator.Rate(),
		),
	}
}

func (a *App) handleRoot(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, map[string]string{
		"service":     "CoreMatch backend",
		"status":      "ok",
		"frontend":    "http://127.0.0.1:5173",
		"snapshotApi": "http://127.0.0.1:8080/api/snapshot",
		"websocket":   "ws://127.0.0.1:8080/ws",
	})
}

func (a *App) handleHealthz(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	writeJSON(w, map[string]string{"status": "ok"})
}

func (a *App) handleWS(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := a.Hub.ServeWS(w, r); err != nil {
		log.Printf("websocket upgrade failed: %v", err)
	}
}

func (a *App) handleSnapshot(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	writeJSON(w, a.BuildSnapshot())
}

func (a *App) handleGeneratorStart(w http.ResponseWriter, r *http.Request) {
	if !requirePost(w, r) {
		return
	}
	if a.Generator.Start() {
		a.Stats.ResetMaxTradesPerSecond()
	}
	writeJSON(w, map[string]string{"status": a.Generator.Status()})
}

func (a *App) handleGeneratorStop(w http.ResponseWriter, r *http.Request) {
	if !requirePost(w, r) {
		return
	}
	if a.Generator.Stop() {
		a.Stats.ResetMaxTradesPerSecond()
	}
	writeJSON(w, map[string]string{"status": a.Generator.Status()})
}

func (a *App) handleGeneratorMode(w http.ResponseWriter, r *http.Request) {
	if !requirePost(w, r) {
		return
	}

	var request struct {
		Mode string `json:"mode"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}
	if err := a.Generator.SetMode(request.Mode); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	writeJSON(w, map[string]string{"mode": a.Generator.Mode()})
}

func (a *App) handleGeneratorRate(w http.ResponseWriter, r *http.Request) {
	if !requirePost(w, r) {
		return
	}

	var request struct {
		Rate int `json:"rate"`
	}
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "invalid JSON body", http.StatusBadRequest)
		return
	}

	if request.Rate == 0 {
		if raw := r.URL.Query().Get("rate"); raw != "" {
			parsed, err := strconv.Atoi(raw)
			if err != nil {
				http.Error(w, "invalid rate", http.StatusBadRequest)
				return
			}
			request.Rate = parsed
		}
	}

	if err := a.Generator.SetRate(request.Rate); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	a.Stats.ResetMaxTradesPerSecond()
	writeJSON(w, map[string]int{"rate": request.Rate})
}

func requirePost(w http.ResponseWriter, r *http.Request) bool {
	if r.Method == http.MethodPost {
		return true
	}
	http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	return false
}

func writeJSON(w http.ResponseWriter, value any) {
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(value); err != nil {
		log.Printf("response encode failed: %v", err)
	}
}
