import {
  PlayCircleOutlined,
  SearchOutlined,
  EyeOutlined,
  DeleteOutlined,
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
  SyncOutlined,
  FunctionOutlined,
  DatabaseOutlined,
  FilterOutlined,
  ExportOutlined,
} from '@ant-design/icons'
import { useQuery } from '@tanstack/react-query'
import {
  Table,
  Button,
  Space,
  Input,
  Typography,
  Row,
  Col,
  Card,
  Avatar,
  Tag,
  Select,
  Tooltip,
  Modal,
  message,
  Badge,
  Progress,
  type TablePaginationConfig,
} from 'antd'
import dayjs from 'dayjs'
import duration from 'dayjs/plugin/duration'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { containerApi } from '@/api/containers'
import { executionApi } from '@/api/executions'
import StatCard from '@/components/ui/StatCard'
import type { Execution } from '@/types/api'
import { ExecutionState } from '@/types/api'

dayjs.extend(duration)

const { Title, Text } = Typography
const { Search } = Input
const { Option } = Select

const ExecutionList = () => {
  const navigate = useNavigate()
  const [searchText, setSearchText] = useState('')
  const [selectedRowKeys, setSelectedRowKeys] = useState<React.Key[]>([])
  const [stateFilter, setStateFilter] = useState<ExecutionState | undefined>()
  const [algorithmFilter, setAlgorithmFilter] = useState<string | undefined>()
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  })

  // Fetch executions
  const { data: executionsData, isLoading, refetch } = useQuery({
    queryKey: ['executions', pagination.current, pagination.pageSize, searchText, stateFilter, algorithmFilter],
    queryFn: () =>
      executionApi.getExecutions({
        page: pagination.current,
        size: pagination.pageSize,
        state: stateFilter,
      }),
  })

  // Fetch algorithms for filter
  const { data: algorithmsData } = useQuery({
    queryKey: ['algorithms'],
    queryFn: () => containerApi.getContainers({ type: 'Algorithm' }),
  })

  // Statistics
  const stats = {
    total: executionsData?.total || 0,
    running: executionsData?.data.filter(e => e.state === ExecutionState.RUNNING).length || 0,
    completed: executionsData?.data.filter(e => e.state === ExecutionState.COMPLETED).length || 0,
    failed: executionsData?.data.filter(e => e.state === ExecutionState.ERROR).length || 0,
  }

  const handleTableChange = (newPagination: TablePaginationConfig) => {
    setPagination({
      ...pagination,
      current: newPagination.current || 1,
      pageSize: newPagination.pageSize || 10,
    })
  }

  const handleSearch = (value: string) => {
    setSearchText(value)
    setPagination({ ...pagination, current: 1 })
  }

  const handleStateFilter = (state: ExecutionState | undefined) => {
    setStateFilter(state)
    setPagination({ ...pagination, current: 1 })
  }

  const handleAlgorithmFilter = (algorithmId: string | undefined) => {
    setAlgorithmFilter(algorithmId)
    setPagination({ ...pagination, current: 1 })
  }

  const handleViewExecution = (id: number) => {
    navigate(`/executions/${id}`)
  }

  const handleDeleteExecution = (id: number) => {
    Modal.confirm({
      title: 'Delete Execution',
      content: 'Are you sure you want to delete this execution? This action cannot be undone.',
      okText: 'Yes, delete it',
      okButtonProps: { danger: true },
      cancelText: 'Cancel',
      onOk: async () => {
        try {
          await executionApi.batchDelete([id])
          message.success('Execution deleted successfully')
          refetch()
        } catch (error) {
          message.error('Failed to delete execution')
        }
      },
    })
  }

  const handleBatchDelete = () => {
    if (selectedRowKeys.length === 0) {
      message.warning('Please select executions to delete')
      return
    }

    Modal.confirm({
      title: 'Batch Delete Executions',
      content: `Are you sure you want to delete ${selectedRowKeys.length} executions?`,
      okText: 'Yes, delete them',
      okButtonProps: { danger: true },
      cancelText: 'Cancel',
      onOk: async () => {
        try {
          await executionApi.batchDelete(selectedRowKeys as number[])
          message.success(`${selectedRowKeys.length} executions deleted successfully`)
          setSelectedRowKeys([])
          refetch()
        } catch (error) {
          message.error('Failed to delete executions')
        }
      },
    })
  }

  const handleCreateExecution = () => {
    navigate('/executions/new')
  }

  const formatDuration = (seconds?: number) => {
    if (!seconds) return '-'
    const d = dayjs.duration(seconds, 'seconds')
    if (d.asHours() >= 1) {
      return `${Math.floor(d.asHours())}h ${d.minutes()}m ${d.seconds()}s`
    } else if (d.asMinutes() >= 1) {
      return `${d.minutes()}m ${d.seconds()}s`
    } else {
      return `${d.seconds()}s`
    }
  }

  const getStateColor = (state: ExecutionState) => {
    switch (state) {
      case ExecutionState.PENDING:
        return '#d1d5db'
      case ExecutionState.RUNNING:
        return '#3b82f6'
      case ExecutionState.COMPLETED:
        return '#10b981'
      case ExecutionState.ERROR:
        return '#ef4444'
      default:
        return '#6b7280'
    }
  }

  const getStateIcon = (state: ExecutionState) => {
    switch (state) {
      case ExecutionState.PENDING:
        return <ClockCircleOutlined />
      case ExecutionState.RUNNING:
        return <SyncOutlined spin />
      case ExecutionState.COMPLETED:
        return <CheckCircleOutlined />
      case ExecutionState.ERROR:
        return <CloseCircleOutlined />
      default:
        return <ClockCircleOutlined />
    }
  }

  const rowSelection = {
    selectedRowKeys,
    onChange: setSelectedRowKeys,
  }

  const columns = [
    {
      title: 'Execution',
      dataIndex: 'id',
      key: 'id',
      width: '12%',
      render: (id: number, record: Execution) => (
        <Space direction="vertical" size={0}>
          <Text strong>#{id}</Text>
          <Text type="secondary" style={{ fontSize: '0.75rem' }}>
            {record.algorithm?.name || 'Unknown Algorithm'}
          </Text>
        </Space>
      ),
    },
    {
      title: 'Algorithm',
      dataIndex: ['algorithm', 'name'],
      key: 'algorithm',
      width: '20%',
      render: (_: string, record: Execution) => (
        <Space>
          <Avatar
            size="small"
            style={{ backgroundColor: '#f59e0b' }}
            icon={<FunctionOutlined />}
          />
          <div>
            <Text strong>{record.algorithm?.name || 'Unknown'}</Text>
            <br />
            <Text type="secondary" style={{ fontSize: '0.75rem' }}>
              v{record.algorithm_version}
            </Text>
          </div>
        </Space>
      ),
    },
    {
      title: 'Datapack',
      dataIndex: ['datapack', 'id'],
      key: 'datapack',
      width: '15%',
      render: (datapackId: string) => (
        <Space>
          <DatabaseOutlined style={{ color: '#3b82f6' }} />
          <Text code>{datapackId?.substring(0, 8) || 'N/A'}</Text>
        </Space>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'state',
      key: 'state',
      width: '12%',
      render: (state: ExecutionState) => (
        <Badge
          status={
            state === ExecutionState.COMPLETED ? 'success' :
            state === ExecutionState.ERROR ? 'error' :
            state === ExecutionState.RUNNING ? 'processing' :
            'default'
          }
          text={
            <Space>
              {getStateIcon(state)}
              <Text strong style={{ color: getStateColor(state) }}>
                {state === ExecutionState.PENDING ? 'Pending' :
                 state === ExecutionState.RUNNING ? 'Running' :
                 state === ExecutionState.COMPLETED ? 'Completed' :
                 state === ExecutionState.ERROR ? 'Error' :
                 'Unknown'}
              </Text>
            </Space>
          }
        />
      ),
      filters: [
        { text: 'Pending', value: ExecutionState.PENDING },
        { text: 'Running', value: ExecutionState.RUNNING },
        { text: 'Completed', value: ExecutionState.COMPLETED },
        { text: 'Error', value: ExecutionState.ERROR },
      ],
      onFilter: (value: number, record: Execution) => record.state === value,
    },
    {
      title: 'Duration',
      dataIndex: 'execution_duration',
      key: 'duration',
      width: '10%',
      render: (duration: number) => (
        <Text code>{formatDuration(duration)}</Text>
      ),
    },
    {
      title: 'Progress',
      key: 'progress',
      width: '12%',
      render: (_: string, record: Execution) => {
        const progress = record.state === ExecutionState.COMPLETED ? 100 :
                        record.state === ExecutionState.ERROR ? 0 :
                        record.state === ExecutionState.RUNNING ? 50 : 0
        return (
          <Progress
            percent={progress}
            status={
              record.state === ExecutionState.ERROR ? 'exception' :
              record.state === ExecutionState.COMPLETED ? 'success' :
              'active'
            }
            size="small"
            format={percent => `${percent}%`}
          />
        )
      },
    },
    {
      title: 'Labels',
      dataIndex: 'labels',
      key: 'labels',
      width: '12%',
      render: (labels: any[] = []) => (
        <Space size="small" wrap>
          {labels.slice(0, 2).map((label, index) => (
            <Tag key={index} style={{ fontSize: '0.75rem' }}>
              {label.key}
            </Tag>
          ))}
          {labels.length > 2 && (
            <Tooltip title={`${labels.length - 2} more labels`}>
              <Tag style={{ fontSize: '0.75rem' }}>+{labels.length - 2}</Tag>
            </Tooltip>
          )}
        </Space>
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
      width: '10%',
      render: (_: string, record: Execution) => (
        <Space size="small">
          <Tooltip title="View Details">
            <Button
              type="text"
              icon={<EyeOutlined />}
              onClick={() => handleViewExecution(record.id)}
            />
          </Tooltip>
          <Tooltip title="Delete Execution">
            <Button
              type="text"
              danger
              icon={<DeleteOutlined />}
              onClick={() => handleDeleteExecution(record.id)}
            />
          </Tooltip>
        </Space>
      ),
    },
  ]

  return (
    <div className="execution-list">
      {/* Page Header */}
      <div className="page-header">
        <div className="page-header-left">
          <Title level={2} className="page-title">
            Algorithm Executions
          </Title>
          <Text type="secondary">
            Monitor and manage RCA algorithm executions
          </Text>
        </div>
        <div className="page-header-right">
          <Button
            type="primary"
            size="large"
            icon={<PlayCircleOutlined />}
            onClick={handleCreateExecution}
          >
            New Execution
          </Button>
        </div>
      </div>

      {/* Statistics Cards */}
      <Row gutter={[16, 16]} style={{ marginBottom: 24 }}>
        <Col xs={24} sm={12} md={6}>
          <StatCard
            title="Total Executions"
            value={stats.total}
            icon={<FunctionOutlined />}
            color="#3b82f6"
          />
        </Col>
        <Col xs={24} sm={12} md={6}>
          <StatCard
            title="Running"
            value={stats.running}
            icon={<SyncOutlined />}
            color="#f59e0b"
          />
        </Col>
        <Col xs={24} sm={12} md={6}>
          <StatCard
            title="Completed"
            value={stats.completed}
            icon={<CheckCircleOutlined />}
            color="#10b981"
          />
        </Col>
        <Col xs={24} sm={12} md={6}>
          <StatCard
            title="Failed"
            value={stats.failed}
            icon={<CloseCircleOutlined />}
            color="#ef4444"
          />
        </Col>
      </Row>

      {/* Filters and Actions */}
      <Card style={{ marginBottom: 16 }}>
        <Row gutter={[16, 16]} align="middle">
          <Col xs={24} sm={12} md={6}>
            <Search
              placeholder="Search executions..."
              allowClear
              enterButton={<SearchOutlined />}
              onSearch={handleSearch}
              style={{ width: '100%' }}
            />
          </Col>
          <Col xs={24} sm={12} md={4}>
            <Select
              placeholder="Filter by status"
              allowClear
              style={{ width: '100%' }}
              onChange={handleStateFilter}
              value={stateFilter}
            >
              <Option value={0}>Pending</Option>
              <Option value={1}>Running</Option>
              <Option value={2}>Completed</Option>
              <Option value={-1}>Error</Option>
            </Select>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Select
              placeholder="Filter by algorithm"
              allowClear
              style={{ width: '100%' }}
              onChange={handleAlgorithmFilter}
              value={algorithmFilter}
            >
              {algorithmsData?.data.map(algo => (
                <Option key={algo.id} value={algo.id}>
                  {algo.name}
                </Option>
              ))}
            </Select>
          </Col>
          <Col xs={24} sm={24} md={8} style={{ textAlign: 'right' }}>
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
              <Button icon={<ExportOutlined />}>
                Export
              </Button>
              <Button icon={<FilterOutlined />}>
                Advanced Filter
              </Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* Executions Table */}
      <Card>
        <Table
          rowKey="id"
          rowSelection={rowSelection}
          columns={columns}
          dataSource={executionsData?.data || []}
          loading={isLoading}
          pagination={{
            ...pagination,
            total: executionsData?.data.total || 0,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `${range[0]}-${range[1]} of ${total} executions`,
          }}
          onChange={handleTableChange}
        />
      </Card>
    </div>
  )
}

export default ExecutionList