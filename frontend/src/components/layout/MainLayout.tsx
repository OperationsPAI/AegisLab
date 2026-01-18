import {
  ApiOutlined,
  BarChartOutlined,
  BulbOutlined,
  CloudServerOutlined,
  ContainerOutlined,
  DashboardOutlined,
  DatabaseOutlined,
  ExperimentOutlined,
  LogoutOutlined,
  MenuFoldOutlined,
  MenuUnfoldOutlined,
  PlayCircleOutlined,
  ProjectOutlined,
  SafetyCertificateOutlined,
  SettingOutlined,
  UnorderedListOutlined,
  UserOutlined,
} from '@ant-design/icons';
import {
  Avatar,
  Button,
  Dropdown,
  Layout,
  Menu,
  type MenuProps,
  Space,
  Typography,
} from 'antd';
import { useEffect, useState } from 'react';
import { Outlet, useLocation, useNavigate } from 'react-router-dom';


import ThemeToggle from '@/components/ui/ThemeToggle';
import { useAuthStore } from '@/store/auth';
import { useThemeStore } from '@/store/theme';

import './MainLayout.css';

const { Header, Sider, Content } = Layout;
const { Text } = Typography;

const MainLayout: React.FC = () => {
  const navigate = useNavigate();
  const location = useLocation();
  const { user, logout } = useAuthStore();
  const { sidebarCollapsed, toggleSidebar } = useThemeStore();
  const [selectedKeys, setSelectedKeys] = useState<string[]>([]);
  const [openKeys, setOpenKeys] = useState<string[]>([]);

  // Menu items
  const menuItems: MenuProps['items'] = [
    {
      key: '/dashboard',
      icon: <DashboardOutlined />,
      label: 'Dashboard',
    },
    {
      type: 'divider',
      style: { margin: '8px 0' },
    },
    {
      key: 'experiments',
      icon: <ExperimentOutlined />,
      label: 'Experiments',
      children: [
        {
          key: '/projects',
          icon: <ProjectOutlined />,
          label: 'Projects',
        },
        {
          key: '/injections',
          icon: <BulbOutlined />,
          label: 'Fault Injections',
        },
        {
          key: '/executions',
          icon: <PlayCircleOutlined />,
          label: 'Algorithm Runs',
        },
      ],
    },
    {
      key: 'infrastructure',
      icon: <CloudServerOutlined />,
      label: 'Infrastructure',
      children: [
        {
          key: '/containers',
          icon: <ContainerOutlined />,
          label: 'Containers',
        },
        {
          key: '/datasets',
          icon: <DatabaseOutlined />,
          label: 'Datasets',
        },
        {
          key: '/datapacks',
          icon: <DatabaseOutlined />,
          label: 'Datapacks',
        },
      ],
    },
    {
      key: 'analysis',
      icon: <BarChartOutlined />,
      label: 'Analysis',
      children: [
        {
          key: '/evaluations',
          icon: <SafetyCertificateOutlined />,
          label: 'Evaluations',
        },
        {
          key: '/tasks',
          icon: <UnorderedListOutlined />,
          label: 'Task Monitor',
        },
      ],
    },
    {
      type: 'divider',
      style: { margin: '8px 0' },
    },
    {
      key: 'system',
      icon: <SettingOutlined />,
      label: 'System',
      children: [
        {
          key: '/system',
          icon: <ApiOutlined />,
          label: 'API & Config',
        },
        {
          key: '/settings',
          icon: <SettingOutlined />,
          label: 'Settings',
        },
      ],
    },
  ];

  // User dropdown menu
  const userMenuItems: MenuProps['items'] = [
    {
      key: 'profile',
      icon: <UserOutlined />,
      label: 'Profile',
    },
    {
      type: 'divider',
    },
    {
      key: 'logout',
      icon: <LogoutOutlined />,
      label: 'Sign Out',
      danger: true,
    },
  ];

  const handleMenuClick = ({ key }: { key: string }) => {
    if (key.startsWith('/')) {
      navigate(key);
    }
  };

  const handleUserMenuClick = async ({ key }: { key: string }) => {
    if (key === 'logout') {
      await logout();
      navigate('/login');
    } else if (key === 'profile') {
      navigate('/settings/profile');
    }
  };

  const handleOpenChange = (keys: string[]) => {
    setOpenKeys(keys);
  };

  // Update selected keys based on location
  useEffect(() => {
    const path = location.pathname;
    setSelectedKeys([path]);

    // Set open keys for parent menus
    if (
      path.startsWith('/projects') ||
      path.startsWith('/injections') ||
      path.startsWith('/executions')
    ) {
      setOpenKeys(['experiments']);
    } else if (path.startsWith('/containers') || path.startsWith('/datasets')) {
      setOpenKeys(['infrastructure']);
    } else if (path.startsWith('/evaluations') || path.startsWith('/tasks')) {
      setOpenKeys(['analysis']);
    } else if (path.startsWith('/system') || path.startsWith('/settings')) {
      setOpenKeys(['system']);
    }
  }, [location.pathname]);

  return (
    <Layout className='main-layout'>
      {/* Header */}
      <Header className='main-header'>
        <div className='header-left'>
          <Button
            type='text'
            icon={
              sidebarCollapsed ? <MenuUnfoldOutlined /> : <MenuFoldOutlined />
            }
            onClick={toggleSidebar}
            aria-label={sidebarCollapsed ? 'Expand sidebar' : 'Collapse sidebar'}
            className='sidebar-toggle'
          />
          <div className='logo-section' onClick={() => navigate('/dashboard')} style={{ cursor: 'pointer' }}>
            <div className='logo-icon'>
              <svg width='32' height='32' viewBox='0 0 32 32' fill='none' role='img' aria-label='AegisLab logo'>
                <title>AegisLab</title>
                <path
                  d='M16 2L30 8.5V23.5L16 30L2 23.5V8.5L16 2Z'
                  stroke='currentColor'
                  strokeWidth='2'
                  strokeLinejoin='round'
                />
                <path
                  d='M16 16L30 8.5M16 16V30M16 16L2 8.5'
                  stroke='currentColor'
                  strokeWidth='2'
                  strokeLinejoin='round'
                  opacity='0.3'
                />
                <circle cx='16' cy='16' r='4' fill='currentColor' />
              </svg>
            </div>
            {!sidebarCollapsed && (
              <div className='logo-text'>
                <span className='logo-title'>AegisLab</span>
                <span className='logo-subtitle'>RCA Benchmark Platform</span>
              </div>
            )}
          </div>
        </div>

        <div className='header-right'>
          <ThemeToggle />
          <Dropdown
            menu={{ items: userMenuItems, onClick: handleUserMenuClick }}
            placement='bottomRight'
            arrow
            overlayClassName='user-dropdown'
          >
            <Space className='user-section' role='button' aria-label='User menu' tabIndex={0}>
              <Avatar
                size='small'
                icon={<UserOutlined />}
                style={{ backgroundColor: 'var(--color-primary-500)' }}
              />
              {!sidebarCollapsed && (
                <Text className='username'>{user?.username || 'User'}</Text>
              )}
            </Space>
          </Dropdown>
        </div>
      </Header>

      <Layout className='main-body'>
        {/* Sidebar */}
        <Sider
          width={240}
          collapsed={sidebarCollapsed}
          collapsedWidth={64}
          className='main-sidebar'
        >
          <div className='sidebar-content'>
            <Menu
              mode='inline'
              selectedKeys={selectedKeys}
              openKeys={openKeys}
              items={menuItems}
              onClick={handleMenuClick}
              onOpenChange={handleOpenChange}
              className='sidebar-menu'
              inlineCollapsed={sidebarCollapsed}
            />
          </div>

          {/* Sidebar Footer */}
          <div className='sidebar-footer'>
            <div className='system-status'>
              <div className='status-indicator' />
              <span className='status-text'>System Online</span>
            </div>
          </div>
        </Sider>

        {/* Main Content */}
        <Layout className='main-content-wrapper'>
          <Content className='main-content'>
            <div className='content-inner'>
              <Outlet />
            </div>
          </Content>
        </Layout>
      </Layout>
    </Layout>
  );
};

export default MainLayout;
