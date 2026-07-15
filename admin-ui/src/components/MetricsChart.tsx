import Chart from 'react-apexcharts';
import type { ApexOptions } from 'apexcharts';

export default function MetricsChart({ 
  data, 
  timeWindow, 
  onTimeWindowChange 
}: { 
  data: any[]; 
  timeWindow?: string; 
  onTimeWindowChange?: (window: string) => void;
}) {
  const series = [
    {
      name: 'Req / Sec',
      type: 'area',
      data: data.map(d => d.rps)
    },
    {
      name: 'P95 Latency',
      type: 'area',
      data: data.map(d => d.latency)
    }
  ];

  const categories = data.map(d => d.time);

  const options: ApexOptions = {
    chart: {
      type: 'area',
      height: '100%',
      fontFamily: 'Inter, system-ui, sans-serif',
      toolbar: { show: false },
      background: 'transparent',
      animations: {
        enabled: true,
        speed: 800,
        dynamicAnimation: {
          enabled: true,
          speed: 350
        }
      }
    },
    colors: ['#4f46e5', '#f59e0b'], // primary and warning colors
    dataLabels: {
      enabled: false
    },
    stroke: {
      curve: 'smooth',
      width: [3, 3]
    },
    fill: {
      type: 'gradient',
      gradient: {
        shadeIntensity: 1,
        inverseColors: false,
        opacityFrom: 0.45,
        opacityTo: 0.05,
        stops: [20, 100]
      },
    },
    xaxis: {
      categories: categories,
      labels: {
        style: {
          colors: '#9ca3af',
          fontSize: '12px'
        },
      },
      axisBorder: { show: false },
      axisTicks: { show: false },
      tickAmount: 8,
    },
    yaxis: [
      {
        title: {
          text: 'Requests / Sec',
          style: { color: '#9ca3af', fontWeight: 500 }
        },
        labels: {
          style: { colors: '#9ca3af' },
          formatter: (val) => val.toFixed(0)
        }
      },
      {
        opposite: true,
        title: {
          text: 'P95 Latency (ms)',
          style: { color: '#9ca3af', fontWeight: 500 }
        },
        labels: {
          style: { colors: '#9ca3af' },
          formatter: (val) => val.toFixed(1)
        }
      }
    ],
    grid: {
      borderColor: 'rgba(255, 255, 255, 0.1)',
      strokeDashArray: 4,
      xaxis: { lines: { show: false } },
      yaxis: { lines: { show: true } }
    },
    legend: {
      position: 'top',
      horizontalAlign: 'right',
      labels: { colors: '#d1d5db' },
      offsetY: 0,
    },
    tooltip: {
      theme: 'dark',
      y: {
        formatter: function (val: number, opts?: any) {
          if (opts && opts.seriesIndex === 0) return val.toFixed(0);
          return val.toFixed(1) + " ms";
        }
      }
    }
  };

  const timeWindows = ['5m', '15m', '30m', '1h', '24h'];

  return (
    <div className="card" style={{ gridColumn: '1 / -1', height: '450px', padding: '1.5rem 2rem', display: 'flex', flexDirection: 'column', position: 'relative' }}>
      <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center', marginBottom: '1rem', position: 'relative', zIndex: 10 }}>
        <h3 style={{ margin: 0, color: 'var(--text-primary)', fontSize: '1.25rem', fontWeight: 600 }}>
          Traffic Overview
        </h3>
        
        {timeWindow && onTimeWindowChange && (
          <div style={{ display: 'flex', gap: '0.25rem', background: 'var(--surface-color)', padding: '4px', borderRadius: '8px', border: '1px solid var(--border-color)' }}>
            {timeWindows.map(tw => (
              <button
                key={tw}
                onClick={() => onTimeWindowChange(tw)}
                style={{
                  background: timeWindow === tw ? 'var(--primary-color)' : 'transparent',
                  color: timeWindow === tw ? '#fff' : 'var(--text-secondary)',
                  border: 'none',
                  padding: '4px 12px',
                  borderRadius: '6px',
                  fontSize: '0.875rem',
                  fontWeight: timeWindow === tw ? 600 : 500,
                  cursor: 'pointer',
                  transition: 'all 0.2s'
                }}
              >
                {tw}
              </button>
            ))}
          </div>
        )}
      </div>

      <div style={{ flex: 1, minHeight: 0, marginLeft: '-10px', marginTop: '-10px' }}>
        <Chart options={options} series={series} type="area" height="100%" />
      </div>
    </div>
  );
}
