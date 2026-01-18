
import type { EChartsOption } from 'echarts';
import ReactECharts from 'echarts-for-react';
import { memo, type CSSProperties } from 'react';

interface LabChartProps {
  option: EChartsOption;
  style?: CSSProperties;
  className?: string;
  showLoading?: boolean;
  onEvents?: Record<string, (...args: unknown[]) => void>;
  theme?: 'light' | 'dark';
}

const LabChart = ({
  option,
  style = { height: '400px' },
  className = '',
  showLoading = false,
  onEvents = {},
  theme,
}: LabChartProps) => {
  // Base chart configuration for AegisLab theme
  const baseOption: EChartsOption = {
    backgroundColor: 'transparent',
    textStyle: {
      fontFamily: 'var(--font-family-body)',
      color: 'var(--color-secondary-700)',
    },
    title: {
      textStyle: {
        fontFamily: 'var(--font-family-display)',
        fontWeight: 600,
        color: 'var(--color-secondary-800)',
      },
      subtextStyle: {
        color: 'var(--color-secondary-500)',
      },
    },
    legend: {
      textStyle: {
        color: 'var(--color-secondary-600)',
      },
    },
    tooltip: {
      backgroundColor: 'var(--color-secondary-800)',
      borderColor: 'var(--color-secondary-700)',
      textStyle: {
        color: 'var(--color-secondary-100)',
      },
      axisPointer: {
        type: 'shadow',
        shadowStyle: {
          color: 'rgba(0, 0, 0, 0.1)',
        },
      },
    },
    grid: {
      left: 60,
      right: 20,
      bottom: 30,
      top: 60,
      containLabel: true,
    },
    xAxis: {
      axisLine: {
        lineStyle: {
          color: 'var(--color-secondary-300)',
        },
      },
      axisLabel: {
        color: 'var(--color-secondary-600)',
      },
      splitLine: {
        lineStyle: {
          color: 'var(--color-secondary-200)',
        },
      },
    },
    yAxis: {
      axisLine: {
        lineStyle: {
          color: 'var(--color-secondary-300)',
        },
      },
      axisLabel: {
        color: 'var(--color-secondary-600)',
      },
      splitLine: {
        lineStyle: {
          color: 'var(--color-secondary-200)',
        },
      },
    },
    series: {
      itemStyle: {
        borderRadius: [2, 2, 0, 0],
      },
    },
    color: [
      'var(--color-primary-500)',
      'var(--color-success)',
      'var(--color-warning)',
      'var(--color-error)',
      'var(--color-info)',
      '#8b5cf6',
      '#ec4899',
      '#f97316',
    ],
  };

  // Dark theme overrides
  const darkThemeOverrides: EChartsOption = {
    textStyle: {
      color: 'var(--color-secondary-300)',
    },
    title: {
      textStyle: {
        color: 'var(--color-secondary-100)',
      },
      subtextStyle: {
        color: 'var(--color-secondary-400)',
      },
    },
    legend: {
      textStyle: {
        color: 'var(--color-secondary-400)',
      },
    },
    tooltip: {
      backgroundColor: 'var(--color-secondary-900)',
      borderColor: 'var(--color-secondary-700)',
      textStyle: {
        color: 'var(--color-secondary-200)',
      },
    },
    xAxis: {
      axisLine: {
        lineStyle: {
          color: 'var(--color-secondary-600)',
        },
      },
      axisLabel: {
        color: 'var(--color-secondary-400)',
      },
      splitLine: {
        lineStyle: {
          color: 'var(--color-secondary-800)',
        },
      },
    },
    yAxis: {
      axisLine: {
        lineStyle: {
          color: 'var(--color-secondary-600)',
        },
      },
      axisLabel: {
        color: 'var(--color-secondary-400)',
      },
      splitLine: {
        lineStyle: {
          color: 'var(--color-secondary-800)',
        },
      },
    },
    color: [
      'var(--color-primary-400)',
      '#34d399',
      '#fbbf24',
      '#f87171',
      '#22d3ee',
      '#a78bfa',
      '#f472b6',
      '#fb923c',
    ],
  };

  // Merge options
  const mergedOption = {
    ...baseOption,
    ...(theme === 'dark' && darkThemeOverrides),
    ...option,
  };

  return (
    <ReactECharts
      option={mergedOption}
      style={style}
      className={`lab-chart ${className}`}
      showLoading={showLoading}
      onEvents={onEvents}
      opts={{ renderer: 'svg' }}
    />
  );
};

export default memo(LabChart);
