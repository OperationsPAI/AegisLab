
import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  DeleteOutlined,
  EditOutlined,
  ExperimentOutlined,
  PlusOutlined,
  SearchOutlined,
  SyncOutlined,
} from '@ant-design/icons';
import type { InjectionDetailResp as Injection } from '@rcabench/client';
import { useQuery } from '@tanstack/react-query';
import {
  Avatar,
  Button,
  Card,
  Col,
  Input,
  Progress,
  Row,
  Space,
  Table,
  type TablePaginationConfig,
  Tag,
  Tooltip,
  Typography,
} from 'antd';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import type { EChartsOption } from 'echarts';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { injectionApi } from '@/api/injections';
import LabChart from '@/components/charts/LabChart';
import StatCard from '@/components/ui/StatCard';
import StatusBadge, {
  type StatusBadgeProps,
} from '@/components/ui/StatusBadge';
import { useSSE } from '@/hooks/useSSE';
import { InjectionState, InjectionType } from '@/types/api';

import './InjectionList.css';

const { Title, Text } = Typography;
const { Search } = Input;

dayjs.extend(relativeTime);

const InjectionList = () => {
  const navigate = useNavigate();
  const [searchText, setSearchText] = useState('');
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([]);
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  // Fetch injections
  const { data: injectionsData, isLoading, refetch } = useQuery({
    queryKey: [
      'injections',
      pagination.current,
      pagination.pageSize,
      searchText,
    ],
    queryFn: () =>
      injectionApi.getInjections({
        page: pagination.current,
        size: pagination.pageSize,
      }),
  });

  // Real-time updates via SSE
  useSSE({
    url: '/api/v2/notifications/stream',
    enabled: true,
    onMessage: (data) => {
      // Refetch injections when relevant events are received
      if (data.type === 'injection_completed' || data.type === 'datapack_ready') {
        refetch();
      }
    },
  });

  // Calculate success rate from real data
  const totalInjections = injectionsData?.items?.length || 0;
  const successfulInjections =
    injectionsData?.items?.filter(
      (i) => i.state === 'build_success' || i.state === 'inject_success'
    ).length || 0;
  const calculatedSuccessRate =
    totalInjections > 0
      ? Math.round((successfulInjections / totalInjections) * 100)
      : 0;

  // Calculate statistics from real data
  const stats = {
    total: injectionsData?.pagination?.total || 0,
    running:
      injectionsData?.items?.filter((i) => i.state === 'initial').length || 0,
    successRate: calculatedSuccessRate,
    avgDuration: 0, // No duration data available in API response
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

  const handleCreateInjection = () => {
    navigate('/injections/create');
  };

  const handleEditInjection = (id: number) => {
    navigate(`/injections/${id}`);
  };

  const getInjectionTypeColor = (type: InjectionType) => {
    const colors = {
      [InjectionType.NETWORK]: 'blue',
      [InjectionType.CPU]: 'orange',
      [InjectionType.MEMORY]: 'purple',
      [InjectionType.DISK]: 'green',
      [InjectionType.PROCESS]: 'red',
      [InjectionType.KUBERNETES]: 'cyan',
    };
    return colors[type] || 'default';
  };

  const getInjectionTypeIcon = (type: InjectionType) => {
    const icons = {
      [InjectionType.NETWORK]: '🌐',
      [InjectionType.CPU]: '💻',
      [InjectionType.MEMORY]: '🧠',
      [InjectionType.DISK]: '💾',
      [InjectionType.PROCESS]: '⚙️',
      [InjectionType.KUBERNETES]: '☸️',
    };
    return icons[type] || '🔧';
  };

  // Injection timeline chart - based on real data grouped by fault type
  const faultTypeCounts = injectionsData?.items?.reduce(
    (acc, item) => {
      const type = item.fault_type || 'Other';
      acc[type] = (acc[type] || 0) + 1;
      return acc;
    },
    {} as Record<string, number>
  ) || {};

  const timelineData: EChartsOption = {
    title: {
      text: 'Injections by Fault Type',
      left: 'center',
      top: 0,
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
      orient: 'vertical',
      right: 10,
      top: 'center',
      type: 'scroll',
    },
    series: [
      {
        name: 'Fault Types',
        type: 'pie',
        radius: ['40%', '70%'],
        center: ['40%', '55%'],
        avoidLabelOverlap: true,
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
        data: Object.entries(faultTypeCounts).map(([name, value]) => ({
          name,
          value,
        })),
      },
    ],
  };

  // Success rate chart - based on real data
  const successRateData: EChartsOption = {
    title: {
      text: 'Success Rate',
      left: 'center',
      textStyle: {
        fontSize: 16,
        fontWeight: 600,
      },
    },
    tooltip: {
      trigger: 'item',
      formatter: '{b}: {c}%',
    },
    series: [
      {
        name: 'Success Rate',
        type: 'gauge',
        startAngle: 180,
        endAngle: 0,
        min: 0,
        max: 100,
        splitNumber: 5,
        itemStyle: {
          color: '#10b981',
        },
        progress: {
          show: true,
          width: 30,
        },
        pointer: {
          show: false,
        },
        axisLine: {
          lineStyle: {
            width: 30,
            color: [
              [0.3, '#ef4444'],
              [0.7, '#f59e0b'],
              [1, '#10b981'],
            ],
          },
        },
        axisTick: {
          show: false,
        },
        splitLine: {
          show: false,
        },
        axisLabel: {
          show: false,
        },
        anchor: {
          show: false,
        },
        title: {
          offsetCenter: [0, '30%'],
          fontSize: 14,
        },
        detail: {
          valueAnimation: true,
          width: '60%',
          lineHeight: 40,
          borderRadius: 8,
          offsetCenter: [0, '-10%'],
          fontSize: 30,
          fontWeight: 'bolder',
          formatter: '{value}%',
          color: 'inherit',
        },
        data: [
          {
            value: calculatedSuccessRate,
            name: 'Overall',
          },
        ],
      },
    ],
  };

  const columns = [
    {
      title: 'Injection',
      dataIndex: 'name',
      key: 'name',
      width: '25%',
      render: (name: string, record: Injection) => (
        <Space>
          <Avatar
            size='large'
            style={{
              backgroundColor: getInjectionTypeColor(InjectionType.NETWORK),
              fontSize: '1.25rem',
            }}
          >
            {getInjectionTypeIcon(InjectionType.NETWORK)}
          </Avatar>
          <div>
            <Text strong style={{ fontSize: '1rem' }}>
              {name}
            </Text>
            <br />
            <Tag color={getInjectionTypeColor(InjectionType.NETWORK)}>
              {record.fault_type || 'Unknown'}
            </Tag>
          </div>
        </Space>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'state',
      key: 'state',
      width: '12%',
      render: (state: InjectionState) => {
        const statusMap = {
          [InjectionState.PENDING]: { text: 'Pending', color: 'warning' },
          [InjectionState.RUNNING]: { text: 'Running', color: 'info' },
          [InjectionState.COMPLETED]: { text: 'Completed', color: 'success' },
          [InjectionState.ERROR]: { text: 'Error', color: 'error' },
          [InjectionState.STOPPED]: { text: 'Stopped', color: 'default' },
        };
        const config = statusMap[state] || {
          text: 'Unknown',
          color: 'default',
        };
        return (
          <StatusBadge
            status={config.color as StatusBadgeProps['status']}
            text={config.text}
          />
        );
      },
    },
    {
      title: 'Progress',
      dataIndex: 'progress',
      key: 'progress',
      width: '15%',
      render: (progress: number, record: Injection) => (
        <div>
          <Progress
            percent={progress || 0}
            size='small'
            status={
              record.state === '3' ? 'exception' : 'active' // ERROR = 3
            }
            strokeColor={
              record.state === '2' ? '#10b981' : undefined // COMPLETED = 2
            }
          />
          <Text type='secondary' style={{ fontSize: '0.75rem' }}>
            {progress || 0}% Complete
          </Text>
        </div>
      ),
    },
    {
      title: 'Duration',
      dataIndex: 'duration',
      key: 'duration',
      width: '12%',
      render: (duration: number) => (
        <Text>
          <ClockCircleOutlined /> {duration ? `${duration}s` : '-'}
        </Text>
      ),
    },
    {
      title: 'Target',
      dataIndex: 'target',
      key: 'target',
      width: '15%',
      render: (target: string) => (
        <Tooltip title={target}>
          <Text ellipsis style={{ maxWidth: 150 }}>
            {target || 'All Services'}
          </Text>
        </Tooltip>
      ),
    },
    {
      title: 'Started',
      dataIndex: 'created_at',
      key: 'created_at',
      width: '12%',
      render: (date: string) => (
        <Text type='secondary'>{dayjs(date).fromNow()}</Text>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: '12%',
      render: (_: string, record: Injection) => (
        <Space>
          <Button
            type='text'
            icon={<EditOutlined />}
            onClick={() => handleEditInjection(record.id || 0)}
            title='View Injection'
          />
          <Button
            type='text'
            danger
            icon={<DeleteOutlined />}
            title='Delete Injection'
          />
        </Space>
      ),
    },
  ];

  const rowSelection = {
    selectedRowKeys,
    onChange: (newSelectedRowKeys: React.Key[]) => {
      setSelectedRowKeys(newSelectedRowKeys);
    },
  };

  return (
    <div className='injection-list'>
      {/* Page Header */}
      <div className='page-header'>
        <div className='page-header-left'>
          <Title level={2} className='page-title'>
            Fault Injections
          </Title>
          <Text type='secondary'>
            Manage chaos engineering experiments for your microservices
          </Text>
        </div>
        <Button
          type='primary'
          size='large'
          icon={<PlusOutlined />}
          onClick={handleCreateInjection}
          className='create-button'
        >
          New Injection
        </Button>
      </div>

      {/* Statistics Cards */}
      <Row gutter={[{ xs: 8, sm: 16, lg: 24 }, { xs: 8, sm: 16, lg: 24 }]} className='stats-row'>
        <Col xs={12} sm={12} lg={6}>
          <StatCard
            title='Total Injections'
            value={stats?.total || 0}
            prefix={<ExperimentOutlined />}
            color='primary'
          />
        </Col>
        <Col xs={12} sm={12} lg={6}>
          <StatCard
            title='Running Now'
            value={stats?.running || 0}
            prefix={<SyncOutlined spin={stats?.running > 0} />}
            color='info'
          />
        </Col>
        <Col xs={12} sm={12} lg={6}>
          <StatCard
            title='Success Rate'
            value={`${stats?.successRate || 0}%`}
            prefix={<CheckCircleOutlined />}
            color='success'
          />
        </Col>
        <Col xs={12} sm={12} lg={6}>
          <StatCard
            title='Avg Duration'
            value={`${stats?.avgDuration || 0}s`}
            prefix={<ClockCircleOutlined />}
            color='warning'
          />
        </Col>
      </Row>

      {/* Charts */}
      <Row gutter={[{ xs: 8, sm: 16, lg: 24 }, { xs: 8, sm: 16, lg: 24 }]} className='charts-row'>
        <Col xs={24} lg={16}>
          <Card className='chart-card'>
            <LabChart option={timelineData} style={{ height: '300px' }} />
          </Card>
        </Col>
        <Col xs={24} lg={8}>
          <Card className='chart-card'>
            <LabChart option={successRateData} style={{ height: '300px' }} />
          </Card>
        </Col>
      </Row>

      {/* Search and Bulk Actions */}
      <Card className='search-card'>
        <Row gutter={[24, 24]} align='middle'>
          <Col flex='auto'>
            <Search
              placeholder='Search injections by name, type, or target...'
              allowClear
              enterButton={<SearchOutlined />}
              size='large'
              onSearch={handleSearch}
              style={{ maxWidth: 400 }}
            />
          </Col>
          <Col>
            <Space>
              {selectedRowKeys.length > 0 && (
                <Button size='large' danger>
                  Delete Selected ({selectedRowKeys.length})
                </Button>
              )}
              <Button size='large'>Filter by Type</Button>
              <Button size='large'>Export</Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* Injections Table */}
      <Card className='table-card'>
        <Table
          rowSelection={rowSelection}
          columns={columns}
          dataSource={injectionsData?.items || []}
          loading={isLoading}
          pagination={{
            ...pagination,
            total: injectionsData?.pagination?.total || 0,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `Total ${total} injections`,
          }}
          onChange={handleTableChange}
          rowKey='id'
          className='injections-table'
          rowClassName='injection-row'
        />

        {/* Mobile Card Layout */}
        <div className='injection-card-mobile'>
          <Space direction='vertical' style={{ width: '100%' }} size='middle'>
            {injectionsData?.items?.map((injection: Injection) => (
              <Card
                key={injection.id}
                size='small'
                style={{ cursor: 'pointer' }}
                onClick={() => handleEditInjection(injection.id || 0)}
              >
                <Space direction='vertical' style={{ width: '100%' }} size='small'>
                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <Text strong style={{ fontSize: '1rem' }}>{injection.name}</Text>
                    <Tag color={getInjectionTypeColor(InjectionType.NETWORK)}>
                      {injection.fault_type || 'Unknown'}
                    </Tag>
                  </div>

                  <div style={{ display: 'flex', justifyContent: 'space-between', alignItems: 'center' }}>
                    <StatusBadge
                      status={
                        injection.state === InjectionState.COMPLETED ? 'success' :
                        injection.state === InjectionState.RUNNING ? 'info' :
                        injection.state === InjectionState.ERROR ? 'error' : 'warning'
                      }
                      text={
                        injection.state === InjectionState.COMPLETED ? 'Completed' :
                        injection.state === InjectionState.RUNNING ? 'Running' :
                        injection.state === InjectionState.ERROR ? 'Error' : 'Pending'
                      }
                    />
                    <Text type='secondary' style={{ fontSize: '0.875rem' }}>
                      {dayjs(injection.created_at).fromNow()}
                    </Text>
                  </div>

                  <Progress
                    percent={injection.progress || 0}
                    size='small'
                    status={injection.state === InjectionState.ERROR ? 'exception' : 'active'}
                    strokeColor={injection.state === InjectionState.COMPLETED ? '#10b981' : undefined}
                  />
                </Space>
              </Card>
            ))}
          </Space>
        </div>
      </Card>
    </div>
  );
};

export default InjectionList;
