import { Badge, Tag } from 'antd';
import type { CSSProperties } from 'react';


interface StatusBadgeProps {
  status:
    | 'pending'
    | 'running'
    | 'completed'
    | 'error'
    | 'warning'
    | 'info'
    | 'success'
    | 'default';
  text?: string;
  size?: 'small' | 'default';
  showDot?: boolean;
  style?: CSSProperties;
  className?: string;
}

const statusConfig: Record<
  StatusBadgeProps['status'],
  { color: string; text: string; icon: string; pulse?: boolean }
> = {
  pending: {
    color: '#f59e0b',
    text: 'Pending',
    icon: '⏳',
  },
  running: {
    color: '#3b82f6',
    text: 'Running',
    icon: '⚡',
    pulse: true,
  },
  completed: {
    color: '#10b981',
    text: 'Completed',
    icon: '✅',
  },
  error: {
    color: '#ef4444',
    text: 'Error',
    icon: '❌',
  },
  warning: {
    color: '#f59e0b',
    text: 'Warning',
    icon: '⚠️',
  },
  info: {
    color: '#06b6d4',
    text: 'Info',
    icon: 'ℹ️',
  },
  success: {
    color: '#10b981',
    text: 'Success',
    icon: '✓',
  },
  default: {
    color: '#6b7280',
    text: 'Default',
    icon: '•',
  },
};

const StatusBadge = ({
  status,
  text,
  size = 'default',
  showDot = true,
  style,
  className = '',
}: StatusBadgeProps) => {
  const config = statusConfig[status];
  const displayText = text || config.text;

  if (showDot) {
    return (
      <Badge
        color={config.color}
        text={
          <span
            className={`status-badge-text ${className}`}
            style={{
              color: config.color,
              fontSize: size === 'small' ? '0.875rem' : '1rem',
              fontWeight: 500,
              ...style,
            }}
          >
            {config.icon} {displayText}
          </span>
        }
      />
    );
  }

  return (
    <Tag
      color={config.color}
      style={{
        borderRadius: '9999px',
        fontSize: size === 'small' ? '0.75rem' : '0.875rem',
        padding: size === 'small' ? '0.125rem 0.5rem' : '0.25rem 0.75rem',
        border: `1px solid ${config.color}40`,
        backgroundColor: `${config.color}10`,
        color: config.color,
        fontWeight: 500,
        ...style,
      }}
      className={`status-tag ${status === 'running' ? 'pulse' : ''} ${className}`}
    >
      {config.icon} {displayText}
    </Tag>
  );
};

export default StatusBadge;
export type { StatusBadgeProps };
