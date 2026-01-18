import { ArrowDownOutlined, ArrowRightOutlined, ArrowUpOutlined } from '@ant-design/icons';
import { Card } from 'antd';
import { memo, type CSSProperties, type ReactNode } from 'react';


import './StatCard.css';

interface StatCardProps {
  title: string;
  value: string | number;
  prefix?: ReactNode;
  suffix?: ReactNode;
  icon?: ReactNode;
  trend?: 'up' | 'down' | 'neutral';
  trendValue?: string;
  color?: 'primary' | 'success' | 'warning' | 'error' | 'info';
  className?: string;
  style?: CSSProperties;
  loading?: boolean;
  onClick?: () => void;
}

const colorMap = {
  primary: 'var(--color-primary-500)',
  success: 'var(--color-success)',
  warning: 'var(--color-warning)',
  error: 'var(--color-error)',
  info: 'var(--color-info)',
};

const StatCard = ({
  title,
  value,
  prefix,
  suffix,
  icon,
  trend,
  trendValue,
  color = 'primary',
  className = '',
  style,
  loading = false,
  onClick,
}: StatCardProps) => {
  const colorValue = colorMap[color];

  return (
    <Card
      className={`stat-card ${className}`}
      style={
        {
          ...style,
          cursor: onClick ? 'pointer' : 'default',
          transition: 'all 0.3s ease',
          border: '1px solid var(--color-secondary-200)',
          '--card-accent-color': colorValue,
        } as CSSProperties
      }
      loading={loading}
      onClick={onClick}
    >
      <div className='stat-card-content'>
        <div className='stat-card-header'>
          <span className='stat-card-title'>{title}</span>
          {prefix && <span className='stat-card-prefix'>{prefix}</span>}
        </div>

        <div className='stat-card-body'>
          {icon && <span className='stat-card-icon'>{icon}</span>}
          <span className='stat-card-value' style={{ color: colorValue }}>
            {value}
          </span>
          {suffix && <span className='stat-card-suffix'>{suffix}</span>}
        </div>

        {trend && trendValue && (
          <div className={`stat-card-trend trend-${trend}`}>
            <span className='trend-icon'>
              {trend === 'up' ? <ArrowUpOutlined /> : trend === 'down' ? <ArrowDownOutlined /> : <ArrowRightOutlined />}
            </span>
            <span className='trend-value'>{trendValue}</span>
          </div>
        )}
      </div>

      {/* Animated background effect */}
      <div className='stat-card-bg-effect' />
    </Card>
  );
};

export default memo(StatCard);
