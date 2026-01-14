import {
  ArrowLeftOutlined,
  ClockCircleOutlined,
  DeleteOutlined,
  DownloadOutlined,
  EditOutlined,
  PlusOutlined,
  TagsOutlined,
} from '@ant-design/icons';
import type { DatasetVersionResp } from '@rcabench/client';
import { useQuery } from '@tanstack/react-query';
import {
  Badge,
  Button,
  Card,
  Col,
  Descriptions,
  Divider,
  message,
  Modal,
  Row,
  Space,
  Table,
  Tabs,
  Tag,
  Tooltip,
  Typography,
} from 'antd';
import dayjs from 'dayjs';
import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';


import { datasetApi } from '@/api/datasets';


const { Title, Text } = Typography;
// Removed deprecated TabPane destructuring - using items prop instead

const DatasetDetail = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const datasetId = Number(id);
  const [activeTab, setActiveTab] = useState('overview');

  // Fetch dataset details
  const { data: dataset, isLoading } = useQuery({
    queryKey: ['dataset', datasetId],
    queryFn: () => datasetApi.getDataset(datasetId),
    enabled: !!datasetId,
  });

  // Fetch versions
  const { data: versionsData, isLoading: versionsLoading } = useQuery({
    queryKey: ['dataset-versions', datasetId],
    queryFn: () => datasetApi.getVersions(datasetId),
    enabled: !!datasetId,
  });
  const versions = versionsData?.items || [];

  const handleEdit = () => {
    navigate(`/datasets/${datasetId}/edit`);
  };

  const handleDelete = () => {
    Modal.confirm({
      title: 'Delete Dataset',
      content: `Are you sure you want to delete dataset "${dataset?.name}"? This action cannot be undone.`,
      okText: 'Yes, delete it',
      okButtonProps: { danger: true },
      cancelText: 'Cancel',
      onOk: async () => {
        try {
          await datasetApi.deleteDataset(datasetId);
          message.success('Dataset deleted successfully');
          navigate('/datasets');
        } catch (error) {
          message.error('Failed to delete dataset');
        }
      },
    });
  };

  const handleDownloadVersion = async (version: DatasetVersionResp) => {
    if (!version.id) {
      message.error('Version ID is missing');
      return;
    }
    try {
      await datasetApi.downloadVersion(
        datasetId,
        version.id,
        `${dataset?.name || 'dataset'}-${version.name}.zip`
      );
      message.success('Download started');
    } catch (error) {
      message.error('Failed to download version');
    }
  };

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'Trace':
        return '#3b82f6';
      case 'Log':
        return '#10b981';
      case 'Metric':
        return '#f59e0b';
      default:
        return '#6b7280';
    }
  };

  const versionColumns = [
    {
      title: 'Version',
      dataIndex: 'name',
      key: 'name',
      width: 120,
      render: (name: string) => (
        <Badge
          count={name}
          style={{ backgroundColor: '#3b82f6', fontWeight: 'bold' }}
        />
      ),
    },
    {
      title: 'File Count',
      dataIndex: 'file_count',
      key: 'file_count',
      width: 100,
      render: (count?: number) => <Text code>{count || 0}</Text>,
    },
    {
      title: 'Checksum',
      dataIndex: 'checksum',
      key: 'checksum',
      width: 120,
      render: (checksum?: string) =>
        checksum ? (
          <Tooltip title={checksum}>
            <Text ellipsis style={{ maxWidth: 100 }} code>
              {checksum.substring(0, 8)}...
            </Text>
          </Tooltip>
        ) : (
          <Text type='secondary'>-</Text>
        ),
    },
    {
      title: 'Updated',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 150,
      render: (date?: string) => (
        <Space>
          <ClockCircleOutlined />
          <Text>{date ? dayjs(date).format('MMM D, YYYY HH:mm') : '-'}</Text>
        </Space>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 80,
      fixed: 'right' as const,
      render: (_: unknown, record: DatasetVersionResp) => (
        <Tooltip title='Download'>
          <Button
            type='text'
            icon={<DownloadOutlined />}
            onClick={() => handleDownloadVersion(record)}
          />
        </Tooltip>
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

  if (!dataset) {
    return (
      <div style={{ padding: 24, textAlign: 'center' }}>
        <Text type='secondary'>Dataset not found</Text>
      </div>
    );
  }

  const datasetData = dataset;

  return (
    <div style={{ padding: 24 }}>
      {/* Header */}
      <div style={{ marginBottom: 24 }}>
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/datasets')}
          >
            Back to List
          </Button>
          <Title level={2} style={{ margin: 0 }}>
            {datasetData.name}
          </Title>
        </Space>
      </div>

      {/* Actions */}
      <Card style={{ marginBottom: 24 }}>
        <Row justify='space-between' align='middle'>
          <Col>
            <Space>
              <Button
                type='primary'
                icon={<EditOutlined />}
                onClick={handleEdit}
              >
                Edit Dataset
              </Button>
              <Button
                icon={<PlusOutlined />}
                onClick={() => {
                  // TODO: Navigate to version creation
                  message.info('Version creation will be implemented soon');
                }}
              >
                Add Version
              </Button>
            </Space>
          </Col>
          <Col>
            <Button danger icon={<DeleteOutlined />} onClick={handleDelete}>
              Delete Dataset
            </Button>
          </Col>
        </Row>
      </Card>

      {/* Tabs */}
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={[
          {
            key: 'overview',
            label: 'Overview',
            children: (
              <>
                <Row gutter={[16, 16]}>
                  <Col xs={24} lg={16}>
                    <Card title='Dataset Information'>
                      <Descriptions column={2} bordered>
                        <Descriptions.Item label='ID'>
                          {datasetData.id}
                        </Descriptions.Item>
                        <Descriptions.Item label='Type'>
                          <Tag
                            color={getTypeColor(datasetData.type || '')}
                            style={{ fontWeight: 500, fontSize: '1rem' }}
                          >
                            {datasetData.type}
                          </Tag>
                        </Descriptions.Item>
                        <Descriptions.Item label='Visibility'>
                          <Tag
                            color={datasetData.is_public ? 'green' : 'default'}
                          >
                            {datasetData.is_public ? 'Public' : 'Private'}
                          </Tag>
                        </Descriptions.Item>
                        <Descriptions.Item label='Created'>
                          <Space>
                            <ClockCircleOutlined />
                            {dayjs(datasetData.created_at).format(
                              'MMMM D, YYYY HH:mm'
                            )}
                          </Space>
                        </Descriptions.Item>
                        <Descriptions.Item label='Updated'>
                          <Space>
                            <ClockCircleOutlined />
                            {dayjs(datasetData.updated_at).format(
                              'MMMM D, YYYY HH:mm'
                            )}
                          </Space>
                        </Descriptions.Item>
                        <Descriptions.Item label='Labels'>
                          {datasetData.labels?.length ? (
                            <Space wrap>
                              {datasetData.labels.map((label, index) => (
                                <Tag key={index} icon={<TagsOutlined />}>
                                  {label.key}: {label.value}
                                </Tag>
                              ))}
                            </Space>
                          ) : (
                            <Text type='secondary'>No labels</Text>
                          )}
                        </Descriptions.Item>
                      </Descriptions>
                    </Card>
                  </Col>
                  <Col xs={24} lg={8}>
                    <Card title='Quick Stats'>
                      <Space direction='vertical' style={{ width: '100%' }}>
                        <div>
                          <Text type='secondary'>Total Versions</Text>
                          <br />
                          <Title
                            level={3}
                            style={{ margin: 0, color: '#3b82f6' }}
                          >
                            {versions.length}
                          </Title>
                        </div>
                        <Divider />
                        <div>
                          <Text type='secondary'>Latest Version</Text>
                          <br />
                          <Text strong style={{ fontSize: '1.25rem' }}>
                            {versions[0]?.name || 'N/A'}
                          </Text>
                        </div>
                        <Divider />
                        <div>
                          <Text type='secondary'>Total Files</Text>
                          <br />
                          <Text strong style={{ fontSize: '1.25rem' }}>
                            {versions.reduce(
                              (sum, v) => sum + (v.file_count || 0),
                              0
                            )}
                          </Text>
                        </div>
                      </Space>
                    </Card>
                  </Col>
                </Row>

                {datasetData.description && (
                  <Card title='Description' style={{ marginTop: 16 }}>
                    <Text>{datasetData.description}</Text>
                  </Card>
                )}
              </>
            ),
          },
          {
            key: 'versions',
            label: 'Versions',
            children: (
              <Card
                title='Dataset Versions'
                extra={
                  <Button
                    type='primary'
                    icon={<PlusOutlined />}
                    onClick={() => {
                      // TODO: Navigate to version creation page
                      message.info('Version creation will be implemented soon');
                    }}
                  >
                    Add Version
                  </Button>
                }
              >
                <Table
                  rowKey='id'
                  columns={versionColumns}
                  dataSource={versions}
                  loading={versionsLoading}
                  pagination={{
                    pageSize: 10,
                    showSizeChanger: true,
                    showQuickJumper: true,
                    showTotal: (total, range) =>
                      `${range[0]}-${range[1]} of ${total} versions`,
                  }}
                />
              </Card>
            ),
          },
          {
            key: 'usage',
            label: 'Usage',
            children: (
              <Card title='Dataset Usage'>
                <Text>
                  This dataset can be used in experiments and evaluations.
                </Text>
                <Divider />
                <Text strong>How to use this dataset:</Text>
                <ul style={{ marginTop: 8 }}>
                  <li>Select this dataset when creating an experiment</li>
                  <li>Dataset will be automatically loaded during execution</li>
                  <li>Results can be compared with other datasets</li>
                </ul>
              </Card>
            ),
          },
        ]}
      />
    </div>
  );
};

export default DatasetDetail;
