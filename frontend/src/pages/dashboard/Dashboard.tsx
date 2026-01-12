import { Row, Col, Card, Statistic, Table, Tag, Typography, Empty } from 'antd'
import { useQuery } from '@tanstack/react-query'
import {
  ProjectOutlined,
  ExperimentOutlined,
  PlayCircleOutlined,
  ClockCircleOutlined,
} from '@ant-design/icons'
import ReactECharts from 'echarts-for-react'
import type { EChartsOption } from 'echarts'
import { projectApi } from '@/api/projects'
import { injectionApi } from '@/api/injections'
import { executionApi } from '@/api/executions'
import { taskApi } from '@/api/tasks'
import { TaskState, InjectionState, ExecutionState } from '@/types/api'
import dayjs from 'dayjs'
import relativeTime from 'dayjs/plugin/relativeTime'

dayjs.extend(relativeTime)

const { Title, Text } = Typography

const Dashboard = () => {
  // Fetch data
  const { data: projects } = useQuery({
    queryKey: ['projects', { page: 1, size: 5 }],
    queryFn: () => projectApi.getProjects({ page: 1, size: 5 }),
  })

  const { data: injections } = useQuery({
    queryKey: ['injections', { page: 1, size: 5 }],
    queryFn: () => injectionApi.getInjections({ page: 1, size: 5 }),
  })

  const { data: executions } = useQuery({
    queryKey: ['executions', { page: 1, size: 5 }],
    queryFn: () => executionApi.getExecutions({ page: 1, size: 5 }),
  })

  const { data: tasks } = useQuery({
    queryKey: ['tasks', { page: 1, size: 100 }],
    queryFn: () => taskApi.getTasks({ page: 1, size: 100 }),
  })

  // Calculate stats
  const projectCount = projects?.data.total || 0
  const activeInjections =
    injections?.data.data.filter((i) => i.state < InjectionState.COMPLETED)
      .length || 0
  const pendingTasks =
    tasks?.data.data.filter((t) => t.state === TaskState.PENDING).length || 0
  const todayExecutions =
    executions?.data.data.filter(
      (e) => dayjs(e.created_at).isAfter(dayjs().startOf('day'))
    ).length || 0

  // Task state distribution chart
  const taskStateData = [
    {
      value:
        tasks?.data.data.filter((t) => t.state === TaskState.PENDING).length ||
        0,
      name: '待处理',
    },
    {
      value:
        tasks?.data.data.filter((t) => t.state === TaskState.RUNNING).length ||
        0,
      name: '运行中',
    },
    {
      value:
        tasks?.data.data.filter((t) => t.state === TaskState.COMPLETED).length ||
        0,
      name: '已完成',
    },
    {
      value:
        tasks?.data.data.filter((t) => t.state === TaskState.ERROR).length || 0,
      name: '错误',
    },
  ]

  const taskChartOption: EChartsOption = {
    title: {
      text: '任务状态分布',
      left: 'center',
      textStyle: {
        fontSize: 14,
        fontWeight: 600,
        color: '#374151',
      },
    },
    tooltip: {
      trigger: 'item',
    },
    legend: {
      bottom: '5%',
      left: 'center',
    },
    series: [
      {
        name: '任务数量',
        type: 'pie',
        radius: ['40%', '70%'],
        avoidLabelOverlap: false,
        itemStyle: {
          borderRadius: 10,
          borderColor: '#fff',
          borderWidth: 2,
        },
        label: {
          show: false,
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 14,
            fontWeight: 'bold',
          },
        },
        data: taskStateData,
        color: ['#9ca3af', '#3b82f6', '#10b981', '#ef4444'],
      },
    ],
  }

  // Recent activity columns
  const activityColumns = [
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      width: 100,
      render: (type: string) => {
        const typeMap: Record<string, { color: string; text: string }> = {
          project: { color: 'blue', text: '项目' },
          injection: { color: 'purple', text: '注入' },
          execution: { color: 'green', text: '执行' },
        }
        const config = typeMap[type] || { color: 'default', text: type }
        return <Tag color={config.color}>{config.text}</Tag>
      },
    },
    {
      title: '名称',
      dataIndex: 'name',
      key: 'name',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => {
        const statusMap: Record<string, string> = {
          completed: 'success',
          running: 'processing',
          error: 'error',
          pending: 'default',
        }
        return <Tag color={statusMap[status]}>{status}</Tag>
      },
    },
    {
      title: '时间',
      dataIndex: 'time',
      key: 'time',
      width: 150,
      render: (time: string) => (
        <Text type="secondary">{dayjs(time).fromNow()}</Text>
      ),
    },
  ]

  // Combine recent activities
  const recentActivities = [
    ...(projects?.data.data.slice(0, 3).map((p) => ({
      key: `project-${p.id}`,
      type: 'project',
      name: p.name,
      status: 'completed',
      time: p.created_at,
    })) || []),
    ...(injections?.data.data.slice(0, 3).map((i) => ({
      key: `injection-${i.id}`,
      type: 'injection',
      name: i.name,
      status:
        i.state === InjectionState.COMPLETED
          ? 'completed'
          : i.state === InjectionState.ERROR
          ? 'error'
          : 'running',
      time: i.created_at,
    })) || []),
  ].sort((a, b) => dayjs(b.time).unix() - dayjs(a.time).unix())

  return (
    <div>
      <Title level={3} style={{ marginBottom: '24px' }}>
        仪表盘
      </Title>

      {/* Key Metrics */}
      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="项目总数"
              value={projectCount}
              prefix={<ProjectOutlined />}
              valueStyle={{ color: '#2563eb' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="活跃实验"
              value={activeInjections}
              prefix={<ExperimentOutlined />}
              valueStyle={{ color: '#8b5cf6' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="待处理任务"
              value={pendingTasks}
              prefix={<ClockCircleOutlined />}
              valueStyle={{ color: '#f59e0b' }}
            />
          </Card>
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <Card>
            <Statistic
              title="今日执行"
              value={todayExecutions}
              prefix={<PlayCircleOutlined />}
              valueStyle={{ color: '#10b981' }}
            />
          </Card>
        </Col>
      </Row>

      {/* Charts and Tables */}
      <Row gutter={[16, 16]}>
        <Col xs={24} lg={10}>
          <Card>
            {taskStateData.every((d) => d.value === 0) ? (
              <Empty description="暂无任务数据" />
            ) : (
              <ReactECharts option={taskChartOption} style={{ height: '300px' }} />
            )}
          </Card>
        </Col>

        <Col xs={24} lg={14}>
          <Card title="最近活动" style={{ height: '100%' }}>
            <Table
              columns={activityColumns}
              dataSource={recentActivities}
              pagination={false}
              size="small"
            />
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default Dashboard
