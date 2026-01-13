import {
  UserOutlined,
  MailOutlined,
  PhoneOutlined,
  EditOutlined,
  SaveOutlined,
  CloseOutlined,
  KeyOutlined,
  HistoryOutlined,
  GlobalOutlined,
  EyeOutlined,
  EyeInvisibleOutlined,
} from '@ant-design/icons'
import {
  Card,
  Form,
  Input,
  Button,
  Space,
  Typography,
  Row,
  Col,
  Avatar,
  Descriptions,
  Divider,
  Modal,
  message,
  Tabs,
  Timeline,
  Tag,
  Statistic,
  Progress,
  Switch,
} from 'antd'
import dayjs from 'dayjs'
import { useState } from 'react'

const { Title, Text } = Typography
const { TabPane } = Tabs

const UserProfile = () => {
  const [isEditing, setIsEditing] = useState(false)
  const [form] = Form.useForm()
  const [passwordModalVisible, setPasswordModalVisible] = useState(false)
  const [passwordForm] = Form.useForm()
  const [showOldPassword, setShowOldPassword] = useState(false)
  const [showNewPassword, setShowNewPassword] = useState(false)
  const [showConfirmPassword, setShowConfirmPassword] = useState(false)

  // Mock user data
  const userData = {
    id: 1,
    username: 'john_doe',
    email: 'john.doe@example.com',
    fullName: 'John Doe',
    phone: '+1 (555) 123-4567',
    avatar: null,
    role: 'Administrator',
    department: 'Engineering',
    createdAt: '2023-01-15T10:30:00Z',
    lastLoginAt: '2024-01-15T09:30:00Z',
    status: 'active',
    twoFactorEnabled: true,
    emailVerified: true,
    phoneVerified: false,
  }

  // Mock activity data
  const recentActivity = [
    {
      id: 1,
      action: 'Created new project',
      description: 'Project: Microservice RCA Analysis',
      timestamp: '2024-01-15T14:30:00Z',
      type: 'project',
    },
    {
      id: 2,
      action: 'Ran fault injection experiment',
      description: 'Experiment #123 on service payment-service',
      timestamp: '2024-01-15T13:15:00Z',
      type: 'experiment',
    },
    {
      id: 3,
      action: 'Uploaded dataset',
      description: 'Dataset: Production Traces Q4 2023',
      timestamp: '2024-01-15T11:45:00Z',
      type: 'dataset',
    },
    {
      id: 4,
      action: 'Algorithm execution completed',
      description: 'MicroRank algorithm on datapack dp-789012',
      timestamp: '2024-01-15T10:20:00Z',
      type: 'execution',
    },
  ]

  // Mock statistics
  const userStats = {
    totalProjects: 12,
    totalExperiments: 45,
    totalDatasets: 8,
    successRate: 87,
    avgExperimentDuration: '23m 45s',
    last30DaysActivity: 28,
  }

  const handleEditProfile = () => {
    setIsEditing(true)
    form.setFieldsValue({
      fullName: userData.fullName,
      email: userData.email,
      phone: userData.phone,
    })
  }

  const handleSaveProfile = async (values: Record<string, unknown>) => {
    try {
      // TODO: Implement API call to update profile
      console.log('Updating profile:', values)
      message.success('Profile updated successfully')
      setIsEditing(false)
    } catch (error) {
      message.error('Failed to update profile')
      console.error('Update profile error:', error)
    }
  }

  const handleCancelEdit = () => {
    setIsEditing(false)
    form.resetFields()
  }

  const handleChangePassword = async (values: Record<string, unknown>) => {
    try {
      // TODO: Implement API call to change password
      console.log('Changing password:', values)
      message.success('Password changed successfully')
      setPasswordModalVisible(false)
      passwordForm.resetFields()
    } catch (error) {
      message.error('Failed to change password')
      console.error('Change password error:', error)
    }
  }

  const getActivityColor = (type: string) => {
    switch (type) {
      case 'project':
        return '#3b82f6'
      case 'experiment':
        return '#10b981'
      case 'dataset':
        return '#f59e0b'
      case 'execution':
        return '#8b5cf6'
      default:
        return '#6b7280'
    }
  }

  const getActivityIcon = (type: string) => {
    switch (type) {
      case 'project':
        return '🔧'
      case 'experiment':
        return '🧪'
      case 'dataset':
        return '📊'
      case 'execution':
        return '⚡'
      default:
        return '📝'
    }
  }

  return (
    <div style={{ padding: 24 }}>
      {/* Header */}
      <div style={{ marginBottom: 24 }}>
        <Title level={2} style={{ margin: 0 }}>
          <UserOutlined style={{ marginRight: 8 }} />
          User Profile
        </Title>
        <Text type="secondary">
          Manage your profile information and account settings
        </Text>
      </div>

      {/* Profile Overview */}
      <Card style={{ marginBottom: 24 }}>
        <Row gutter={[24, 24]} align="middle">
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
                  {userData.fullName}
                </Title>
                <Text type="secondary">@{userData.username}</Text>
              </div>
            </div>
          </Col>
          <Col xs={24} sm={18} md={20}>
            <Descriptions
              title="Profile Information"
              bordered
              column={{ xs: 1, sm: 2, md: 3 }}
              extra={
                !isEditing && (
                  <Button type="primary" icon={<EditOutlined />} onClick={handleEditProfile}>
                    Edit Profile
                  </Button>
                )
              }
            >
              <Descriptions.Item label="Full Name">
                {isEditing ? (
                  <Form form={form} layout="inline" onFinish={handleSaveProfile}>
                    <Form.Item name="fullName" style={{ margin: 0 }}>
                      <Input />
                    </Form.Item>
                  </Form>
                ) : (
                  userData.fullName
                )}
              </Descriptions.Item>
              <Descriptions.Item label="Email">
                {isEditing ? (
                  <Form form={form} layout="inline">
                    <Form.Item name="email" style={{ margin: 0 }}>
                      <Input />
                    </Form.Item>
                  </Form>
                ) : (
                  <Space>
                    {userData.email}
                    {userData.emailVerified && (
                      <Tag color="green" icon={<GlobalOutlined />}>
                        Verified
                      </Tag>
                    )}
                  </Space>
                )}
              </Descriptions.Item>
              <Descriptions.Item label="Phone">
                {isEditing ? (
                  <Form form={form} layout="inline">
                    <Form.Item name="phone" style={{ margin: 0 }}>
                      <Input />
                    </Form.Item>
                  </Form>
                ) : (
                  <Space>
                    {userData.phone}
                    {userData.phoneVerified ? (
                      <Tag color="green" icon={<PhoneOutlined />}>
                        Verified
                      </Tag>
                    ) : (
                      <Tag color="orange">Not Verified</Tag>
                    )}
                  </Space>
                )}
              </Descriptions.Item>
              <Descriptions.Item label="Role">
                <Tag color="blue">{userData.role}</Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Department">
                {userData.department}
              </Descriptions.Item>
              <Descriptions.Item label="Status">
                <Tag color={userData.status === 'active' ? 'green' : 'orange'}>
                  {userData.status.toUpperCase()}
                </Tag>
              </Descriptions.Item>
              <Descriptions.Item label="Member Since">
                {dayjs(userData.createdAt).format('MMMM D, YYYY')}
              </Descriptions.Item>
              <Descriptions.Item label="Last Login">
                {dayjs(userData.lastLoginAt).format('MMMM D, YYYY HH:mm')}
              </Descriptions.Item>
              <Descriptions.Item label="Two-Factor Auth">
                <Switch
                  checked={userData.twoFactorEnabled}
                  checkedChildren="Enabled"
                  unCheckedChildren="Disabled"
                  onChange={(checked) => {
                    // TODO: Implement 2FA toggle
                    message.info(`2FA ${checked ? 'enabled' : 'disabled'}`)
                  }}
                />
              </Descriptions.Item>
            </Descriptions>

            {isEditing && (
              <div style={{ marginTop: 16, textAlign: 'right' }}>
                <Space>
                  <Button icon={<CloseOutlined />} onClick={handleCancelEdit}>
                    Cancel
                  </Button>
                  <Button
                    type="primary"
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

      {/* Statistics */}
      <Card style={{ marginBottom: 24 }}>
        <Title level={4} style={{ marginBottom: 16 }}>
          Activity Statistics
        </Title>
        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12} md={6}>
            <Card size="small">
              <Statistic
                title="Total Projects"
                value={userStats.totalProjects}
                prefix={<span style={{ fontSize: 20 }}>📁</span>}
                valueStyle={{ color: '#3b82f6' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card size="small">
              <Statistic
                title="Total Experiments"
                value={userStats.totalExperiments}
                prefix={<span style={{ fontSize: 20 }}>🧪</span>}
                valueStyle={{ color: '#10b981' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card size="small">
              <Statistic
                title="Total Datasets"
                value={userStats.totalDatasets}
                prefix={<span style={{ fontSize: 20 }}>📊</span>}
                valueStyle={{ color: '#f59e0b' }}
              />
            </Card>
          </Col>
          <Col xs={24} sm={12} md={6}>
            <Card size="small">
              <Statistic
                title="Success Rate"
                value={userStats.successRate}
                suffix="%"
                prefix={<span style={{ fontSize: 20 }}>✅</span>}
                valueStyle={{ color: '#10b981' }}
              />
            </Card>
          </Col>
        </Row>

        <Divider />

        <Row gutter={[16, 16]}>
          <Col xs={24} sm={12} md={12}>
            <div style={{ marginBottom: 8 }}>
              <Text>Experiment Success Rate</Text>
            </div>
            <Progress
              percent={userStats.successRate}
              status={userStats.successRate > 80 ? 'success' : 'active'}
              format={percent => `${percent}%`}
            />
          </Col>
          <Col xs={24} sm={12} md={12}>
            <div style={{ marginBottom: 8 }}>
              <Text>Activity in Last 30 Days</Text>
            </div>
            <Progress
              percent={(userStats.last30DaysActivity / 30) * 100}
              format={percent => `${userStats.last30DaysActivity}/30 days`}
            />
          </Col>
        </Row>
      </Card>

      {/* Tabs */}
      <Tabs activeKey="activity" onChange={() => {}}>
        <TabPane
          tab={
            <span>
              <HistoryOutlined />
              Recent Activity
            </span>
          }
          key="activity"
        >
          <Card title="Recent Activity">
            <Timeline>
              {recentActivity.map((activity) => (
                <Timeline.Item
                  key={activity.id}
                  color={getActivityColor(activity.type)}
                  dot={
                    <span style={{ fontSize: 20 }}>
                      {getActivityIcon(activity.type)}
                    </span>
                  }
                >
                  <div>
                    <Text strong>{activity.action}</Text>
                    <br />
                    <Text type="secondary">{activity.description}</Text>
                    <br />
                    <Text type="secondary" style={{ fontSize: '0.75rem' }}>
                      {dayjs(activity.timestamp).format('MMMM D, YYYY HH:mm')}
                    </Text>
                  </div>
                </Timeline.Item>
              ))}
            </Timeline>
          </Card>
        </TabPane>

        <TabPane
          tab={
            <span>
              <KeyOutlined />
              Security
            </span>
          }
          key="security"
        >
          <Card
            title="Security Settings"
            extra={
              <Button
                type="primary"
                icon={<KeyOutlined />}
                onClick={() => setPasswordModalVisible(true)}
              >
                Change Password
              </Button>
            }
          >
            <Descriptions bordered column={1}>
              <Descriptions.Item label="Two-Factor Authentication">
                <Switch
                  checked={userData.twoFactorEnabled}
                  checkedChildren="Enabled"
                  unCheckedChildren="Disabled"
                  onChange={(checked) => {
                    // TODO: Implement 2FA setup
                    message.info(
                      checked
                        ? 'Two-factor authentication setup initiated'
                        : 'Two-factor authentication disabled'
                    )
                  }}
                />
              </Descriptions.Item>
              <Descriptions.Item label="Email Verification">
                {userData.emailVerified ? (
                  <Tag color="green" icon={<MailOutlined />}>
                    Verified
                  </Tag>
                ) : (
                  <Button type="link" size="small">
                    Verify Email
                  </Button>
                )}
              </Descriptions.Item>
              <Descriptions.Item label="Phone Verification">
                {userData.phoneVerified ? (
                  <Tag color="green" icon={<PhoneOutlined />}>
                    Verified
                  </Tag>
                ) : (
                  <Button type="link" size="small">
                    Verify Phone
                  </Button>
                )}
              </Descriptions.Item>
              <Descriptions.Item label="Last Password Change">
                {dayjs().subtract(30, 'days').format('MMMM D, YYYY')}
              </Descriptions.Item>
              <Descriptions.Item label="Active Sessions">
                2 sessions active
                <Button type="link" size="small" style={{ marginLeft: 8 }}>
                  View Sessions
                </Button>
              </Descriptions.Item>
            </Descriptions>
          </Card>
        </TabPane>
      </Tabs>

      {/* Change Password Modal */}
      <Modal
        title="Change Password"
        visible={passwordModalVisible}
        onCancel={() => {
          setPasswordModalVisible(false)
          passwordForm.resetFields()
        }}
        footer={[
          <Button
            key="cancel"
            onClick={() => {
              setPasswordModalVisible(false)
              passwordForm.resetFields()
            }}
          >
            Cancel
          </Button>,
          <Button
            key="submit"
            type="primary"
            icon={<SaveOutlined />}
            onClick={() => passwordForm.submit()}
          >
            Change Password
          </Button>,
        ]}
      >
        <Form
          form={passwordForm}
          layout="vertical"
          onFinish={handleChangePassword}
        >
          <Form.Item
            label="Current Password"
            name="oldPassword"
            rules={[{ required: true, message: 'Please enter your current password' }]}
          >
            <Input.Password
              placeholder="Enter current password"
              iconRender={(visible) =>
                visible ? <EyeOutlined /> : <EyeInvisibleOutlined />
              }
            />
          </Form.Item>

          <Form.Item
            label="New Password"
            name="newPassword"
            rules={[
              { required: true, message: 'Please enter a new password' },
              { min: 8, message: 'Password must be at least 8 characters' },
            ]}
          >
            <Input.Password
              placeholder="Enter new password"
              iconRender={(visible) =>
                visible ? <EyeOutlined /> : <EyeInvisibleOutlined />
              }
            />
          </Form.Item>

          <Form.Item
            label="Confirm New Password"
            name="confirmPassword"
            dependencies={['newPassword']}
            rules={[
              { required: true, message: 'Please confirm your new password' },
              ({ getFieldValue }) => ({
                validator(_, value) {
                  if (!value || getFieldValue('newPassword') === value) {
                    return Promise.resolve()
                  }
                  return Promise.reject(new Error('Passwords do not match'))
                },
              }),
            ]}
          >
            <Input.Password
              placeholder="Confirm new password"
              iconRender={(visible) =>
                visible ? <EyeOutlined /> : <EyeInvisibleOutlined />
              }
            />
          </Form.Item>
        </Form>
      </Modal>
    </div>
  )
}

export default UserProfile