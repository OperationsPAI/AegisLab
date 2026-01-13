import {
  ArrowLeftOutlined,
  SyncOutlined,
  ClockCircleOutlined,
  CheckCircleOutlined,
  CloseCircleOutlined,
  PauseCircleOutlined,
  PlayCircleOutlined,
  DatabaseOutlined,
  FunctionOutlined,
  DashboardOutlined,
  TagsOutlined,
  CopyOutlined,
  DownloadOutlined,
  ReloadOutlined,
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
  Modal,
  message,
  Tabs,
  Badge,
  Divider,
  Progress,
  Empty,
  Timeline,
  Switch,
} from 'antd'
import dayjs from 'dayjs'
import duration from 'dayjs/plugin/duration'
import { useEffect, useState } from 'react'
import { useNavigate, useParams } from 'react-router-dom'

import { taskApi } from '@/api/tasks'
import StatusBadge from '@/components/ui/StatusBadge'
import type { Task, TaskState, TaskType } from '@/types/api'

dayjs.extend(duration)

const { Title, Text } = Typography
const { TabPane } = Tabs

const TaskDetail = () => {
  const navigate = useNavigate()
  const { id } = useParams<{ id: string }>()
  const taskId = id!
  const [activeTab, setActiveTab] = useState('overview')
  const [logs, setLogs] = useState<string[]>([])
  const [autoRefresh, setAutoRefresh] = useState(true)

  // Fetch task details
  const { data: task, isLoading, refetch } = useQuery({
    queryKey: ['task', taskId],
    queryFn: () => taskApi.getTask(taskId),
    refetchInterval: autoRefresh ? 2000 : false,
  })

  // Real-time log streaming via SSE
  useEffect(() => {
    if (!task || !task.data || task.data.state !== 1) return // 1 = RUNNING

    const eventSource = new EventSource(`/api/v2/traces/${task.data.trace_id}/stream`)

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data)
        if (data.type === 'log') {
          setLogs(prev => [...prev, `[${dayjs().format('HH:mm:ss')}] ${data.message}`])
        } else if (data.type === 'task_update') {
          refetch()
        }
      } catch (error) {
        console.error('Error parsing SSE data:', error)
      }
    }

    eventSource.onerror = (error) => {
      console.error('SSE error:', error)
      eventSource.close()
    }

    return () => {
      eventSource.close()
    }
  }, [task, refetch])

  const handleCancelTask = () => {
    if (task?.data?.state !== 1 && task?.data?.state !== 0) { // Not RUNNING or PENDING
      message.warning('Only running or pending tasks can be cancelled')
      return
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
          message.success('Task cancellation requested')
          refetch()
        } catch (error) {
          message.error('Failed to cancel task')
        }
      },
    })
  }

  const handleRetryTask = () => {
    if (task?.data.state !== TaskState.ERROR) {
      message.warning('Only failed tasks can be retried')
      return
    }

    Modal.confirm({
      title: 'Retry Task',
      content: `Are you sure you want to retry task "${taskId}"?`,
      okText: 'Yes, retry it',
      cancelText: 'No',
      onOk: async () => {
        try {
          // TODO: Implement task retry when API is ready
          message.success('Task retry requested')
          refetch()
        } catch (error) {
          message.error('Failed to retry task')
        }
      },
    })
  }

  const handleDownloadLogs = () => {
    const logContent = logs.join('\n')
    const blob = new Blob([logContent], { type: 'text/plain' })
    const url = URL.createObjectURL(blob)
    const a = document.createElement('a')
    a.href = url
    a.download = `task-${taskId}-logs.txt`
    document.body.appendChild(a)
    a.click()
    document.body.removeChild(a)
    URL.revokeObjectURL(url)
    message.success('Logs downloaded successfully')
  }

  const getTaskTypeIcon = (type: TaskType) => {
    switch (type) {
      case 'SubmitInjection':
        return <PlayCircleOutlined style={{ color: '#3b82f6' }} />
      case 'BuildDatapack':
        return <DashboardOutlined style={{ color: '#10b981' }} />
      case 'FaultInjection':
        return <SyncOutlined style={{ color: '#f59e0b' }} />
      case 'CollectResult':
        return <DatabaseOutlined style={{ color: '#8b5cf6' }} />
      case 'AlgorithmExecution':
        return <FunctionOutlined style={{ color: '#ec4899' }} />
      default:
        return <ClockCircleOutlined />
    }
  }

  const getTaskTypeColor = (type: TaskType) => {
    switch (type) {
      case 'SubmitInjection':
        return '#3b82f6'
      case 'BuildDatapack':
        return '#10b981'
      case 'FaultInjection':
        return '#f59e0b'
      case 'CollectResult':
        return '#8b5cf6'
      case 'AlgorithmExecution':
        return '#ec4899'
      default:
        return '#6b7280'
    }
  }

  const getStateColor = (state: TaskState) => {
    switch (state) {
      case 0: // PENDING
        return '#d1d5db'
      case 1: // RUNNING
        return '#3b82f6'
      case 2: // COMPLETED
        return '#10b981'
      case 3: // ERROR
        return '#ef4444'
      case 4: // CANCELLED
        return '#6b7280'
      default:
        return '#6b7280'
    }
  }

  const getStateIcon = (state: TaskState) => {
    switch (state) {
      case 0: // PENDING
        return <ClockCircleOutlined />
      case 1: // RUNNING
        return <SyncOutlined spin />
      case 2: // COMPLETED
        return <CheckCircleOutlined />
      case 3: // ERROR
        return <CloseCircleOutlined />
      case 4: // CANCELLED
        return <PauseCircleOutlined />
      default:
        return <ClockCircleOutlined />
    }
  }

  const formatDuration = (start?: string, end?: string) => {
    if (!start) return '-'
    const startTime = dayjs(start)
    const endTime = end ? dayjs(end) : dayjs()
    const duration = dayjs.duration(endTime.diff(startTime))

    if (duration.asHours() >= 1) {
      return `${Math.floor(duration.asHours())}h ${duration.minutes()}m ${duration.seconds()}s`
    } else if (duration.asMinutes() >= 1) {
      return `${duration.minutes()}m ${duration.seconds()}s`
    } else {
      return `${duration.seconds()}s`
    }
  }

  const getTaskProgress = (task: Task) => {
    if (task.state === 2) return 100 // COMPLETED
    if (task.state === 3 || task.state === 4) return 0 // ERROR or CANCELLED
    if (task.state === 1) return 50 // RUNNING
    return 0
  }

  if (isLoading) {
    return (
      <div style={{ padding: 24 }}>
        <Card loading>
          <div style={{ minHeight: 400 }} />
        </Card>
      </div>
    )
  }

  if (!task) {
    return (
      <div style={{ padding: 24, textAlign: 'center' }}>
        <Text type="secondary">Task not found</Text>
      </div>
    )
  }

  const taskData = task?.data
  const progress = getTaskProgress(taskData)

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
              task?.data?.state === 2 ? 'success' : // COMPLETED
              task?.data?.state === 3 ? 'error' : // ERROR
              task?.data?.state === 1 ? 'processing' : // RUNNING
              task?.data?.state === 4 ? 'warning' : // CANCELLED
              'default'
            }
            text={
              <Space>
                {getStateIcon(task?.data?.state || 0)}
                <Text strong style={{ color: getStateColor(task?.data?.state || 0) }}>
                  {task?.data?.state === 0 ? 'Pending' : // PENDING
                   task?.data?.state === 1 ? 'Running' : // RUNNING
                   task?.data?.state === 2 ? 'Completed' : // COMPLETED
                   task?.data?.state === 3 ? 'Error' : // ERROR
                   task?.data?.state === 4 ? 'Cancelled' : // CANCELLED
                   'Unknown'}
                </Text>
              </Space>
            }
          />
        </Space>
      </div>

      {/* Actions */}
      <Card style={{ marginBottom: 24 }}>
        <Row justify="space-between" align="middle">
          <Col>
            <Space>
              {(taskData?.state === 1 || taskData?.state === 0) && ( // RUNNING or PENDING
                <Button
                  danger
                  icon={<PauseCircleOutlined />}
                  onClick={handleCancelTask}
                >
                  Cancel Task
                </Button>
              )}
              {taskData?.state === 3 && ( // ERROR
                <Button
                  type="primary"
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
                  navigator.clipboard.writeText(taskId)
                  message.success('Task ID copied to clipboard')
                }}
              >
                Copy ID
              </Button>
            </Space>
          </Col>
          <Col>
            <Space>
              <Text type="secondary">Auto-refresh:</Text>
              <Switch
                checked={autoRefresh}
                onChange={setAutoRefresh}
                checkedChildren="ON"
                unCheckedChildren="OFF"
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
            taskData.state === TaskState.ERROR ? 'exception' :
            taskData.state === TaskState.COMPLETED ? 'success' :
            'active'
          }
          strokeColor={getStateColor(taskData.state)}
          format={percent => (
            <Space>
              {getStateIcon(taskData.state)}
              <Text>{percent}%</Text>
            </Space>
          )}
        />
      </Card>

      {/* Tabs */}
      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane tab="Overview" key="overview">
          <Row gutter={[16, 16]}>
            <Col xs={24} lg={16}>
              <Card title="Task Information">
                <Descriptions column={2} bordered>
                  <Descriptions.Item label="Task ID">
                    <Space>
                      <Text code>{taskId}</Text>
                      <Button
                        type="text"
                        size="small"
                        icon={<CopyOutlined />}
                        onClick={() => {
                          navigator.clipboard.writeText(taskId)
                          message.success('Task ID copied to clipboard')
                        }}
                      />
                    </Space>
                  </Descriptions.Item>
                  <Descriptions.Item label="Type">
                    <Tag
                      color={getTaskTypeColor(taskData.type)}
                      style={{ fontWeight: 500, fontSize: '1rem' }}
                    >
                      <Space>
                        {getTaskTypeIcon(taskData.type)}
                        {taskData.type}
                      </Space>
                    </Tag>
                  </Descriptions.Item>
                  <Descriptions.Item label="Status">
                    <StatusBadge
                      status={
                        taskData.state === TaskState.COMPLETED ? 'success' :
                        taskData.state === TaskState.ERROR ? 'error' :
                        taskData.state === TaskState.RUNNING ? 'processing' :
                        taskData.state === TaskState.CANCELLED ? 'warning' :
                        'default'
                      }
                      text={
                        taskData.state === TaskState.PENDING ? 'Pending' :
                        taskData.state === TaskState.RUNNING ? 'Running' :
                        taskData.state === TaskState.COMPLETED ? 'Completed' :
                        taskData.state === TaskState.ERROR ? 'Error' :
                        taskData.state === TaskState.CANCELLED ? 'Cancelled' :
                        'Unknown'
                      }
                    />
                  </Descriptions.Item>
                  <Descriptions.Item label="Retry Count">
                    <Text code>
                      {taskData.retry_count}/{taskData.max_retry}
                    </Text>
                  </Descriptions.Item>
                  <Descriptions.Item label="Immediate">
                    <Text>{taskData.immediate ? 'Yes' : 'No'}</Text>
                  </Descriptions.Item>
                  <Descriptions.Item label="Trace ID">
                    <Text code>{taskData.trace_id}</Text>
                  </Descriptions.Item>
                  <Descriptions.Item label="Group ID">
                    <Text code>{taskData.group_id}</Text>
                  </Descriptions.Item>
                  {taskData.parent_id && (
                    <Descriptions.Item label="Parent ID">
                      <Text code>{taskData.parent_id}</Text>
                    </Descriptions.Item>
                  )}
                  <Descriptions.Item label="Project ID">
                    <Text>{taskData.project_id || 'N/A'}</Text>
                  </Descriptions.Item>
                  <Descriptions.Item label="Status Code">
                    <Text code>{taskData.status}</Text>
                  </Descriptions.Item>
                </Descriptions>
              </Card>
            </Col>
            <Col xs={24} lg={8}>
              <Card title="Timing Information">
                <Space direction="vertical" style={{ width: '100%' }}>
                  <div>
                    <Text type="secondary">Created</Text>
                    <br />
                    <Text strong>
                      {dayjs(taskData.created_at).format('MMM D, YYYY HH:mm:ss')}
                    </Text>
                  </div>
                  <Divider />
                  {taskData.started_at && (
                    <>
                      <div>
                        <Text type="secondary">Started</Text>
                        <br />
                        <Text strong>
                          {dayjs(taskData.started_at).format('MMM D, YYYY HH:mm:ss')}
                        </Text>
                      </div>
                      <Divider />
                    </>
                  )}
                  {taskData.finished_at && (
                    <>
                      <div>
                        <Text type="secondary">Finished</Text>
                        <br />
                        <Text strong>
                          {dayjs(taskData.finished_at).format('MMM D, YYYY HH:mm:ss')}
                        </Text>
                      </div>
                      <Divider />
                    </>
                  )}
                  <div>
                    <Text type="secondary">Duration</Text>
                    <br />
                    <Title level={3} style={{ margin: 0, color: '#3b82f6' }}>
                      {formatDuration(taskData.started_at, taskData.finished_at)}
                    </Title>
                  </div>
                </Space>
              </Card>
            </Col>
          </Row>

          {taskData.labels && taskData.labels.length > 0 && (
            <Card title="Labels" style={{ marginTop: 16 }}>
              <Space wrap>
                {taskData.labels.map((label, index) => (
                  <Tag key={index} icon={<TagsOutlined />}>
                    {label.key}: {label.value}
                  </Tag>
                ))}
              </Space>
            </Card>
          )}

          {taskData.payload && (
            <Card title="Payload" style={{ marginTop: 16 }}>
              <pre style={{ margin: 0, fontSize: '0.875rem', whiteSpace: 'pre-wrap' }}>
                {JSON.stringify(taskData.payload, null, 2)}
              </pre>
            </Card>
          )}
        </TabPane>

        <TabPane tab="Logs" key="logs">
          <Card
            title="Task Logs"
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
              <div style={{ background: '#f5f5f5', padding: 16, borderRadius: 4, maxHeight: 400, overflow: 'auto' }}>
                <pre style={{ margin: 0, fontSize: '0.875rem', fontFamily: 'monospace' }}>
                  {logs.join('\n')}
                </pre>
              </div>
            ) : (
              <Empty
                description="No logs available. Logs will appear when the task starts running."
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              />
            )}
          </Card>
        </TabPane>

        <TabPane tab="Timeline" key="timeline">
          <Card title="Task Execution Timeline">
            <Timeline>
              <Timeline.Item
                color="blue"
                dot={<ClockCircleOutlined />}
              >
                <Text strong>Task Created</Text>
                <br />
                <Text type="secondary">
                  {dayjs(taskData.created_at).format('MMM D, YYYY HH:mm:ss')}
                </Text>
              </Timeline.Item>

              {taskData.started_at && (
                <Timeline.Item
                  color="green"
                  dot={<PlayCircleOutlined />}
                >
                  <Text strong>Task Started</Text>
                  <br />
                  <Text type="secondary">
                    {dayjs(taskData.started_at).format('MMM D, YYYY HH:mm:ss')}
                  </Text>
                </Timeline.Item>
              )}

              {taskData.state === TaskState.RUNNING && (
                <Timeline.Item
                  color="blue"
                  dot={<SyncOutlined spin />}
                >
                  <Text strong>Task Running</Text>
                  <br />
                  <Text type="secondary">In progress...</Text>
                </Timeline.Item>
              )}

              {taskData.state === TaskState.COMPLETED && taskData.finished_at && (
                <Timeline.Item
                  color="green"
                  dot={<CheckCircleOutlined />}
                >
                  <Text strong>Task Completed</Text>
                  <br />
                  <Text type="secondary">
                    {dayjs(taskData.finished_at).format('MMM D, YYYY HH:mm:ss')}
                  </Text>
                </Timeline.Item>
              )}

              {taskData.state === TaskState.ERROR && taskData.finished_at && (
                <Timeline.Item
                  color="red"
                  dot={<CloseCircleOutlined />}
                >
                  <Text strong>Task Failed</Text>
                  <br />
                  <Text type="secondary">
                    {dayjs(taskData.finished_at).format('MMM D, YYYY HH:mm:ss')}
                  </Text>
                </Timeline.Item>
              )}

              {taskData.state === TaskState.CANCELLED && taskData.finished_at && (
                <Timeline.Item
                  color="orange"
                  dot={<PauseCircleOutlined />}
                >
                  <Text strong>Task Cancelled</Text>
                  <br />
                  <Text type="secondary">
                    {dayjs(taskData.finished_at).format('MMM D, YYYY HH:mm:ss')}
                  </Text>
                </Timeline.Item>
              )}
            </Timeline>
          </Card>
        </TabPane>
      </Tabs>
    </div>
  )
}

export default TaskDetail