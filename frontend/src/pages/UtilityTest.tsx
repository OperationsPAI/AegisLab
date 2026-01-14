import { CheckCircleOutlined, ClockCircleOutlined, CloseCircleOutlined } from '@ant-design/icons';
import { Card, Table, Tag, Button, Switch, Row, Col, Divider } from 'antd';
import type React from 'react';
import { useState } from 'react';

const UtilityTest: React.FC = () => {
  const [darkTheme, setDarkTheme] = useState(false);

  const toggleTheme = () => {
    const newTheme = !darkTheme;
    setDarkTheme(newTheme);
    document.documentElement.setAttribute('data-theme', newTheme ? 'dark' : 'light');
  };

  const tableColumns = [
    {
      title: 'ID',
      dataIndex: 'id',
      key: 'id',
    },
    {
      title: '任务',
      dataIndex: 'task',
      key: 'task',
    },
    {
      title: '状态',
      dataIndex: 'status',
      key: 'status',
      render: (status: string) => {
        const statusConfig = {
          已完成: { color: 'success', icon: <CheckCircleOutlined /> },
          进行中: { color: 'processing', icon: <ClockCircleOutlined /> },
          失败: { color: 'error', icon: <CloseCircleOutlined /> },
        };
        const config = statusConfig[status as keyof typeof statusConfig] || { color: 'default', icon: <ClockCircleOutlined /> };
        return (
          <Tag color={config.color}>
            {config.icon} {status}
          </Tag>
        );
      },
    },
    {
      title: '优先级',
      dataIndex: 'priority',
      key: 'priority',
    },
    {
      title: '创建时间',
      dataIndex: 'createdAt',
      key: 'createdAt',
    },
    {
      title: '更新时间',
      dataIndex: 'updatedAt',
      key: 'updatedAt',
    },
  ];

  const tableData = [
    {
      id: '1',
      task: '运行注入实验',
      status: '已完成',
      priority: '高',
      createdAt: '2024-01-13 09:30:00',
      updatedAt: '2024-01-13 09:45:00',
    },
    {
      id: '2',
      task: '执行根因分析',
      status: '进行中',
      priority: '中',
      createdAt: '2024-01-13 10:00:00',
      updatedAt: '2024-01-13 10:15:00',
    },
    {
      id: '3',
      task: '生成报告',
      status: '失败',
      priority: '低',
      createdAt: '2024-01-13 11:00:00',
      updatedAt: '2024-01-13 11:30:00',
    },
  ];

  return (
    <div className="container">
      <div className="page-header">
        <div className="page-header-left">
          <h1 className="page-title">工具类系统验证</h1>
          <p className="page-description">
            验证新添加的工具类系统，包括表格滚动条、卡片共享样式、响应式断点和布局工具类。
          </p>
        </div>
        <div className="flex items-center gap-2">
          <span className="text-gray-500 text-sm">浅色</span>
          <Switch checked={darkTheme} onChange={toggleTheme} />
          <span className="text-gray-500 text-sm">深色</span>
        </div>
      </div>

      <Divider />

      {/* 响应式网格布局 */}
      <h2 className="text-xl font-bold mb-4">响应式网格布局</h2>
      <Row gutter={16} className="mb-8">
        <Col xs={24} sm={12} md={8} lg={6} className="flex">
          <Card className="w-full card-hover">
            <div className="text-2xl font-bold text-primary-500">24</div>
            <div className="text-sm text-gray-500">总项目</div>
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8} lg={6} className="flex">
          <Card className="w-full card-hover">
            <div className="text-2xl font-bold text-warning">12</div>
            <div className="text-sm text-gray-500">进行中</div>
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8} lg={6} className="flex">
          <Card className="w-full card-hover">
            <div className="text-2xl font-bold text-success">8</div>
            <div className="text-sm text-gray-500">已完成</div>
          </Card>
        </Col>
        <Col xs={24} sm={12} md={8} lg={6} className="flex">
          <Card className="w-full card-hover">
            <div className="text-2xl font-bold text-error">4</div>
            <div className="text-sm text-gray-500">失败</div>
          </Card>
        </Col>
      </Row>

      {/* 表格滚动 */}
      <h2 className="text-xl font-bold mb-4">表格滚动验证</h2>
      <div className="table-scroll-container mb-8">
        <Table
          columns={tableColumns}
          dataSource={tableData}
          rowKey="id"
          pagination={false}
          scroll={{ x: 800 }}
          className="projects-table"
        />
      </div>

      {/* 垂直滚动表格 */}
      <h2 className="text-xl font-bold mb-4">垂直滚动表格</h2>
      <div className="table-scroll-vertical mb-8">
        <Table
          columns={tableColumns}
          dataSource={[...tableData, ...tableData, ...tableData, ...tableData, ...tableData]}
          rowKey="id"
          pagination={false}
          className="projects-table"
        />
      </div>

      {/* 卡片共享样式 */}
      <h2 className="text-xl font-bold mb-4">卡片共享样式</h2>
      <Row gutter={16} className="mb-8">
        <Col xs={24} sm={12} md={8} className="flex">
          <div className="card w-full card-padding flex-1">
            <div className="card-header">
              <h3 className="card-title">基础卡片</h3>
            </div>
            <div className="card-body">
              <p className="text-gray-500">这是一个基础卡片，具有默认样式和阴影效果。</p>
            </div>
            <div className="card-footer">
              <Button type="primary">操作</Button>
            </div>
          </div>
        </Col>
        <Col xs={24} sm={12} md={8} className="flex">
          <div className="card w-full card-padding flex-1 card-shadow-lg card-border-primary">
            <div className="card-header">
              <h3 className="card-title">强调卡片</h3>
            </div>
            <div className="card-body">
              <p className="text-gray-500">这个卡片有更大的阴影和蓝色边框，用于突出显示。</p>
            </div>
            <div className="card-footer">
              <Button type="default">查看</Button>
              <Button type="primary">编辑</Button>
            </div>
          </div>
        </Col>
        <Col xs={24} sm={12} md={8} className="flex">
          <div className="card w-full card-padding flex-1 card-shadow-sm card-rounded-xl">
            <div className="card-header">
              <h3 className="card-title">圆角卡片</h3>
            </div>
            <div className="card-body">
              <p className="text-gray-500">这个卡片具有更大的圆角，外观更柔和。</p>
            </div>
            <div className="card-footer">
              <Button type="text">了解更多</Button>
            </div>
          </div>
        </Col>
      </Row>

      {/* 布局工具类 */}
      <h2 className="text-xl font-bold mb-4">布局工具类</h2>
      <div className="flex flex-wrap gap-4 mb-4">
        <div className="p-4 bg-primary-50 border border-primary-200 rounded-lg">
          <div className="font-medium">Flex 布局</div>
          <div className="text-xs text-gray-500">flex flex-wrap gap-4</div>
        </div>
        <div className="p-4 bg-warning-light border border-warning-200 rounded-lg">
          <div className="font-medium">Grid 布局</div>
          <div className="text-xs text-gray-500">grid grid-cols-3 gap-4</div>
        </div>
        <div className="p-4 bg-success-light border border-success-200 rounded-lg">
          <div className="font-medium">间距工具</div>
          <div className="text-xs text-gray-500">mx-4 my-2 p-3</div>
        </div>
      </div>

      {/* 响应式显示/隐藏 */}
      <h2 className="text-xl font-bold mb-4">响应式显示/隐藏</h2>
      <div className="flex flex-wrap gap-2">
        <span className="hidden-xs p-2 bg-gray-100 rounded">仅在 xs 以上显示</span>
        <span className="hidden-sm p-2 bg-gray-100 rounded">仅在 sm 以上显示</span>
        <span className="hidden-md p-2 bg-gray-100 rounded">仅在 md 以上显示</span>
        <span className="hidden-lg p-2 bg-gray-100 rounded">仅在 lg 以上显示</span>
        <span className="hidden-xl p-2 bg-gray-100 rounded">仅在 xl 以上显示</span>
      </div>

      <Divider />

      <div className="text-center text-gray-500 text-sm">
        工具类系统验证页面 - 支持浅色/深色主题切换，所有工具类基于 CSS 变量构建
      </div>
    </div>
  );
};

export default UtilityTest;
