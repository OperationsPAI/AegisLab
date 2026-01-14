import {
  ArrowLeftOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
  CopyOutlined,
  DashboardOutlined,
  DatabaseOutlined,
  DownloadOutlined,
  FunctionOutlined,
  PauseCircleOutlined,
  PlayCircleOutlined,
  ReloadOutlined,
  SyncOutlined,
} from '@ant-design/icons';
import { TaskState, TaskType } from '@rcabench/client';
import type { TaskDetailResp } from '@rcabench/client';
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
  Modal,
  Progress,
  Row,
  Space,
  Switch,
  Tabs,
  Tag,
  Timeline,
  Typography,
} from 'antd';
import dayjs from 'dayjs';
import duration from 'dayjs/plugin/duration';
import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';

import { taskApi } from '@/api/tasks';
import StatusBadge from '@/components/ui/StatusBadge';


dayjs.extend(duration);

const { Title, Text } = Typography;
// Removed deprecated TabPane destructuring - using items prop instead

const TaskDetail = () => {
  const navigate = useNavigate();
  const { id } = useParams<{ id: string }>();
  const [activeTab, setActiveTab] = useState('overview');
  const [logs, setLogs] = useState<string[]>([]);
  const [autoRefresh, setAutoRefresh] = useState(true);

  // Validate taskId exists - must be before hooks
  const taskId = id;

  // Fetch task details
  const {
    data: task,
    isLoading,
    refetch,
  } = useQuery({
    queryKey: ['task', taskId],
    queryFn: async () => {
      if (!taskId) {
        throw new Error('Task ID is required');
      }
      return taskApi.getTask(taskId);
    },
    refetchInterval: autoRefresh ? 2000 : false,
    enabled: !!taskId,
  });

  // Real-time log streaming via SSE
  useEffect(() => {
    if (!task || task.state !== '1') return; // 1 = RUNNING

    const eventSource = new EventSource(
      `/api/v2/traces/${task.trace_id}/stream`
    );

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data);
        if (data.type === 'log') {
          setLogs((prev) => [
            ...prev,
            `[${dayjs().format('HH:mm:ss')}] ${data.message}`,
          ]);
        } else if (data.type === 'task_update') {
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

    return () => {
      eventSource.close();
    };
  }, [task, refetch]);

  const handleCancelTask = () => {
    if (task?.state !== '1' && task?.state !== '0') {
      // Not RUNNING or PENDING
      message.warning('Only running or pending tasks can be cancelled');
      return;
    }

    Modal.confirm({
      title: 'Cancel Task',
      content: `Are you sure you want to cancel task "${taskId}"?`,
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

  const handleRetryTask = () => {
    if (getTaskState(task?.state) !== TaskState.Error) {
      message.warning('Only failed tasks can be retried');
      return;
    }

    Modal.confirm({
      title: 'Retry Task',
      content: `Are you sure you want to retry task "${taskId}"?`,
      okText: 'Yes, retry it',
      cancelText: 'No',
      onOk: async () => {
        try {
          // TODO: Implement task retry when API is ready
          message.success('Task retry requested');
          refetch();
        } catch (error) {
          message.error('Failed to retry task');
        }
      },
    });
  };

  const handleDownloadLogs = () => {
    const logContent = logs.join('\n');
    const blob = new Blob([logContent], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `task-${taskId}-logs.txt`;
    document.body.appendChild(a);
    a.click();
    document.body.removeChild(a);
    URL.revokeObjectURL(url);
    message.success('Logs downloaded successfully');
  };

  const getTaskTypeIcon = (type: TaskType) => {
    switch (type) {
      case TaskType.BuildContainer:
        return <DashboardOutlined style={{ color: '#10b981' }} />;
      case TaskType.RestartPedestal:
        return <ReloadOutlined style={{ color: '#3b82f6' }} />;
      case TaskType.FaultInjection:
        return <SyncOutlined style={{ color: '#f59e0b' }} />;
      case TaskType.RunAlgorithm:
        return <FunctionOutlined style={{ color: '#ec4899' }} />;
      case TaskType.BuildDatapack:
        return <DatabaseOutlined style={{ color: '#8b5cf6' }} />;
      case TaskType.CollectResult:
        return <DatabaseOutlined style={{ color: '#8b5cf6' }} />;
      case TaskType.CronJob:
        return <ClockCircleOutlined style={{ color: '#6b7280' }} />;
      default:
        return <ClockCircleOutlined />;
    }
  };

  const getTaskTypeColor = (type: TaskType) => {
    switch (type) {
      case TaskType.BuildContainer:
        return '#10b981';
      case TaskType.RestartPedestal:
        return '#3b82f6';
      case TaskType.FaultInjection:
        return '#f59e0b';
      case TaskType.RunAlgorithm:
        return '#ec4899';
      case TaskType.BuildDatapack:
        return '#8b5cf6';
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
      case TaskState.Pending:
        return '#d1d5db';
      case TaskState.Rescheduled:
        return '#9ca3af';
      case TaskState.Running:
        return '#3b82f6';
      case TaskState.Completed:
        return '#10b981';
      case TaskState.Error:
        return '#ef4444';
      case TaskState.Cancelled:
        return '#6b7280';
      default:
        return '#6b7280';
    }
  };

  const getStateIcon = (state: TaskState) => {
    switch (state) {
      case TaskState.Pending:
        return <ClockCircleOutlined />;
      case TaskState.Rescheduled:
        return <ClockCircleOutlined style={{ color: '#9ca3af' }} />;
      case TaskState.Running:
        return <SyncOutlined spin />;
      case TaskState.Completed:
        return <CheckCircleOutlined />;
      case TaskState.Error:
        return <CloseCircleOutlined />;
      case TaskState.Cancelled:
        return <PauseCircleOutlined />;
      default:
        return <ClockCircleOutlined />;
    }
  };

  if (isLoading) {
    return (
      <div style={{ padding: 24 }}>
        <Card loading>
          <div style={{ minHeight: 400 }} />
        </Card>
      </div>
    );
  }

  if (!taskId) {
    return (
      <div style={{ padding: 24, textAlign: 'center' }}>
        <Text type='secondary'>Task ID not provided</Text>
      </div>
    );
  }

  if (!task) {
    return (
      <div style={{ padding: 24, textAlign: 'center' }}>
        <Text type='secondary'>Task not found</Text>
      </div>
    );
  }

  const taskData: TaskDetailResp | undefined = task;

  // Helper to convert string state to TaskState enum
  const getTaskState = (state?: string): TaskState => {
    const numState = parseInt(state || '0', 10);
    if (Object.values(TaskState).includes(numState as TaskState)) {
      return numState as TaskState;
    }
    return TaskState.Pending;
  };

  // Helper to convert string type to TaskType enum
  const getTaskType = (type?: string): TaskType => {
    const numType = parseInt(type || '0', 10);
    if (Object.values(TaskType).includes(numType as TaskType)) {
      return numType as TaskType;
    }
    return TaskType.BuildContainer;
  };

  // Safe version of getTaskProgress
  const getSafeTaskProgress = (taskData?: TaskDetailResp): number => {
    if (!taskData) return 0;
    const state = getTaskState(taskData.state);
    if (state === TaskState.Completed) return 100;
    if (state === TaskState.Error || state === TaskState.Cancelled) return 0;
    if (state === TaskState.Running) return 50;
    if (state === TaskState.Rescheduled) return 25;
    return 0;
  };

  const progress = getSafeTaskProgress(taskData);

  return (
    <div style={{ padding: 24 }}>
      {/* Header */}
      <div style={{ marginBottom: 24 }}>
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/tasks')}
          >
            Back to List
          </Button>
          <Title level={2} style={{ margin: 0 }}>
            Task {taskId.substring(0, 8)}
          </Title>
          <Badge
            status={
              (getTaskState(taskData?.state) === TaskState.Completed
                ? 'success' // COMPLETED
                : getTaskState(taskData?.state) === TaskState.Error
                  ? 'error' // ERROR
                  : getTaskState(taskData?.state) === TaskState.Running
                    ? 'processing' // RUNNING
                    : getTaskState(taskData?.state) === TaskState.Cancelled
                      ? 'warning' // CANCELLED
                      : 'default') as
                | 'success'
                | 'processing'
                | 'error'
                | 'default'
                | 'warning'
            }
            text={
              <Space>
                {getStateIcon(getTaskState(taskData?.state))}
                <Text
                  strong
                  style={{
                    color: getStateColor(getTaskState(taskData?.state)),
                  }}
                >
                  {getTaskState(taskData?.state) === TaskState.Pending
                    ? 'Pending' // PENDING
                    : getTaskState(taskData?.state) === TaskState.Running
                      ? 'Running' // RUNNING
                      : getTaskState(taskData?.state) === TaskState.Completed
                        ? 'Completed' // COMPLETED
                        : getTaskState(taskData?.state) === TaskState.Error
                          ? 'Error' // ERROR
                          : getTaskState(taskData?.state) ===
                              TaskState.Cancelled
                            ? 'Cancelled' // CANCELLED
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
              {(getTaskState(taskData?.state) === TaskState.Running ||
                getTaskState(taskData?.state) === TaskState.Pending) && ( // RUNNING or PENDING
                <Button
                  danger
                  icon={<PauseCircleOutlined />}
                  onClick={handleCancelTask}
                >
                  Cancel Task
                </Button>
              )}
              {getTaskState(taskData?.state) === TaskState.Error && ( // ERROR
                <Button
                  type='primary'
                  icon={<ReloadOutlined />}
                  onClick={handleRetryTask}
                >
                  Retry Task
                </Button>
              )}
              <Button
                icon={<DownloadOutlined />}
                onClick={handleDownloadLogs}
                disabled={logs.length === 0}
              >
                Download Logs
              </Button>
              <Button
                icon={<CopyOutlined />}
                onClick={() => {
                  navigator.clipboard.writeText(taskId);
                  message.success('Task ID copied to clipboard');
                }}
              >
                Copy ID
              </Button>
            </Space>
          </Col>
          <Col>
            <Space>
              <Text type='secondary'>Auto-refresh:</Text>
              <Switch
                checked={autoRefresh}
                onChange={setAutoRefresh}
                checkedChildren='ON'
                unCheckedChildren='OFF'
              />
            </Space>
          </Col>
        </Row>
      </Card>

      {/* Progress */}
      <Card style={{ marginBottom: 24 }}>
        <div style={{ marginBottom: 16 }}>
          <Text strong>Task Progress</Text>
        </div>
        <Progress
          percent={progress}
          status={
            getTaskState(taskData?.state) === TaskState.Error
              ? 'exception'
              : getTaskState(taskData?.state) === TaskState.Completed
                ? 'success'
                : 'active'
          }
          strokeColor={getStateColor(getTaskState(taskData?.state))}
          format={(percent) => (
            <Space>
              {getStateIcon(getTaskState(taskData?.state))}
              <Text>{percent}%</Text>
            </Space>
          )}
        />
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
                    <Card title='Task Information'>
                      <Descriptions column={2} bordered>
                        <Descriptions.Item label='Task ID'>
                          <Space>
                            <Text code>{taskId}</Text>
                            <Button
                              type='text'
                              size='small'
                              icon={<CopyOutlined />}
                              onClick={() => {
                                navigator.clipboard.writeText(taskId);
                                message.success('Task ID copied to clipboard');
                              }}
                            />
                          </Space>
                        </Descriptions.Item>
                        <Descriptions.Item label='Type'>
                          <Tag
                            color={getTaskTypeColor(
                              getTaskType(taskData?.type)
                            )}
                            style={{ fontWeight: 500, fontSize: '1rem' }}
                          >
                            <Space>
                              {getTaskTypeIcon(getTaskType(taskData?.type))}
                              {taskData?.type || 'Unknown'}
                            </Space>
                          </Tag>
                        </Descriptions.Item>
                        <Descriptions.Item label='Status'>
                          <StatusBadge
                            status={
                              getTaskState(taskData?.state) ===
                              TaskState.Completed
                                ? 'completed'
                                : getTaskState(taskData?.state) ===
                                    TaskState.Error
                                  ? 'error'
                                  : getTaskState(taskData?.state) ===
                                      TaskState.Running
                                    ? 'running'
                                    : getTaskState(taskData?.state) ===
                                        TaskState.Cancelled
                                      ? 'warning'
                                      : 'pending'
                            }
                            text={
                              getTaskState(taskData?.state) ===
                              TaskState.Pending
                                ? 'Pending'
                                : getTaskState(taskData?.state) ===
                                    TaskState.Running
                                  ? 'Running'
                                  : getTaskState(taskData?.state) ===
                                      TaskState.Completed
                                    ? 'Completed'
                                    : getTaskState(taskData?.state) ===
                                        TaskState.Error
                                      ? 'Error'
                                      : getTaskState(taskData?.state) ===
                                          TaskState.Cancelled
                                        ? 'Cancelled'
                                        : 'Unknown'
                            }
                          />
                        </Descriptions.Item>
                        <Descriptions.Item label='Retry Count'>
                          <Text code>N/A</Text>
                        </Descriptions.Item>
                        <Descriptions.Item label='Immediate'>
                          <Text>{taskData?.immediate ? 'Yes' : 'No'}</Text>
                        </Descriptions.Item>
                        <Descriptions.Item label='Trace ID'>
                          <Text code>{taskData?.trace_id || 'N/A'}</Text>
                        </Descriptions.Item>
                        <Descriptions.Item label='Group ID'>
                          <Text code>{taskData?.group_id || 'N/A'}</Text>
                        </Descriptions.Item>
                        <Descriptions.Item label='Project ID'>
                          <Text>{taskData?.project_id || 'N/A'}</Text>
                        </Descriptions.Item>
                        <Descriptions.Item label='Status Code'>
                          <Text code>{taskData?.status || 'N/A'}</Text>
                        </Descriptions.Item>
                      </Descriptions>
                    </Card>
                  </Col>
                  <Col xs={24} lg={8}>
                    <Card title='Timing Information'>
                      <Space direction='vertical' style={{ width: '100%' }}>
                        <div>
                          <Text type='secondary'>Created</Text>
                          <br />
                          <Text strong>
                            {taskData?.created_at
                              ? dayjs(taskData.created_at).format(
                                  'MMM D, YYYY HH:mm:ss'
                                )
                              : 'N/A'}
                          </Text>
                        </div>
                        <Divider />
                        <div>
                          <Text type='secondary'>Duration</Text>
                          <br />
                          <Title
                            level={3}
                            style={{ margin: 0, color: '#3b82f6' }}
                          >
                            N/A
                          </Title>
                        </div>
                      </Space>
                    </Card>
                  </Col>
                </Row>

                {taskData?.payload !== undefined &&
                  taskData?.payload !== null && (
                    <Card title='Payload' style={{ marginTop: 16 }}>
                      <pre
                        style={{
                          margin: 0,
                          fontSize: '0.875rem',
                          whiteSpace: 'pre-wrap',
                        }}
                      >
                        {JSON.stringify(taskData.payload, null, 2)}
                      </pre>
                    </Card>
                  )}
              </>
            ),
          },

          {
            key: 'logs',
            label: 'Logs',
            children: (
              <Card
                title='Task Logs'
                extra={
                  <Space>
                    <Button
                      icon={<ReloadOutlined />}
                      onClick={() => setLogs([])}
                    >
                      Clear Logs
                    </Button>
                    <Button
                      icon={<DownloadOutlined />}
                      onClick={handleDownloadLogs}
                      disabled={logs.length === 0}
                    >
                      Download
                    </Button>
                  </Space>
                }
              >
                {logs.length > 0 ? (
                  <div
                    style={{
                      background: '#f5f5f5',
                      padding: 16,
                      borderRadius: 4,
                      maxHeight: 400,
                      overflow: 'auto',
                    }}
                  >
                    <pre
                      style={{
                        margin: 0,
                        fontSize: '0.875rem',
                        fontFamily: 'monospace',
                      }}
                    >
                      {logs.join('\n')}
                    </pre>
                  </div>
                ) : (
                  <Empty
                    description='No logs available. Logs will appear when the task starts running.'
                    image={Empty.PRESENTED_IMAGE_SIMPLE}
                  />
                )}
              </Card>
            ),
          },

          {
            key: 'timeline',
            label: 'Timeline',
            children: (
              <Card title='Task Execution Timeline'>
                <Timeline>
                  <Timeline.Item color='blue' dot={<ClockCircleOutlined />}>
                    <Text strong>Task Created</Text>
                    <br />
                    <Text type='secondary'>
                      {taskData?.created_at
                        ? dayjs(taskData.created_at).format(
                            'MMM D, YYYY HH:mm:ss'
                          )
                        : 'N/A'}
                    </Text>
                  </Timeline.Item>

                  {getTaskState(taskData?.state) === TaskState.Running && (
                    <Timeline.Item color='green' dot={<PlayCircleOutlined />}>
                      <Text strong>Task Started</Text>
                      <br />
                      <Text type='secondary'>
                        {taskData?.updated_at
                          ? dayjs(taskData.updated_at).format(
                              'MMM D, YYYY HH:mm:ss'
                            )
                          : 'N/A'}
                      </Text>
                    </Timeline.Item>
                  )}

                  {getTaskState(taskData?.state) === TaskState.Running && (
                    <Timeline.Item color='blue' dot={<SyncOutlined spin />}>
                      <Text strong>Task Running</Text>
                      <br />
                      <Text type='secondary'>In progress...</Text>
                    </Timeline.Item>
                  )}

                  {getTaskState(taskData?.state) === TaskState.Completed && (
                      <Timeline.Item
                        color='green'
                        dot={<CheckCircleOutlined />}
                      >
                        <Text strong>Task Completed</Text>
                        <br />
                        <Text type='secondary'>
                          {taskData?.updated_at
                            ? dayjs(taskData.updated_at).format(
                                'MMM D, YYYY HH:mm:ss'
                              )
                            : 'N/A'}
                        </Text>
                      </Timeline.Item>
                    )}

                  {getTaskState(taskData?.state) === TaskState.Error && (
                      <Timeline.Item color='red' dot={<CloseCircleOutlined />}>
                        <Text strong>Task Failed</Text>
                        <br />
                        <Text type='secondary'>
                          {taskData?.updated_at
                            ? dayjs(taskData.updated_at).format(
                                'MMM D, YYYY HH:mm:ss'
                              )
                            : 'N/A'}
                        </Text>
                      </Timeline.Item>
                    )}

                  {getTaskState(taskData?.state) === TaskState.Cancelled && (
                      <Timeline.Item
                        color='orange'
                        dot={<PauseCircleOutlined />}
                      >
                        <Text strong>Task Cancelled</Text>
                        <br />
                        <Text type='secondary'>
                          {taskData?.updated_at
                            ? dayjs(taskData.updated_at).format(
                                'MMM D, YYYY HH:mm:ss'
                              )
                            : 'N/A'}
                        </Text>
                      </Timeline.Item>
                    )}
                </Timeline>
              </Card>
            ),
          },
        ]}
      />
    </div>
  );
};

export default TaskDetail;
