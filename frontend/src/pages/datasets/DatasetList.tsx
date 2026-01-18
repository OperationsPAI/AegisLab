import {
  ClockCircleOutlined,
  CloudUploadOutlined,
  DatabaseOutlined,
  DeleteOutlined,
  EditOutlined,
  EyeOutlined,
  FileTextOutlined,
  LineChartOutlined,
  PlusOutlined,
  SearchOutlined,
  SettingOutlined,
  UploadOutlined,
} from '@ant-design/icons';
import type { DatasetResp } from '@rcabench/client';
import { useQuery } from '@tanstack/react-query';
import {
  Avatar,
  Badge,
  Button,
  Card,
  Col,
  Input,
  message,
  Modal,
  Progress,
  Row,
  Select,
  Space,
  Table,
  type TablePaginationConfig,
  Tag,
  Tooltip,
  Typography,
  Upload,
} from 'antd';
import dayjs from 'dayjs';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';


import { datasetApi } from '@/api/datasets';
import StatCard from '@/components/ui/StatCard';


// DatasetType for internal use
type DatasetType = 'Trace' | 'Log' | 'Metric';

const { Title, Text } = Typography;
const { Search } = Input;
const { Option } = Select;

const DatasetList = () => {
  const navigate = useNavigate();
  const [searchText, setSearchText] = useState('');
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [typeFilter, setTypeFilter] = useState<DatasetType | undefined>();
  const [uploadModalVisible, setUploadModalVisible] = useState(false);
  const [uploadingFile, setUploadingFile] = useState<File | null>(null);
  const [uploadProgress, setUploadProgress] = useState(0);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  // Fetch datasets
  const {
    data: datasetsResponse,
    isLoading,
    refetch,
  } = useQuery({
    queryKey: [
      'datasets',
      pagination.current,
      pagination.pageSize,
      searchText,
      typeFilter,
    ],
    queryFn: () =>
      datasetApi.getDatasets({
        page: pagination.current,
        size: pagination.pageSize,
        type: typeFilter,
      }),
  });

  const datasetsData = datasetsResponse;

  // Statistics
  const stats = {
    total: datasetsData?.pagination?.total || 0,
    trace: datasetsData?.items?.filter((d) => d.type === 'Trace').length || 0,
    log: datasetsData?.items?.filter((d) => d.type === 'Log').length || 0,
    metric: datasetsData?.items?.filter((d) => d.type === 'Metric').length || 0,
  };

  const handleTableChange = (newPagination: TablePaginationConfig) => {
    setPagination({
      ...pagination,
      current: newPagination.current || 1,
      pageSize: newPagination.pageSize || 10,
    });
  };

  const handleSearch = (value: string) => {
    setSearchText(value);
    setPagination({ ...pagination, current: 1 });
  };

  const handleTypeFilter = (type: DatasetType | undefined) => {
    setTypeFilter(type);
    setPagination({ ...pagination, current: 1 });
  };

  const handleCreateDataset = () => {
    navigate('/datasets/new');
  };

  const handleViewDataset = (id: number) => {
    navigate(`/datasets/${id}`);
  };

  const handleEditDataset = (id: number) => {
    navigate(`/datasets/${id}/edit`);
  };

  const handleDeleteDataset = (id: number) => {
    Modal.confirm({
      title: 'Delete Dataset',
      content:
        'Are you sure you want to delete this dataset? This action cannot be undone.',
      okText: 'Yes, delete it',
      okButtonProps: { danger: true },
      cancelText: 'Cancel',
      onOk: async () => {
        try {
          await datasetApi.deleteDataset(id);
          message.success('Dataset deleted successfully');
          refetch();
        } catch (error) {
          message.error('Failed to delete dataset');
        }
      },
    });
  };

  const handleUploadDataset = () => {
    setUploadModalVisible(true);
  };

  const handleFileSelect = (file: File) => {
    setUploadingFile(file);
    return false; // Prevent auto upload
  };

  const handleUpload = async () => {
    if (!uploadingFile) return;

    // Simulate upload progress
    setUploadProgress(0);
    const interval = setInterval(() => {
      setUploadProgress((prev) => {
        if (prev >= 90) {
          clearInterval(interval);
          return 90;
        }
        return prev + 10;
      });
    }, 200);

    try {
      // TODO: Implement actual file upload when API is ready
      message.success('Dataset uploaded successfully');
      setUploadModalVisible(false);
      setUploadingFile(null);
      setUploadProgress(0);
      refetch();
    } catch (error) {
      message.error('Failed to upload dataset');
      setUploadProgress(0);
    } finally {
      clearInterval(interval);
    }
  };

  const handleBatchDelete = () => {
    if (selectedRowKeys.length === 0) {
      message.warning('Please select datasets to delete');
      return;
    }

    Modal.confirm({
      title: 'Batch Delete Datasets',
      content: `Are you sure you want to delete ${selectedRowKeys.length} datasets?`,
      okText: 'Yes, delete them',
      okButtonProps: { danger: true },
      cancelText: 'Cancel',
      onOk: async () => {
        try {
          await datasetApi.batchDelete(selectedRowKeys as number[]);
          message.success(
            `${selectedRowKeys.length} datasets deleted successfully`
          );
          setSelectedRowKeys([]);
          refetch();
        } catch (error) {
          message.error('Failed to delete datasets');
        }
      },
    });
  };

  const getTypeIcon = (type: DatasetType) => {
    switch (type) {
      case 'Trace':
        return <DatabaseOutlined style={{ color: '#3b82f6' }} />;
      case 'Log':
        return <FileTextOutlined style={{ color: '#10b981' }} />;
      case 'Metric':
        return <LineChartOutlined style={{ color: '#f59e0b' }} />;
      default:
        return <DatabaseOutlined />;
    }
  };

  const getTypeColor = (type: DatasetType) => {
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

  const formatFileSize = (bytes: number) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return `${parseFloat((bytes / Math.pow(k, i)).toFixed(2))} ${sizes[i]}`;
  };

  const rowSelection = {
    selectedRowKeys,
    onChange: setSelectedRowKeys,
  };

  const columns = [
    {
      title: 'Dataset',
      dataIndex: 'name',
      key: 'name',
      width: '30%',
      render: (name: string, record: DatasetResp) => (
        <Space>
          <Avatar
            size='large'
            style={{
              backgroundColor: getTypeColor(record.type as DatasetType),
              fontSize: '1.25rem',
            }}
            icon={getTypeIcon(record.type as DatasetType)}
          />
          <div>
            <Text strong style={{ fontSize: '1rem' }}>
              {name}
            </Text>
            <br />
            <Text type='secondary' style={{ fontSize: '0.875rem' }}>
              ID: {record.id}
            </Text>
          </div>
        </Space>
      ),
    },
    {
      title: 'Type',
      dataIndex: 'type',
      key: 'type',
      width: '15%',
      render: (type: string) => (
        <Tag color={getTypeColor(type as DatasetType)} style={{ fontWeight: 500 }}>
          {type}
        </Tag>
      ),
      filters: [
        { text: 'Trace', value: 'Trace' },
        { text: 'Log', value: 'Log' },
        { text: 'Metric', value: 'Metric' },
      ],
      onFilter: (value: unknown, record: DatasetResp) =>
        record.type === (value as string),
    },
    {
      title: 'Public',
      dataIndex: 'is_public',
      key: 'is_public',
      width: '10%',
      render: (isPublic: boolean) => (
        <Tag color={isPublic ? 'green' : 'default'}>
          {isPublic ? 'Public' : 'Private'}
        </Tag>
      ),
    },
    {
      title: 'Versions',
      dataIndex: 'versions',
      key: 'versions',
      width: '10%',
      render: (versions: Array<{ version?: string }> = []) => (
        <Badge
          count={versions.length}
          showZero
          style={{ backgroundColor: '#3b82f6' }}
        />
      ),
    },
    {
      title: 'Description',
      dataIndex: 'description',
      key: 'description',
      width: '20%',
      render: (description?: string) =>
        description ? (
          <Tooltip title={description}>
            <Text ellipsis style={{ maxWidth: 200 }}>
              {description}
            </Text>
          </Tooltip>
        ) : (
          <Text type='secondary'>No description</Text>
        ),
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      width: '12%',
      render: (date: string) => (
        <Space>
          <ClockCircleOutlined />
          <Text>{dayjs(date).format('MMM D, YYYY')}</Text>
        </Space>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: '18%',
      render: (_: unknown, record: DatasetResp) => (
        <Space>
          <Tooltip title='View Details'>
            <Button
              type='text'
              icon={<EyeOutlined />}
              onClick={() => record.id && handleViewDataset(record.id)}
            />
          </Tooltip>
          <Tooltip title='Edit Dataset'>
            <Button
              type='text'
              icon={<EditOutlined />}
              onClick={() => record.id && handleEditDataset(record.id)}
            />
          </Tooltip>
          <Tooltip title='Manage Versions'>
            <Button
              type='text'
              icon={<SettingOutlined />}
              onClick={() => navigate(`/datasets/${record.id}/versions`)}
            />
          </Tooltip>
          <Tooltip title='Delete Dataset'>
            <Button
              type='text'
              danger
              icon={<DeleteOutlined />}
              onClick={() => record.id && handleDeleteDataset(record.id)}
            />
          </Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <div className='dataset-list'>
      {/* Page Header */}
      <div className='page-header'>
        <div className='page-header-left'>
          <Title level={2} className='page-title'>
            Dataset Management
          </Title>
          <Text type='secondary'>
            Manage your Trace, Log, and Metric datasets
          </Text>
        </div>
        <div className='page-header-right'>
          <Space>
            <Button
              icon={<CloudUploadOutlined />}
              size='large'
              onClick={handleUploadDataset}
            >
              Upload Dataset
            </Button>
            <Button
              type='primary'
              size='large'
              icon={<PlusOutlined />}
              onClick={handleCreateDataset}
            >
              Create Dataset
            </Button>
          </Space>
        </div>
      </div>

      {/* Statistics Cards */}
      <Row gutter={[{ xs: 8, sm: 16, lg: 24 }, { xs: 8, sm: 16, lg: 24 }]} className='stats-row'>
        <Col xs={12} sm={12} lg={6}>
          <StatCard
            title='Total Datasets'
            value={stats.total}
            prefix={<DatabaseOutlined />}
            color='primary'
          />
        </Col>
        <Col xs={12} sm={12} lg={6}>
          <StatCard
            title='Trace Datasets'
            value={stats.trace}
            prefix={<DatabaseOutlined />}
            color='primary'
          />
        </Col>
        <Col xs={12} sm={12} lg={6}>
          <StatCard
            title='Log Datasets'
            value={stats.log}
            prefix={<FileTextOutlined />}
            color='success'
          />
        </Col>
        <Col xs={12} sm={12} lg={6}>
          <StatCard
            title='Metric Datasets'
            value={stats.metric}
            prefix={<LineChartOutlined />}
            color='warning'
          />
        </Col>
      </Row>

      {/* Filters and Actions */}
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={[16, 16]} align='middle'>
          <Col xs={24} sm={12} md={8}>
            <Search
              placeholder='Search datasets by name or description...'
              allowClear
              enterButton={<SearchOutlined />}
              onSearch={handleSearch}
              style={{ width: '100%' }}
            />
          </Col>
          <Col xs={24} sm={12} md={4}>
            <Select
              placeholder='Filter by type'
              allowClear
              style={{ width: '100%' }}
              onChange={handleTypeFilter}
              value={typeFilter}
            >
              <Option value='Trace'>Trace</Option>
              <Option value='Log'>Log</Option>
              <Option value='Metric'>Metric</Option>
            </Select>
          </Col>
          <Col xs={24} sm={24} md={12} style={{ textAlign: 'right' }}>
            <Space>
              {selectedRowKeys.length > 0 && (
                <Button
                  danger
                  icon={<DeleteOutlined />}
                  onClick={handleBatchDelete}
                >
                  Delete Selected ({selectedRowKeys.length})
                </Button>
              )}
              <Button icon={<UploadOutlined />}>Import Dataset</Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* Dataset Table */}
      <Card className='table-card'>
        <Table
          rowKey='id'
          rowSelection={rowSelection}
          columns={columns}
          dataSource={datasetsData?.items || []}
          loading={isLoading}
          className='datasets-table'
          pagination={{
            ...pagination,
            total: datasetsData?.pagination?.total || 0,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `${range[0]}-${range[1]} of ${total} datasets`,
          }}
          onChange={handleTableChange}
        />
      </Card>

      {/* Upload Modal */}
      <Modal
        title='Upload Dataset'
        open={uploadModalVisible}
        onCancel={() => {
          setUploadModalVisible(false);
          setUploadingFile(null);
          setUploadProgress(0);
        }}
        footer={[
          <Button
            key='cancel'
            onClick={() => {
              setUploadModalVisible(false);
              setUploadingFile(null);
              setUploadProgress(0);
            }}
          >
            Cancel
          </Button>,
          <Button
            key='upload'
            type='primary'
            icon={<UploadOutlined />}
            onClick={handleUpload}
            disabled={!uploadingFile || uploadProgress > 0}
            loading={uploadProgress > 0}
          >
            Upload
          </Button>,
        ]}
      >
        <Upload.Dragger
          accept='.csv,.json,.parquet,.zip'
          maxCount={1}
          beforeUpload={handleFileSelect}
          showUploadList={false}
        >
          <p className='ant-upload-drag-icon'>
            <CloudUploadOutlined style={{ fontSize: 48, color: '#3b82f6' }} />
          </p>
          <p className='ant-upload-text'>
            Click or drag dataset file to this area
          </p>
          <p className='ant-upload-hint'>
            Support for single file upload. File types: .csv, .json, .parquet,
            .zip
          </p>
        </Upload.Dragger>
        {uploadingFile && (
          <div style={{ marginTop: 16 }}>
            <Text strong>Selected file: </Text>
            <Text>{uploadingFile.name}</Text>
            <br />
            <Text type='secondary'>
              Size: {formatFileSize(uploadingFile.size)}
            </Text>
            {uploadProgress > 0 && (
              <Progress
                percent={uploadProgress}
                status={uploadProgress === 100 ? 'success' : 'active'}
                style={{ marginTop: 8 }}
              />
            )}
          </div>
        )}
      </Modal>
    </div>
  );
};

export default DatasetList;
