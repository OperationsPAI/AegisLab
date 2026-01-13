import { useState } from 'react';

import {
  ClockCircleOutlined,
  CloudOutlined,
  DatabaseOutlined,
  GlobalOutlined,
  InfoCircleOutlined,
  LockOutlined,
  MailOutlined,
  NotificationOutlined,
  ReloadOutlined,
  SafetyOutlined,
  SaveOutlined,
  SettingOutlined,
  UserOutlined,
} from '@ant-design/icons';
import {
  Alert,
  Avatar,
  Button,
  Card,
  Col,
  Divider,
  Form,
  Input,
  InputNumber,
  List,
  message,
  Modal,
  Progress,
  Row,
  Select,
  Space,
  Statistic,
  Switch,
  Tabs,
  Tag,
  Typography,
} from 'antd';

const { Title, Text } = Typography;
const { TabPane } = Tabs;
const { Option } = Select;

const SystemSettings = () => {
  const [activeTab, setActiveTab] = useState('general');
  const [form] = Form.useForm();
  const [loading, setLoading] = useState(false);
  const [testEmailLoading, setTestEmailLoading] = useState(false);

  // Mock system statistics
  const systemStats = {
    totalUsers: 156,
    activeUsers: 142,
    totalProjects: 89,
    totalExperiments: 234,
    systemUptime: '99.9%',
    diskUsage: 68,
    memoryUsage: 45,
    cpuUsage: 23,
  };

  // Mock user list
  const users = [
    {
      id: 1,
      name: 'John Doe',
      email: 'john@example.com',
      role: 'Admin',
      status: 'active',
      lastLogin: '2024-01-15 10:30:00',
    },
    {
      id: 2,
      name: 'Jane Smith',
      email: 'jane@example.com',
      role: 'User',
      status: 'active',
      lastLogin: '2024-01-15 09:15:00',
    },
    {
      id: 3,
      name: 'Bob Johnson',
      email: 'bob@example.com',
      role: 'User',
      status: 'inactive',
      lastLogin: '2024-01-10 14:20:00',
    },
  ];

  const handleSaveSettings = async (values: Record<string, unknown>) => {
    setLoading(true);
    try {
      // TODO: Implement API call to save settings
      console.error('Saving settings:', values);
      message.success('Settings saved successfully');
    } catch (error) {
      message.error('Failed to save settings');
      console.error('Save settings error:', error);
    } finally {
      setLoading(false);
    }
  };

  const handleTestEmail = async () => {
    setTestEmailLoading(true);
    try {
      // TODO: Implement API call to test email configuration
      message.success('Test email sent successfully');
    } catch (error) {
      message.error('Failed to send test email');
      console.error('Test email error:', error);
    } finally {
      setTestEmailLoading(false);
    }
  };

  const handleUserAction = (_userId: number, action: string) => {
    Modal.confirm({
      title: `Confirm ${action}`,
      content: `Are you sure you want to ${action.toLowerCase()} this user?`,
      onOk: () => {
        message.success(`User ${action.toLowerCase()}d successfully`);
      },
    });
  };

  const getRoleColor = (role: string) => {
    switch (role.toLowerCase()) {
      case 'admin':
        return 'red';
      case 'user':
        return 'blue';
      default:
        return 'default';
    }
  };

  const getStatusColor = (status: string) => {
    switch (status.toLowerCase()) {
      case 'active':
        return 'green';
      case 'inactive':
        return 'orange';
      default:
        return 'default';
    }
  };

  return (
    <div style={{ padding: 24 }}>
      {/* Header */}
      <div style={{ marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0 }}>
          <GlobalOutlined style={{ marginRight: 8 }} />
          System Settings
        </Title>
        <Text type='secondary'>
          Configure system-wide settings and manage users
        </Text>
      </div>

      {/* System Overview */}
      <Card style={{ marginBottom: 24 }}>
        <Title level={4} style={{ marginBottom: 16 }}>
          System Overview
        </Title>
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12} md={6}>
            <Card size='small'>
              <Statistic
                title='Total Users'
                value={systemStats.totalUsers}
                prefix={<UserOutlined />}
                valueStyle={{ color: '#3b82f6' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card size='small'>
              <Statistic
                title='Active Users'
                value={systemStats.activeUsers}
                prefix={<UserOutlined />}
                valueStyle={{ color: '#10b981' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card size='small'>
              <Statistic
                title='System Uptime'
                value={systemStats.systemUptime}
                prefix={<GlobalOutlined />}
                valueStyle={{ color: '#10b981' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card size='small'>
              <Statistic
                title='Total Projects'
                value={systemStats.totalProjects}
                prefix={<DatabaseOutlined />}
                valueStyle={{ color: '#f59e0b' }}
              />
            </Card>
          </Col>
        </Row>

        <Divider />

        <Title level={4} style={{ marginBottom: 16 }}>
          Resource Usage
        </Title>
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={8}>
            <div style={{ marginBottom: 8 }}>
              <Text>Disk Usage</Text>
            </div>
            <Progress
              percent={systemStats.diskUsage}
              status={systemStats.diskUsage > 80 ? 'exception' : 'active'}
              format={(percent) => `${percent}%`}
            />
          </Col>
          <Col xs={24} sm={8}>
            <div style={{ marginBottom: 8 }}>
              <Text>Memory Usage</Text>
            </div>
            <Progress
              percent={systemStats.memoryUsage}
              status={systemStats.memoryUsage > 80 ? 'exception' : 'active'}
              format={(percent) => `${percent}%`}
            />
          </Col>
          <Col xs={24} sm={8}>
            <div style={{ marginBottom: 8 }}>
              <Text>CPU Usage</Text>
            </div>
            <Progress
              percent={systemStats.cpuUsage}
              status={systemStats.cpuUsage > 80 ? 'exception' : 'active'}
              format={(percent) => `${percent}%`}
            />
          </Col>
        </Row>
      </Card>

      {/* Settings Tabs */}
      <Tabs activeKey={activeTab} onChange={setActiveTab}>
        <TabPane
          tab={
            <span>
              <SettingOutlined />
              General Settings
            </span>
          }
          key='general'
        >
          <Card
            title='General Configuration'
            extra={
              <Button
                type='primary'
                icon={<SaveOutlined />}
                loading={loading}
                onClick={() => form.submit()}
              >
                Save Settings
              </Button>
            }
          >
            <Form
              form={form}
              layout='vertical'
              onFinish={handleSaveSettings}
              initialValues={{
                siteName: 'AegisLab RCABench',
                siteUrl: 'https://rcabench.example.com',
                timezone: 'UTC',
                dateFormat: 'YYYY-MM-DD',
                timeFormat: 'HH:mm:ss',
                enableRegistration: true,
                enablePublicProjects: true,
                maxFileSize: 100,
                maxProjectsPerUser: 10,
                sessionTimeout: 3600,
              }}
            >
              <Row gutter={[24, 24]}>
                <Col xs={24} lg={12}>
                  <Form.Item
                    label='Site Name'
                    name='siteName'
                    rules={[
                      { required: true, message: 'Please enter site name' },
                    ]}
                  >
                    <Input placeholder='Enter site name' />
                  </Form.Item>

                  <Form.Item
                    label='Site URL'
                    name='siteUrl'
                    rules={[
                      { required: true, message: 'Please enter site URL' },
                    ]}
                  >
                    <Input placeholder='https://example.com' />
                  </Form.Item>

                  <Form.Item
                    label='Timezone'
                    name='timezone'
                    rules={[
                      { required: true, message: 'Please select timezone' },
                    ]}
                  >
                    <Select placeholder='Select timezone'>
                      <Option value='UTC'>UTC</Option>
                      <Option value='America/New_York'>America/New_York</Option>
                      <Option value='Europe/London'>Europe/London</Option>
                      <Option value='Asia/Shanghai'>Asia/Shanghai</Option>
                    </Select>
                  </Form.Item>

                  <Form.Item
                    label='Date Format'
                    name='dateFormat'
                    rules={[
                      { required: true, message: 'Please select date format' },
                    ]}
                  >
                    <Select placeholder='Select date format'>
                      <Option value='YYYY-MM-DD'>YYYY-MM-DD</Option>
                      <Option value='MM/DD/YYYY'>MM/DD/YYYY</Option>
                      <Option value='DD/MM/YYYY'>DD/MM/YYYY</Option>
                    </Select>
                  </Form.Item>
                </Col>

                <Col xs={24} lg={12}>
                  <Form.Item
                    label='Time Format'
                    name='timeFormat'
                    rules={[
                      { required: true, message: 'Please select time format' },
                    ]}
                  >
                    <Select placeholder='Select time format'>
                      <Option value='HH:mm:ss'>24-hour (HH:mm:ss)</Option>
                      <Option value='hh:mm:ss A'>12-hour (hh:mm:ss A)</Option>
                    </Select>
                  </Form.Item>

                  <Form.Item
                    label='Max File Size (MB)'
                    name='maxFileSize'
                    rules={[
                      { required: true, message: 'Please enter max file size' },
                    ]}
                  >
                    <InputNumber min={1} max={1000} style={{ width: '100%' }} />
                  </Form.Item>

                  <Form.Item
                    label='Max Projects per User'
                    name='maxProjectsPerUser'
                    rules={[
                      { required: true, message: 'Please enter max projects' },
                    ]}
                  >
                    <InputNumber min={1} max={100} style={{ width: '100%' }} />
                  </Form.Item>

                  <Form.Item
                    label='Session Timeout (seconds)'
                    name='sessionTimeout'
                    rules={[
                      {
                        required: true,
                        message: 'Please enter session timeout',
                      },
                    ]}
                  >
                    <InputNumber
                      min={300}
                      max={86400}
                      step={300}
                      style={{ width: '100%' }}
                    />
                  </Form.Item>
                </Col>
              </Row>

              <Divider />

              <Row gutter={[24, 24]}>
                <Col xs={24} lg={12}>
                  <Form.Item
                    label='Enable User Registration'
                    name='enableRegistration'
                    valuePropName='checked'
                  >
                    <Switch />
                  </Form.Item>

                  <Form.Item
                    label='Enable Public Projects'
                    name='enablePublicProjects'
                    valuePropName='checked'
                  >
                    <Switch />
                  </Form.Item>
                </Col>

                <Col xs={24} lg={12}>
                  <Alert
                    message='Security Settings'
                    description='These settings affect the security and accessibility of your system. Please review carefully before making changes.'
                    type='warning'
                    showIcon
                    icon={<SafetyOutlined />}
                  />
                </Col>
              </Row>
            </Form>
          </Card>
        </TabPane>

        <TabPane
          tab={
            <span>
              <MailOutlined />
              Email Settings
            </span>
          }
          key='email'
        >
          <Card
            title='Email Configuration'
            extra={
              <Space>
                <Button
                  icon={<ReloadOutlined />}
                  onClick={handleTestEmail}
                  loading={testEmailLoading}
                >
                  Test Email
                </Button>
                <Button
                  type='primary'
                  icon={<SaveOutlined />}
                  loading={loading}
                  onClick={() => form.submit()}
                >
                  Save Settings
                </Button>
              </Space>
            }
          >
            <Form
              form={form}
              layout='vertical'
              onFinish={handleSaveSettings}
              initialValues={{
                smtpHost: 'smtp.gmail.com',
                smtpPort: 587,
                smtpUser: '',
                smtpPassword: '',
                smtpEncryption: 'tls',
                fromEmail: 'noreply@example.com',
                fromName: 'AegisLab RCABench',
                enableEmailNotifications: true,
              }}
            >
              <Row gutter={[24, 24]}>
                <Col xs={24} lg={12}>
                  <Form.Item
                    label='SMTP Host'
                    name='smtpHost'
                    rules={[
                      { required: true, message: 'Please enter SMTP host' },
                    ]}
                  >
                    <Input placeholder='smtp.gmail.com' />
                  </Form.Item>

                  <Form.Item
                    label='SMTP Port'
                    name='smtpPort'
                    rules={[
                      { required: true, message: 'Please enter SMTP port' },
                    ]}
                  >
                    <InputNumber
                      min={1}
                      max={65535}
                      style={{ width: '100%' }}
                    />
                  </Form.Item>

                  <Form.Item
                    label='SMTP User'
                    name='smtpUser'
                    rules={[
                      { required: true, message: 'Please enter SMTP user' },
                    ]}
                  >
                    <Input placeholder='your-email@gmail.com' />
                  </Form.Item>

                  <Form.Item
                    label='SMTP Password'
                    name='smtpPassword'
                    rules={[
                      { required: true, message: 'Please enter SMTP password' },
                    ]}
                  >
                    <Input.Password placeholder='Enter SMTP password' />
                  </Form.Item>
                </Col>

                <Col xs={24} lg={12}>
                  <Form.Item
                    label='Encryption Method'
                    name='smtpEncryption'
                    rules={[
                      {
                        required: true,
                        message: 'Please select encryption method',
                      },
                    ]}
                  >
                    <Select placeholder='Select encryption method'>
                      <Option value='none'>None</Option>
                      <Option value='tls'>TLS</Option>
                      <Option value='ssl'>SSL</Option>
                    </Select>
                  </Form.Item>

                  <Form.Item
                    label='From Email'
                    name='fromEmail'
                    rules={[
                      { required: true, message: 'Please enter from email' },
                    ]}
                  >
                    <Input placeholder='noreply@example.com' />
                  </Form.Item>

                  <Form.Item
                    label='From Name'
                    name='fromName'
                    rules={[
                      { required: true, message: 'Please enter from name' },
                    ]}
                  >
                    <Input placeholder='AegisLab RCABench' />
                  </Form.Item>

                  <Form.Item
                    label='Enable Email Notifications'
                    name='enableEmailNotifications'
                    valuePropName='checked'
                  >
                    <Switch />
                  </Form.Item>
                </Col>
              </Row>
            </Form>
          </Card>
        </TabPane>

        <TabPane
          tab={
            <span>
              <UserOutlined />
              User Management
            </span>
          }
          key='users'
        >
          <Card title='User Management'>
            <List
              itemLayout='horizontal'
              dataSource={users}
              renderItem={(item) => (
                <List.Item
                  key={item.id}
                  actions={[
                    <Button
                      key='edit'
                      type='link'
                      size='small'
                      onClick={() => handleUserAction(item.id, 'Edit')}
                    >
                      Edit
                    </Button>,
                    item.status === 'active' ? (
                      <Button
                        key='deactivate'
                        type='link'
                        danger
                        size='small'
                        onClick={() => handleUserAction(item.id, 'Deactivate')}
                      >
                        Deactivate
                      </Button>
                    ) : (
                      <Button
                        key='activate'
                        type='link'
                        size='small'
                        onClick={() => handleUserAction(item.id, 'Activate')}
                      >
                        Activate
                      </Button>
                    ),
                    <Button
                      key='delete'
                      type='link'
                      danger
                      size='small'
                      onClick={() => handleUserAction(item.id, 'Delete')}
                    >
                      Delete
                    </Button>,
                  ]}
                >
                  <List.Item.Meta
                    avatar={<Avatar icon={<UserOutlined />} />}
                    title={
                      <Space>
                        <Text strong>{item.name}</Text>
                        <Tag color={getRoleColor(item.role)}>{item.role}</Tag>
                        <Tag color={getStatusColor(item.status)}>
                          {item.status}
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
                            Last login: {item.lastLogin}
                          </Text>
                        </Space>
                      </Space>
                    }
                  />
                </List.Item>
              )}
            />
          </Card>
        </TabPane>

        <TabPane
          tab={
            <span>
              <CloudOutlined />
              Integration Settings
            </span>
          }
          key='integration'
        >
          <Card title='Third-party Integrations'>
            <Alert
              message='Integration Settings'
              description='Configure integrations with external services and APIs.'
              type='info'
              showIcon
              icon={<InfoCircleOutlined />}
              style={{ marginBottom: 24 }}
            />

            <Row gutter={[24, 24]}>
              <Col xs={24} lg={12}>
                <Card
                  title='Kubernetes Integration'
                  extra={<Switch defaultChecked />}
                >
                  <Form.Item label='Cluster URL'>
                    <Input placeholder='https://kubernetes.example.com' />
                  </Form.Item>
                  <Form.Item label='Namespace'>
                    <Input placeholder='default' />
                  </Form.Item>
                  <Form.Item label='Service Account'>
                    <Input placeholder='rcabench-sa' />
                  </Form.Item>
                </Card>
              </Col>

              <Col xs={24} lg={12}>
                <Card title='Docker Registry' extra={<Switch defaultChecked />}>
                  <Form.Item label='Registry URL'>
                    <Input placeholder='https://registry.example.com' />
                  </Form.Item>
                  <Form.Item label='Username'>
                    <Input placeholder='username' />
                  </Form.Item>
                  <Form.Item label='Password'>
                    <Input.Password placeholder='password' />
                  </Form.Item>
                </Card>
              </Col>

              <Col xs={24} lg={12}>
                <Card
                  title='Monitoring & Logging'
                  extra={<Switch defaultChecked />}
                >
                  <Form.Item label='Prometheus URL'>
                    <Input placeholder='http://prometheus:9090' />
                  </Form.Item>
                  <Form.Item label='Grafana URL'>
                    <Input placeholder='http://grafana:3000' />
                  </Form.Item>
                  <Form.Item label='Jaeger URL'>
                    <Input placeholder='http://jaeger:16686' />
                  </Form.Item>
                </Card>
              </Col>

              <Col xs={24} lg={12}>
                <Card title='External Storage' extra={<Switch />}>
                  <Form.Item label='Storage Type'>
                    <Select placeholder='Select storage type'>
                      <Option value='s3'>Amazon S3</Option>
                      <Option value='gcs'>Google Cloud Storage</Option>
                      <Option value='azure'>Azure Blob Storage</Option>
                      <Option value='minio'>MinIO</Option>
                    </Select>
                  </Form.Item>
                  <Form.Item label='Bucket Name'>
                    <Input placeholder='my-bucket' />
                  </Form.Item>
                  <Form.Item label='Access Key'>
                    <Input placeholder='access key' />
                  </Form.Item>
                  <Form.Item label='Secret Key'>
                    <Input.Password placeholder='secret key' />
                  </Form.Item>
                </Card>
              </Col>
            </Row>
          </Card>
        </TabPane>

        <TabPane
          tab={
            <span>
              <NotificationOutlined />
              Notification Settings
            </span>
          }
          key='notifications'
        >
          <Card title='Notification Configuration'>
            <Form
              layout='vertical'
              initialValues={{
                enableSystemNotifications: true,
                enableEmailNotifications: true,
                enableSlackNotifications: false,
                enableWebhookNotifications: false,
                notificationEmail: 'admin@example.com',
                slackWebhook: '',
                webhookUrl: '',
              }}
            >
              <Row gutter={[24, 24]}>
                <Col xs={24} lg={12}>
                  <Form.Item
                    label='Enable System Notifications'
                    name='enableSystemNotifications'
                    valuePropName='checked'
                  >
                    <Switch />
                  </Form.Item>

                  <Form.Item
                    label='Enable Email Notifications'
                    name='enableEmailNotifications'
                    valuePropName='checked'
                  >
                    <Switch />
                  </Form.Item>

                  <Form.Item
                    label='Enable Slack Notifications'
                    name='enableSlackNotifications'
                    valuePropName='checked'
                  >
                    <Switch />
                  </Form.Item>

                  <Form.Item
                    label='Enable Webhook Notifications'
                    name='enableWebhookNotifications'
                    valuePropName='checked'
                  >
                    <Switch />
                  </Form.Item>
                </Col>

                <Col xs={24} lg={12}>
                  <Form.Item
                    label='Notification Email'
                    name='notificationEmail'
                    rules={[
                      { type: 'email', message: 'Please enter valid email' },
                    ]}
                  >
                    <Input
                      placeholder='admin@example.com'
                      prefix={<MailOutlined />}
                    />
                  </Form.Item>

                  <Form.Item label='Slack Webhook URL' name='slackWebhook'>
                    <Input placeholder='https://hooks.slack.com/services/...' />
                  </Form.Item>

                  <Form.Item label='Webhook URL' name='webhookUrl'>
                    <Input placeholder='https://your-webhook-url.com' />
                  </Form.Item>
                </Col>
              </Row>

              <Divider />

              <Title level={5}>Notification Events</Title>
              <Form.Item
                label='Experiment Completed'
                name='notifyExperimentCompleted'
                valuePropName='checked'
              >
                <Switch />
              </Form.Item>

              <Form.Item
                label='Experiment Failed'
                name='notifyExperimentFailed'
                valuePropName='checked'
              >
                <Switch />
              </Form.Item>

              <Form.Item
                label='System Error'
                name='notifySystemError'
                valuePropName='checked'
              >
                <Switch />
              </Form.Item>

              <Form.Item
                label='User Registration'
                name='notifyUserRegistration'
                valuePropName='checked'
              >
                <Switch />
              </Form.Item>
            </Form>
          </Card>
        </TabPane>

        <TabPane
          tab={
            <span>
              <SafetyOutlined />
              Security Settings
            </span>
          }
          key='security'
        >
          <Card title='Security Configuration'>
            <Form
              layout='vertical'
              initialValues={{
                enableTwoFactorAuth: false,
                enableCaptcha: true,
                maxLoginAttempts: 5,
                lockoutDuration: 900,
                passwordMinLength: 8,
                passwordRequireUppercase: true,
                passwordRequireLowercase: true,
                passwordRequireNumbers: true,
                passwordRequireSpecialChars: true,
                sessionTimeout: 3600,
                enableAuditLogging: true,
              }}
            >
              <Row gutter={[24, 24]}>
                <Col xs={24} lg={12}>
                  <Title level={5}>Authentication</Title>
                  <Form.Item
                    label='Enable Two-Factor Authentication'
                    name='enableTwoFactorAuth'
                    valuePropName='checked'
                  >
                    <Switch />
                  </Form.Item>

                  <Form.Item
                    label='Enable CAPTCHA'
                    name='enableCaptcha'
                    valuePropName='checked'
                  >
                    <Switch />
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

                  <Form.Item
                    label='Lockout Duration (seconds)'
                    name='lockoutDuration'
                    rules={[
                      {
                        required: true,
                        message: 'Please enter lockout duration',
                      },
                    ]}
                  >
                    <InputNumber
                      min={60}
                      max={3600}
                      step={60}
                      style={{ width: '100%' }}
                    />
                  </Form.Item>
                </Col>

                <Col xs={24} lg={12}>
                  <Title level={5}>Password Policy</Title>
                  <Form.Item
                    label='Minimum Password Length'
                    name='passwordMinLength'
                    rules={[
                      {
                        required: true,
                        message: 'Please enter minimum password length',
                      },
                    ]}
                  >
                    <InputNumber min={6} max={20} style={{ width: '100%' }} />
                  </Form.Item>

                  <Form.Item
                    label='Require Uppercase Letters'
                    name='passwordRequireUppercase'
                    valuePropName='checked'
                  >
                    <Switch />
                  </Form.Item>

                  <Form.Item
                    label='Require Lowercase Letters'
                    name='passwordRequireLowercase'
                    valuePropName='checked'
                  >
                    <Switch />
                  </Form.Item>

                  <Form.Item
                    label='Require Numbers'
                    name='passwordRequireNumbers'
                    valuePropName='checked'
                  >
                    <Switch />
                  </Form.Item>

                  <Form.Item
                    label='Require Special Characters'
                    name='passwordRequireSpecialChars'
                    valuePropName='checked'
                  >
                    <Switch />
                  </Form.Item>
                </Col>
              </Row>

              <Divider />

              <Row gutter={[24, 24]}>
                <Col xs={24} lg={12}>
                  <Title level={5}>Session Management</Title>
                  <Form.Item
                    label='Session Timeout (seconds)'
                    name='sessionTimeout'
                    rules={[
                      {
                        required: true,
                        message: 'Please enter session timeout',
                      },
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
                    label='Enable Audit Logging'
                    name='enableAuditLogging'
                    valuePropName='checked'
                  >
                    <Switch />
                  </Form.Item>
                </Col>

                <Col xs={24} lg={12}>
                  <Alert
                    message='Security Recommendations'
                    description='Enable two-factor authentication for enhanced security. Use strong password policies and regularly review audit logs.'
                    type='info'
                    showIcon
                    icon={<LockOutlined />}
                  />
                </Col>
              </Row>

              <Form.Item>
                <Button
                  type='primary'
                  icon={<SaveOutlined />}
                  loading={loading}
                  onClick={() => form.submit()}
                >
                  Save Security Settings
                </Button>
              </Form.Item>
            </Form>
          </Card>
        </TabPane>
      </Tabs>
    </div>
  );
};

export default SystemSettings;
