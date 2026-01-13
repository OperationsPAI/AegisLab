import {
  EditOutlined,
  DeleteOutlined,
  ArrowLeftOutlined,
  DatabaseOutlined,
  FileTextOutlined,
  LineChartOutlined,
  ClockCircleOutlined,
  TagsOutlined,
  DownloadOutlined,
  EyeOutlined,
  PlusOutlined,
} from '@ant-design/icons'
import { useQuery } from '@tanstack/react-query'
import {
  Card,
  Button,
  Space,
  Typography,
  Row,
  Col,
  Tag,
  Descriptions,
  Table,
  Tooltip,
  Modal,
  message,
  Tabs,
  Badge,
  Divider,
} from 'antd'
import dayjs from 'dayjs'
import { useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

import { datasetApi } from '@/api/datasets'
import StatusBadge from '@/components/ui/StatusBadge'
import type { DatasetVersion } from '@/types/api'

const { Title, Text } = Typography
const { TabPane } = Tabs

const DatasetDetail = () => {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const datasetId = Number(id)
  const [activeTab, setActiveTab] = useState('overview')

  // Fetch dataset details
  const { data: dataset, isLoading } = useQuery({
    queryKey: ['dataset', datasetId],
    queryFn: () => datasetApi.getDataset(datasetId),
    enabled: !!datasetId,
  })

  // Fetch versions
  const { data: versions = [], isLoading: versionsLoading } = useQuery({
    queryKey: ['dataset-versions', datasetId],
    queryFn: () => datasetApi.getVersions(datasetId),
    enabled: !!datasetId,
  })

  const handleEdit = () => {
    navigate(`/datasets/${datasetId}/edit`)
  }

  const handleDelete = () => {
    Modal.confirm({
      title: 'Delete Dataset',
      content: `Are you sure you want to delete dataset "${dataset?.name}"? This action cannot be undone.`,
      okText: 'Yes, delete it',
      okButtonProps: { danger: true },
      cancelText: 'Cancel',
      onOk: async () => {
        try {
          await datasetApi.deleteDataset(datasetId)
          message.success('Dataset deleted successfully')
          navigate('/datasets')
        } catch (error) {
          message.error('Failed to delete dataset')
        }
      },
    })
  }

  const handleDownloadVersion = (version: DatasetVersion) => {
    // TODO: Implement download logic
    message.info('Download functionality will be implemented soon')
  }

  const handlePreviewVersion = (version: DatasetVersion) => {
    // TODO: Implement preview logic
    message.info('Preview functionality will be implemented soon')
  }

  const getTypeIcon = (type: string) => {
    switch (type) {
      case 'Trace':
        return <DatabaseOutlined style={{ color: '#3b82f6' }} />
      case 'Log':
        return <FileTextOutlined style={{ color: '#10b981' }} />
      case 'Metric':
        return <LineChartOutlined style={{ color: '#f59e0b' }} />
      default:
        return <DatabaseOutlined />
    }
  }

  const getTypeColor = (type: string) => {
    switch (type) {
      case 'Trace':
        return '#3b82f6'
      case 'Log':
        return '#10b981'
      case 'Metric':
        return '#f59e0b'
      default:
        return '#6b7280'
    }
  }

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes'
    const k = 1024
    const sizes = ['Bytes', 'KB', 'MB', 'GB', 'TB']
    const i = Math.floor(Math.log(bytes) / Math.log(k))
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))  } ${  sizes[i]}`
  }

  const versionColumns = [
    {
      title: 'Version',
      dataIndex: 'version',
      key: 'version',
      width: 120,
      render: (version: string) => (
        <Badge
          count={version}
          style={{ backgroundColor: '#3b82f6', fontWeight: 'bold' }}
        />
      ),
    },
    {
      title: 'File Path',
      dataIndex: 'file_path',
      key: 'file_path',
      render: (filePath: string) => (
        <Tooltip title={filePath}>
          <Text ellipsis style={{ maxWidth: 200 }}>
            {filePath}
          </Text>
        </Tooltip>
      ),
    },
    {
      title: 'Size',
      dataIndex: 'size',
      key: 'size',
      width: 100,
      render: (size: number) => (
        <Text code>{formatFileSize(size)}</Text>
      ),
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
          <Text type="secondary">-</Text>
        ),
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 150,
      render: (date: string) => (
        <Space>
          <ClockCircleOutlined />
          <Text>{dayjs(date).format('MMM D, YYYY HH:mm')}</Text>
        </Space>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: 120,
      fixed: 'right' as const,
      render: (_: string, record: DatasetVersion) => (
        <Space size="small">
          <Tooltip title="Preview">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => handlePreviewVersion(record)}
            />
          </Tooltip>
          <Tooltip title="Download">
            <Button
              type="text"
              icon={<DownloadOutlined />}
              onClick={() => handleDownloadVersion(record)}
            />
          </Tooltip>
        </Space>
      ),
    },
  ]

  if (isLoading) {
    return (
      <div style={{ padding: 24 }}>
        <Card loading>
          <div style={{ minHeight: 400 }} />
        </Card>
      </div>
    )
  }

  if (!dataset) {
    return (
      <div style={{ padding: 24, textAlign: 'center' }}>
        <Text type="secondary">Dataset not found</Text>
      </div>
    )
  }

  const datasetData = dataset.data

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
        <Row justify="space-between" align="middle">
          <Col>
            <Space>
              <Button
                type="primary"
                icon={<EditOutlined />}
                onClick={handleEdit}
              >
                Edit Dataset
              </Button>
              <Button
                icon={<PlusOutlined />}
                onClick={() => {
                  // TODO: Navigate to version creation
                  message.info('Version creation will be implemented soon')
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
      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane tab="Overview" key="overview">
          <Row gutter={[16, 16]}>
            <Col xs={24} lg={16}>
              <Card title="Dataset Information">
                <Descriptions column={2} bordered>
                  <Descriptions.Item label="ID">{datasetData.id}</Descriptions.Item>
                  <Descriptions.Item label="Type">
                    <Tag
                      color={getTypeColor(datasetData.type)}
                      style={{ fontWeight: 500, fontSize: '1rem' }}
                    >
                      {datasetData.type}
                    </Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label="Visibility">
                    <StatusBadge
                      status={datasetData.is_public ? 'success' : 'default'}
                      text={datasetData.is_public ? 'Public' : 'Private'}
                    />
                  </Descriptions.Item>
                  <Descriptions.Item label="Created">
                    <Space>
                      <ClockCircleOutlined />
                      {dayjs(datasetData.created_at).format('MMMM D, YYYY HH:mm')}
                    </Space>
                  </Descriptions.Item>
                  <Descriptions.Item label="Updated">
                    <Space>
                      <ClockCircleOutlined />
                      {dayjs(datasetData.updated_at).format('MMMM D, YYYY HH:mm')}
                    </Space>
                  </Descriptions.Item>
                  <Descriptions.Item label="Labels">
                    {datasetData.labels?.length ? (
                      <Space wrap>
                        {datasetData.labels.map((label, index) => (
                          <Tag key={index} icon={<TagsOutlined />}>
                            {label.key}: {label.value}
                          </Tag>
                        ))}
                      </Space>
                    ) : (
                      <Text type="secondary">No labels</Text>
                    )}
                  </Descriptions.Item>
                </Descriptions>
              </Card>
            </Col>
            <Col xs={24} lg={8}>
              <Card title="Quick Stats">
                <Space direction="vertical" style={{ width: '100%' }}>
                  <div>
                    <Text type="secondary">Total Versions</Text>
                    <br />
                    <Title level={3} style={{ margin: 0, color: '#3b82f6' }}>
                      {versions.length}
                    </Title>
                  </div>
                  <Divider />
                  <div>
                    <Text type="secondary">Latest Version</Text>
                    <br />
                    <Text strong style={{ fontSize: '1.25rem' }}>
                      {versions[0]?.version || 'N/A'}
                    </Text>
                  </div>
                  <Divider />
                  <div>
                    <Text type="secondary">Total Size</Text>
                    <br />
                    <Text strong style={{ fontSize: '1.25rem' }}>
                      {formatFileSize(versions.reduce((sum, v) => sum + (v.size || 0), 0))}
                    </Text>
                  </div>
                </Space>
              </Card>
            </Col>
          </Row>

          {datasetData.description && (
            <Card title="Description" style={{ marginTop: 16 }}>
              <Text>{datasetData.description}</Text>
            </Card>
          )}
        </TabPane>

        <TabPane tab="Versions" key="versions">
          <Card
            title="Dataset Versions"
            extra={
              <Button
                type="primary"
                icon={<PlusOutlined />}
                onClick={() => {
                  // TODO: Navigate to version creation page
                  message.info('Version creation will be implemented soon')
                }}
              >
                Add Version
              </Button>
            }
          >
            <Table
              rowKey="id"
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
        </TabPane>

        <TabPane tab="Preview" key="preview">
          <Card title="Dataset Preview">
            <Text type="secondary">
              Dataset preview functionality will be implemented soon.
            </Text>
            <div style={{ marginTop: 16 }}>
              <Text>Sample data preview will appear here...</Text>
            </div>
          </Card>
        </TabPane>

        <TabPane tab="Usage" key="usage">
          <Card title="Dataset Usage">
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
        </TabPane>
      </Tabs>
    </div>
  )
}

export default DatasetDetail