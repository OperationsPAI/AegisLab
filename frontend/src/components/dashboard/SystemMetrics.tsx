import {
  ClockCircleOutlined,
  CloudServerOutlined,
  DatabaseOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import { Col, Progress, Row, Typography } from 'antd';

const { Text } = Typography;

const SystemMetrics = () => {
  const metrics = [
    {
      label: 'CPU Usage',
      value: 68,
      max: 100,
      unit: '%',
      icon: <ThunderboltOutlined />,
      color: '#3b82f6',
      status: 'normal',
    },
    {
      label: 'Memory',
      value: 4.2,
      max: 16,
      unit: 'GB',
      icon: <DatabaseOutlined />,
      color: '#10b981',
      status: 'normal',
    },
    {
      label: 'Storage',
      value: 234,
      max: 500,
      unit: 'GB',
      icon: <CloudServerOutlined />,
      color: '#f59e0b',
      status: 'warning',
    },
    {
      label: 'Queue Time',
      value: 2.3,
      max: 10,
      unit: 's',
      icon: <ClockCircleOutlined />,
      color: '#8b5cf6',
      status: 'normal',
    },
  ];

  return (
    <div className='system-metrics'>
      <h3 className='activity-feed-title'>System Metrics</h3>
      <Row gutter={[24, 24]}>
        {metrics.map((metric) => (
          <Col key={metric.label} xs={24} md={12}>
            <div className='metric-card'>
              <div className='metric-header'>
                <div
                  className='metric-icon'
                  style={{
                    backgroundColor: `${metric.color}20`,
                    color: metric.color,
                  }}
                >
                  {metric.icon}
                </div>
                <div className='metric-info'>
                  <Text className='metric-label'>{metric.label}</Text>
                  <Text className='metric-value' strong>
                    {metric.value} {metric.unit}
                  </Text>
                </div>
              </div>
              <Progress
                percent={(metric.value / metric.max) * 100}
                strokeColor={metric.color}
                showInfo={false}
                status={
                  metric.status as 'normal' | 'exception' | 'active' | 'success'
                }
              />
              <div className='metric-footer'>
                <Text type='secondary' style={{ fontSize: '0.75rem' }}>
                  {metric.max} {metric.unit} max
                </Text>
                <Text
                  style={{
                    fontSize: '0.75rem',
                    color: metric.color,
                    fontWeight: 500,
                  }}
                >
                  {Math.round((metric.value / metric.max) * 100)}%
                </Text>
              </div>
            </div>
          </Col>
        ))}
      </Row>

      <div style={{ marginTop: '2rem' }}>
        <h4 style={{ marginBottom: '1rem', fontSize: '1rem', fontWeight: 600 }}>
          Active Components
        </h4>
        <Row gutter={[16, 16]}>
          <Col span={8}>
            <div className='metric-card'>
              <Text className='metric-label'>Kubernetes Nodes</Text>
              <Text
                className='metric-value'
                strong
                style={{ fontSize: '1.25rem' }}
              >
                12/12
              </Text>
              <Text type='success' style={{ fontSize: '0.75rem' }}>
                All healthy
              </Text>
            </div>
          </Col>
          <Col span={8}>
            <div className='metric-card'>
              <Text className='metric-label'>Active Jobs</Text>
              <Text
                className='metric-value'
                strong
                style={{ fontSize: '1.25rem' }}
              >
                24
              </Text>
              <Text type='secondary' style={{ fontSize: '0.75rem' }}>
                3 pending
              </Text>
            </div>
          </Col>
          <Col span={8}>
            <div className='metric-card'>
              <Text className='metric-label'>Success Rate</Text>
              <Text
                className='metric-value'
                strong
                style={{ fontSize: '1.25rem' }}
              >
                98.2%
              </Text>
              <Text type='success' style={{ fontSize: '0.75rem' }}>
                +0.3% from last week
              </Text>
            </div>
          </Col>
        </Row>
      </div>
    </div>
  );
};

export default SystemMetrics;
