import {
  CheckCircleOutlined,
  ClockCircleOutlined,
  CloseCircleOutlined,
  ExperimentOutlined,
  PlayCircleOutlined,
  ProjectOutlined,
} from '@ant-design/icons';
import type {
  ExecutionDetailResp as Execution,
  InjectionDetailResp as Injection,
  ProjectResp as Project,
} from '@rcabench/client';
import { Typography } from 'antd';
import dayjs from 'dayjs';
import relativeTime from 'dayjs/plugin/relativeTime';

dayjs.extend(relativeTime);

const { Text } = Typography;

interface ActivityFeedProps {
  projects: Project[];
  injections: Injection[];
  executions: Execution[];
}

const ActivityFeed = ({
  projects,
  injections,
  executions,
}: ActivityFeedProps) => {
  // Combine and sort all activities
  const activities = [
    ...projects.map((project) => ({
      id: `project-${project.id}`,
      type: 'project',
      title: project.name,
      description: 'Project created',
      time: project.created_at,
      status: 'completed',
      icon: <ProjectOutlined />,
      color: '#3b82f6',
    })),
    ...injections.map((injection) => ({
      id: `injection-${injection.id}`,
      type: 'injection',
      title: injection.name,
      description: `Fault injection ${injection.state}`,
      time: injection.created_at,
      status:
        injection.state === 'COMPLETED'
          ? 'completed'
          : injection.state === 'ERROR'
            ? 'error'
            : 'running',
      icon: <ExperimentOutlined />,
      color: '#f59e0b',
    })),
    ...executions.map((execution) => ({
      id: `execution-${execution.id}`,
      type: 'execution',
      title: `Execution #${execution.id}`,
      description: `Algorithm execution ${execution.state}`,
      time: execution.created_at,
      status:
        execution.state === 'COMPLETED'
          ? 'completed'
          : execution.state === 'ERROR'
            ? 'error'
            : 'running',
      icon: <PlayCircleOutlined />,
      color: '#10b981',
    })),
  ]
    .sort((a, b) => dayjs(b.time).unix() - dayjs(a.time).unix())
    .slice(0, 8);

  const getStatusIcon = (status: string) => {
    switch (status) {
      case 'completed':
        return <CheckCircleOutlined style={{ color: '#10b981' }} />;
      case 'error':
        return <CloseCircleOutlined style={{ color: '#ef4444' }} />;
      case 'running':
        return <ClockCircleOutlined style={{ color: '#3b82f6' }} />;
      default:
        return null;
    }
  };

  return (
    <div className='activity-feed'>
      <h3 className='activity-feed-title'>Recent Activity</h3>
      <div className='activity-list'>
        {activities.map((activity) => (
          <div key={activity.id} className='activity-item'>
            <div
              className='activity-icon'
              style={{
                backgroundColor: `${activity.color}20`,
                color: activity.color,
              }}
            >
              {activity.icon}
            </div>
            <div className='activity-content'>
              <div className='activity-header'>
                <Text className='activity-title' strong>
                  {activity.title}
                </Text>
                {getStatusIcon(activity.status)}
              </div>
              <Text className='activity-desc' type='secondary'>
                {activity.description}
              </Text>
              <Text className='activity-time' type='secondary'>
                {dayjs(activity.time).fromNow()}
              </Text>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default ActivityFeed;
