
import {
  CalendarOutlined,
  CheckCircleOutlined,
  DeleteOutlined,
  EditOutlined,
  ExperimentOutlined,
  FolderOutlined,
  PlayCircleOutlined,
  PlusOutlined,
  SearchOutlined,
  TeamOutlined,
} from '@ant-design/icons';
import type { ProjectResp as Project } from '@rcabench/client';
import { useQuery } from '@tanstack/react-query';
import {
  Avatar,
  Button,
  Card,
  Col,
  Input,
  Row,
  Space,
  Table,
  type TablePaginationConfig,
  Typography,
} from 'antd';
import dayjs from 'dayjs';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { projectApi } from '@/api/projects';
import StatCard from '@/components/ui/StatCard';
import StatusBadge, {
  type StatusBadgeProps,
} from '@/components/ui/StatusBadge';
import { ProjectState } from '@/types/api';

import './ProjectList.css';

const { Title, Text } = Typography;
const { Search } = Input;

const ProjectList = () => {
  const navigate = useNavigate();
  const [searchText, setSearchText] = useState('');
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  // Fetch projects
  const { data: projectsData, isLoading } = useQuery({
    queryKey: ['projects', pagination.current, pagination.pageSize, searchText],
    queryFn: () =>
      projectApi.getProjects({
        page: pagination.current,
        size: pagination.pageSize,
      }),
  });

  // Statistics - mock data for now
  const stats = {
    total: projectsData?.pagination?.total || 0,
    active:
      projectsData?.items?.filter((p: any) => p.state === ProjectState.ACTIVE)
        .length || 0,
    completedThisMonth:
      projectsData?.items?.filter((p: any) =>
        dayjs(p.created_at).isAfter(dayjs().subtract(1, 'month'))
      ).length || 0,
    totalExperiments:
      projectsData?.items?.reduce(
        (sum: number, p: any) => sum + (p.experiment_count || 0),
        0
      ) || 0,
  };

  const handleTableChange = (newPagination: TablePaginationConfig) => {
    setPagination({
      ...pagination,
      current: newPagination.current || 1,
      pageSize: newPagination.pageSize || 10,
    });
  };

  const handleSearch = (value: string) => {
    setSearchText(value);
    setPagination({ ...pagination, current: 1 });
  };

  const handleCreateProject = () => {
    navigate('/projects/new');
  };

  const handleEditProject = (id: number | undefined) => {
    if (id) {
      navigate(`/projects/${id}/edit`);
    }
  };

  const handleRunExperiment = (project: Project) => {
    navigate('/injections/create', {
      state: { projectId: project.id, projectName: project.name },
    });
  };

  const columns = [
    {
      title: 'Project',
      dataIndex: 'name',
      key: 'name',
      width: '30%',
      render: (name: string, record: Project) => (
        <Space>
          <Avatar
            size='large'
            style={{
              backgroundColor: '#3b82f6',
              fontSize: '1.25rem',
            }}
            icon={<FolderOutlined />}
          />
          <div>
            <Text strong style={{ fontSize: '1rem' }}>
              {name}
            </Text>
            <br />
            <Text type='secondary' style={{ fontSize: '0.875rem' }}>
              ID: {record.id}
            </Text>
          </div>
        </Space>
      ),
    },
    {
      title: 'Status',
      dataIndex: 'state',
      key: 'state',
      width: '15%',
      render: (state: ProjectState) => {
        const statusMap = {
          [ProjectState.ACTIVE]: { text: 'Active', color: 'success' },
          [ProjectState.PAUSED]: { text: 'Paused', color: 'warning' },
          [ProjectState.COMPLETED]: { text: 'Completed', color: 'info' },
          [ProjectState.ARCHIVED]: { text: 'Archived', color: 'default' },
        };
        const config = statusMap[state] || {
          text: 'Unknown',
          color: 'default',
        };
        return (
          <StatusBadge
            status={config.color as StatusBadgeProps['status']}
            text={config.text}
          />
        );
      },
    },
    {
      title: 'Experiments',
      dataIndex: 'experiment_count',
      key: 'experiment_count',
      width: '12%',
      render: (count: number) => (
        <Text>
          <ExperimentOutlined /> {count || 0}
        </Text>
      ),
    },
    {
      title: 'Team',
      dataIndex: 'team_size',
      key: 'team_size',
      width: '12%',
      render: (size: number) => (
        <Text>
          <TeamOutlined /> {size || 1} member{size !== 1 ? 's' : ''}
        </Text>
      ),
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      width: '15%',
      render: (date: string) => (
        <Space>
          <CalendarOutlined />
          <Text>{dayjs(date).format('MMM D, YYYY')}</Text>
        </Space>
      ),
    },
    {
      title: 'Actions',
      key: 'actions',
      width: '16%',
      render: (_: unknown, record: Project) => (
        <Space>
          <Button
            type='text'
            icon={<PlayCircleOutlined />}
            onClick={() => handleRunExperiment(record)}
            title='Run Experiment'
          />
          <Button
            type='text'
            icon={<EditOutlined />}
            onClick={() => handleEditProject(record.id)}
            title='Edit Project'
          />
          <Button
            type='text'
            danger
            icon={<DeleteOutlined />}
            title='Delete Project'
          />
        </Space>
      ),
    },
  ];

  return (
    <div className='project-list'>
      {/* Page Header */}
      <div className='page-header'>
        <div className='page-header-left'>
          <Title level={2} className='page-title'>
            Projects
          </Title>
          <Text type='secondary'>
            Manage your RCA benchmarking projects and experiments
          </Text>
        </div>
        <Button
          type='primary'
          size='large'
          icon={<PlusOutlined />}
          onClick={handleCreateProject}
          className='create-button'
        >
          New Project
        </Button>
      </div>

      {/* Statistics Cards */}
      <Row gutter={[{ xs: 8, sm: 16, lg: 24 }, { xs: 8, sm: 16, lg: 24 }]} className='stats-row'>
        <Col xs={12} sm={12} lg={6}>
          <StatCard
            title='Total Projects'
            value={stats?.total || 0}
            prefix={<FolderOutlined />}
            color='primary'
          />
        </Col>
        <Col xs={12} sm={12} lg={6}>
          <StatCard
            title='Active Projects'
            value={stats?.active || 0}
            prefix={<ExperimentOutlined />}
            color='success'
          />
        </Col>
        <Col xs={12} sm={12} lg={6}>
          <StatCard
            title='Completed This Month'
            value={stats?.completedThisMonth || 0}
            prefix={<CheckCircleOutlined />}
            color='info'
          />
        </Col>
        <Col xs={12} sm={12} lg={6}>
          <StatCard
            title='Total Experiments'
            value={stats?.totalExperiments || 0}
            prefix={<PlayCircleOutlined />}
            color='warning'
          />
        </Col>
      </Row>

      {/* Search and Filters */}
      <Card className='search-card'>
        <Row gutter={[24, 24]} align='middle'>
          <Col flex='auto'>
            <Search
              placeholder='Search projects by name or ID...'
              allowClear
              enterButton={<SearchOutlined />}
              size='large'
              onSearch={handleSearch}
              style={{ maxWidth: 400 }}
            />
          </Col>
          <Col>
            <Space>
              <Button size='large'>Filter</Button>
              <Button size='large'>Sort</Button>
            </Space>
          </Col>
        </Row>
      </Card>

      {/* Projects Table */}
      <Card className='table-card'>
        <Table
          columns={columns}
          dataSource={projectsData?.items || []}
          loading={isLoading}
          pagination={{
            ...pagination,
            total: projectsData?.pagination?.total || 0,
            showSizeChanger: true,
            showQuickJumper: true,
            showTotal: (total) => `Total ${total} projects`,
          }}
          onChange={handleTableChange}
          rowKey='id'
          className='projects-table'
          rowClassName='project-row'
        />
      </Card>
    </div>
  );
};

export default ProjectList;
