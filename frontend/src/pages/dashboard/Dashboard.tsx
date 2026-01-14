import {
  ExperimentOutlined,
  PlayCircleOutlined,
  ProjectOutlined,
  SyncOutlined,
} from '@ant-design/icons';
import type {
  ExecutionResp,
  InjectionResp,
  TaskResp,
} from '@rcabench/client';
import { useQuery } from '@tanstack/react-query';
import { Col, Row, Typography } from 'antd';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import type { EChartsOption } from 'echarts';

import { executionApi } from '@/api/executions';
import { injectionApi } from '@/api/injections';
import { projectApi } from '@/api/projects';
import { taskApi } from '@/api/tasks';
import LabChart from '@/components/charts/LabChart';
import StatCard from '@/components/ui/StatCard';

import './Dashboard.css';

dayjs.extend(relativeTime);

const { Title, Text } = Typography;

const Dashboard = () => {
  // Fetch data
  const { data: projects } = useQuery({
    queryKey: ['projects', { page: 1, size: 10 }],
    queryFn: () => projectApi.getProjects({ page: 1, size: 10 }),
  });

  const { data: injections } = useQuery({
    queryKey: ['injections', { page: 1, size: 10 }],
    queryFn: () => injectionApi.getInjections({ page: 1, size: 10 }),
  });

  const { data: executions } = useQuery({
    queryKey: ['executions', { page: 1, size: 10 }],
    queryFn: () => executionApi.getExecutions({ page: 1, size: 10 }),
  });

  const { data: tasks } = useQuery({
    queryKey: ['tasks', { page: 1, size: 50 }],
    queryFn: () => taskApi.getTasks({ page: 1, size: 50 }),
  });

  // Calculate statistics
  const stats = {
    totalProjects: projects?.pagination?.total || 0,
    activeInjections:
      injections?.items?.filter((i: InjectionResp) => i.state === '1').length || 0, // RUNNING
    pendingTasks:
      tasks?.items?.filter((t: TaskResp) => t.state === '0').length || 0, // PENDING
    runningTasks:
      tasks?.items?.filter((t: TaskResp) => t.state === '2').length || 0, // RUNNING
    completedTasks:
      tasks?.items?.filter((t: TaskResp) => t.state === '3').length || 0, // COMPLETED
    errorTasks: tasks?.items?.filter((t: TaskResp) => t.state === '-1').length || 0, // ERROR
    todayExecutions:
      executions?.items?.filter((e: ExecutionResp) =>
        dayjs(e.created_at).isAfter(dayjs().startOf('day'))
      ).length || 0,
  };

  // Task distribution chart
  const taskDistributionOption: EChartsOption = {
    title: {
      text: 'Task Distribution',
      left: 'center',
      textStyle: {
        fontSize: 16,
        fontWeight: 600,
      },
    },
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c} ({d}%)',
    },
    legend: {
      bottom: '5%',
      left: 'center',
    },
    series: [
      {
        name: 'Tasks',
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
          position: 'center',
        },
        emphasis: {
          label: {
            show: true,
            fontSize: 20,
            fontWeight: 'bold',
          },
          itemStyle: {
            shadowBlur: 10,
            shadowOffsetX: 0,
            shadowColor: 'rgba(0, 0, 0, 0.5)',
          },
        },
        labelLine: {
          show: false,
        },
        data: [
          {
            value: stats.pendingTasks,
            name: 'Pending',
            itemStyle: { color: '#f59e0b' },
          },
          {
            value: stats.runningTasks,
            name: 'Running',
            itemStyle: { color: '#3b82f6' },
          },
          {
            value: stats.completedTasks,
            name: 'Completed',
            itemStyle: { color: '#10b981' },
          },
          {
            value: stats.errorTasks,
            name: 'Error',
            itemStyle: { color: '#ef4444' },
          },
        ],
      },
    ],
  };

  // System health chart
  const systemHealthOption: EChartsOption = {
    title: {
      text: 'System Health (24h)',
      left: 'center',
      textStyle: {
        fontSize: 16,
        fontWeight: 600,
      },
    },
    tooltip: {
      trigger: 'axis',
      axisPointer: {
        type: 'cross',
      },
    },
    xAxis: {
      type: 'category',
      boundaryGap: false,
      data: Array.from({ length: 24 }, (_, i) => {
        const hour = dayjs()
          .subtract(23 - i, 'hour')
          .hour();
        return `${hour}:00`;
      }),
    },
    yAxis: {
      type: 'value',
      name: 'CPU %',
      min: 0,
      max: 100,
    },
    series: [
      {
        name: 'CPU Usage',
        type: 'line',
        smooth: true,
        symbol: 'none',
        areaStyle: {
          opacity: 0.3,
        },
        lineStyle: {
          width: 3,
        },
        data: Array.from(
          { length: 24 },
          () => Math.floor(Math.random() * 40) + 30
        ),
      },
      {
        name: 'Memory Usage',
        type: 'line',
        smooth: true,
        symbol: 'none',
        areaStyle: {
          opacity: 0.3,
        },
        lineStyle: {
          width: 3,
        },
        data: Array.from(
          { length: 24 },
          () => Math.floor(Math.random() * 30) + 50
        ),
      },
    ],
  };

  return (
    <div className='dashboard'>
      <div className='dashboard-header'>
        <Title level={2} className='dashboard-title'>
          Dashboard
        </Title>
        <Text className='dashboard-subtitle'>
          Welcome back! Here&apos;s what&apos;s happening with your RCA
          experiments.
        </Text>
      </div>

      {/* Key Metrics */}
      <Row gutter={[24, 24]} className='metrics-row'>
        <Col xs={24} sm={12} lg={6}>
          <StatCard
            title='Total Projects'
            value={stats.totalProjects}
            prefix={<ProjectOutlined />}
            color='primary'
            trend='up'
            trendValue='+12%'
          />
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <StatCard
            title='Active Injections'
            value={stats.activeInjections}
            prefix={<ExperimentOutlined />}
            color='warning'
            trend='neutral'
            trendValue='Running'
          />
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <StatCard
            title='Running Tasks'
            value={stats.runningTasks}
            prefix={<SyncOutlined spin={stats.runningTasks > 0} />}
            color='info'
            trend='up'
            trendValue='Active'
          />
        </Col>
        <Col xs={24} sm={12} lg={6}>
          <StatCard
            title="Today's Executions"
            value={stats.todayExecutions}
            prefix={<PlayCircleOutlined />}
            color='success'
            trend='up'
            trendValue='+5'
          />
        </Col>
      </Row>

      {/* Charts and Visualizations */}
      <Row gutter={[24, 24]} className='charts-row'>
        <Col xs={24} lg={12}>
          <div className='chart-container'>
            <LabChart
              option={taskDistributionOption}
              style={{ height: '350px' }}
            />
          </div>
        </Col>
        <Col xs={24} lg={12}>
          <div className='chart-container'>
            <LabChart option={systemHealthOption} style={{ height: '350px' }} />
          </div>
        </Col>
      </Row>

      {/* System Metrics and Activity */}
      <Row gutter={[24, 24]} className='bottom-row'>
        <Col xs={24} lg={12}>
          <div className='chart-container'>
            <LabChart option={systemHealthOption} style={{ height: '350px' }} />
          </div>
        </Col>
        <Col xs={24} lg={12}>
          <div className='chart-container'>
            <LabChart
              option={taskDistributionOption}
              style={{ height: '350px' }}
            />
          </div>
        </Col>
      </Row>
    </div>
  );
};

export default Dashboard;
