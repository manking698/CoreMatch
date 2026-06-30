import { Pause, Play } from 'lucide-react';

interface Props {
  currentRate: number;
  engineStatus: string;
  onCommand: (path: string, body?: object) => Promise<void>;
}

const rates = [50, 100, 1000, 5000, 25000, 50000, 100000];

export default function GeneratorControls({ currentRate, engineStatus, onCommand }: Props) {
  const isRunning = engineStatus === 'RUNNING';

  return (
    <section className="panel generator-panel">
      <div className="panel-header">
        <h2>Generator</h2>
      </div>

      <div className="generator-summary">
        <div className="generator-target-row">
          <span>Target</span>
          <div className="target-actions">
            <select
              className="target-select"
              value={currentRate}
              onChange={(event) => onCommand('/api/generator/rate', { rate: Number(event.target.value) })}
            >
              {rates.map((rate) => (
                <option key={rate} value={rate}>
                  {rate.toLocaleString()} orders/sec
                </option>
              ))}
            </select>
            <button
              type="button"
              className={`control-button primary ${isRunning ? '' : 'selected'}`}
              onClick={() => onCommand(isRunning ? '/api/generator/stop' : '/api/generator/start')}
            >
              {isRunning ? <Pause size={16} /> : <Play size={16} />}
              <span>{isRunning ? 'Stop' : 'Start'}</span>
            </button>
          </div>
        </div>
      </div>
    </section>
  );
}
