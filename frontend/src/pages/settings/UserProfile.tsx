import {
  CloseOutlined,
  EditOutlined,
  EyeInvisibleOutlined,
  EyeOutlined,
  KeyOutlined,
  MailOutlined,
  PhoneOutlined,
  SaveOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import {
  Avatar,
  Button,
  Card,
  Col,
  Descriptions,
  Form,
  Input,
  message,
  Modal,
  Row,
  Skeleton,
  Space,
  Tag,
  Typography,
} from 'antd';
import dayjs from 'dayjs';
import { useState } from 'react';

import { authApi } from '../../api/auth';

const { Title, Text } = Typography;

const UserProfile = () => {
  const [isEditing, setIsEditing] = useState(false);
  const [form] = Form.useForm();
  const [passwordModalVisible, setPasswordModalVisible] = useState(false);
  const [passwordForm] = Form.useForm();
  const [changingPassword, setChangingPassword] = useState(false);

  // 使用 TanStack Query 获取用户数据
  const {
    data: userData,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['profile'],
    queryFn: () => authApi.getProfile(),
  });

  const handleEditProfile = () => {
    if (!userData) return;
    setIsEditing(true);
    form.setFieldsValue({
      full_name: userData.full_name,
      email: userData.email,
      phone: userData.phone,
    });
  };

  const handleSaveProfile = async (values: Record<string, unknown>) => {
    try {
      // TODO: 实现更新资料 API（后端需要添加）
      console.log('Updating profile:', values);
      message.success('Profile updated successfully');
      setIsEditing(false);
    } catch (error) {
      message.error('Failed to update profile');
      console.error('Update profile error:', error);
    }
  };

  const handleCancelEdit = () => {
    setIsEditing(false);
    form.resetFields();
  };

  const handleChangePassword = async (values: {
    oldPassword: string;
    newPassword: string;
  }) => {
    setChangingPassword(true);
    try {
      await authApi.changePassword({
        old_password: values.oldPassword,
        new_password: values.newPassword,
      });
      message.success('Password changed successfully');
      setPasswordModalVisible(false);
      passwordForm.resetFields();
    } catch {
      // 错误已在 apiClient 拦截器中处理
    } finally {
      setChangingPassword(false);
    }
  };

  if (isLoading) {
    return (
      <div style={{ padding: 24 }}>
        <Skeleton active avatar paragraph={{ rows: 10 }} />
      </div>
    );
  }

  if (error || !userData) {
    return (
      <div style={{ padding: 24 }}>
        <Card>
          <Text type='danger'>Failed to load user profile</Text>
        </Card>
      </div>
    );
  }

  return (
    <div style={{ padding: 24 }}>
      {/* Header */}
      <div style={{ marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0 }}>
          <UserOutlined style={{ marginRight: 8 }} />
          User Profile
        </Title>
        <Text type='secondary'>
          Manage your profile information and account settings
        </Text>
      </div>

      {/* Profile Overview */}
      <Card style={{ marginBottom: 24 }}>
        <Row gutter={[24, 24]} align='middle'>
          <Col xs={24} sm={6} md={4}>
            <div style={{ textAlign: 'center' }}>
              <Avatar
                size={128}
                icon={<UserOutlined />}
                src={userData.avatar}
                style={{ backgroundColor: '#3b82f6' }}
              />
              <div style={{ marginTop: 16 }}>
                <Title level={4} style={{ margin: 0 }}>
                  {userData.full_name || userData.username}
                </Title>
                <Text type='secondary'>@{userData.username}</Text>
              </div>
            </div>
          </Col>
          <Col xs={24} sm={18} md={20}>
            <Descriptions
              title='Profile Information'
              bordered
              column={{ xs: 1, sm: 2, md: 3 }}
              extra={
                !isEditing && (
                  <Button
                    type='primary'
                    icon={<EditOutlined />}
                    onClick={handleEditProfile}
                  >
                    Edit Profile
                  </Button>
                )
              }
            >
              <Descriptions.Item label='Full Name'>
                {isEditing ? (
                  <Form
                    form={form}
                    layout='inline'
                    onFinish={handleSaveProfile}
                  >
                    <Form.Item name='full_name' style={{ margin: 0 }}>
                      <Input />
                    </Form.Item>
                  </Form>
                ) : (
                  userData.full_name || '-'
                )}
              </Descriptions.Item>
              <Descriptions.Item label='Email'>
                {isEditing ? (
                  <Form form={form} layout='inline'>
                    <Form.Item name='email' style={{ margin: 0 }}>
                      <Input />
                    </Form.Item>
                  </Form>
                ) : (
                  <Space>
                    <MailOutlined />
                    {userData.email}
                  </Space>
                )}
              </Descriptions.Item>
              <Descriptions.Item label='Phone'>
                {isEditing ? (
                  <Form form={form} layout='inline'>
                    <Form.Item name='phone' style={{ margin: 0 }}>
                      <Input />
                    </Form.Item>
                  </Form>
                ) : (
                  <Space>
                    <PhoneOutlined />
                    {userData.phone || '-'}
                  </Space>
                )}
              </Descriptions.Item>
              <Descriptions.Item label='Status'>
                <Tag color={userData.is_active ? 'green' : 'orange'}>
                  {userData.is_active ? 'Active' : 'Inactive'}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label='Member Since'>
                {userData.created_at
                  ? dayjs(userData.created_at).format('MMMM D, YYYY')
                  : '-'}
              </Descriptions.Item>
              <Descriptions.Item label='Last Login'>
                {userData.last_login_at
                  ? dayjs(userData.last_login_at).format('MMMM D, YYYY HH:mm')
                  : 'Never'}
              </Descriptions.Item>
            </Descriptions>

            {isEditing && (
              <div style={{ marginTop: 16, textAlign: 'right' }}>
                <Space>
                  <Button icon={<CloseOutlined />} onClick={handleCancelEdit}>
                    Cancel
                  </Button>
                  <Button
                    type='primary'
                    icon={<SaveOutlined />}
                    onClick={() => form.submit()}
                  >
                    Save Changes
                  </Button>
                </Space>
              </div>
            )}
          </Col>
        </Row>
      </Card>

      {/* Security Settings */}
      <Card
        title={
          <Space>
            <KeyOutlined />
            Security Settings
          </Space>
        }
        extra={
          <Button
            type='primary'
            icon={<KeyOutlined />}
            onClick={() => setPasswordModalVisible(true)}
          >
            Change Password
          </Button>
        }
      >
        <Descriptions bordered column={1}>
          <Descriptions.Item label='Password'>
            <Text type='secondary'>••••••••••••</Text>
          </Descriptions.Item>
          <Descriptions.Item label='Account Status'>
            <Tag color={userData.is_active ? 'green' : 'red'}>
              {userData.is_active ? 'Active' : 'Inactive'}
            </Tag>
          </Descriptions.Item>
        </Descriptions>
      </Card>

      {/* Change Password Modal */}
      <Modal
        title='Change Password'
        open={passwordModalVisible}
        onCancel={() => {
          setPasswordModalVisible(false);
          passwordForm.resetFields();
        }}
        footer={[
          <Button
            key='cancel'
            onClick={() => {
              setPasswordModalVisible(false);
              passwordForm.resetFields();
            }}
          >
            Cancel
          </Button>,
          <Button
            key='submit'
            type='primary'
            icon={<SaveOutlined />}
            loading={changingPassword}
            onClick={() => passwordForm.submit()}
          >
            Change Password
          </Button>,
        ]}
      >
        <Form
          form={passwordForm}
          layout='vertical'
          onFinish={handleChangePassword}
        >
          <Form.Item
            label='Current Password'
            name='oldPassword'
            rules={[
              { required: true, message: 'Please enter your current password' },
            ]}
          >
            <Input.Password
              placeholder='Enter current password'
              iconRender={(visible) =>
                visible ? <EyeOutlined /> : <EyeInvisibleOutlined />
              }
            />
          </Form.Item>

          <Form.Item
            label='New Password'
            name='newPassword'
            rules={[
              { required: true, message: 'Please enter a new password' },
              { min: 8, message: 'Password must be at least 8 characters' },
            ]}
          >
            <Input.Password
              placeholder='Enter new password'
              iconRender={(visible) =>
                visible ? <EyeOutlined /> : <EyeInvisibleOutlined />
              }
            />
          </Form.Item>

          <Form.Item
            label='Confirm New Password'
            name='confirmPassword'
            dependencies={['newPassword']}
            rules={[
              { required: true, message: 'Please confirm your new password' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('newPassword') === value) {
                    return Promise.resolve();
                  }
                  return Promise.reject(new Error('Passwords do not match'));
                },
              }),
            ]}
          >
            <Input.Password
              placeholder='Confirm new password'
              iconRender={(visible) =>
                visible ? <EyeOutlined /> : <EyeInvisibleOutlined />
              }
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default UserProfile;
