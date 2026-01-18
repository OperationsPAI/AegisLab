import {
  CheckCircleOutlined,
  DatabaseOutlined,
  SearchOutlined,
} from '@ant-design/icons';
import type { InjectionDetailResp as Datapack } from '@rcabench/client';
import { useQuery } from '@tanstack/react-query';
import {
  Card,
  Col,
  Input,
  Row,
  Space,
  Table,
  type TablePaginationConfig,
  Tag,
  Typography,
} from 'antd';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { datapackApi } from '@/api/datapacks';
import StatCard from '@/components/ui/StatCard';

const { Title } = Typography;
const { Search } = Input;

dayjs.extend(relativeTime);

const DatapackList = () => {
  const navigate = useNavigate();
  const [searchText, setSearchText] = useState('');
  const [pagination, setPagination] = useState({
    current: 1,
    pageSize: 10,
    total: 0,
  });

  const { data: datapacksData, isLoading } = useQuery({
    queryKey: ['datapacks', pagination.current, pagination.pageSize, searchText],
    queryFn: () =>
      datapackApi.getDatapacks({
        page: pagination.current,
        size: pagination.pageSize,
      }),
  });

  const stats = {
    total: datapacksData?.pagination?.total || 0,
    ready:
      datapacksData?.items?.filter((d) => d.state === 'build_success').length || 0,
    building:
      datapacksData?.items?.filter((d) => d.state === 'building').length || 0,
  };

  const handleTableChange = (newPagination: TablePaginationConfig) => {
    setPagination({
      ...pagination,
      current: newPagination.current || 1,
      pageSize: newPagination.pageSize || 10,
    });
  };

  const getStateTag = (state: string) => {
    const stateMap: Record<string, { color: string; text: string }> = {
      build_success: { color: 'success', text: 'Ready' },
      building: { color: 'processing', text: 'Building' },
      build_failed: { color: 'error', text: 'Failed' },
      initial: { color: 'default', text: 'Pending' },
    };
    const config = stateMap[state] || { color: 'default', text: state };
    return <Tag color={config.color}>{config.text}</Tag>;
  };

  const columns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
      width: 80,
    },
    {
      title: 'Name',
      dataIndex: 'name',
      key: 'name',
      render: (text: string, record: Datapack) => (
        <a onClick={() => navigate(`/datapacks/${record.id}`)}>{text}</a>
      ),
    },
    {
      title: 'Fault Type',
      dataIndex: 'fault_type',
      key: 'fault_type',
      render: (type: string) => <Tag>{type}</Tag>,
    },
    {
      title: 'Benchmark',
      dataIndex: 'benchmark',
      key: 'benchmark',
      render: (benchmark: { name: string }) => benchmark?.name || '-',
    },
    {
      title: 'State',
      dataIndex: 'state',
      key: 'state',
      render: (state: string) => getStateTag(state),
    },
    {
      title: 'Created',
      dataIndex: 'created_at',
      key: 'created_at',
      render: (date: string) => dayjs(date).fromNow(),
    },
  ];

  return (
    <div style={{ padding: '24px' }}>
      <Title level={2}>
        <DatabaseOutlined /> Datapacks
      </Title>

      <Row gutter={[16, 16]} style={{ marginBottom: '24px' }}>
        <Col span={8}>
          <StatCard
            title="Total Datapacks"
            value={stats.total}
            icon={<DatabaseOutlined />}
            color="#1890ff"
          />
        </Col>
        <Col span={8}>
          <StatCard
            title="Ready"
            value={stats.ready}
            icon={<CheckCircleOutlined />}
            color="#52c41a"
          />
        </Col>
        <Col span={8}>
          <StatCard
            title="Building"
            value={stats.building}
            icon={<DatabaseOutlined />}
            color="#faad14"
          />
        </Col>
      </Row>

      <Card>
        <Space direction="vertical" size="large" style={{ width: '100%' }}>
          <Search
            placeholder="Search datapacks"
            allowClear
            enterButton={<SearchOutlined />}
            size="large"
            onSearch={setSearchText}
          />

          <Table
            columns={columns}
            dataSource={datapacksData?.items || []}
            loading={isLoading}
            rowKey="id"
            pagination={{
              current: pagination.current,
              pageSize: pagination.pageSize,
              total: datapacksData?.pagination?.total || 0,
              showSizeChanger: true,
              showTotal: (total) => `Total ${total} datapacks`,
            }}
            onChange={handleTableChange}
          />
        </Space>
      </Card>
    </div>
  );
};

export default DatapackList;
