
import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
  DashboardOutlined,
  DatabaseOutlined,
  DeleteOutlined,
  ExportOutlined,
  EyeOutlined,
  FilterOutlined,
  FunctionOutlined,
  PauseCircleOutlined,
  PlayCircleOutlined,
  ReloadOutlined,
  SearchOutlined,
  SyncOutlined,
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
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
  Statistic,
  Table,
  type TablePaginationConfig,
  Tag,
  Tooltip,
  Typography,
} from 'antd';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { taskApi } from '@/api/tasks';
import type { Task, TaskState, TaskType } from '@/types/api';

dayjs.extend(relativeTime);

const { Title, Text } = Typography;
const { Search } = Input;
const { Option } = Select;

const TaskList = () => {
  const navigate = useNavigate();
  const [searchText, setSearchText] = useState('');
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [typeFilter, setTypeFilter] = useState<TaskType | undefined>();
  const [stateFilter, setStateFilter] = useState<TaskState | undefined>();
  const [autoRefresh, setAutoRefresh] = useState(true);
  const [refreshInterval] = useState(5000);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  // Fetch tasks with real-time updates
  const {
    data: tasksData,
    isLoading,
    refetch,
  } = useQuery({
    queryKey: [
      'tasks',
      pagination.current,
      pagination.pageSize,
      searchText,
      typeFilter,
      stateFilter,
    ],
    queryFn: () =>
      taskApi.getTasks({
        page: pagination.current,
        size: pagination.pageSize,
        taskType: typeFilter ? ({
          'SubmitInjection': 0,
          'BuildDatapack': 1,
          'FaultInjection': 2,
          'CollectResult': 3,
          'AlgorithmExecution': 4,
        }[typeFilter]) : undefined,
        state: stateFilter,
      }),
    refetchInterval: autoRefresh ? refreshInterval : false,
  });

  // Real-time updates via SSE for running tasks
  useEffect(() => {
    if (!autoRefresh) return;

    const runningTasks = tasksData?.data?.items?.filter(
      (t) => t.state === 'RUNNING' || t.state === 1
    ); // RUNNING
    if (!runningTasks?.length) return;

    // Create SSE connections for each running task
    const eventSources: EventSource[] = [];

    runningTasks.forEach((task) => {
      const eventSource = new EventSource(
        `/api/v2/traces/${task.trace_id}/stream`
      );

      eventSource.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          // Update task status based on SSE data
          if (data.type === 'task_update') {
            message.info(`Task ${task.id} update: ${data.message}`);
            refetch();
          }
        } catch (error) {
          console.error('Error parsing SSE data:', error);
        }
      };

      eventSource.onerror = (error) => {
        console.error('SSE error:', error);
        eventSource.close();
      };

      eventSources.push(eventSource);
    });

    return () => {
      eventSources.forEach((es) => es.close());
    };
  }, [autoRefresh, tasksData, refetch]);

  // Statistics
  const stats = {
    total: tasksData?.data?.pagination?.total || 0,
    pending:
      tasksData?.data?.items?.filter(
        (t) => t.state === 0 || t.state === 'PENDING'
      ).length || 0, // PENDING
    running:
      tasksData?.data?.items?.filter(
        (t) => t.state === 1 || t.state === 'RUNNING'
      ).length || 0, // RUNNING
    completed:
      tasksData?.data?.items?.filter(
        (t) => t.state === 2 || t.state === 'COMPLETED'
      ).length || 0, // COMPLETED
    error:
      tasksData?.data?.items?.filter(
        (t) => t.state === 3 || t.state === 'ERROR'
      ).length || 0, // ERROR
    cancelled:
      tasksData?.data?.items?.filter(
        (t) => t.state === 4 || t.state === 'CANCELLED'
      ).length || 0, // CANCELLED
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

  const handleTypeFilter = (type: TaskType | undefined) => {
    setTypeFilter(type);
    setPagination({ ...pagination, current: 1 });
  };

  const handleStateFilter = (state: TaskState | undefined) => {
    setStateFilter(state);
    setPagination({ ...pagination, current: 1 });
  };

  const handleViewTask = (id: string) => {
    navigate(`/tasks/${id}`);
  };

  const handleCancelTask = (task: Task) => {
    if (task.state !== 1 && task.state !== 0) {
      // Not RUNNING or PENDING
      message.warning('Only running or pending tasks can be cancelled');
      return;
    }

    Modal.confirm({
      title: 'Cancel Task',
      content: `Are you sure you want to cancel task "${task.id}"?`,
      okText: 'Yes, cancel it',
      okButtonProps: { danger: true },
      cancelText: 'No',
      onOk: async () => {
        try {
          // TODO: Implement task cancellation when API is ready
          message.success('Task cancellation requested');
          refetch();
        } catch (error) {
          message.error('Failed to cancel task');
        }
      },
    });
  };

  const handleDeleteTask = (id: string) => {
    Modal.confirm({
      title: 'Delete Task',
      content:
        'Are you sure you want to delete this task? This action cannot be undone.',
      okText: 'Yes, delete it',
      okButtonProps: { danger: true },
      cancelText: 'Cancel',
      onOk: async () => {
        try {
          await taskApi.batchDelete([id]);
          message.success('Task deleted successfully');
          refetch();
        } catch (error) {
          message.error('Failed to delete task');
        }
      },
    });
  };

  const handleBatchDelete = () => {
    if (selectedRowKeys.length === 0) {
      message.warning('Please select tasks to delete');
      return;
    }

    Modal.confirm({
      title: 'Batch Delete Tasks',
      content: `Are you sure you want to delete ${selectedRowKeys.length} tasks?`,
      okText: 'Yes, delete them',
      okButtonProps: { danger: true },
      cancelText: 'Cancel',
      onOk: async () => {
        try {
          await taskApi.batchDelete(selectedRowKeys as string[]);
          message.success(
            `${selectedRowKeys.length} tasks deleted successfully`
          );
          setSelectedRowKeys([]);
          refetch();
        } catch (error) {
          message.error('Failed to delete tasks');
        }
      },
    });
  };

  const handleManualRefresh = () => {
    refetch();
    message.success('Tasks refreshed');
  };

  const getTaskTypeIcon = (type: TaskType) => {
    switch (type) {
      case 'SubmitInjection':
        return <PlayCircleOutlined style={{ color: '#3b82f6' }} />;
      case 'BuildDatapack':
        return <DashboardOutlined style={{ color: '#10b981' }} />;
      case 'FaultInjection':
        return <SyncOutlined style={{ color: '#f59e0b' }} />;
      case 'CollectResult':
        return <DatabaseOutlined style={{ color: '#8b5cf6' }} />;
      case 'AlgorithmExecution':
        return <FunctionOutlined style={{ color: '#ec4899' }} />;
      default:
        return <ClockCircleOutlined />;
    }
  };

  const getTaskTypeColor = (type: TaskType) => {
    switch (type) {
      case 'SubmitInjection':
        return '#3b82f6';
      case 'BuildDatapack':
        return '#10b981';
      case 'FaultInjection':
        return '#f59e0b';
      case 'CollectResult':
        return '#8b5cf6';
      case 'AlgorithmExecution':
        return '#ec4899';
      default:
        return '#6b7280';
    }
  };

  const getStateColor = (state: TaskState) => {
    switch (state) {
      case 0: // PENDING
        return '#d1d5db';
      case 1: // RUNNING
        return '#3b82f6';
      case 2: // COMPLETED
        return '#10b981';
      case 3: // ERROR
        return '#ef4444';
      case 4: // CANCELLED
        return '#6b7280';
      default:
        return '#6b7280';
    }
  };

  const getStateIcon = (state: TaskState) => {
    switch (state) {
      case 0: // PENDING
        return <ClockCircleOutlined />;
      case 1: // RUNNING
        return <SyncOutlined spin />;
      case 2: // COMPLETED
        return <CheckCircleOutlined />;
      case 3: // ERROR
        return <CloseCircleOutlined />;
      case 4: // CANCELLED
        return <PauseCircleOutlined />;
      default:
        return <ClockCircleOutlined />;
    }
  };

  const formatRetryInfo = (retryCount: number, maxRetry: number) => {
    if (retryCount === 0) return '0/0';
    return `${retryCount}/${maxRetry}`;
  };

  const getTaskProgress = (task: Task) => {
    if (task.state === 2) return 100; // COMPLETED
    if (task.state === 3 || task.state === 4) return 0; // ERROR or CANCELLED
    if (task.state === 1) return 50; // RUNNING
    return 0;
  };

  const rowSelection = {
    selectedRowKeys,
    onChange: setSelectedRowKeys,
  };

  const columns = [
    {
      title: 'Task',
      dataIndex: 'id',
      key: 'id',
      width: '15%',
      render: (id: string, record: Task) => (
        <Space>
          <Avatar
            size='small'
            style={{ backgroundColor: getTaskTypeColor(record.type) }}
            icon={getTaskTypeIcon(record.type)}
          />
          <div>
            <Text strong style={{ fontSize: '0.875rem' }}>
              {id.substring(0, 8)}
            </Text>
            <br />
            <Text type='secondary' style={{ fontSize: '0.75rem' }}>
              {record.type}
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
      render: (type: TaskType) => (
        <Tag color={getTaskTypeColor(type)} style={{ fontWeight: 500 }}>
          {type}
        </Tag>
      ),
      filters: [
        { text: 'Submit Injection', value: 'SubmitInjection' },
        { text: 'Build Datapack', value: 'BuildDatapack' },
        { text: 'Fault Injection', value: 'FaultInjection' },
        { text: 'Collect Result', value: 'CollectResult' },
        { text: 'Algorithm Execution', value: 'AlgorithmExecution' },
      ],
      onFilter: (value: React.Key, record: Task) => record.type === value,
    },
    {
      title: 'Status',
      dataIndex: 'state',
      key: 'state',
      width: '12%',
      render: (state: TaskState) => (
        <Badge
          status={
            state === 2
              ? 'success' // COMPLETED
              : state === 3
                ? 'error' // ERROR
                : state === 1
                  ? 'processing' // RUNNING
                  : state === 4
                    ? 'warning' // CANCELLED
                    : 'default'
          }
          text={
            <Space size='small'>
              {getStateIcon(state)}
              <Text
                strong
                style={{ color: getStateColor(state), fontSize: '0.875rem' }}
              >
                {state === 0
                  ? 'Pending' // PENDING
                  : state === 1
                    ? 'Running' // RUNNING
                    : state === 2
                      ? 'Completed' // COMPLETED
                      : state === 3
                        ? 'Error' // ERROR
                        : state === 4
                          ? 'Cancelled' // CANCELLED
                          : 'Unknown'}
              </Text>
            </Space>
          }
        />
      ),
      filters: [
        { text: 'Pending', value: 0 },
        { text: 'Running', value: 1 },
        { text: 'Completed', value: 2 },
        { text: 'Error', value: 3 },
        { text: 'Cancelled', value: 4 },
      ],
      onFilter: (value: React.Key, record: Task) => record.state === value,
    },
    {
      title: 'Progress',
      key: 'progress',
      width: '10%',
      render: (_: string, record: Task) => (
        <Progress
          percent={getTaskProgress(record)}
          status={
            record.state === 3
              ? 'exception' // ERROR
              : record.state === 2
                ? 'success' // COMPLETED
                : 'active'
          }
          size='small'
          format={(percent) => `${percent}%`}
        />
      ),
    },
    {
      title: 'Retries',
      key: 'retries',
      width: '8%',
      render: (_: string, record: Task) => (
        <Text code style={{ fontSize: '0.75rem' }}>
          {formatRetryInfo(record.retry_count, record.max_retry)}
        </Text>
      ),
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      width: '12%',
      render: (date: string) => (
        <Tooltip title={dayjs(date).format('YYYY-MM-DD HH:mm:ss')}>
          <Text style={{ fontSize: '0.75rem' }}>{dayjs(date).fromNow()}</Text>
        </Tooltip>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: '12%',
      render: (_: string, record: Task) => (
        <Space size='small'>
          <Tooltip title='View Details'>
            <Button
              type='text'
              size='small'
              icon={<EyeOutlined />}
              onClick={() => handleViewTask(record.id)}
            />
          </Tooltip>
          {record.state === 1 && ( // RUNNING
            <Tooltip title='Cancel Task'>
              <Button
                type='text'
                size='small'
                danger
                icon={<CloseCircleOutlined />}
                onClick={() => handleCancelTask(record)}
              />
            </Tooltip>
          )}
          <Tooltip title='Delete Task'>
            <Button
              type='text'
              size='small'
              danger
              icon={<DeleteOutlined />}
              onClick={() => handleDeleteTask(record.id)}
            />
          </Tooltip>
        </Space>
      ),
    },
  ];

  return (
    <div className='task-list'>
      {/* Page Header */}
      <div className='page-header'>
        <div className='page-header-left'>
          <Title level={2} className='page-title'>
            Task Monitor
          </Title>
          <Text type='secondary'>
            Monitor and manage background tasks with real-time updates
          </Text>
        </div>
        <div className='page-header-right'>
          <Space>
            <Button icon={<ReloadOutlined />} onClick={handleManualRefresh}>
              Refresh
            </Button>
            <Button
              type={autoRefresh ? 'primary' : 'default'}
              icon={<SyncOutlined spin={autoRefresh} />}
              onClick={() => setAutoRefresh(!autoRefresh)}
            >
              {autoRefresh ? 'Auto-refresh ON' : 'Auto-refresh OFF'}
            </Button>
          </Space>
        </div>
      </div>

      {/* Statistics Cards */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} md={4}>
          <Card>
            <Statistic
              title='Total Tasks'
              value={stats.total}
              prefix={<DashboardOutlined />}
              valueStyle={{ color: '#3b82f6' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={4}>
          <Card>
            <Statistic
              title='Pending'
              value={stats.pending}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: '#6b7280' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={4}>
          <Card>
            <Statistic
              title='Running'
              value={stats.running}
              prefix={<SyncOutlined />}
              valueStyle={{ color: '#3b82f6' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={4}>
          <Card>
            <Statistic
              title='Completed'
              value={stats.completed}
              prefix={<CheckCircleOutlined />}
              valueStyle={{ color: '#10b981' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={4}>
          <Card>
            <Statistic
              title='Error'
              value={stats.error}
              prefix={<CloseCircleOutlined />}
              valueStyle={{ color: '#ef4444' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} md={4}>
          <Card>
            <Statistic
              title='Cancelled'
              value={stats.cancelled}
              prefix={<PauseCircleOutlined />}
              valueStyle={{ color: '#6b7280' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Filters and Actions */}
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={[16, 16]} align='middle'>
          <Col xs={24} sm={12} md={6}>
            <Search
              placeholder='Search tasks by ID or type...'
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
              <Option value='SubmitInjection'>Submit Injection</Option>
              <Option value='BuildDatapack'>Build Datapack</Option>
              <Option value='FaultInjection'>Fault Injection</Option>
              <Option value='CollectResult'>Collect Result</Option>
              <Option value='AlgorithmExecution'>Algorithm Execution</Option>
            </Select>
          </Col>
          <Col xs={24} sm={12} md={4}>
            <Select
              placeholder='Filter by status'
              allowClear
              style={{ width: '100%' }}
              onChange={handleStateFilter}
              value={stateFilter}
            >
              <Option value={0}>Pending</Option>
              <Option value={1}>Running</Option>
              <Option value={2}>Completed</Option>
              <Option value={3}>Error</Option>
              <Option value={4}>Cancelled</Option>
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
              <Button icon={<ExportOutlined />}>Export</Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* Task Table */}
      <Card>
        <Table
          rowKey='id'
          rowSelection={rowSelection}
          columns={columns}
          dataSource={(tasksData?.data?.items as Task[] | undefined) || []}
          loading={isLoading}
          pagination={{
            ...pagination,
            total: (tasksData?.data as any)?.total || 0,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `${range[0]}-${range[1]} of ${total} tasks`,
          }}
          onChange={handleTableChange}
          locale={{
            emptyText: <Empty description='No tasks found' />,
          }}
        />
      </Card>

      {/* Real-time Status Indicator */}
      {autoRefresh && (
        <div style={{ position: 'fixed', bottom: 24, right: 24 }}>
          <Card size='small' style={{ width: 200 }}>
            <Space>
              <Badge status='processing' />
              <Text type='secondary'>Real-time updates active</Text>
            </Space>
          </Card>
        </div>
      )}
    </div>
  );
};

export default TaskList;
