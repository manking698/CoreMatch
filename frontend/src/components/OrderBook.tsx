import type { CSSProperties } from 'react';
import { formatPrice, formatQuantity } from '../format';
import type { BookLevel } from '../types';

interface Props {
  asks: BookLevel[];
  bids: BookLevel[];
}

type BookSide = 'ask' | 'bid';
const ORDER_BOOK_DEPTH = 30;

export default function OrderBook({ asks, bids }: Props) {
  const visibleBids = bids.slice(0, ORDER_BOOK_DEPTH);
  const visibleAsks = asks.slice(0, ORDER_BOOK_DEPTH);

  return (
    <section className="panel orderbook-panel">
      <div className="panel-header">
        <h2>Order Book</h2>
      </div>

      <div className="book-split">
        <BookSideTable title="Bid" side="bid" levels={visibleBids} />
        <BookSideTable title="Ask" side="ask" levels={visibleAsks} />
      </div>
    </section>
  );
}

function BookSideTable({ title, side, levels }: { title: string; side: BookSide; levels: BookLevel[] }) {
  const maxDepth = Math.max(...levels.map((level) => level.quantity), 0);

  return (
    <div className={`book-side ${side}-side`}>
      <div className={`book-side-title ${side}`}>
        <span>{title}</span>
        <span>{side === 'bid' ? 'Buy Depth' : 'Sell Depth'}</span>
      </div>
      <BookHeader />
      <div className="book-side-rows">
        {levels.map((level) => (
          <BookRow key={`${side}-${level.price}`} level={level} side={side} maxDepth={maxDepth} />
        ))}
      </div>
    </div>
  );
}

function BookHeader() {
  return (
    <div className="book-row book-header">
      <span>Price</span>
      <span>Qty</span>
      <span>Total</span>
    </div>
  );
}

function BookRow({ level, side, maxDepth }: { level: BookLevel; side: BookSide; maxDepth: number }) {
  const depth = maxDepth > 0 ? Math.max(4, Math.min(100, (level.quantity / maxDepth) * 100)) : 0;
  const style = { '--depth': `${depth}%` } as CSSProperties & { '--depth': string };

  return (
    <div className={`book-row depth-row ${side}`} style={style}>
      <span>{formatPrice(level.price)}</span>
      <span>{formatQuantity(level.quantity)}</span>
      <span>{formatQuantity(level.total)}</span>
    </div>
  );
}
