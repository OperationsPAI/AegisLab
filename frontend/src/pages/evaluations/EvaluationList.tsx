
import {
  BarChartOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  DatabaseOutlined,
  DeleteOutlined,
  DownloadOutlined,
  EyeOutlined,
  FilterOutlined,
  FunctionOutlined,
  PlayCircleOutlined,
  SearchOutlined,
  SyncOutlined,
} from '@ant-design/icons';
import { useMutation, useQuery } from '@tanstack/react-query';
import {
  Avatar,
  Badge,
  Button,
  Card,
  Col,
  Empty,
  Input,
  message,
  Modal,
  Progress,
  Row,
  Select,
  Space,
  Table,
  type TablePaginationConfig,
  Tooltip,
  Typography,
} from 'antd';
import dayjs from 'dayjs';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { containerApi } from '@/api/containers';
import { datasetApi } from '@/api/datasets';
import { evaluationApi } from '@/api/evaluations';
import StatCard from '@/components/ui/StatCard';
import type {
  DatapackEvaluationResult,
  DatapackEvaluationSpec,
} from '@/types/api';

const { Title, Text } = Typography;
const { Search } = Input;
const { Option } = Select;

// Mock evaluation data for demonstration
const mockEvaluations: DatapackEvaluationResult[] = [
  {
    algorithm: 'MicroRank',
    algorithm_version: 'v1.0.0',
    datapack: 'dp-12345678',
    groundtruths: [],
    execution_refs: [
      {
        execution_id: 42,
        execution_duration: 120,
        detector_results: [],
        predictions: {},
        executed_at: '2024-01-15T10:30:00Z',
      },
    ],
  },
  {
    algorithm: 'TraceRCA',
    algorithm_version: 'v2.1.0',
    datapack: 'dp-87654321',
    groundtruths: [],
    execution_refs: [
      {
        execution_id: 43,
        execution_duration: 95,
        detector_results: [],
        predictions: {},
        executed_at: '2024-01-15T11:45:00Z',
      },
    ],
  },
];

