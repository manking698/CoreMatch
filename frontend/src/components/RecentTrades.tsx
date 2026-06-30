import { formatPrice, formatQuantity, formatTime } from '../format';
import type { TradeView } from '../types';

interface Props {
  trades: TradeView[];
}

const RECENT_TRADE_LIMIT = 30;

export default function RecentTrades({ trades }: Props) {
  const visibleTrades = trades.slice(0, RECENT_TRADE_LIMIT);

  return (
    <section className="panel trades-panel">
      <div className="panel-header">
        <h2>Recent Trades</h2>
        <span className="panel-caption">Latest 30</span>
      </div>

      <div className="trades-table">
        <div className="trade-row trade-header">
          <span>Time</span>
          <span>Price</span>
          <span>Qty</span>
          <span>Side</span>
        </div>
        {visibleTrades.map((trade) => (
          <div className={`trade-row ${trade.side.toLowerCase()}`} key={trade.tradeId}>
            <span>{formatTime(trade.time)}</span>
            <span>{formatPrice(trade.price)}</span>
            <span>{formatQuantity(trade.quantity)}</span>
            <span>{trade.side}</span>
          </div>
        ))}
      </div>
    </section>
  );
}
