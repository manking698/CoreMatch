const PAIR_QUANTITY_SCALE = 1_000_000;
const CASH_SCALE = 1_000_000;

export function formatPrice(value: number): string {
  if (!value) return '-';
  return (value / CASH_SCALE).toLocaleString(undefined, {
    minimumFractionDigits: 3,
    maximumFractionDigits: 3,
  });
}

export function formatCash(value: number): string {
  if (!value) return '0.00';
  return (value / CASH_SCALE).toLocaleString(undefined, {
    minimumFractionDigits: 2,
    maximumFractionDigits: 2,
  });
}

export function formatQuantity(value: number): string {
  if (!value) return '0';
  return (value / PAIR_QUANTITY_SCALE).toLocaleString(undefined, {
    minimumFractionDigits: 0,
    maximumFractionDigits: 0,
  });
}

export function formatNumber(value: number): string {
  return value.toLocaleString();
}

export function formatLatency(value: number): string {
  return `${value.toFixed(2)} ms`;
}

export function formatUptime(totalSeconds: number): string {
  const safeSeconds = Math.max(0, Math.floor(totalSeconds));
  const hours = Math.floor(safeSeconds / 3600);
  const minutes = Math.floor((safeSeconds % 3600) / 60);
  const seconds = safeSeconds % 60;
  return [hours, minutes, seconds].map((part) => part.toString().padStart(2, '0')).join(':');
}

export function formatTime(value: number): string {
  if (!value) return '-';
  return new Date(value).toLocaleTimeString();
}