const EvaluationList = () => {
  const navigate = useNavigate();
  const [_searchText, setSearchText] = useState('');
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [algorithmFilter, setAlgorithmFilter] = useState<string | undefined>();
  const [typeFilter, setTypeFilter] = useState<
    'datapack' | 'dataset' | undefined
  >();
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });
  const [evaluations, setEvaluations] =
    useState<DatapackEvaluationResult[]>(mockEvaluations);
  const [isLoading, _setIsLoading] = useState(false);

  // Fetch available algorithms and datasets for filters
  const { data: algorithmsData } = useQuery({
    queryKey: ['algorithms'],
    queryFn: () => containerApi.getContainers({ type: 2 }), // Algorithm = 2
  });

  const { data: _datasetsData } = useQuery({
    queryKey: ['datasets'],
    queryFn: () => datasetApi.getDatasets(),
  });

  // Evaluate datapack mutation
  const evaluateDatapackMutation = useMutation({
    mutationFn: (specs: DatapackEvaluationSpec[]) =>
      evaluationApi.evaluateDatapacks(specs as Array<Record<string, unknown>>),
    onSuccess: (data) => {
      message.success('Evaluation completed successfully');
      // Add new evaluations to the list
      setEvaluations((prev) => [...prev, ...data]);
    },
    onError: (error) => {
      message.error('Failed to evaluate datapack');
      console.error('Evaluation error:', error);
    },
  });

  // Evaluate dataset mutation
  const evaluateDatasetMutation = useMutation({
    mutationFn: (specs: DatapackEvaluationSpec[]) =>
      evaluationApi.evaluateDatasets(specs as Array<Record<string, unknown>>),
    onSuccess: (data) => {
      message.success('Evaluation completed successfully');
      // Add new evaluations to the list
      setEvaluations((prev) => [...prev, ...data]);
    },
    onError: (error) => {
      message.error('Failed to evaluate dataset');
      console.error('Evaluation error:', error);
    },
  });

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

  const handleAlgorithmFilter = (algorithm: string | undefined) => {
    setAlgorithmFilter(algorithm);
    setPagination({ ...pagination, current: 1 });
  };

  const handleTypeFilter = (type: 'datapack' | 'dataset' | undefined) => {
    setTypeFilter(type);
    setPagination({ ...pagination, current: 1 });
  };

  const handleViewEvaluation = (_evaluation: DatapackEvaluationResult) => {
    // TODO: Navigate to detailed evaluation view when implemented
    message.info('Detailed evaluation view will be implemented soon');
  };

  const handleDeleteEvaluation = (index: number) => {
    Modal.confirm({
      title: 'Delete Evaluation',
      content: 'Are you sure you want to delete this evaluation result?',
      okText: 'Yes, delete it',
      okButtonProps: { danger: true },
      cancelText: 'Cancel',
      onOk: () => {
        setEvaluations((prev) => prev.filter((_, i) => i !== index));
        message.success('Evaluation deleted successfully');
      },
    });
  };

  const handleBatchDelete = () => {
    if (selectedRowKeys.length === 0) {
      message.warning('Please select evaluations to delete');
      return;
    }

    Modal.confirm({
      title: 'Batch Delete Evaluations',
      content: `Are you sure you want to delete ${selectedRowKeys.length} evaluations?`,
      okText: 'Yes, delete them',
      okButtonProps: { danger: true },
      cancelText: 'Cancel',
      onOk: () => {
        setEvaluations((prev) =>
          prev.filter((_, i) => !selectedRowKeys.includes(i))
        );
        setSelectedRowKeys([]);
        message.success(
          `${selectedRowKeys.length} evaluations deleted successfully`
        );
      },
    });
  };

  const handleCreateEvaluation = () => {
    navigate('/evaluations/new');
  };

  const handleExportResults = () => {
    // Export evaluation results as CSV
    const csvContent = [
      'Algorithm,Version,Datapack,Dataset,Execution Count,Created',
      ...evaluations.map(
        (e) =>
          `${e.algorithm},${e.algorithm_version},${e.datapack},${e.groundtruths.length},${e.execution_refs[0]?.executed_at || ''}`
      ),
    ].join('\n');

    const blob = new Blob([csvContent], { type: 'text/csv' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `evaluation-results-${dayjs().format('YYYY-MM-DD')}.csv`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    message.success('Evaluation results exported successfully');
  };

  const getMetricColor = (value: number) => {
    if (value >= 0.9) return '#10b981';
    if (value >= 0.7) return '#f59e0b';
    return '#ef4444';
  };

  const getMetricStatus = (value: number) => {
    if (value >= 0.9) return 'success';
    if (value >= 0.7) return 'warning';
    return 'error';
  };

  const rowSelection = {
    selectedRowKeys,
    onChange: setSelectedRowKeys,
  };

  const columns = [
    {
      title: 'Algorithm',
      key: 'algorithm',
      width: '20%',
      render: (_: string, record: DatapackEvaluationResult) => (
        <Space>
          <Avatar
            size='small'
            style={{ backgroundColor: '#f59e0b' }}
            icon={<FunctionOutlined />}
          />
          <div>
            <Text strong>{record.algorithm}</Text>
            <br />
            <Text type='secondary' style={{ fontSize: '0.75rem' }}>
              v{record.algorithm_version}
            </Text>
          </div>
        </Space>
      ),
    },
    {
      title: 'Datapack',
      dataIndex: 'datapack',
      key: 'datapack',
      width: '15%',
      render: (datapackId: string) => (
        <Space>
          <DatabaseOutlined style={{ color: '#3b82f6' }} />
          <Text code>{datapackId?.substring(0, 8)}</Text>
        </Space>
      ),
    },
    {
      title: 'Dataset',
      dataIndex: 'dataset_id',
      key: 'dataset_id',
      width: '15%',
      render: (datasetId: string) => (
        <Space>
          <DatabaseOutlined style={{ color: '#10b981' }} />
          <Text code>{datasetId?.substring(0, 8)}</Text>
        </Space>
      ),
    },
    {
      title: 'Precision',
      dataIndex: ['metrics', 'precision'],
      key: 'precision',
      width: '10%',
      render: (precision?: number) => (
        <Progress
          percent={(precision || 0) * 100}
          size='small'
          strokeColor={getMetricColor(precision || 0)}
          format={(percent) => `${percent.toFixed(1)}%`}
        />
      ),
    },
    {
      title: 'Recall',
      dataIndex: ['metrics', 'recall'],
      key: 'recall',
      width: '10%',
      render: (recall?: number) => (
        <Progress
          percent={(recall || 0) * 100}
          size='small'
          strokeColor={getMetricColor(recall || 0)}
          format={(percent) => `${percent.toFixed(1)}%`}
        />
      ),
    },
    {
      title: 'F1-Score',
      dataIndex: ['metrics', 'f1_score'],
      key: 'f1_score',
      width: '10%',
      render: (f1Score?: number) => (
        <Progress
          percent={(f1Score || 0) * 100}
          size='small'
          strokeColor={getMetricColor(f1Score || 0)}
          format={(percent) => `${percent.toFixed(1)}%`}
        />
      ),
    },
    {
      title: 'Accuracy',
      dataIndex: ['metrics', 'accuracy'],
      key: 'accuracy',
      width: '10%',
      render: (accuracy: number) => (
        <Badge
          status={
            getMetricStatus(accuracy) as
              | 'success'
              | 'error'
              | 'warning'
              | 'processing'
              | 'default'
          }
          text={
            <Progress
              percent={(accuracy || 0) * 100}
              size='small'
              strokeColor={getMetricColor(accuracy || 0)}
              format={(percent) => `${percent.toFixed(1)}%`}
            />
          }
        />
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
          <Text>{dayjs(date).format('MMM D, HH:mm')}</Text>
        </Space>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: '8%',
      render: (_: string, record: DatapackEvaluationResult, index: number) => (
        <Space size='small'>
          <Tooltip title='View Details'>
            <Button
              type='text'
              icon={<EyeOutlined />}
              onClick={() => handleViewEvaluation(record)}
            />
          </Tooltip>
          <Tooltip title='Delete'>
            <Button
              type='text'
              danger
              icon={<DeleteOutlined />}
              onClick={() => handleDeleteEvaluation(index)}
            />
          </Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <div className='evaluation-list'>
      {/* Page Header */}
      <div className='page-header'>
        <div className='page-header-left'>
          <Title level={2} className='page-title'>
            Evaluation Results
          </Title>
          <Text type='secondary'>
            Compare and analyze RCA algorithm performance
          </Text>
        </div>
        <div className='page-header-right'>
          <Space>
            <Button icon={<DownloadOutlined />} onClick={handleExportResults}>
              Export Results
            </Button>
            <Button
              type='primary'
              size='large'
              icon={<PlayCircleOutlined />}
              onClick={handleCreateEvaluation}
            >
              New Evaluation
            </Button>
          </Space>
        </div>
      </div>

      {/* Statistics Cards */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} md={6}>
          <StatCard
            title='Total Evaluations'
            value={evaluations.length}
            icon={<BarChartOutlined />}
            color='primary'
          />
        </Col>
        <Col xs={24} sm={12} md={6}>
          <StatCard
            title='Algorithms Evaluated'
            value={new Set(evaluations.map((e) => e.algorithm)).size}
            icon={<FunctionOutlined />}
            color='success'
          />
        </Col>
        <Col xs={24} sm={12} md={6}>
          <StatCard
            title='Avg F1-Score'
            value={
              evaluations.length > 0
                ? `${(evaluations.length * 85).toFixed(1)}%`
                : '0%'
            }
            icon={<CheckCircleOutlined />}
            color='warning'
          />
        </Col>
        <Col xs={24} sm={12} md={6}>
          <StatCard
            title='Best Accuracy'
            value={
              evaluations.length > 0
                ? `${(95).toFixed(1)}%`
                : '0%'
            }
            icon={<CheckCircleOutlined />}
            color='error'
          />
        </Col>
      </Row>

      {/* Filters and Actions */}
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={[16, 16]} align='middle'>
          <Col xs={24} sm={12} md={6}>
            <Search
              placeholder='Search evaluations...'
              allowClear
              enterButton={<SearchOutlined />}
              onSearch={handleSearch}
              style={{ width: '100%' }}
            />
          </Col>
          <Col xs={24} sm={12} md={4}>
            <Select
              placeholder='Filter by algorithm'
              allowClear
              style={{ width: '100%' }}
              onChange={handleAlgorithmFilter}
              value={algorithmFilter}
            >
              {algorithmsData?.data?.data?.map((algo) => (
                <Option key={algo.id} value={algo.name}>
                  {algo.name}
                </Option>
              ))}
            </Select>
          </Col>
          <Col xs={24} sm={12} md={4}>
            <Select
              placeholder='Filter by type'
              allowClear
              style={{ width: '100%' }}
              onChange={handleTypeFilter}
              value={typeFilter}
            >
              <Option value='datapack'>Datapack</Option>
              <Option value='dataset'>Dataset</Option>
            </Select>
          </Col>
          <Col xs={24} sm={24} md={10} style={{ textAlign: 'right' }}>
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
              <Button icon={<FilterOutlined />}>Advanced Filter</Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* Evaluation Table */}
      <Card>
        <Table
          rowKey={(record, index) =>
            `${record.algorithm}-${record.datapack}-${index}`
          }
          rowSelection={rowSelection}
          columns={columns}
          dataSource={evaluations}
          loading={
            isLoading ||
            evaluateDatapackMutation.isPending ||
            evaluateDatasetMutation.isPending
          }
          pagination={{
            ...pagination,
            total: evaluations.length,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `${range[0]}-${range[1]} of ${total} evaluations`,
          }}
          onChange={handleTableChange}
          locale={{
            emptyText: <Empty description='No evaluations found' />,
          }}
        />
      </Card>

      {/* Evaluation in Progress */}
      {(evaluateDatapackMutation.isPending ||
        evaluateDatasetMutation.isPending) && (
        <Card style={{ marginTop: 16 }}>
          <Space>
            <SyncOutlined spin />
            <Text>Evaluation in progress...</Text>
          </Space>
        </Card>
      )}
    </div>
  );
};

export default EvaluationList;
