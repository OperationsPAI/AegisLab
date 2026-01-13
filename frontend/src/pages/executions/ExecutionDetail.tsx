
import {
  ArrowLeftOutlined,
  BarChartOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
  DatabaseOutlined,
  DownloadOutlined,
  EyeOutlined,
  FunctionOutlined,
  SyncOutlined,
  TagsOutlined,
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import {
  Badge,
  Button,
  Card,
  Col,
  Descriptions,
  Divider,
  Empty,
  message,
  Progress,
  Row,
  Space,
  Table,
  Tabs,
  Tag,
  Typography,
} from 'antd';
import dayjs from 'dayjs';
import duration from 'dayjs/plugin/duration';
import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';

import { executionApi } from '@/api/executions';
import StatusBadge from '@/components/ui/StatusBadge';
import type { RankResult, ExecutionState } from '@/types/api';

dayjs.extend(duration);

const { Title, Text } = Typography;
// Removed deprecated TabPane destructuring - using items prop instead

const ExecutionDetail = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const executionId = Number(id);
  const [activeTab, setActiveTab] = useState('overview');

  // Fetch execution details
  const { data: execution, isLoading } = useQuery({
    queryKey: ['execution', executionId],
    queryFn: () => executionApi.getExecution(executionId),
    enabled: !!executionId,
  });

  const handleDownloadResults = () => {
    // TODO: Implement download logic
    message.info('Download functionality will be implemented soon');
  };

  const handleViewGranularity = (type: string, _results: RankResult[]) => {
    // TODO: Implement detailed view
    message.info(`View ${type} granularity results`);
  };

  const formatDuration = (seconds?: number) => {
    if (!seconds) return '-';
    const d = dayjs.duration(seconds, 'seconds');
    if (d.asHours() >= 1) {
      return `${Math.floor(d.asHours())}h ${d.minutes()}m ${d.seconds()}s`;
    } else if (d.asMinutes() >= 1) {
      return `${d.minutes()}m ${d.seconds()}s`;
    } else {
      return `${d.seconds()}s`;
    }
  };

  const getStateColor = (state: ExecutionState) => {
    switch (state) {
      case ExecutionState.PENDING:
        return '#d1d5db';
      case ExecutionState.RUNNING:
        return '#3b82f6';
      case ExecutionState.COMPLETED:
        return '#10b981';
      case ExecutionState.ERROR:
        return '#ef4444';
      default:
        return '#6b7280';
    }
  };

  const getStateIcon = (state: ExecutionState) => {
    switch (state) {
      case ExecutionState.PENDING:
        return <ClockCircleOutlined />;
      case ExecutionState.RUNNING:
        return <SyncOutlined spin />;
      case ExecutionState.COMPLETED:
        return <CheckCircleOutlined />;
      case ExecutionState.ERROR:
        return <CloseCircleOutlined />;
      default:
        return <ClockCircleOutlined />;
    }
  };

  // Detector Results Table
  const detectorColumns = [
    {
      title: 'Span Name',
      dataIndex: 'span_name',
      key: 'span_name',
      width: '25%',
    },
    {
      title: 'Anomaly Type',
      dataIndex: 'anomaly_type',
      key: 'anomaly_type',
      width: '15%',
      render: (type: string) => (
        <Tag color={type === 'latency' ? 'orange' : 'red'}>{type}</Tag>
      ),
    },
    {
      title: 'Normal Avg Latency',
      dataIndex: 'normal_avg_latency',
      key: 'normal_avg_latency',
      width: '15%',
      render: (value: number) => `${value.toFixed(2)}ms`,
    },
    {
      title: 'Abnormal Avg Latency',
      dataIndex: 'abnormal_avg_latency',
      key: 'abnormal_avg_latency',
      width: '15%',
      render: (value: number) => `${value.toFixed(2)}ms`,
    },
    {
      title: 'Normal Success Rate',
      dataIndex: 'normal_success_rate',
      key: 'normal_success_rate',
      width: '15%',
      render: (value: number) => `${(value * 100).toFixed(1)}%`,
    },
    {
      title: 'Abnormal Success Rate',
      dataIndex: 'abnormal_success_rate',
      key: 'abnormal_success_rate',
      width: '15%',
      render: (value: number) => `${(value * 100).toFixed(1)}%`,
    },
  ];

  // Granularity Results Table
  const granularityColumns = [
    {
      title: 'Rank',
      dataIndex: 'rank',
      key: 'rank',
      width: '10%',
      render: (rank: number) => (
        <Badge
          count={rank}
          style={{
            backgroundColor:
              rank === 1 ? '#10b981' : rank === 2 ? '#f59e0b' : '#6b7280',
          }}
        />
      ),
    },
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      width: '40%',
    },
    {
      title: 'Confidence',
      dataIndex: 'confidence',
      key: 'confidence',
      width: '20%',
      render: (confidence: number) => (
        <Progress
          percent={confidence * 100}
          size='small'
          format={(percent) => `${percent.toFixed(1)}%`}
        />
      ),
    },
    {
      title: 'Ground Truth',
      dataIndex: 'is_ground_truth',
      key: 'is_ground_truth',
      width: '15%',
      render: (isGroundTruth: boolean) => (
        <Tag color={isGroundTruth ? 'green' : 'default'}>
          {isGroundTruth ? 'Yes' : 'No'}
        </Tag>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: '15%',
      render: (_: string, record: RankResult, type: string) => (
        <Button
          type='link'
          icon={<EyeOutlined />}
          onClick={() => handleViewGranularity(type, [record])}
        >
          View
        </Button>
      ),
    },
  ];

  if (isLoading) {
    return (
      <div style={{ padding: 24 }}>
        <Card loading>
          <div style={{ minHeight: 400 }} />
        </Card>
      </div>
    );
  }

  if (!execution) {
    return (
      <div style={{ padding: 24, textAlign: 'center' }}>
        <Text type='secondary'>Execution not found</Text>
      </div>
    );
  }

  const executionData = execution;
  const progress =
    executionData.state === ExecutionState.COMPLETED
      ? 100
      : executionData.state === ExecutionState.ERROR
        ? 0
        : executionData.state === ExecutionState.RUNNING
          ? 50
          : 0;

  return (
    <div style={{ padding: 24 }}>
      {/* Header */}
      <div style={{ marginBottom: 24 }}>
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/executions')}
          >
            Back to List
          </Button>
          <Title level={2} style={{ margin: 0 }}>
            Execution #{executionData.id}
          </Title>
          <Badge
            status={
              executionData.state === ExecutionState.COMPLETED
                ? 'success'
                : executionData.state === ExecutionState.ERROR
                  ? 'error'
                  : executionData.state === ExecutionState.RUNNING
                    ? 'processing'
                    : 'default'
            }
            text={
              <Space>
                {getStateIcon(executionData.state)}
                <Text
                  strong
                  style={{ color: getStateColor(executionData.state) }}
                >
                  {executionData.state === ExecutionState.PENDING
                    ? 'Pending'
                    : executionData.state === ExecutionState.RUNNING
                      ? 'Running'
                      : executionData.state === ExecutionState.COMPLETED
                        ? 'Completed'
                        : executionData.state === ExecutionState.ERROR
                          ? 'Error'
                          : 'Unknown'}
                </Text>
              </Space>
            }
          />
        </Space>
      </div>

      {/* Actions */}
      <Card style={{ marginBottom: 24 }}>
        <Row justify='space-between' align='middle'>
          <Col>
            <Space>
              <Button
                type='primary'
                icon={<DownloadOutlined />}
                onClick={handleDownloadResults}
                disabled={executionData.state !== ExecutionState.COMPLETED}
              >
                Download Results
              </Button>
              <Button
                icon={<EyeOutlined />}
                onClick={() => {
                  // TODO: View logs
                  message.info('Log viewing will be implemented soon');
                }}
              >
                View Logs
              </Button>
            </Space>
          </Col>
          <Col>
            <Text type='secondary'>
              Duration: {formatDuration(executionData.execution_duration)}
            </Text>
          </Col>
        </Row>
      </Card>

      {/* Progress */}
      <Card style={{ marginBottom: 24 }}>
        <div style={{ marginBottom: 16 }}>
          <Text strong>Execution Progress</Text>
        </div>
        <Progress
          percent={progress}
          status={
            executionData.state === ExecutionState.ERROR
              ? 'exception'
              : executionData.state === ExecutionState.COMPLETED
                ? 'success'
                : 'active'
          }
          strokeColor={getStateColor(executionData.state)}
          format={(percent) => (
            <Space>
              {getStateIcon(executionData.state)}
              <Text>{percent}%</Text>
            </Space>
          )}
        />
      </Card>

      {/* Tabs */}
      <Tabs activeKey={activeTab} onChange={setActiveTab} items={[
        {
          key: 'overview',
          label: 'Overview',
          children: (
            <Row gutter={[16, 16]}>
              <Col xs={24} lg={16}>
                <Card title='Execution Information'>
                  <Descriptions column={2} bordered>
                    <Descriptions.Item label='Execution ID'>
                      {executionData.id}
                    </Descriptions.Item>
                    <Descriptions.Item label='Algorithm'>
                      <Space>
                        <FunctionOutlined style={{ color: '#f59e0b' }} />
                        <Text strong>
                          {executionData.algorithm?.name || 'Unknown'}
                        </Text>
                      </Space>
                    </Descriptions.Item>
                    <Descriptions.Item label='Algorithm Version'>
                      <Tag color='blue'>v{executionData.algorithm_version}</Tag>
                    </Descriptions.Item>
                    <Descriptions.Item label='Datapack'>
                      <Space>
                        <DatabaseOutlined style={{ color: '#3b82f6' }} />
                        <Text code>
                          {executionData.datapack_id?.substring(0, 8)}
                        </Text>
                      </Space>
                    </Descriptions.Item>
                    <Descriptions.Item label='Status'>
                      <StatusBadge
                        status={
                          executionData.state === ExecutionState.COMPLETED
                            ? 'success'
                            : executionData.state === ExecutionState.ERROR
                              ? 'error'
                              : executionData.state === ExecutionState.RUNNING
                                ? 'processing'
                                : 'default'
                        }
                        text={
                          executionData.state === ExecutionState.PENDING
                            ? 'Pending'
                            : executionData.state === ExecutionState.RUNNING
                              ? 'Running'
                              : executionData.state === ExecutionState.COMPLETED
                                ? 'Completed'
                                : executionData.state === ExecutionState.ERROR
                                  ? 'Error'
                                  : 'Unknown'
                        }
                      />
                    </Descriptions.Item>
                    <Descriptions.Item label='Duration'>
                      <Text code>
                        {formatDuration(executionData.execution_duration)}
                      </Text>
                    </Descriptions.Item>
                    <Descriptions.Item label='Created'>
                      <Space>
                        <ClockCircleOutlined />
                        {dayjs(executionData.created_at).format(
                          'MMMM D, YYYY HH:mm:ss'
                        )}
                      </Space>
                    </Descriptions.Item>
                    <Descriptions.Item label='Updated'>
                      <Space>
                        <ClockCircleOutlined />
                        {dayjs(executionData.updated_at).format(
                          'MMMM D, YYYY HH:mm:ss'
                        )}
                      </Space>
                    </Descriptions.Item>
                  </Descriptions>
                </Card>
              </Col>
              <Col xs={24} lg={8}>
                <Card title='Quick Stats'>
                  <Space direction='vertical' style={{ width: '100%' }}>
                    <div>
                      <Text type='secondary'>Algorithm</Text>
                      <br />
                      <Title level={4} style={{ margin: 0, color: '#f59e0b' }}>
                        {executionData.algorithm?.name || 'Unknown'}
                      </Title>
                    </div>
                    <Divider />
                    <div>
                      <Text type='secondary'>Datapack ID</Text>
                      <br />
                      <Text code style={{ fontSize: '0.875rem' }}>
                        {executionData.datapack_id?.substring(0, 16)}
                      </Text>
                    </div>
                    <Divider />
                    <div>
                      <Text type='secondary'>Labels</Text>
                      <br />
                      {executionData.labels?.length ? (
                        <Space wrap>
                          {executionData.labels.map((label, index) => (
                            <Tag key={index} icon={<TagsOutlined />}>
                              {label.key}: {label.value}
                            </Tag>
                          ))}
                        </Space>
                      ) : (
                        <Text type='secondary'>No labels</Text>
                      )}
                    </div>
                  </Space>
                </Card>
              </Col>
            </Row>
          )
        },
        {
          key: 'detector',
          label: 'Detector Results',
          children: (
            <Card
              title='Anomaly Detection Results'
              extra={
                <Button
                  icon={<DownloadOutlined />}
                  onClick={handleDownloadResults}
                  disabled={executionData.state !== ExecutionState.COMPLETED}
                >
                  Export
                </Button>
              }
            >
              {executionData.detector_results?.length ? (
                <Table
                  rowKey='span_name'
                  columns={detectorColumns}
                  dataSource={executionData.detector_results}
                  pagination={{
                    pageSize: 10,
                    showSizeChanger: true,
                    showQuickJumper: true,
                  }}
                />
              ) : (
                <Empty description='No detector results available' />
              )}
            </Card>
          )
        },
        {
          key: 'granularity',
          label: 'Granularity Results',
          children: (
            <Space direction='vertical' style={{ width: '100%' }} size='large'>
              {executionData.granularity_results?.service_results && (
                <Card
                  title='Service Level Results'
                  extra={
                    <Button
                      icon={<BarChartOutlined />}
                      onClick={() =>
                        handleViewGranularity(
                          'service',
                          executionData.granularity_results?.service_results || []
                        )
                      }
                    >
                      View Chart
                    </Button>
                  }
                >
                  <Table
                    rowKey='name'
                    columns={granularityColumns}
                    dataSource={executionData.granularity_results.service_results}
                    pagination={false}
                  />
                </Card>
              )}

              {executionData.granularity_results?.pod_results && (
                <Card
                  title='Pod Level Results'
                  extra={
                    <Button
                      icon={<BarChartOutlined />}
                      onClick={() =>
                        handleViewGranularity(
                          'pod',
                          executionData.granularity_results?.pod_results || []
                        )
                      }
                    >
                      View Chart
                    </Button>
                  }
                >
                  <Table
                    rowKey='name'
                    columns={granularityColumns}
                    dataSource={executionData.granularity_results.pod_results}
                    pagination={false}
                  />
                </Card>
              )}

              {executionData.granularity_results?.span_results && (
                <Card
                  title='Span Level Results'
                  extra={
                    <Button
                      icon={<BarChartOutlined />}
                      onClick={() =>
                        handleViewGranularity(
                          'span',
                          executionData.granularity_results?.span_results || []
                        )
                      }
                    >
                      View Chart
                    </Button>
                  }
                >
                  <Table
                    rowKey='name'
                    columns={granularityColumns}
                    dataSource={executionData.granularity_results.span_results}
                    pagination={false}
                  />
                </Card>
              )}

              {executionData.granularity_results?.metric_results && (
                <Card
                  title='Metric Level Results'
                  extra={
                    <Button
                      icon={<BarChartOutlined />}
                      onClick={() =>
                        handleViewGranularity(
                          'metric',
                          executionData.granularity_results?.metric_results || []
                        )
                      }
                    >
                      View Chart
                    </Button>
                  }
                >
                  <Table
                    rowKey='name'
                    columns={granularityColumns}
                    dataSource={executionData.granularity_results.metric_results}
                    pagination={false}
                  />
                </Card>
              )}

              {!executionData.granularity_results && (
                <Empty description='No granularity results available' />
              )}
            </Space>
          )
        },
        {
          key: 'logs',
          label: 'Logs',
          children: (
            <Card title='Execution Logs'>
              <Text type='secondary'>
                Execution logs will be displayed here when available.
              </Text>
              <div
                style={{
                  marginTop: 16,
                  background: '#f5f5f5',
                  padding: 16,
                  borderRadius: 4,
                }}
              >
                <pre style={{ margin: 0, fontSize: '0.875rem' }}>
                  {`[${dayjs().format('YYYY-MM-DD HH:mm:ss')}] Execution started...
[${dayjs().format('YYYY-MM-DD HH:mm:ss')}] Loading algorithm: ${executionData.algorithm?.name}
[${dayjs().format('YYYY-MM-DD HH:mm:ss')}] Loading datapack: ${executionData.datapack_id}
[${dayjs().format('YYYY-MM-DD HH:mm:ss')}] Running RCA algorithm...
[${dayjs().format('YYYY-MM-DD HH:mm:ss')}] Generating results...
[${dayjs().format('YYYY-MM-DD HH:mm:ss')}] Execution completed successfully`}
                </pre>
              </div>
            </Card>
          )
        }
      ]} />
    </div>
  );
};

export default ExecutionDetail;
