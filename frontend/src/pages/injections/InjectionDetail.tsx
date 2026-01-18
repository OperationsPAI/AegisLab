import {
  ArrowLeftOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  ExperimentOutlined,
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import {
  Button,
  Card,
  Col,
  Descriptions,
  Row,
  Space,
  Spin,
  Tag,
  Typography,
} from 'antd';
import dayjs from 'dayjs';
import { useNavigate, useParams } from 'react-router-dom';

import { injectionApi } from '@/api/injections';

const { Title } = Typography;

const InjectionDetail = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const { data: injection, isLoading } = useQuery({
    queryKey: ['injection', id],
    queryFn: () => injectionApi.getInjection(Number(id)),
    enabled: !!id,
  });

  if (isLoading) {
    return (
      <div style={{ textAlign: 'center', padding: '100px' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!injection) {
    return <div>Injection not found</div>;
  }

  const getStateTag = (state: string) => {
    const stateMap: Record<string, { color: string; text: string; icon: JSX.Element }> = {
      build_success: { color: 'success', text: 'Build Success', icon: <CheckCircleOutlined /> },
      building: { color: 'processing', text: 'Building', icon: <ClockCircleOutlined /> },
      build_failed: { color: 'error', text: 'Build Failed', icon: <ClockCircleOutlined /> },
      inject_success: { color: 'success', text: 'Inject Success', icon: <CheckCircleOutlined /> },
      inject_failed: { color: 'error', text: 'Inject Failed', icon: <ClockCircleOutlined /> },
      initial: { color: 'default', text: 'Initial', icon: <ClockCircleOutlined /> },
    };
    const config = stateMap[state] || { color: 'default', text: state, icon: <ClockCircleOutlined /> };
    return <Tag color={config.color} icon={config.icon}>{config.text}</Tag>;
  };

  return (
    <div style={{ padding: '24px' }}>
      <Space direction="vertical" size="large" style={{ width: '100%' }}>
        <div>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/injections')}
            style={{ marginBottom: '16px' }}
          >
            Back to Injections
          </Button>
          <Title level={2}>
            <ExperimentOutlined /> {injection.name}
          </Title>
        </div>

        <Row gutter={[16, 16]}>
          <Col span={24}>
            <Card title="Basic Information">
              <Descriptions column={2} bordered>
                <Descriptions.Item label="ID">{injection.id}</Descriptions.Item>
                <Descriptions.Item label="State">
                  {getStateTag(injection.state)}
                </Descriptions.Item>
                <Descriptions.Item label="Fault Type">
                  <Tag>{injection.fault_type}</Tag>
                </Descriptions.Item>
                <Descriptions.Item label="Category">
                  <Tag>{injection.category}</Tag>
                </Descriptions.Item>
                <Descriptions.Item label="Benchmark">
                  {injection.benchmark?.name || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="Pedestal">
                  {injection.pedestal?.name || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="Pre Duration">
                  {injection.pre_duration}s
                </Descriptions.Item>
                <Descriptions.Item label="Task ID">
                  {injection.task_id || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="Created At">
                  {dayjs(injection.created_at).format('YYYY-MM-DD HH:mm:ss')}
                </Descriptions.Item>
                <Descriptions.Item label="Updated At">
                  {dayjs(injection.updated_at).format('YYYY-MM-DD HH:mm:ss')}
                </Descriptions.Item>
                {injection.description && (
                  <Descriptions.Item label="Description" span={2}>
                    {injection.description}
                  </Descriptions.Item>
                )}
              </Descriptions>
            </Card>
          </Col>

          {injection.labels && injection.labels.length > 0 && (
            <Col span={24}>
              <Card title="Labels">
                <Space wrap>
                  {injection.labels.map((label) => (
                    <Tag key={label.key} color={label.color || 'blue'}>
                      {label.key}: {label.value}
                    </Tag>
                  ))}
                </Space>
              </Card>
            </Col>
          )}

          {injection.groundtruths && injection.groundtruths.length > 0 && (
            <Col span={24}>
              <Card title="Ground Truths">
                <Space direction="vertical" style={{ width: '100%' }}>
                  {injection.groundtruths.map((gt, index) => (
                    <Card key={index} size="small" type="inner">
                      <Descriptions column={2} size="small">
                        <Descriptions.Item label="Type">{gt.type}</Descriptions.Item>
                        <Descriptions.Item label="Value">{gt.value}</Descriptions.Item>
                      </Descriptions>
                    </Card>
                  ))}
                </Space>
              </Card>
            </Col>
          )}

          {injection.state === 'build_success' && (
            <Col span={24}>
              <Card title="Datapack Information">
                <Descriptions bordered>
                  <Descriptions.Item label="Status">
                    <Tag color="success">Ready for Execution</Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label="View Datapack">
                    <Button
                      type="link"
                      onClick={() => navigate(`/datapacks/${injection.id}`)}
                    >
                      View Datapack Details
                    </Button>
                  </Descriptions.Item>
                </Descriptions>
              </Card>
            </Col>
          )}
        </Row>
      </Space>
    </div>
  );
};

export default InjectionDetail;
