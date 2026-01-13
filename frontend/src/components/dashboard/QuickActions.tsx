
import {
  BarChartOutlined,
  ExperimentOutlined,
  PlayCircleOutlined,
  PlusOutlined,
} from '@ant-design/icons';
import { Button, Col, Row } from 'antd';
import { useNavigate } from 'react-router-dom';

const QuickActions = () => {
  const navigate = useNavigate();

  const actions = [
    {
      key: 'new-project',
      icon: <PlusOutlined />,
      label: 'New Project',
      description: 'Create a new RCA experiment',
      color: 'var(--color-primary-500)',
      path: '/projects/new',
    },
    {
      key: 'run-injection',
      icon: <ExperimentOutlined />,
      label: 'Fault Injection',
      description: 'Start a chaos experiment',
      color: 'var(--color-warning)',
      path: '/injections/create',
    },
    {
      key: 'execute-algo',
      icon: <PlayCircleOutlined />,
      label: 'Run Algorithm',
      description: 'Execute RCA algorithm',
      color: 'var(--color-success)',
      path: '/executions/new',
    },
    {
      key: 'view-reports',
      icon: <BarChartOutlined />,
      label: 'View Reports',
      description: 'Analyze results',
      color: 'var(--color-info)',
      path: '/evaluations',
    },
  ];

  return (
    <div className='quick-actions'>
      <h3 className='quick-actions-title'>Quick Actions</h3>
      <Row gutter={[16, 16]} className='quick-actions-grid'>
        {actions.map((action) => (
          <Col key={action.key} xs={24} sm={12} md={6}>
            <Button
              className='quick-action-btn'
              onClick={() => navigate(action.path)}
              style={{
                background: `linear-gradient(135deg, ${action.color}15, ${action.color}25)`,
                borderColor: `${action.color}40`,
              }}
            >
              <div
                className='quick-action-icon'
                style={{ color: action.color }}
              >
                {action.icon}
              </div>
              <div className='quick-action-content'>
                <div className='quick-action-label'>{action.label}</div>
                <div className='quick-action-desc'>{action.description}</div>
              </div>
            </Button>
          </Col>
        ))}
      </Row>
    </div>
  );
};

export default QuickActions;
