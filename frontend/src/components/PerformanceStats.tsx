import { useEffect, useState } from 'react';
import type { EngineStats } from '../types';
import { formatLatency, formatNumber } from '../format';

interface Props {
  stats: EngineStats;
}

export default function PerformanceStats({ stats }: Props) {
  const targetRate = stats.generatorTargetRate || stats.generatorRate || 0;
  const targetDenominator = targetRate || 1;
  const displayOrdersPerSecond = capDisplayRate(stats.ordersPerSecond, targetRate);
  const displayTradesPerSecond = capDisplayRate(stats.tradesPerSecond, targetRate);
  const displayMaxTradesPerSecond = capDisplayRate(stats.maxTradesPerSecond, targetRate);
  const isRunning = stats.engineStatus === 'RUNNING';
  const animatedMaxTradesPerSecond = useAnimatedPeak(displayMaxTradesPerSecond, isRunning);

  return (
    <section className="panel performance-panel">
      <div className="panel-header panel-header-strong">
        <h2>Performance</h2>
        <span className={`status-badge ${stats.engineStatus.toLowerCase()}`}>{stats.engineStatus}</span>
      </div>

      <div className="hero-metrics">
        <MetricCard
          label="Orders/sec"
          value={formatNumber(displayOrdersPerSecond)}
          ratio={displayOrdersPerSecond / targetDenominator}
          tone="green"
        />
        <MetricCard
          label="Trades/sec"
          value={formatNumber(displayTradesPerSecond)}
          ratio={displayTradesPerSecond / targetDenominator}
          tone="blue"
        />
      </div>

      <div className="stats-grid compact-stats">
        <StatCell label="Avg Latency" value={formatLatency(stats.avgLatencyMs)} />
        <StatCell label="Generated" value={formatNumber(stats.ordersGenerated)} />
        <StatCell label="Trades" value={formatNumber(stats.tradesExecuted)} />
        <StatCell label="Max Trades/sec" value={isRunning ? formatNumber(animatedMaxTradesPerSecond) : '-'} highlight />
      </div>
    </section>
  );
}

function capDisplayRate(value: number, targetRate: number): number {
  if (targetRate <= 0) {
    return Math.max(0, value);
  }
  return Math.max(0, Math.min(value, targetRate));
}

function useAnimatedPeak(targetValue: number, animate: boolean): number {
  const [displayValue, setDisplayValue] = useState(targetValue);

  useEffect(() => {
    if (!animate) {
      setDisplayValue(0);
      return;
    }

    if (targetValue <= displayValue) {
      setDisplayValue(targetValue);
      return;
    }

    const distance = targetValue - displayValue;
    const step = Math.max(1, Math.ceil(distance / 24));
    const timer = window.setInterval(() => {
      setDisplayValue((current) => {
        if (current >= targetValue) {
          window.clearInterval(timer);
          return targetValue;
        }
        return Math.min(targetValue, current + step);
      });
    }, 80);

    return () => window.clearInterval(timer);
  }, [animate, displayValue, targetValue]);

  return displayValue;
}

function MetricCard({ label, value, ratio, tone }: { label: string; value: string; ratio: number; tone: 'green' | 'blue' }) {
  const percent = Math.max(0, Math.min(100, Math.round(ratio * 100)));

  return (
    <div className={`metric-card ${tone}`}>
      <div className="metric-head">
        <span>{label}</span>
        <strong>{value}</strong>
      </div>
      <div className="metric-bar" aria-label={`${label} target ratio ${percent}%`}>
        <span style={{ width: `${percent}%` }} />
      </div>
      <small>{percent}% of target</small>
    </div>
  );
}

function StatCell({ label, value, highlight = false }: { label: string; value: string; highlight?: boolean }) {
  return (
    <div className={`stat-cell ${highlight ? 'highlight-stat' : ''}`}>
      <span>{label}</span>
      <strong>{value}</strong>
    </div>
  );
}
