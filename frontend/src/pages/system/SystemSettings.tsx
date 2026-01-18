import {
  ClockCircleOutlined,
  DeleteOutlined,
  GlobalOutlined,
  MailOutlined,
  UserAddOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import {
  Avatar,
  Button,
  Card,
  Col,
  Form,
  Input,
  List,
  message,
  Modal,
  Row,
  Skeleton,
  Space,
  Statistic,
  Switch,
  Tag,
  Typography,
} from 'antd';
import dayjs from 'dayjs';
import { useState } from 'react';

import {
  usersApi,
  type UserResp,
  type CreateUserReq,
  type UpdateUserReq,
} from '../../api/users';

const { Title, Text } = Typography;

const SystemSettings = () => {
  const queryClient = useQueryClient();
  const [createModalVisible, setCreateModalVisible] = useState(false);
  const [editModalVisible, setEditModalVisible] = useState(false);
  const [editingUser, setEditingUser] = useState<{
    id: number;
    username: string;
    email: string;
    full_name?: string;
  } | null>(null);
  const [createForm] = Form.useForm();
  const [editForm] = Form.useForm();

  // 获取用户列表
  const {
    data: usersData,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['users'],
    queryFn: () => usersApi.getUsers({ page: 1, size: 50 }),
  });

  // 创建用户
  const createUserMutation = useMutation({
    mutationFn: (data: CreateUserReq) => usersApi.createUser(data),
    onSuccess: () => {
      message.success('User created successfully');
      setCreateModalVisible(false);
      createForm.resetFields();
      queryClient.invalidateQueries({ queryKey: ['users'] });
    },
    onError: () => {
      // 错误已在 apiClient 拦截器中处理
    },
  });

  // 更新用户
  const updateUserMutation = useMutation({
    mutationFn: ({ id, data }: { id: number; data: UpdateUserReq }) =>
      usersApi.updateUser(id, data),
    onSuccess: () => {
      message.success('User updated successfully');
      setEditModalVisible(false);
      setEditingUser(null);
      editForm.resetFields();
      queryClient.invalidateQueries({ queryKey: ['users'] });
    },
    onError: () => {
      // 错误已在 apiClient 拦截器中处理
    },
  });

  // 删除用户
  const deleteUserMutation = useMutation({
    mutationFn: (id: number) => usersApi.deleteUser(id),
    onSuccess: () => {
      message.success('User deleted successfully');
      queryClient.invalidateQueries({ queryKey: ['users'] });
    },
    onError: () => {
      // 错误已在 apiClient 拦截器中处理
    },
  });

  const handleCreateUser = (values: {
    username: string;
    email: string;
    password: string;
    full_name?: string;
  }) => {
    createUserMutation.mutate({
      username: values.username,
      email: values.email,
      password: values.password,
      full_name: values.full_name || '',
    });
  };

  const handleEditUser = (values: {
    email?: string;
    full_name?: string;
    is_active?: boolean;
  }) => {
    if (!editingUser) return;
    updateUserMutation.mutate({
      id: editingUser.id,
      data: values,
    });
  };

  const handleDeleteUser = (userId: number, username: string) => {
    Modal.confirm({
      title: 'Delete User',
      content: `Are you sure you want to delete user "${username}"? This action cannot be undone.`,
      okText: 'Yes, Delete',
      okButtonProps: { danger: true },
      cancelText: 'Cancel',
      onOk: () => deleteUserMutation.mutate(userId),
    });
  };

  const handleToggleUserStatus = (
    userId: number,
    currentStatus: boolean,
    username: string
  ) => {
    const action = currentStatus ? 'deactivate' : 'activate';
    Modal.confirm({
      title: `${currentStatus ? 'Deactivate' : 'Activate'} User`,
      content: `Are you sure you want to ${action} user "${username}"?`,
      okText: `Yes, ${currentStatus ? 'Deactivate' : 'Activate'}`,
      okButtonProps: { danger: currentStatus },
      cancelText: 'Cancel',
      onOk: () =>
        updateUserMutation.mutate({
          id: userId,
          data: { is_active: !currentStatus },
        }),
    });
  };

  const openEditModal = (user: {
    id: number;
    username: string;
    email: string;
    full_name?: string;
    is_active?: boolean;
  }) => {
    setEditingUser(user);
    editForm.setFieldsValue({
      email: user.email,
      full_name: user.full_name,
      is_active: user.is_active,
    });
    setEditModalVisible(true);
  };

  const getStatusColor = (isActive: boolean) => {
    return isActive ? 'green' : 'orange';
  };

  if (isLoading) {
    return (
      <div style={{ padding: 24 }}>
        <Skeleton active paragraph={{ rows: 10 }} />
      </div>
    );
  }

  if (error) {
    return (
      <div style={{ padding: 24 }}>
        <Card>
          <Text type='danger'>Failed to load system settings</Text>
        </Card>
      </div>
    );
  }

  const users: UserResp[] = usersData?.items || [];
  const totalUsers = usersData?.pagination?.total || 0;
  const activeUsers = users.filter((u: UserResp) => u.is_active).length;

  return (
    <div style={{ padding: 24 }}>
      {/* Header */}
      <div style={{ marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0 }}>
          <GlobalOutlined style={{ marginRight: 8 }} />
          System Settings
        </Title>
        <Text type='secondary'>Manage users and system configuration</Text>
      </div>

      {/* System Overview */}
      <Card style={{ marginBottom: 24 }}>
        <Title level={4} style={{ marginBottom: 16 }}>
          System Overview
        </Title>
        <Row gutter={[{ xs: 8, sm: 16, lg: 24 }, { xs: 8, sm: 16, lg: 24 }]} className='stats-row'>
          <Col xs={12} sm={12} lg={8}>
            <Card size='small'>
              <Statistic
                title='Total Users'
                value={totalUsers}
                prefix={<UserOutlined />}
                valueStyle={{ color: '#3b82f6' }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={12} lg={8}>
            <Card size='small'>
              <Statistic
                title='Active Users'
                value={activeUsers}
                prefix={<UserOutlined />}
                valueStyle={{ color: '#10b981' }}
              />
            </Card>
          </Col>
          <Col xs={12} sm={12} lg={8}>
            <Card size='small'>
              <Statistic
                title='Inactive Users'
                value={totalUsers - activeUsers}
                prefix={<UserOutlined />}
                valueStyle={{ color: '#f59e0b' }}
              />
            </Card>
          </Col>
        </Row>
      </Card>

      {/* User Management */}
      <Card
        title='User Management'
        extra={
          <Button
            type='primary'
            icon={<UserAddOutlined />}
            onClick={() => setCreateModalVisible(true)}
          >
            Create User
          </Button>
        }
      >
        <List<UserResp>
          itemLayout='horizontal'
          dataSource={users}
          locale={{ emptyText: 'No users found' }}
          renderItem={(item: UserResp) => (
            <List.Item
              key={item.id}
              actions={[
                <Button
                  key='edit'
                  type='link'
                  size='small'
                  onClick={() =>
                    openEditModal({
                      id: item.id!,
                      username: item.username!,
                      email: item.email!,
                      full_name: item.full_name,
                      is_active: item.is_active,
                    })
                  }
                >
                  Edit
                </Button>,
                <Button
                  key='toggle'
                  type='link'
                  danger={item.is_active}
                  size='small'
                  onClick={() =>
                    handleToggleUserStatus(
                      item.id!,
                      item.is_active!,
                      item.username!
                    )
                  }
                >
                  {item.is_active ? 'Deactivate' : 'Activate'}
                </Button>,
                <Button
                  key='delete'
                  type='link'
                  danger
                  size='small'
                  icon={<DeleteOutlined />}
                  onClick={() => handleDeleteUser(item.id!, item.username!)}
                >
                  Delete
                </Button>,
              ]}
            >
              <List.Item.Meta
                avatar={<Avatar icon={<UserOutlined />} src={item.avatar} />}
                title={
                  <Space>
                    <Text strong>{item.full_name || item.username}</Text>
                    <Text type='secondary'>@{item.username}</Text>
                    <Tag color={getStatusColor(item.is_active!)}>
                      {item.is_active ? 'Active' : 'Inactive'}
                    </Tag>
                  </Space>
                }
                description={
                  <Space direction='vertical' size={0}>
                    <Space>
                      <MailOutlined />
                      <Text>{item.email}</Text>
                    </Space>
                    <Space>
                      <ClockCircleOutlined />
                      <Text type='secondary'>
                        Last login:{' '}
                        {item.last_login_at
                          ? dayjs(item.last_login_at).format(
                              'YYYY-MM-DD HH:mm'
                            )
                          : 'Never'}
                      </Text>
                    </Space>
                  </Space>
                }
              />
            </List.Item>
          )}
        />
      </Card>

      {/* Create User Modal */}
      <Modal
        title='Create User'
        open={createModalVisible}
        onCancel={() => {
          setCreateModalVisible(false);
          createForm.resetFields();
        }}
        footer={[
          <Button
            key='cancel'
            onClick={() => {
              setCreateModalVisible(false);
              createForm.resetFields();
            }}
          >
            Cancel
          </Button>,
          <Button
            key='submit'
            type='primary'
            loading={createUserMutation.isPending}
            onClick={() => createForm.submit()}
          >
            Create
          </Button>,
        ]}
      >
        <Form form={createForm} layout='vertical' onFinish={handleCreateUser}>
          <Form.Item
            label='Username'
            name='username'
            rules={[
              { required: true, message: 'Please enter username' },
              { min: 3, message: 'Username must be at least 3 characters' },
            ]}
          >
            <Input placeholder='Enter username' />
          </Form.Item>

          <Form.Item
            label='Email'
            name='email'
            rules={[
              { required: true, message: 'Please enter email' },
              { type: 'email', message: 'Please enter a valid email' },
            ]}
          >
            <Input placeholder='Enter email' />
          </Form.Item>

          <Form.Item
            label='Password'
            name='password'
            rules={[
              { required: true, message: 'Please enter password' },
              { min: 8, message: 'Password must be at least 8 characters' },
            ]}
          >
            <Input.Password placeholder='Enter password' />
          </Form.Item>

          <Form.Item label='Full Name' name='full_name'>
            <Input placeholder='Enter full name (optional)' />
          </Form.Item>
        </Form>
      </Modal>

      {/* Edit User Modal */}
      <Modal
        title={`Edit User: ${editingUser?.username}`}
        open={editModalVisible}
        onCancel={() => {
          setEditModalVisible(false);
          setEditingUser(null);
          editForm.resetFields();
        }}
        footer={[
          <Button
            key='cancel'
            onClick={() => {
              setEditModalVisible(false);
              setEditingUser(null);
              editForm.resetFields();
            }}
          >
            Cancel
          </Button>,
          <Button
            key='submit'
            type='primary'
            loading={updateUserMutation.isPending}
            onClick={() => editForm.submit()}
          >
            Save
          </Button>,
        ]}
      >
        <Form form={editForm} layout='vertical' onFinish={handleEditUser}>
          <Form.Item
            label='Email'
            name='email'
            rules={[{ type: 'email', message: 'Please enter a valid email' }]}
          >
            <Input placeholder='Enter email' />
          </Form.Item>

          <Form.Item label='Full Name' name='full_name'>
            <Input placeholder='Enter full name' />
          </Form.Item>

          <Form.Item
            label='Active Status'
            name='is_active'
            valuePropName='checked'
          >
            <Switch checkedChildren='Active' unCheckedChildren='Inactive' />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default SystemSettings;
