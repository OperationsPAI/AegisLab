import {
  ArrowLeftOutlined,
  ClockCircleOutlined,
  DeleteOutlined,
  EditOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import type { ContainerVersionResp } from '@rcabench/client';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Badge,
  Button,
  Card,
  Form,
  Input,
  message,
  Modal,
  Popconfirm,
  Space,
  Table,
  Tag,
  Tooltip,
  Typography,
} from 'antd';
import type { ColumnsType } from 'antd/es/table';
import dayjs from 'dayjs';
import { useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';

import { containerApi } from '@/api/containers';

const { Title, Text } = Typography;
const { TextArea } = Input;

interface VersionFormData {
  name: string;
  image_ref: string;
  command?: string;
}

const ContainerVersions = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const { id } = useParams<{ id: string }>();
  const containerId = Number(id);
  const [isModalVisible, setIsModalVisible] = useState(false);
  const [editingVersion, setEditingVersion] = useState<ContainerVersionResp | null>(
    null
  );
  const [form] = Form.useForm<VersionFormData>();

  // Fetch container details
  const { data: container } = useQuery({
    queryKey: ['container', containerId],
    queryFn: () => containerApi.getContainer(containerId),
    enabled: !!containerId,
  });

  // Fetch versions
  const { data: versionsData, isLoading: versionsLoading } = useQuery({
    queryKey: ['container-versions', containerId],
    queryFn: () => containerApi.getVersions(containerId),
    enabled: !!containerId,
  });
  const versions = versionsData?.items || [];

  // Create version mutation
  const createVersionMutation = useMutation({
    mutationFn: (data: VersionFormData) =>
      containerApi.createVersion(containerId, data),
    onSuccess: () => {
      message.success('版本创建成功');
      queryClient.invalidateQueries({
        queryKey: ['container-versions', containerId],
      });
      setIsModalVisible(false);
      form.resetFields();
    },
    onError: () => {
      message.error('版本创建失败');
    },
  });

  // Update version mutation
  const updateVersionMutation = useMutation({
    mutationFn: ({
      versionId,
      data,
    }: {
      versionId: number;
      data: Partial<VersionFormData>;
    }) => containerApi.updateVersion(containerId, versionId, data),
    onSuccess: () => {
      message.success('版本更新成功');
      queryClient.invalidateQueries({
        queryKey: ['container-versions', containerId],
      });
      setIsModalVisible(false);
      setEditingVersion(null);
      form.resetFields();
    },
    onError: () => {
      message.error('版本更新失败');
    },
  });

  // Delete version mutation
  const deleteVersionMutation = useMutation({
    mutationFn: (versionId: number) =>
      containerApi.deleteVersion(containerId, versionId),
    onSuccess: () => {
      message.success('版本删除成功');
      queryClient.invalidateQueries({
        queryKey: ['container-versions', containerId],
      });
    },
    onError: () => {
      message.error('版本删除失败');
    },
  });

  const handleCreateVersion = () => {
    setEditingVersion(null);
    form.resetFields();
    setIsModalVisible(true);
  };

  const handleEditVersion = (version: ContainerVersionResp) => {
    setEditingVersion(version);
    form.setFieldsValue({
      name: version.name || '',
      image_ref: version.image_ref || '',
    });
    setIsModalVisible(true);
  };

  const handleDeleteVersion = (versionId: number | undefined) => {
    if (versionId !== undefined) {
      deleteVersionMutation.mutate(versionId);
    }
  };

  const handleModalOk = () => {
    form.validateFields().then((values) => {
      if (editingVersion && editingVersion.id !== undefined) {
        updateVersionMutation.mutate({
          versionId: editingVersion.id,
          data: values,
        });
      } else {
        createVersionMutation.mutate(values);
      }
    });
  };

  const handleModalCancel = () => {
    setIsModalVisible(false);
    setEditingVersion(null);
    form.resetFields();
  };

  const columns: ColumnsType<ContainerVersionResp> = [
    {
      title: '版本',
      dataIndex: 'name',
      key: 'name',
      width: 120,
      render: (name: string) => (
        <Badge
          count={name}
          style={{ backgroundColor: '#3b82f6', fontWeight: 'bold' }}
        />
      ),
    },
    {
      title: '镜像引用',
      dataIndex: 'image_ref',
      key: 'image_ref',
      render: (imageRef: string) => (
        <Tooltip title={imageRef}>
          <Text code ellipsis style={{ maxWidth: 300, fontSize: '0.875rem' }}>
            {imageRef}
          </Text>
        </Tooltip>
      ),
    },
    {
      title: '使用次数',
      dataIndex: 'usage',
      key: 'usage',
      width: 100,
      render: (usage?: number) => (
        <Tag color='blue'>{usage ?? 0}</Tag>
      ),
    },
    {
      title: '更新时间',
      dataIndex: 'updated_at',
      key: 'updated_at',
      width: 180,
      render: (date: string) => (
        <Space>
          <ClockCircleOutlined />
          <Text>{dayjs(date).format('YYYY-MM-DD HH:mm')}</Text>
        </Space>
      ),
    },
    {
      title: '操作',
      key: 'actions',
      width: 150,
      fixed: 'right' as const,
      render: (_, record) => (
        <Space>
          <Button
            type='link'
            size='small'
            icon={<EditOutlined />}
            onClick={() => handleEditVersion(record)}
          >
            编辑
          </Button>
          <Popconfirm
            title='确认删除'
            description='确定要删除这个版本吗？'
            onConfirm={() => handleDeleteVersion(record.id)}
            okText='确认'
            cancelText='取消'
          >
            <Button
              type='link'
              size='small'
              danger
              icon={<DeleteOutlined />}
              loading={deleteVersionMutation.isPending}
            />
          </Popconfirm>
        </Space>
      ),
    },
  ];

  return (
    <div style={{ padding: 24 }}>
      {/* Header */}
      <div style={{ marginBottom: 24 }}>
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate(`/containers/${containerId}`)}
          >
            返回容器详情
          </Button>
          <Title level={2} style={{ margin: 0 }}>
            {container?.name} - 版本管理
          </Title>
        </Space>
      </div>

      {/* Versions Table */}
      <Card
        title='容器版本列表'
        extra={
          <Button
            type='primary'
            icon={<PlusOutlined />}
            onClick={handleCreateVersion}
          >
            添加版本
          </Button>
        }
      >
        <Table
          rowKey='id'
          columns={columns}
          dataSource={versions}
          loading={versionsLoading}
          pagination={{
            pageSize: 10,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total, range) =>
              `${range[0]}-${range[1]} / ${total} 个版本`,
          }}
        />
      </Card>

      {/* Create/Edit Version Modal */}
      <Modal
        title={editingVersion ? '编辑版本' : '添加版本'}
        open={isModalVisible}
        onOk={handleModalOk}
        onCancel={handleModalCancel}
        confirmLoading={
          createVersionMutation.isPending || updateVersionMutation.isPending
        }
        width={600}
      >
        <Form form={form} layout='vertical' style={{ marginTop: 24 }}>
          <Form.Item
            label='版本名称'
            name='name'
            rules={[
              { required: true, message: '请输入版本名称' },
              {
                pattern: /^[a-zA-Z0-9._-]+$/,
                message: '版本名称只能包含字母、数字、点、下划线和连字符',
              },
            ]}
          >
            <Input placeholder='v1.0.0 或 latest' />
          </Form.Item>

          <Form.Item
            label='镜像引用'
            name='image_ref'
            rules={[{ required: true, message: '请输入完整的镜像引用' }]}
          >
            <Input placeholder='docker.io/username/image:tag' />
          </Form.Item>

          <Form.Item label='启动命令（可选）' name='command'>
            <TextArea
              rows={3}
              placeholder='容器启动命令或参数'
            />
          </Form.Item>

          <div
            style={{
              padding: 12,
              background: '#f5f5f5',
              borderRadius: 4,
              marginTop: 16,
            }}
          >
            <Text type='secondary' style={{ fontSize: '0.875rem' }}>
              <strong>提示：</strong> 镜像引用格式为：
              <br />
              <code>registry/repository:tag</code>
              <br />
              例如：<code>docker.io/library/nginx:latest</code>
            </Text>
          </div>
        </Form>
      </Modal>
    </div>
  );
};

export default ContainerVersions;
