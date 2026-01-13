import { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';

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

import { taskApi } from '@/api/tasks';
import { TaskState, TaskType } from '@/types/api';
import type { Task } from '@/types/api';

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
        taskType: typeFilter,
        state: stateFilter,
      }),
    refetchInterval: autoRefresh ? refreshInterval : false,
  });

  // Real-time updates via SSE for running tasks
  useEffect(() => {
    if (!autoRefresh) return;

    const runningTasks = tasksData?.data?.items?.filter(
      (t) =>
        String(t.state) === String(TaskState.RUNNING) ||
        t.state === '2' ||
        t.state === 'RUNNING'
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
        (t) => String(t.state) === String(TaskState.PENDING)
      ).length || 0, // PENDING
    running:
      tasksData?.data?.items?.filter(
        (t) => String(t.state) === String(TaskState.RUNNING)
      ).length || 0, // RUNNING
    completed:
      tasksData?.data?.items?.filter(
        (t) => String(t.state) === String(TaskState.COMPLETED)
      ).length || 0, // COMPLETED
    error:
      tasksData?.data?.items?.filter(
        (t) => String(t.state) === String(TaskState.ERROR)
      ).length || 0, // ERROR
    cancelled:
      tasksData?.data?.items?.filter(
        (t) => String(t.state) === String(TaskState.CANCELLED)
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

  const handleViewTask = (id?: string) => {
    if (id) {
      navigate(`/tasks/${id}`);
    }
  };

  const handleCancelTask = (task: Task) => {
    if (
      String(task.state) !== String(TaskState.RUNNING) &&
      String(task.state) !== String(TaskState.PENDING)
    ) {
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

  const handleDeleteTask = (id?: string) => {
    if (!id) return;

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

  const getTaskTypeIcon = (type: TaskType | string | undefined) => {
    if (!type) return <ClockCircleOutlined />;

    const taskType =
      typeof type === 'string'
        ? (Object.values(TaskType).find((v) => v === type) as TaskType)
        : type;

    switch (taskType) {
      case TaskType.BuildContainer:
        return <FunctionOutlined style={{ color: '#3b82f6' }} />;
      case TaskType.RestartPedestal:
        return <SyncOutlined style={{ color: '#10b981' }} />;
      case TaskType.FaultInjection:
        return <SyncOutlined style={{ color: '#f59e0b' }} />;
      case TaskType.RunAlgorithm:
        return <FunctionOutlined style={{ color: '#ec4899' }} />;
      case TaskType.BuildDatapack:
        return <DashboardOutlined style={{ color: '#10b981' }} />;
      case TaskType.CollectResult:
        return <DatabaseOutlined style={{ color: '#8b5cf6' }} />;
      case TaskType.CronJob:
        return <ClockCircleOutlined style={{ color: '#6b7280' }} />;
      default:
        return <ClockCircleOutlined />;
    }
  };

  const getTaskTypeColor = (type: TaskType | string | undefined): string => {
    if (!type) return '#6b7280';

    const taskType =
      typeof type === 'string'
        ? (Object.values(TaskType).find((v) => v === type) as TaskType)
        : type;

    switch (taskType) {
      case TaskType.BuildContainer:
        return '#3b82f6';
      case TaskType.RestartPedestal:
        return '#10b981';
      case TaskType.FaultInjection:
        return '#f59e0b';
      case TaskType.RunAlgorithm:
        return '#ec4899';
      case TaskType.BuildDatapack:
        return '#10b981';
      case TaskType.CollectResult:
        return '#8b5cf6';
      case TaskType.CronJob:
        return '#6b7280';
      default:
        return '#6b7280';
    }
  };

  const getStateColor = (state: TaskState) => {
    switch (state) {
      case TaskState.PENDING: // PENDING
        return '#d1d5db';
      case TaskState.RUNNING: // RUNNING
        return '#3b82f6';
      case TaskState.COMPLETED: // COMPLETED
        return '#10b981';
      case TaskState.ERROR: // ERROR
        return '#ef4444';
      case TaskState.CANCELLED: // CANCELLED
        return '#6b7280';
      default:
        return '#6b7280';
    }
  };

  const getStateIcon = (state: TaskState) => {
    switch (state) {
      case TaskState.PENDING: // PENDING
        return <ClockCircleOutlined />;
      case TaskState.RUNNING: // RUNNING
        return <SyncOutlined spin />;
      case TaskState.COMPLETED: // COMPLETED
        return <CheckCircleOutlined />;
      case TaskState.ERROR: // ERROR
        return <CloseCircleOutlined />;
      case TaskState.CANCELLED: // CANCELLED
        return <PauseCircleOutlined />;
      default:
        return <ClockCircleOutlined />;
    }
  };

  const formatRetryInfo = (retryCount?: number, maxRetry?: number): string => {
    const count = retryCount ?? 0;
    const max = maxRetry ?? 0;
    return `${count}/${max}`;
  };

  const getTaskProgress = (task: Task) => {
    if (String(task.state) === String(TaskState.COMPLETED)) return 100; // COMPLETED
    if (
      String(task.state) === String(TaskState.ERROR) ||
      String(task.state) === String(TaskState.CANCELLED)
    )
      return 0; // ERROR or CANCELLED
    if (String(task.state) === String(TaskState.RUNNING)) return 50; // RUNNING
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
      onFilter: (value: boolean | React.Key, record: Task) =>
        record.type === value,
    },
    {
      title: 'Status',
      dataIndex: 'state',
      key: 'state',
      width: '12%',
      render: (state: string | TaskState) => {
        const stateStr = String(state);
        let taskState: TaskState;
        if (!isNaN(Number(stateStr))) {
          taskState = Number(stateStr) as TaskState;
        } else {
          taskState = TaskState.PENDING;
        }

        return (
          <Badge
            status={
              taskState === TaskState.COMPLETED
                ? 'success' // COMPLETED
                : taskState === TaskState.ERROR
                  ? 'error' // ERROR
                  : taskState === TaskState.RUNNING
                    ? 'processing' // RUNNING
                    : taskState === TaskState.CANCELLED
                      ? 'warning' // CANCELLED
                      : 'default'
            }
            text={
              <Space size='small'>
                {getStateIcon(taskState)}
                <Text
                  strong
                  style={{
                    color: getStateColor(taskState),
                    fontSize: '0.875rem',
                  }}
                >
                  {taskState === TaskState.PENDING
                    ? 'Pending' // PENDING
                    : taskState === TaskState.RUNNING
                      ? 'Running' // RUNNING
                      : taskState === TaskState.COMPLETED
                        ? 'Completed' // COMPLETED
                        : taskState === TaskState.ERROR
                          ? 'Error' // ERROR
                          : taskState === TaskState.CANCELLED
                            ? 'Cancelled' // CANCELLED
                            : 'Unknown'}
                </Text>
              </Space>
            }
          />
        );
      },
      filters: [
        { text: 'Pending', value: 0 },
        { text: 'Running', value: 1 },
        { text: 'Completed', value: 2 },
        { text: 'Error', value: 3 },
        { text: 'Cancelled', value: 4 },
      ],
      onFilter: (value: boolean | React.Key, record: Task) =>
        String(record.state) === String(value),
    },
    {
      title: 'Progress',
      key: 'progress',
      width: '10%',
      render: (_: unknown, record: Task) => (
        <Progress
          percent={getTaskProgress(record)}
          status={
            String(record.state) === String(TaskState.ERROR)
              ? 'exception' // ERROR
              : String(record.state) === String(TaskState.COMPLETED)
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
      render: (_: unknown, record: Task) => (
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
      render: (_: unknown, record: Task) => (
        <Space size='small'>
          <Tooltip title='View Details'>
            <Button
              type='text'
              size='small'
              icon={<EyeOutlined />}
              onClick={() => handleViewTask(record.id)}
            />
          </Tooltip>
          {String(record.state) === String(TaskState.RUNNING) && ( // RUNNING
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
            total: tasksData?.data?.pagination?.total || 0,
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
