export interface BookLevel {
  price: number;
  quantity: number;
  total: number;
}

export interface TradeView {
  tradeId: number;
  price: number;
  quantity: number;
  side: 'BUY' | 'SELL';
  time: number;
}

export interface EngineStats {
  ordersGenerated: number;
  ordersAccepted: number;
  ordersRejected: number;
  tradesExecuted: number;
  ordersPerSecond: number;
  tradesPerSecond: number;
  maxTradesPerSecond: number;
  openBidOrders: number;
  openAskOrders: number;
  bestBid: number;
  bestAsk: number;
  engineStatus: string;
  generatorMode: string;
  generatorRate: number;
  generatorTargetRate: number;
  startedAt: number;
  uptimeSeconds: number;
  avgLatencyMs: number;
}

export interface MarketSnapshot {
  pairSymbol: string;
  referencePrice: number;
  last24hTradeCash: number;
  last24hTradeQty: number;
  bids: BookLevel[];
  asks: BookLevel[];
  recentTrades: TradeView[];
  stats: EngineStats;
}
