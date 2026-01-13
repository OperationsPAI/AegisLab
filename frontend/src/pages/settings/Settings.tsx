
import {
  ApiOutlined,
  BellOutlined,
  EyeInvisibleOutlined,
  EyeOutlined,
  GlobalOutlined,
  HistoryOutlined,
  KeyOutlined,
  LockOutlined,
  MailOutlined,
  PhoneOutlined,
  SaveOutlined,
  SecurityScanOutlined,
  SettingOutlined,
  UserOutlined,
} from '@ant-design/icons';
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
  InputNumber,
  List,
  message,
  Modal,
  Row,
  Select,
  Space,
  Switch,
  Tabs,
  Tag,
  Timeline,
  Typography,
} from 'antd';
import dayjs from 'dayjs';
import { useState } from 'react';

const { Title, Text } = Typography;
// Removed deprecated TabPane destructuring - using items prop instead
const { TextArea } = Input;

const Settings = () => {
  const [activeTab, setActiveTab] = useState('profile');
  const [profileForm] = Form.useForm();
  const [notificationForm] = Form.useForm();
  const [securityForm] = Form.useForm();
  const [apiForm] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [apiKeyModalVisible, setApiKeyModalVisible] = useState(false);

  // Mock user data
  const userData = {
    id: 1,
    username: 'john_doe',
    email: 'john.doe@example.com',
    fullName: 'John Doe',
    phone: '+1 (555) 123-4567',
    department: 'Engineering',
    role: 'Senior Engineer',
    createdAt: '2023-01-15T10:30:00Z',
    lastLoginAt: '2024-01-15T09:30:00Z',
    avatar: null,
  };

  // Mock API keys
  interface ApiKey {
    id: number;
    name: string;
    key: string;
    createdAt: string;
    lastUsed: string | null;
    status: string;
    permissions: string[];
  }

  const [apiKeys, setApiKeys] = useState<ApiKey[]>([
    {
      id: 1,
      name: 'Production API Key',
      key: 'ak_prod_1234567890abcdef',
      createdAt: '2024-01-01T10:00:00Z',
      lastUsed: '2024-01-15T08:30:00Z',
      status: 'active',
      permissions: ['read', 'write'],
    },
    {
      id: 2,
      name: 'Development API Key',
      key: 'ak_dev_abcdef1234567890',
      createdAt: '2024-01-05T14:30:00Z',
      lastUsed: '2024-01-14T16:45:00Z',
      status: 'active',
      permissions: ['read'],
    },
  ]);

  // Mock recent activity
  const recentActivity = [
    {
      id: 1,
      action: 'Password changed',
      timestamp: '2024-01-15T09:00:00Z',
      type: 'security',
    },
    {
      id: 2,
      action: 'API key created',
      timestamp: '2024-01-14T16:30:00Z',
      type: 'api',
    },
    {
      id: 3,
      action: 'Login from new device',
      timestamp: '2024-01-13T14:20:00Z',
      type: 'login',
    },
    {
      id: 4,
      action: 'Email notification settings updated',
      timestamp: '2024-01-12T11:15:00Z',
      type: 'settings',
    },
  ];

  const handleSaveProfile = async (values: Record<string, unknown>) => {
    setLoading(true);
    try {
      // TODO: Implement API call to update profile
      console.error('Updating profile:', values);
      message.success('Profile updated successfully');
    } catch (error) {
      message.error('Failed to update profile');
      console.error('Update profile error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleSaveNotifications = async (values: Record<string, unknown>) => {
    setLoading(true);
    try {
      // TODO: Implement API call to update notification settings
      console.error('Updating notifications:', values);
      message.success('Notification settings updated successfully');
    } catch (error) {
      message.error('Failed to update notification settings');
      console.error('Update notifications error:', error);
    } finally {
      setLoading(false);
    }
  };


  const handleChangePassword = async (values: Record<string, unknown>) => {
    setLoading(true);
    try {
      // TODO: Implement API call to change password
      console.error('Changing password:', values);
      message.success('Password changed successfully');
      securityForm.resetFields();
    } catch (error) {
      message.error('Failed to change password');
      console.error('Change password error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateApiKey = async (values: Record<string, string | string[]>) => {
    setLoading(true);
    try {
      // Generate a mock API key
      const newKey = {
        id: apiKeys.length + 1,
        name: values.name as string,
        key: `ak_${values.environment}_${Math.random().toString(36).substring(2, 18)}`,
        createdAt: new Date().toISOString(),
        lastUsed: null as string | null,
        status: 'active',
        permissions: values.permissions as string[],
      };
      setApiKeys([...apiKeys, newKey]);
      message.success('API key created successfully');
      setApiKeyModalVisible(false);
      apiForm.resetFields();
    } catch (error) {
      message.error('Failed to create API key');
      console.error('Create API key error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleDeleteApiKey = (keyId: number) => {
    Modal.confirm({
      title: 'Delete API Key',
      content:
        'Are you sure you want to delete this API key? This action cannot be undone.',
      okText: 'Yes, Delete',
      okButtonProps: { danger: true },
      cancelText: 'Cancel',
      onOk: () => {
        setApiKeys(apiKeys.filter((key) => key.id !== keyId));
        message.success('API key deleted successfully');
      },
    });
  };

  const handleRevokeApiKey = (keyId: number) => {
    Modal.confirm({
      title: 'Revoke API Key',
      content: 'Are you sure you want to revoke this API key?',
      okText: 'Yes, Revoke',
      okButtonProps: { danger: true },
      cancelText: 'Cancel',
      onOk: () => {
        setApiKeys(
          apiKeys.map((key) =>
            key.id === keyId ? { ...key, status: 'revoked' } : key
          )
        );
        message.success('API key revoked successfully');
      },
    });
  };

  const getActivityColor = (type: string) => {
    switch (type) {
      case 'security':
        return 'red';
      case 'api':
        return 'blue';
      case 'login':
        return 'green';
      case 'settings':
        return 'orange';
      default:
        return 'default';
    }
  };

  const getActivityIcon = (type: string) => {
    switch (type) {
      case 'security':
        return <SecurityScanOutlined />;
      case 'api':
        return <ApiOutlined />;
      case 'login':
        return <KeyOutlined />;
      case 'settings':
        return <SettingOutlined />;
      default:
        return <GlobalOutlined />;
    }
  };

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
                  {userData.fullName}
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
                  {userData.phone}
                </Space>
              </Descriptions.Item>
              <Descriptions.Item label='Role'>
                <Tag color='blue'>{userData.role}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label='Department'>
                {userData.department}
              </Descriptions.Item>
              <Descriptions.Item label='Member Since'>
                {dayjs(userData.createdAt).format('MMMM D, YYYY')}
              </Descriptions.Item>
              <Descriptions.Item label='Last Login'>
                {dayjs(userData.lastLoginAt).format('MMMM D, YYYY HH:mm')}
              </Descriptions.Item>
            </Descriptions>
          </Col>
        </Row>
      </Card>

      {/* Settings Tabs */}
      <Tabs activeKey={activeTab} onChange={setActiveTab} items={[
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
                fullName: userData.fullName,
                email: userData.email,
                phone: userData.phone,
                department: userData.department,
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
                    name='fullName'
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
                      { required: true, message: 'Please enter your email' },
                      { type: 'email', message: 'Please enter a valid email' },
                    ]}
                  >
                    <Input prefix={<MailOutlined />} />
                  </Form.Item>

                  <Form.Item label='Phone' name='phone'>
                    <Input prefix={<PhoneOutlined />} />
                  </Form.Item>

                  <Form.Item label='Department' name='department'>
                    <Input />
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
                      { required: true, message: 'Please select date format' },
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
                      { required: true, message: 'Please select time format' },
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
    )
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
                marketingEmails: false,
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

              <Form.Item
                label='Marketing Emails'
                name='marketingEmails'
                valuePropName='checked'
              >
                <Switch />
              </Form.Item>
            </Form>
          </Card>
    )
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
              initialValues={{
                twoFactorAuth: true,
                loginAlerts: true,
                sessionTimeout: 3600,
                maxLoginAttempts: 5,
              }}
            >
              <Alert
                message='Security Recommendations'
                description='Enable two-factor authentication and login alerts for enhanced security.'
                type='warning'
                showIcon
                style={{ marginBottom: 24 }}
              />

              <Title level={4}>Authentication</Title>
              <Form.Item
                label='Two-Factor Authentication'
                name='twoFactorAuth'
                valuePropName='checked'
              >
                <Switch />
              </Form.Item>

              <Form.Item
                label='Login Alerts'
                name='loginAlerts'
                valuePropName='checked'
              >
                <Switch />
              </Form.Item>

              <Form.Item
                label='Session Timeout (seconds)'
                name='sessionTimeout'
                rules={[
                  { required: true, message: 'Please enter session timeout' },
                ]}
              >
                <InputNumber
                  min={300}
                  max={86400}
                  step={300}
                  style={{ width: '100%' }}
                />
              </Form.Item>

              <Form.Item
                label='Max Login Attempts'
                name='maxLoginAttempts'
                rules={[
                  {
                    required: true,
                    message: 'Please enter max login attempts',
                  },
                ]}
              >
                <InputNumber min={1} max={10} style={{ width: '100%' }} />
              </Form.Item>

              <Divider />

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
                      { required: true, message: 'Please enter new password' },
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
                    description='Your password must be at least 8 characters long and contain uppercase letters, lowercase letters, numbers, and special characters.'
                    type='info'
                    showIcon
                  />
                </Col>
              </Row>
            </Form>
          </Card>
    )
  },

        {
          key: 'api',
          label: (
            <span>
              <KeyOutlined />
              API Keys
            </span>
          ),
          children: (
          <Card
            title='API Keys'
            extra={
              <Button
                type='primary'
                onClick={() => setApiKeyModalVisible(true)}
              >
                Create API Key
              </Button>
            }
          >
            <List
              itemLayout='horizontal'
              dataSource={apiKeys}
              renderItem={(item) => (
                <List.Item
                  key={item.id}
                  actions={[
                    <Button
                      key='copy'
                      type='link'
                      size='small'
                      onClick={() => {
                        navigator.clipboard.writeText(item.key);
                        message.success('API key copied to clipboard');
                      }}
                    >
                      Copy
                    </Button>,
                    item.status === 'active' ? (
                      <Button
                        key='revoke'
                        type='link'
                        danger
                        size='small'
                        onClick={() => handleRevokeApiKey(item.id)}
                      >
                        Revoke
                      </Button>
                    ) : null,
                    <Button
                      key='delete'
                      type='link'
                      danger
                      size='small'
                      onClick={() => handleDeleteApiKey(item.id)}
                    >
                      Delete
                    </Button>,
                  ]}
                >
                  <List.Item.Meta
                    avatar={<Avatar icon={<KeyOutlined />} />}
                    title={
                      <Space>
                        <Text strong>{item.name}</Text>
                        <Tag color={item.status === 'active' ? 'green' : 'red'}>
                          {item.status.toUpperCase()}
                        </Tag>
                      </Space>
                    }
                    description={
                      <Space direction='vertical' size={0}>
                        <Text code>{item.key}</Text>
                        <Space>
                          <Text type='secondary'>
                            Created:{' '}
                            {dayjs(item.createdAt).format('MMM D, YYYY')}
                          </Text>
                          <Text type='secondary'>
                            Last used:{' '}
                            {item.lastUsed
                              ? dayjs(item.lastUsed).format('MMM D, YYYY HH:mm')
                              : 'Never'}
                          </Text>
                        </Space>
                        <Space>
                          {item.permissions.map((permission) => (
                            <Tag key={permission}>{permission}</Tag>
                          ))}
                        </Space>
                      </Space>
                    }
                  />
                </List.Item>
              )}
            />
          </Card>
    )
  },

        {
          key: 'activity',
          label: (
            <span>
              <HistoryOutlined />
              Activity
            </span>
          ),
          children: (
          <Card title='Recent Activity'>
            <Timeline>
              {recentActivity.map((activity) => (
                <Timeline.Item
                  key={activity.id}
                  color={getActivityColor(activity.type)}
                  dot={getActivityIcon(activity.type)}
                >
                  <Space direction='vertical' size={0}>
                    <Text strong>{activity.action}</Text>
                    <Text type='secondary'>
                      {dayjs(activity.timestamp).format('MMMM D, YYYY HH:mm')}
                    </Text>
                  </Space>
                </Timeline.Item>
              ))}
            </Timeline>
          </Card>
    )
  }
]} />

      {/* Create API Key Modal */}
      <Modal
        title='Create API Key'
        open={apiKeyModalVisible}
        onCancel={() => {
          setApiKeyModalVisible(false);
          apiForm.resetFields();
        }}
        footer={[
          <Button
            key='cancel'
            onClick={() => {
              setApiKeyModalVisible(false);
              apiForm.resetFields();
            }}
          >
            Cancel
          </Button>,
          <Button
            key='submit'
            type='primary'
            loading={loading}
            onClick={() => apiForm.submit()}
          >
            Create Key
          </Button>,
        ]}
      >
        <Form
          form={apiForm}
          layout='vertical'
          onFinish={handleCreateApiKey}
          initialValues={{
            environment: 'development',
            permissions: ['read'],
          }}
        >
          <Form.Item
            label='Key Name'
            name='name'
            rules={[
              {
                required: true,
                message: 'Please enter a name for the API key',
              },
            ]}
          >
            <Input placeholder='e.g., Production API Key' />
          </Form.Item>

          <Form.Item
            label='Environment'
            name='environment'
            rules={[
              { required: true, message: 'Please select an environment' },
            ]}
          >
            <Select>
              <Select.Option value='development'>Development</Select.Option>
              <Select.Option value='staging'>Staging</Select.Option>
              <Select.Option value='production'>Production</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item
            label='Permissions'
            name='permissions'
            rules={[{ required: true, message: 'Please select permissions' }]}
          >
            <Select mode='multiple' placeholder='Select permissions'>
              <Select.Option value='read'>Read</Select.Option>
              <Select.Option value='write'>Write</Select.Option>
              <Select.Option value='delete'>Delete</Select.Option>
              <Select.Option value='admin'>Admin</Select.Option>
            </Select>
          </Form.Item>

          <Form.Item label='Expiration (days)' name='expiration'>
            <InputNumber
              min={1}
              max={365}
              placeholder='Leave empty for no expiration'
              style={{ width: '100%' }}
            />
          </Form.Item>

          <Form.Item label='Description' name='description'>
            <TextArea
              rows={3}
              placeholder='Optional description of the API key usage'
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  );
};

export default Settings;
