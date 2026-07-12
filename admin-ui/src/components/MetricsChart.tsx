import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

export default function MetricsChart({ data }: { data: any[] }) {
  return (
    <div className="card" style={{ gridColumn: '1 / -1', height: '400px', padding: '2rem', display: 'flex', flexDirection: 'column' }}>
      <h3 style={{ marginBottom: '1.5rem', color: 'var(--text-primary)' }}>
        Traffic Overview <span style={{color: 'var(--text-muted)', fontWeight: 400}}>(Last 5 mins)</span>
      </h3>
      <div style={{ flex: 1, minHeight: 0 }}>
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={data} margin={{ top: 10, right: 30, left: 0, bottom: 0 }}>
            <XAxis dataKey="time" stroke="var(--border-color)" tick={{fill: 'var(--text-secondary)', fontSize: 12}} tickLine={false} axisLine={false} />
            <YAxis yAxisId="left" stroke="var(--border-color)" tick={{fill: 'var(--text-secondary)', fontSize: 12}} tickLine={false} axisLine={false} />
            <YAxis yAxisId="right" orientation="right" stroke="var(--border-color)" tick={{fill: 'var(--text-secondary)', fontSize: 12}} tickLine={false} axisLine={false} />
            <CartesianGrid strokeDasharray="3 3" stroke="var(--border-color)" vertical={false} />
            <Tooltip 
              contentStyle={{ backgroundColor: 'var(--surface-hover)', border: '1px solid var(--border-color)', borderRadius: '6px', color: 'var(--text-primary)' }} 
              itemStyle={{ color: 'var(--text-primary)', fontWeight: 500 }}
              labelStyle={{ color: 'var(--text-secondary)', marginBottom: '4px' }}
            />
            <Area yAxisId="left" type="step" dataKey="rps" stroke="var(--text-primary)" strokeWidth={2} fill="none" name="Req / Sec" />
            <Area yAxisId="right" type="step" dataKey="latency" stroke="var(--text-secondary)" strokeWidth={2} fill="none" name="P95 Latency (ms)" />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
