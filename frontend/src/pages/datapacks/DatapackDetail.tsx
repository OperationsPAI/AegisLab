import {
  ArrowLeftOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  DatabaseOutlined,
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

import { datapackApi } from '@/api/datapacks';

const { Title } = Typography;

const DatapackDetail = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();

  const { data: datapack, isLoading } = useQuery({
    queryKey: ['datapack', id],
    queryFn: () => datapackApi.getDatapack(Number(id)),
    enabled: !!id,
  });

  if (isLoading) {
    return (
      <div style={{ textAlign: 'center', padding: '100px' }}>
        <Spin size="large" />
      </div>
    );
  }

  if (!datapack) {
    return <div>Datapack not found</div>;
  }

  const getStateTag = (state: string) => {
    const stateMap: Record<string, { color: string; text: string; icon: JSX.Element }> = {
      build_success: { color: 'success', text: 'Ready', icon: <CheckCircleOutlined /> },
      building: { color: 'processing', text: 'Building', icon: <ClockCircleOutlined /> },
      build_failed: { color: 'error', text: 'Failed', icon: <ClockCircleOutlined /> },
      initial: { color: 'default', text: 'Pending', icon: <ClockCircleOutlined /> },
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
            onClick={() => navigate('/datapacks')}
            style={{ marginBottom: '16px' }}
          >
            Back to Datapacks
          </Button>
          <Title level={2}>
            <DatabaseOutlined /> {datapack.name}
          </Title>
        </div>

        <Row gutter={[16, 16]}>
          <Col span={24}>
            <Card title="Basic Information">
              <Descriptions column={2} bordered>
                <Descriptions.Item label="ID">{datapack.id}</Descriptions.Item>
                <Descriptions.Item label="State">
                  {getStateTag(datapack.state)}
                </Descriptions.Item>
                <Descriptions.Item label="Fault Type">
                  <Tag>{datapack.fault_type}</Tag>
                </Descriptions.Item>
                <Descriptions.Item label="Category">
                  <Tag>{datapack.category}</Tag>
                </Descriptions.Item>
                <Descriptions.Item label="Benchmark">
                  {datapack.benchmark?.name || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="Pedestal">
                  {datapack.pedestal?.name || '-'}
                </Descriptions.Item>
                <Descriptions.Item label="Created At">
                  {dayjs(datapack.created_at).format('YYYY-MM-DD HH:mm:ss')}
                </Descriptions.Item>
                <Descriptions.Item label="Updated At">
                  {dayjs(datapack.updated_at).format('YYYY-MM-DD HH:mm:ss')}
                </Descriptions.Item>
                {datapack.description && (
                  <Descriptions.Item label="Description" span={2}>
                    {datapack.description}
                  </Descriptions.Item>
                )}
              </Descriptions>
            </Card>
          </Col>

          {datapack.labels && datapack.labels.length > 0 && (
            <Col span={24}>
              <Card title="Labels">
                <Space wrap>
                  {datapack.labels.map((label) => (
                    <Tag key={label.key} color={label.color || 'blue'}>
                      {label.key}: {label.value}
                    </Tag>
                  ))}
                </Space>
              </Card>
            </Col>
          )}

          {datapack.groundtruths && datapack.groundtruths.length > 0 && (
            <Col span={24}>
              <Card title="Ground Truths">
                <Space direction="vertical" style={{ width: '100%' }}>
                  {datapack.groundtruths.map((gt, index) => (
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
        </Row>
      </Space>
    </div>
  );
};

export default DatapackDetail;
