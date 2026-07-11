import { AreaChart, Area, XAxis, YAxis, CartesianGrid, Tooltip, ResponsiveContainer } from 'recharts';

export default function MetricsChart({ data }: { data: any[] }) {
  return (
    <div className="card" style={{ gridColumn: '1 / -1', height: '400px', padding: '2rem', display: 'flex', flexDirection: 'column' }}>
      <h3 style={{ marginBottom: '2rem', color: 'var(--text-primary)', fontSize: '1.25rem', textTransform: 'none' }}>
        Traffic Overview <span style={{color: 'var(--text-secondary)', fontSize: '0.875rem', fontWeight: 400}}>(Last 5 mins)</span>
      </h3>
      <div style={{ flex: 1, minHeight: 0 }}>
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={data} margin={{ top: 10, right: 30, left: 0, bottom: 0 }}>
            <defs>
              <linearGradient id="colorRps" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#3b82f6" stopOpacity={0.5}/>
                <stop offset="95%" stopColor="#3b82f6" stopOpacity={0.0}/>
              </linearGradient>
              <linearGradient id="colorLat" x1="0" y1="0" x2="0" y2="1">
                <stop offset="5%" stopColor="#8b5cf6" stopOpacity={0.5}/>
                <stop offset="95%" stopColor="#8b5cf6" stopOpacity={0.0}/>
              </linearGradient>
            </defs>
            <XAxis dataKey="time" stroke="#8b949e" tick={{fill: '#8b949e', fontSize: 12}} tickLine={false} axisLine={false} />
            <YAxis yAxisId="left" stroke="#8b949e" tick={{fill: '#8b949e', fontSize: 12}} tickLine={false} axisLine={false} />
            <YAxis yAxisId="right" orientation="right" stroke="#8b949e" tick={{fill: '#8b949e', fontSize: 12}} tickLine={false} axisLine={false} />
            <CartesianGrid strokeDasharray="3 3" stroke="rgba(255,255,255,0.05)" vertical={false} />
            <Tooltip 
              contentStyle={{ backgroundColor: 'rgba(22, 27, 34, 0.95)', border: '1px solid rgba(48, 54, 61, 0.8)', borderRadius: '12px', color: '#fff', boxShadow: '0 10px 15px -3px rgba(0,0,0,0.5)' }} 
              itemStyle={{ color: '#fff', fontWeight: 500 }}
              labelStyle={{ color: '#8b949e', marginBottom: '8px' }}
            />
            <Area yAxisId="left" type="monotone" dataKey="rps" stroke="#3b82f6" strokeWidth={3} fillOpacity={1} fill="url(#colorRps)" name="Req / Sec" />
            <Area yAxisId="right" type="monotone" dataKey="latency" stroke="#8b5cf6" strokeWidth={3} fillOpacity={1} fill="url(#colorLat)" name="P95 Latency (ms)" />
          </AreaChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
