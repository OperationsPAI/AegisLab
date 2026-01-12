import { Layout, Menu, Avatar, Dropdown, Space, Typography } from 'antd'
import type { MenuProps } from 'antd'
import {
  DashboardOutlined,
  ProjectOutlined,
  ContainerOutlined,
  DatabaseOutlined,
  ExperimentOutlined,
  PlayCircleOutlined,
  BarChartOutlined,
  UnorderedListOutlined,
  SettingOutlined,
  UserOutlined,
  LogoutOutlined,
} from '@ant-design/icons'
import { Outlet, useNavigate, useLocation } from 'react-router-dom'
import { useAuthStore } from '@/store/auth'

const { Header, Sider, Content } = Layout
const { Text } = Typography

const MainLayout = () => {
  const navigate = useNavigate()
  const location = useLocation()
  const { user, logout } = useAuthStore()

  // Menu items
  const menuItems: MenuProps['items'] = [
    {
      key: '/dashboard',
      icon: <DashboardOutlined />,
      label: '仪表盘',
    },
    {
      type: 'divider',
    },
    {
      key: '/projects',
      icon: <ProjectOutlined />,
      label: '项目管理',
    },
    {
      key: '/containers',
      icon: <ContainerOutlined />,
      label: '容器管理',
    },
    {
      key: '/datasets',
      icon: <DatabaseOutlined />,
      label: '数据集管理',
    },
    {
      type: 'divider',
    },
    {
      key: '/injections',
      icon: <ExperimentOutlined />,
      label: '故障注入',
    },
    {
      key: '/executions',
      icon: <PlayCircleOutlined />,
      label: '算法执行',
    },
    {
      key: '/evaluations',
      icon: <BarChartOutlined />,
      label: '评估',
    },
    {
      type: 'divider',
    },
    {
      key: '/tasks',
      icon: <UnorderedListOutlined />,
      label: '任务监控',
    },
    {
      key: '/system',
      icon: <SettingOutlined />,
      label: '系统管理',
    },
  ]

  // User dropdown menu
  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: '个人资料',
    },
    {
      key: 'settings',
      icon: <SettingOutlined />,
      label: '设置',
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: '退出登录',
      danger: true,
    },
  ]

  const handleMenuClick = ({ key }: { key: string }) => {
    navigate(key)
  }

  const handleUserMenuClick = async ({ key }: { key: string }) => {
    if (key === 'logout') {
      await logout()
      navigate('/login')
    } else if (key === 'profile') {
      navigate('/settings/profile')
    } else if (key === 'settings') {
      navigate('/settings')
    }
  }

  // Get current selected key from location
  const selectedKey = '/' + location.pathname.split('/')[1]

  return (
    <Layout style={{ minHeight: '100vh' }}>
      {/* Header */}
      <Header
        style={{
          position: 'fixed',
          zIndex: 1,
          width: '100%',
          display: 'flex',
          alignItems: 'center',
          justifyContent: 'space-between',
          backgroundColor: '#ffffff',
          borderBottom: '1px solid #e5e7eb',
          padding: '0 24px',
        }}
      >
        <div style={{ display: 'flex', alignItems: 'center', gap: '16px' }}>
          <div
            style={{
              fontSize: '20px',
              fontWeight: 600,
              color: '#2563eb',
              cursor: 'pointer',
            }}
            onClick={() => navigate('/dashboard')}
          >
            🔬 RCABench
          </div>
          <Text type="secondary" style={{ fontSize: '12px' }}>
            微服务根因分析基准测试平台
          </Text>
        </div>

        <Dropdown
          menu={{ items: userMenuItems, onClick: handleUserMenuClick }}
          placement="bottomRight"
        >
          <Space style={{ cursor: 'pointer' }}>
            <Avatar icon={<UserOutlined />} style={{ backgroundColor: '#2563eb' }} />
            <Text>{user?.username || '用户'}</Text>
          </Space>
        </Dropdown>
      </Header>

      <Layout style={{ marginTop: 64 }}>
        {/* Sidebar */}
        <Sider
          width={220}
          style={{
            overflow: 'auto',
            height: 'calc(100vh - 64px)',
            position: 'fixed',
            left: 0,
            top: 64,
            backgroundColor: '#f9fafb',
            borderRight: '1px solid #e5e7eb',
          }}
        >
          <Menu
            mode="inline"
            selectedKeys={[selectedKey]}
            items={menuItems}
            onClick={handleMenuClick}
            style={{
              borderRight: 'none',
              backgroundColor: 'transparent',
            }}
          />
        </Sider>

        {/* Main Content */}
        <Layout style={{ marginLeft: 220 }}>
          <Content
            style={{
              margin: '24px',
              padding: '24px',
              minHeight: 'calc(100vh - 64px - 48px)',
              backgroundColor: '#ffffff',
              borderRadius: '8px',
            }}
          >
            <Outlet />
          </Content>
        </Layout>
      </Layout>
    </Layout>
  )
}

export default MainLayout
