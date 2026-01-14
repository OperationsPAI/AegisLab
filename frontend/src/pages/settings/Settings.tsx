import {
  BellOutlined,
  EyeInvisibleOutlined,
  EyeOutlined,
  LockOutlined,
  MailOutlined,
  PhoneOutlined,
  SaveOutlined,
  SettingOutlined,
  UserOutlined,
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import {
  Alert,
  Avatar,
  Button,
  Card,
  Col,
  Descriptions,
  Divider,
  Form,
  Input,
  message,
  Row,
  Select,
  Skeleton,
  Space,
  Switch,
  Tabs,
  Tag,
  Typography,
} from 'antd';
import dayjs from 'dayjs';
import { useState } from 'react';

import { authApi } from '../../api/auth';

const { Title, Text } = Typography;

const Settings = () => {
  const [activeTab, setActiveTab] = useState('profile');
  const [profileForm] = Form.useForm();
  const [notificationForm] = Form.useForm();
  const [securityForm] = Form.useForm();
  const [loading, setLoading] = useState(false);

  // 使用 TanStack Query 获取用户数据
  const {
    data: userData,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['profile'],
    queryFn: () => authApi.getProfile(),
  });

  const handleSaveProfile = async (values: Record<string, unknown>) => {
    setLoading(true);
    try {
      // TODO: 实现更新资料 API（后端需要添加）
      console.log('Updating profile:', values);
      message.success('Profile updated successfully');
    } catch (err) {
      message.error('Failed to update profile');
      console.error('Update profile error:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleSaveNotifications = async (values: Record<string, unknown>) => {
    setLoading(true);
    try {
      // TODO: 实现通知设置 API（后端需要添加）
      console.log('Updating notifications:', values);
      message.success('Notification settings updated successfully');
    } catch (err) {
      message.error('Failed to update notification settings');
      console.error('Update notifications error:', err);
    } finally {
      setLoading(false);
    }
  };

  const handleChangePassword = async (values: {
    oldPassword: string;
    newPassword: string;
  }) => {
    setLoading(true);
    try {
      await authApi.changePassword({
        old_password: values.oldPassword,
        new_password: values.newPassword,
      });
      message.success('Password changed successfully');
      securityForm.resetFields();
    } catch {
      // 错误已在 apiClient 拦截器中处理
    } finally {
      setLoading(false);
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
          <Text type='danger'>Failed to load user settings</Text>
        </Card>
      </div>
    );
  }

  return (
    <div style={{ padding: 24 }}>
      {/* Header */}
      <div style={{ marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0 }}>
          <SettingOutlined style={{ marginRight: 8 }} />
          Settings
        </Title>
        <Text type='secondary'>
          Manage your account settings and preferences
        </Text>
      </div>

      {/* Profile Overview */}
      <Card style={{ marginBottom: 24 }}>
        <Row gutter={[24, 24]} align='middle'>
          <Col xs={24} sm={6} md={4}>
            <div style={{ textAlign: 'center' }}>
              <Avatar
                size={96}
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
              title='Account Overview'
              bordered
              column={{ xs: 1, sm: 2, md: 3 }}
            >
              <Descriptions.Item label='Email'>
                <Space>
                  <MailOutlined />
                  {userData.email}
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label='Phone'>
                <Space>
                  <PhoneOutlined />
                  {userData.phone || '-'}
                </Space>
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
          </Col>
        </Row>
      </Card>

      {/* Settings Tabs */}
      <Tabs
        activeKey={activeTab}
        onChange={setActiveTab}
        items={[
          {
            key: 'profile',
            label: (
              <span>
                <UserOutlined />
                Profile
              </span>
            ),
            children: (
              <Card
                title='Profile Settings'
                extra={
                  <Button
                    type='primary'
                    icon={<SaveOutlined />}
                    loading={loading}
                    onClick={() => profileForm.submit()}
                  >
                    Save Changes
                  </Button>
                }
              >
                <Form
                  form={profileForm}
                  layout='vertical'
                  onFinish={handleSaveProfile}
                  initialValues={{
                    full_name: userData.full_name,
                    email: userData.email,
                    phone: userData.phone,
                    timezone: 'UTC',
                    language: 'en',
                    dateFormat: 'YYYY-MM-DD',
                    timeFormat: '24h',
                  }}
                >
                  <Row gutter={[24, 24]}>
                    <Col xs={24} lg={12}>
                      <Form.Item
                        label='Full Name'
                        name='full_name'
                        rules={[
                          {
                            required: true,
                            message: 'Please enter your full name',
                          },
                        ]}
                      >
                        <Input prefix={<UserOutlined />} />
                      </Form.Item>

                      <Form.Item
                        label='Email'
                        name='email'
                        rules={[
                          {
                            required: true,
                            message: 'Please enter your email',
                          },
                          {
                            type: 'email',
                            message: 'Please enter a valid email',
                          },
                        ]}
                      >
                        <Input prefix={<MailOutlined />} />
                      </Form.Item>

                      <Form.Item label='Phone' name='phone'>
                        <Input prefix={<PhoneOutlined />} />
                      </Form.Item>
                    </Col>

                    <Col xs={24} lg={12}>
                      <Form.Item
                        label='Timezone'
                        name='timezone'
                        rules={[
                          {
                            required: true,
                            message: 'Please select your timezone',
                          },
                        ]}
                      >
                        <Select>
                          <Select.Option value='UTC'>UTC</Select.Option>
                          <Select.Option value='America/New_York'>
                            America/New_York
                          </Select.Option>
                          <Select.Option value='Europe/London'>
                            Europe/London
                          </Select.Option>
                          <Select.Option value='Asia/Shanghai'>
                            Asia/Shanghai
                          </Select.Option>
                        </Select>
                      </Form.Item>

                      <Form.Item
                        label='Language'
                        name='language'
                        rules={[
                          {
                            required: true,
                            message: 'Please select your language',
                          },
                        ]}
                      >
                        <Select>
                          <Select.Option value='en'>English</Select.Option>
                          <Select.Option value='zh'>中文</Select.Option>
                          <Select.Option value='es'>Español</Select.Option>
                        </Select>
                      </Form.Item>

                      <Form.Item
                        label='Date Format'
                        name='dateFormat'
                        rules={[
                          {
                            required: true,
                            message: 'Please select date format',
                          },
                        ]}
                      >
                        <Select>
                          <Select.Option value='YYYY-MM-DD'>
                            YYYY-MM-DD
                          </Select.Option>
                          <Select.Option value='MM/DD/YYYY'>
                            MM/DD/YYYY
                          </Select.Option>
                          <Select.Option value='DD/MM/YYYY'>
                            DD/MM/YYYY
                          </Select.Option>
                        </Select>
                      </Form.Item>

                      <Form.Item
                        label='Time Format'
                        name='timeFormat'
                        rules={[
                          {
                            required: true,
                            message: 'Please select time format',
                          },
                        ]}
                      >
                        <Select>
                          <Select.Option value='24h'>24-hour</Select.Option>
                          <Select.Option value='12h'>12-hour</Select.Option>
                        </Select>
                      </Form.Item>
                    </Col>
                  </Row>
                </Form>
              </Card>
            ),
          },

          {
            key: 'notifications',
            label: (
              <span>
                <BellOutlined />
                Notifications
              </span>
            ),
            children: (
              <Card
                title='Notification Settings'
                extra={
                  <Button
                    type='primary'
                    icon={<SaveOutlined />}
                    loading={loading}
                    onClick={() => notificationForm.submit()}
                  >
                    Save Changes
                  </Button>
                }
              >
                <Form
                  form={notificationForm}
                  layout='vertical'
                  onFinish={handleSaveNotifications}
                  initialValues={{
                    emailNotifications: true,
                    pushNotifications: false,
                    smsNotifications: false,
                    experimentCompleted: true,
                    experimentFailed: true,
                    systemAlerts: true,
                    weeklyReports: false,
                  }}
                >
                  <Alert
                    message='Notification Preferences'
                    description='Choose how you want to be notified about different events.'
                    type='info'
                    showIcon
                    style={{ marginBottom: 24 }}
                  />

                  <Title level={4}>Notification Channels</Title>
                  <Form.Item
                    label='Email Notifications'
                    name='emailNotifications'
                    valuePropName='checked'
                  >
                    <Switch />
                  </Form.Item>

                  <Form.Item
                    label='Push Notifications'
                    name='pushNotifications'
                    valuePropName='checked'
                  >
                    <Switch />
                  </Form.Item>

                  <Form.Item
                    label='SMS Notifications'
                    name='smsNotifications'
                    valuePropName='checked'
                  >
                    <Switch />
                  </Form.Item>

                  <Divider />

                  <Title level={4}>Event Notifications</Title>
                  <Form.Item
                    label='Experiment Completed'
                    name='experimentCompleted'
                    valuePropName='checked'
                  >
                    <Switch />
                  </Form.Item>

                  <Form.Item
                    label='Experiment Failed'
                    name='experimentFailed'
                    valuePropName='checked'
                  >
                    <Switch />
                  </Form.Item>

                  <Form.Item
                    label='System Alerts'
                    name='systemAlerts'
                    valuePropName='checked'
                  >
                    <Switch />
                  </Form.Item>

                  <Form.Item
                    label='Weekly Reports'
                    name='weeklyReports'
                    valuePropName='checked'
                  >
                    <Switch />
                  </Form.Item>
                </Form>
              </Card>
            ),
          },

          {
            key: 'security',
            label: (
              <span>
                <LockOutlined />
                Security
              </span>
            ),
            children: (
              <Card title='Security Settings'>
                <Form
                  form={securityForm}
                  layout='vertical'
                  onFinish={handleChangePassword}
                >
                  <Title level={4}>Change Password</Title>
                  <Row gutter={[24, 24]}>
                    <Col xs={24} lg={12}>
                      <Form.Item
                        label='Current Password'
                        name='oldPassword'
                        rules={[
                          {
                            required: true,
                            message: 'Please enter current password',
                          },
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
                          {
                            required: true,
                            message: 'Please enter new password',
                          },
                          {
                            min: 8,
                            message: 'Password must be at least 8 characters',
                          },
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
                          {
                            required: true,
                            message: 'Please confirm new password',
                          },
                          ({ getFieldValue }) => ({
                            validator(_, value) {
                              if (
                                !value ||
                                getFieldValue('newPassword') === value
                              ) {
                                return Promise.resolve();
                              }
                              return Promise.reject(
                                new Error('Passwords do not match')
                              );
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

                      <Form.Item>
                        <Button
                          type='primary'
                          icon={<SaveOutlined />}
                          loading={loading}
                          htmlType='submit'
                        >
                          Change Password
                        </Button>
                      </Form.Item>
                    </Col>

                    <Col xs={24} lg={12}>
                      <Alert
                        message='Password Requirements'
                        description='Your password must be at least 8 characters long.'
                        type='info'
                        showIcon
                      />
                    </Col>
                  </Row>
                </Form>
              </Card>
            ),
          },
        ]}
      />
    </div>
  );
};

export default Settings;
