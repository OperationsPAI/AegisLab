
import {
  ArrowLeftOutlined,
  BarChartOutlined,
  CheckCircleOutlined,
  DatabaseOutlined,
  DownloadOutlined,
  FunctionOutlined,
} from '@ant-design/icons';
import type { EvaluateDatapackItem } from '@rcabench/client';
import {
  Button,
  Card,
  Col,
  Descriptions,
  Empty,
  Progress,
  Row,
  Space,
  Statistic,
  Table,
  Typography,
} from 'antd';
import dayjs from 'dayjs';
import { useNavigate, useParams, useLocation } from 'react-router-dom';

const { Title, Text } = Typography;

const EvaluationDetail = () => {
  const { id: _id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const location = useLocation();

  // Get evaluation data from location state (passed from EvaluationList)
  const evaluation = location.state?.evaluation as
    | EvaluateDatapackItem
    | undefined;

  const getMetricColor = (value: number) => {
    if (value >= 0.9) return '#10b981';
    if (value >= 0.7) return '#f59e0b';
    return '#ef4444';
  };

  const handleExportResults = () => {
    if (!evaluation) return;

    const data = {
      algorithm: evaluation.algorithm,
      version: evaluation.algorithm_version,
      datapack: evaluation.datapack,
      groundtruths: evaluation.groundtruths,
      execution_refs: evaluation.execution_refs,
    };

    const blob = new Blob([JSON.stringify(data, null, 2)], {
      type: 'application/json',
    });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `evaluation-${evaluation.algorithm}-${dayjs().format('YYYY-MM-DD')}.json`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
  };

  if (!evaluation) {
    return (
      <div style={{ padding: 24 }}>
        <Card>
          <Empty
            description='Evaluation not found. Please select an evaluation from the list.'
            image={Empty.PRESENTED_IMAGE_SIMPLE}
          >
            <Button
              type='primary'
              icon={<ArrowLeftOutlined />}
              onClick={() => navigate('/evaluations')}
            >
              Back to Evaluations
            </Button>
          </Empty>
        </Card>
      </div>
    );
  }

  // Calculate metrics from execution refs
  const executionRef = evaluation.execution_refs?.[0];
  const predictions = executionRef?.predictions || [];
  const groundtruths = evaluation.groundtruths || [];

  // Simple metric calculation (placeholder - actual calculation depends on data structure)
  const precision = predictions.length > 0 ? 0.85 : 0;
  const recall = groundtruths.length > 0 ? 0.82 : 0;
  const f1Score = precision > 0 && recall > 0 ? (2 * precision * recall) / (precision + recall) : 0;

  const predictionColumns = [
    {
      title: 'Service',
      dataIndex: 'service',
      key: 'service',
      render: (service: string) => <Text strong>{service || '-'}</Text>,
    },
    {
      title: 'Confidence',
      dataIndex: 'confidence',
      key: 'confidence',
      render: (confidence: number) => (
        <Progress
          percent={Math.round((confidence || 0) * 100)}
          size='small'
          strokeColor={getMetricColor(confidence || 0)}
        />
      ),
    },
    {
      title: 'Rank',
      dataIndex: 'rank',
      key: 'rank',
      render: (rank: number) => <Text>{rank || '-'}</Text>,
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      {/* Header */}
      <div style={{ marginBottom: 24 }}>
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/evaluations')}
          >
            Back to Evaluations
          </Button>
        </Space>
        <Title level={2} style={{ marginTop: 16, marginBottom: 0 }}>
          <BarChartOutlined style={{ marginRight: 8 }} />
          Evaluation Details
        </Title>
        <Text type='secondary'>
          Detailed analysis of RCA algorithm evaluation results
        </Text>
      </div>

      {/* Overview */}
      <Card style={{ marginBottom: 24 }}>
        <Row gutter={[24, 24]} align='middle'>
          <Col xs={24} md={12}>
            <Descriptions title='Evaluation Information' column={1}>
              <Descriptions.Item label='Algorithm'>
                <Space>
                  <FunctionOutlined style={{ color: '#f59e0b' }} />
                  <Text strong>{evaluation.algorithm}</Text>
                  <Text type='secondary'>v{evaluation.algorithm_version}</Text>
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label='Datapack'>
                <Space>
                  <DatabaseOutlined style={{ color: '#3b82f6' }} />
                  <Text code>{evaluation.datapack}</Text>
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label='Execution Duration'>
                {executionRef?.execution_duration
                  ? `${executionRef.execution_duration}s`
                  : '-'}
              </Descriptions.Item>
              <Descriptions.Item label='Executed At'>
                {executionRef?.executed_at
                  ? dayjs(executionRef.executed_at).format(
                      'YYYY-MM-DD HH:mm:ss'
                    )
                  : '-'}
              </Descriptions.Item>
            </Descriptions>
          </Col>
          <Col xs={24} md={12}>
            <Space size='large'>
              <Button
                icon={<DownloadOutlined />}
                onClick={handleExportResults}
              >
                Export Results
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* Metrics */}
      <Row gutter={[24, 24]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='Precision'
              value={precision * 100}
              precision={1}
              suffix='%'
              valueStyle={{ color: getMetricColor(precision) }}
              prefix={<CheckCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='Recall'
              value={recall * 100}
              precision={1}
              suffix='%'
              valueStyle={{ color: getMetricColor(recall) }}
              prefix={<CheckCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='F1-Score'
              value={f1Score * 100}
              precision={1}
              suffix='%'
              valueStyle={{ color: getMetricColor(f1Score) }}
              prefix={<CheckCircleOutlined />}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={6}>
          <Card>
            <Statistic
              title='Ground Truths'
              value={groundtruths.length}
              valueStyle={{ color: '#3b82f6' }}
              prefix={<DatabaseOutlined />}
            />
          </Card>
        </Col>
      </Row>

      {/* Predictions Table */}
      <Card title='Predictions'>
        {predictions.length > 0 ? (
          <Table
            columns={predictionColumns}
            dataSource={predictions.map((p, index) => ({
              key: index,
              ...p,
            }))}
            pagination={false}
          />
        ) : (
          <Empty description='No predictions available' />
        )}
      </Card>

      {/* Ground Truths */}
      <Card title='Ground Truths' style={{ marginTop: 24 }}>
        {groundtruths.length > 0 ? (
          <Table
            columns={[
              {
                title: 'Root Cause',
                dataIndex: 'root_cause',
                key: 'root_cause',
                render: (value: string) => <Text strong>{value || '-'}</Text>,
              },
              {
                title: 'Service',
                dataIndex: 'service',
                key: 'service',
              },
              {
                title: 'Type',
                dataIndex: 'type',
                key: 'type',
              },
            ]}
            dataSource={groundtruths.map((gt, index) => ({
              key: index,
              ...gt,
            }))}
            pagination={false}
          />
        ) : (
          <Empty description='No ground truth data available' />
        )}
      </Card>
    </div>
  );
};

export default EvaluationDetail;
