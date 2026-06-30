const flow = ['Generator', 'Risk Check', 'FIFO Matching', 'Trades', 'Snapshot', 'UI'];

export default function DemoFlow() {
  return (
    <section className="panel demo-flow-panel">
      <div className="panel-header">
        <h2>Demo Flow</h2>
      </div>
      <div className="flow-strip" aria-label="Live demo flow">
        {flow.map((item, index) => (
          <span key={item}>
            {item}
            {index < flow.length - 1 && <b aria-hidden="true">&rarr;</b>}
          </span>
        ))}
      </div>
    </section>
  );
}

