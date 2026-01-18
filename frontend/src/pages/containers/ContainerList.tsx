import { PlusOutlined, SearchOutlined, DeleteOutlined } from '@ant-design/icons'
import type { ContainerResp } from '@rcabench/client'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Table,
  Button,
  Input,
  Space,
  Tag,
  Typography,
  Select,
  Popconfirm,
  message,
  Card,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import dayjs from 'dayjs'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { containerApi } from '@/api/containers'


const { Title } = Typography

const ContainerList = () => {
  const navigate = useNavigate()
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [size, setSize] = useState(10)
  const [searchText, setSearchText] = useState('')
  const [typeFilter, setTypeFilter] = useState<string | undefined>()

  // Fetch containers
  const { data, isLoading } = useQuery({
    queryKey: ['containers', { page, size, type: typeFilter }],
    queryFn: () => containerApi.getContainers({ page, size, type: typeFilter as any }),
  })

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: (id: number) => containerApi.deleteContainer(id),
    onSuccess: () => {
      message.success('容器删除成功')
      queryClient.invalidateQueries({ queryKey: ['containers'] })
    },
    onError: () => {
      message.error('容器删除失败')
    },
  })

  const columns: ColumnsType<ContainerResp> = [
    {
      title: '容器名称',
      dataIndex: 'name',
      key: 'name',
      filteredValue: searchText ? [searchText] : null,
      onFilter: (value, record) =>
        record.name?.toLowerCase().includes((value as string).toLowerCase()) ?? false,
      render: (text: string) => (
        <Typography.Text strong style={{ color: '#2563eb' }}>
          {text}
        </Typography.Text>
      ),
    },
    {
      title: '类型',
      dataIndex: 'type',
      key: 'type',
      width: 120,
      render: (type: string) => {
        const colorMap: Record<string, string> = {
          Pedestal: 'blue',
          Benchmark: 'green',
          Algorithm: 'purple',
        }
        return <Tag color={colorMap[type] || 'default'}>{type}</Tag>
      },
    },
    {
      title: '可见性',
      dataIndex: 'is_public',
      key: 'is_public',
      width: 100,
      render: (isPublic: boolean) => (
        <Tag color={isPublic ? 'green' : 'orange'}>
          {isPublic ? '公开' : '私有'}
        </Tag>
      ),
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      width: 100,
      render: (status: string) => (
        <Tag color={status === 'active' ? 'green' : 'default'}>
          {status || '-'}
        </Tag>
      ),
    },
    {
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: '操作',
      key: 'actions',
      width: 150,
      render: (_, record) => (
        <Space>
          <Button
            type="link"
            size="small"
            onClick={() => navigate(`/containers/${record.id}`)}
          >
            查看
          </Button>
          <Button
            type="link"
            size="small"
            onClick={() => navigate(`/containers/${record.id}/edit`)}
          >
            编辑
          </Button>
          <Popconfirm
            title="确认删除"
            description="确定要删除这个容器吗?"
            onConfirm={() => record.id !== undefined && deleteMutation.mutate(record.id)}
            okText="确认"
            cancelText="取消"
          >
            <Button
              type="link"
              size="small"
              danger
              icon={<DeleteOutlined />}
              loading={deleteMutation.isPending}
            />
          </Popconfirm>
        </Space>
      ),
    },
  ]

  return (
    <div className="container-list page-container">
      <div className="page-header"
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '24px',
        }}
      >
        <Title level={3} className="page-title" style={{ margin: 0 }}>
          容器管理
        </Title>
        <Space>
          <Select
            placeholder="容器类型"
            style={{ width: 150 }}
            allowClear
            value={typeFilter}
            onChange={(value) => setTypeFilter(value)}
            options={[
              { label: 'Pedestal', value: 'Pedestal' },
              { label: 'Benchmark', value: 'Benchmark' },
              { label: 'Algorithm', value: 'Algorithm' },
            ]}
          />
          <Input
            placeholder="搜索容器名称"
            prefix={<SearchOutlined />}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            style={{ width: 250 }}
            allowClear
          />
          <Button type="primary" icon={<PlusOutlined />} onClick={() => navigate('/containers/new')}>
            创建容器
          </Button>
        </Space>
      </div>

      <Card className='table-card'>
        <Table
          columns={columns}
          dataSource={data?.items || []}
          rowKey="id"
          loading={isLoading}
          className='containers-table'
          pagination={{
            current: page,
            pageSize: size,
            total: data?.pagination?.total || 0,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `共 ${total} 个容器`,
            onChange: (newPage, newSize) => {
              setPage(newPage)
              setSize(newSize)
            },
          }}
        />
      </Card>
    </div>
  )
}

export default ContainerList
