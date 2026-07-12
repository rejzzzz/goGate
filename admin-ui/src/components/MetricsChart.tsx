import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

// Custom Tooltip for a premium look
const CustomTooltip = ({ active, payload, label }: any) => {
  if (active && payload && payload.length) {
    return (
      <div style={{
        backgroundColor: 'var(--surface-hover)',
        border: '1px solid var(--border-color)',
        borderRadius: '8px',
        padding: '12px',
        boxShadow: '0 10px 15px -3px rgba(0, 0, 0, 0.3), 0 4px 6px -2px rgba(0, 0, 0, 0.15)',
        backdropFilter: 'blur(4px)'
      }}>
        <p style={{ color: 'var(--text-secondary)', marginBottom: '8px', fontSize: '13px', fontWeight: 500 }}>{label}</p>
        {payload.map((entry: any, index: number) => (
          <div key={index} style={{ display: 'flex', alignItems: 'center', gap: '8px', marginBottom: '4px' }}>
            <div style={{ width: '8px', height: '8px', borderRadius: '50%', backgroundColor: entry.color }}></div>
            <span style={{ color: 'var(--text-primary)', fontSize: '14px', fontWeight: 600 }}>
              {entry.name}: <span style={{ fontWeight: 400 }}>{entry.value}</span>
            </span>
          </div>
        ))}
      </div>
    );
  }
  return null;
};

export default function MetricsChart({ data }: { data: any[] }) {
  return (
    <div className="card" style={{ gridColumn: '1 / -1', height: '420px', padding: '2rem', display: 'flex', flexDirection: 'column' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '2rem' }}>
        <h3 style={{ margin: 0, color: 'var(--text-primary)', fontSize: '1.25rem', fontWeight: 600 }}>
          Traffic Overview
        </h3>
        <span style={{ color: 'var(--text-muted)', fontSize: '0.875rem', fontWeight: 500, backgroundColor: 'var(--surface-color)', padding: '4px 12px', borderRadius: '20px', border: '1px solid var(--border-color)' }}>
          Last 5 Minutes
        </span>
      </div>

      <div style={{ flex: 1, minHeight: 0 }}>
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={data} margin={{ top: 10, right: 10, left: -20, bottom: 0 }}>
            {/* SVG Definitions for Gradients */}
            <defs>
              <linearGradient id="colorRps" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="var(--primary-color)" stopOpacity={0.4} />
                <stop offset="95%" stopColor="var(--primary-color)" stopOpacity={0.0} />
              </linearGradient>
              <linearGradient id="colorLatency" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="var(--warning-color)" stopOpacity={0.4} />
                <stop offset="95%" stopColor="var(--warning-color)" stopOpacity={0.0} />
              </linearGradient>
            </defs>

            <CartesianGrid strokeDasharray="4 4" stroke="var(--border-color)" vertical={false} opacity={0.5} />
            
            <XAxis 
              dataKey="time" 
              stroke="var(--text-muted)" 
              tick={{ fill: 'var(--text-secondary)', fontSize: 12 }} 
              tickLine={false} 
              axisLine={false} 
              dy={10}
            />
            
            <YAxis 
              yAxisId="left" 
              stroke="var(--text-muted)" 
              tick={{ fill: 'var(--text-secondary)', fontSize: 12 }} 
              tickLine={false} 
              axisLine={false} 
            />
            
            <YAxis 
              yAxisId="right" 
              orientation="right" 
              stroke="var(--text-muted)" 
              tick={{ fill: 'var(--text-secondary)', fontSize: 12 }} 
              tickLine={false} 
              axisLine={false} 
            />
            
            <Tooltip content={<CustomTooltip />} cursor={{ stroke: 'var(--border-color)', strokeWidth: 1, strokeDasharray: '4 4' }} />
            
            <Area 
              yAxisId="left" 
              type="monotone" 
              dataKey="rps" 
              stroke="var(--primary-color)" 
              strokeWidth={3} 
              fillOpacity={1} 
              fill="url(#colorRps)" 
              name="Req / Sec" 
              activeDot={{ r: 6, strokeWidth: 0, fill: 'var(--primary-color)' }}
            />
            
            <Area 
              yAxisId="right" 
              type="monotone" 
              dataKey="latency" 
              stroke="var(--warning-color)" 
              strokeWidth={3} 
              fillOpacity={1} 
              fill="url(#colorLatency)" 
              name="P95 Latency" 
              activeDot={{ r: 6, strokeWidth: 0, fill: 'var(--warning-color)' }}
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
