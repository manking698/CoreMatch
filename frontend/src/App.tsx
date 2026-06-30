import { useCallback, useEffect, useMemo, useState } from 'react';
import { ArrowLeftRight, Moon, Sun } from 'lucide-react';
import DemoFlow from './components/DemoFlow';
import GeneratorControls from './components/GeneratorControls';
import OrderBook from './components/OrderBook';
import PerformanceStats from './components/PerformanceStats';
import RecentTrades from './components/RecentTrades';
import { formatCash, formatPrice, formatQuantity } from './format';
import type { MarketSnapshot } from './types';

const API_BASE = import.meta.env.VITE_API_BASE ?? 'http://127.0.0.1:8080';
const WS_URL = API_BASE.replace(/^http/, 'ws') + '/ws';

const emptySnapshot: MarketSnapshot = {
  pairSymbol: 'NVDA',
  referencePrice: 192_530_000,
  last24hTradeCash: 0,
  last24hTradeQty: 0,
  bids: [],
  asks: [],
  recentTrades: [],
  stats: {
    ordersGenerated: 0,
    ordersAccepted: 0,
    ordersRejected: 0,
    tradesExecuted: 0,
    ordersPerSecond: 0,
    tradesPerSecond: 0,
    maxTradesPerSecond: 0,
    openBidOrders: 0,
    openAskOrders: 0,
    bestBid: 0,
    bestAsk: 0,
    engineStatus: 'CONNECTING',
    generatorMode: 'RANDOM',
    generatorRate: 25000,
    generatorTargetRate: 25000,
    startedAt: 0,
    uptimeSeconds: 0,
    avgLatencyMs: 0,
  },
};

type ConnectionState = 'CONNECTING' | 'LIVE' | 'DISCONNECTED';
type ThemeMode = 'black' | 'light';

function readInitialTheme(): ThemeMode {
  if (typeof window === 'undefined') return 'black';
  return window.localStorage.getItem('corematch-theme') === 'light' ? 'light' : 'black';
}

export default function App() {
  const [snapshot, setSnapshot] = useState<MarketSnapshot>(emptySnapshot);
  const [connection, setConnection] = useState<ConnectionState>('CONNECTING');
  const [error, setError] = useState<string>('');
  const [theme, setTheme] = useState<ThemeMode>(readInitialTheme);

  useEffect(() => {
    document.documentElement.dataset.theme = theme;
    window.localStorage.setItem('corematch-theme', theme);
  }, [theme]);

  const fetchSnapshot = useCallback(async () => {
    const response = await fetch(`${API_BASE}/api/snapshot`);
    if (!response.ok) {
      throw new Error(`snapshot request failed: ${response.status}`);
    }
    setSnapshot(await response.json());
  }, []);

  useEffect(() => {
    fetchSnapshot().catch((requestError: Error) => setError(requestError.message));
  }, [fetchSnapshot]);

  useEffect(() => {
    let socket: WebSocket | null = null;
    let reconnectTimer = 0;
    let closedByEffect = false;

    const connect = () => {
      setConnection('CONNECTING');
      socket = new WebSocket(WS_URL);

      socket.onopen = () => {
        setConnection('LIVE');
        setError('');
      };

      socket.onmessage = (event) => {
        setSnapshot(JSON.parse(event.data));
      };

      socket.onerror = () => {
        setError('WebSocket connection error');
      };

      socket.onclose = () => {
        if (closedByEffect) return;
        setConnection('DISCONNECTED');
        reconnectTimer = window.setTimeout(connect, 1000);
      };
    };

    connect();

    return () => {
      closedByEffect = true;
      window.clearTimeout(reconnectTimer);
      socket?.close();
    };
  }, []);

  const sendControl = useCallback(
    async (path: string, body?: object) => {
      setError('');
      const response = await fetch(`${API_BASE}${path}`, {
        method: 'POST',
        headers: body ? { 'Content-Type': 'application/json' } : undefined,
        body: body ? JSON.stringify(body) : undefined,
      });

      if (!response.ok) {
        throw new Error(await response.text());
      }
      await fetchSnapshot();
    },
    [fetchSnapshot],
  );

  const lastPrice = useMemo(
    () => snapshot.recentTrades[0]?.price ?? snapshot.referencePrice,
    [snapshot.recentTrades, snapshot.referencePrice],
  );
  const previousPrice = useMemo(
    () => snapshot.recentTrades[1]?.price ?? snapshot.referencePrice,
    [snapshot.recentTrades, snapshot.referencePrice],
  );
  const priceDirection = lastPrice > previousPrice ? 'up' : lastPrice < previousPrice ? 'down' : 'flat';

  return (
    <main className="app-shell">
      <header className="topbar">
        <div>
          <p className="eyebrow">CoreMatch</p>
          <h1 className="brand-title">
            <ArrowLeftRight size={22} aria-hidden="true" />
            <span>In-Memory Matching Engine</span>
          </h1>
        </div>
        <div className={`market-price-center price-${priceDirection}`}>
          <div className="market-price-line">
            <span>{snapshot.pairSymbol}</span>
            <strong>${formatPrice(lastPrice)}</strong>
          </div>
          <div className="market-volume-strip">
            <div>
              <span>24h USD</span>
              <strong>{formatCash(snapshot.last24hTradeCash)}</strong>
            </div>
            <div>
              <span>24h QTY</span>
              <strong>{formatQuantity(snapshot.last24hTradeQty)}</strong>
            </div>
          </div>
        </div>
        <div className="topbar-actions">
          <button
            className="theme-toggle"
            type="button"
            title={`Switch to ${theme === 'black' ? 'light' : 'dark'} mode`}
            onClick={() => setTheme(theme === 'black' ? 'light' : 'black')}
          >
            {theme === 'black' ? <Sun size={16} /> : <Moon size={16} />}
            <span>{theme === 'black' ? 'Light' : 'Dark'}</span>
          </button>
        </div>
      </header>

      {error && <div className="error-banner">{error}</div>}

      <section className="dashboard-grid">
        <div className="left-column">
          <PerformanceStats stats={snapshot.stats} />
          <GeneratorControls
            currentRate={snapshot.stats.generatorTargetRate || snapshot.stats.generatorRate}
            engineStatus={snapshot.stats.engineStatus}
            onCommand={sendControl}
          />
          <DemoFlow />
        </div>
        <OrderBook asks={snapshot.asks} bids={snapshot.bids} />
        <RecentTrades trades={snapshot.recentTrades} />
      </section>
    </main>
  );
}





