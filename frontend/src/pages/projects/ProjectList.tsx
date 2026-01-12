import { useState } from 'react'
import {
  Table,
  Button,
  Input,
  Space,
  Tag,
  Typography,
  Popconfirm,
  message,
} from 'antd'
import type { ColumnsType } from 'antd/es/table'
import { PlusOutlined, SearchOutlined, DeleteOutlined } from '@ant-design/icons'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { projectApi } from '@/api/projects'
import type { Project } from '@/types/api'
import dayjs from 'dayjs'

const { Title } = Typography

const ProjectList = () => {
  const queryClient = useQueryClient()
  const [page, setPage] = useState(1)
  const [size, setSize] = useState(10)
  const [searchText, setSearchText] = useState('')

  // Fetch projects
  const { data, isLoading } = useQuery({
    queryKey: ['projects', { page, size }],
    queryFn: () => projectApi.getProjects({ page, size }),
  })

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: (id: number) => projectApi.deleteProject(id),
    onSuccess: () => {
      message.success('项目删除成功')
      queryClient.invalidateQueries({ queryKey: ['projects'] })
    },
    onError: () => {
      message.error('项目删除失败')
    },
  })

  const columns: ColumnsType<Project> = [
    {
      title: '项目名称',
      dataIndex: 'name',
      key: 'name',
      filteredValue: searchText ? [searchText] : null,
      onFilter: (value, record) =>
        record.name.toLowerCase().includes((value as string).toLowerCase()),
      render: (text: string) => (
        <Typography.Text strong style={{ color: '#2563eb' }}>
          {text}
        </Typography.Text>
      ),
    },
    {
      title: '描述',
      dataIndex: 'description',
      key: 'description',
      ellipsis: true,
      render: (text: string) => (
        <Typography.Text type="secondary">{text || '-'}</Typography.Text>
      ),
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
      title: '创建时间',
      dataIndex: 'created_at',
      key: 'created_at',
      width: 180,
      render: (time: string) => dayjs(time).format('YYYY-MM-DD HH:mm'),
    },
    {
      title: '标签',
      dataIndex: 'labels',
      key: 'labels',
      width: 200,
      render: (labels: { key: string; value: string }[]) => (
        <Space size={[0, 4]} wrap>
          {labels?.map((label, index) => (
            <Tag key={index} style={{ fontSize: '12px' }}>
              {label.key}: {label.value}
            </Tag>
          )) || '-'}
        </Space>
      ),
    },
    {
      title: '操作',
      key: 'actions',
      width: 150,
      render: (_, record) => (
        <Space>
          <Button type="link" size="small">
            查看
          </Button>
          <Button type="link" size="small">
            编辑
          </Button>
          <Popconfirm
            title="确认删除"
            description="确定要删除这个项目吗?"
            onConfirm={() => deleteMutation.mutate(record.id)}
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
    <div>
      <div
        style={{
          display: 'flex',
          justifyContent: 'space-between',
          alignItems: 'center',
          marginBottom: '24px',
        }}
      >
        <Title level={3} style={{ margin: 0 }}>
          项目管理
        </Title>
        <Space>
          <Input
            placeholder="搜索项目名称"
            prefix={<SearchOutlined />}
            value={searchText}
            onChange={(e) => setSearchText(e.target.value)}
            style={{ width: 250 }}
            allowClear
          />
          <Button type="primary" icon={<PlusOutlined />}>
            创建项目
          </Button>
        </Space>
      </div>

      <Table
        columns={columns}
        dataSource={data?.data.data || []}
        rowKey="id"
        loading={isLoading}
        pagination={{
          current: page,
          pageSize: size,
          total: data?.data.total || 0,
          showSizeChanger: true,
          showTotal: (total) => `共 ${total} 个项目`,
          onChange: (newPage, newSize) => {
            setPage(newPage)
            setSize(newSize)
          },
        }}
      />
    </div>
  )
}

export default ProjectList
