
import {
  ClockCircleOutlined,
  CloudServerOutlined,
  DatabaseOutlined,
  DisconnectOutlined,
  PauseCircleOutlined,
  QuestionCircleOutlined,
  StopOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import { useQuery } from '@tanstack/react-query';
import { Card, Empty, List, Spin, Tag, Tooltip } from 'antd';
import { useState } from 'react';
import type React from 'react';

import { injectionApi } from '../../../api/injections';
import type { FaultType } from '../../../types/api';

import './FaultTypePanel.css';

interface FaultTypePanelProps {
  onFaultSelect: (fault: FaultType) => void;
}

const faultTypeIcons: Record<string, React.ReactNode> = {
  cpu: <ThunderboltOutlined className='fault-type-icon' />,
  memory: <DatabaseOutlined className='fault-type-icon' />,
  disk: <CloudServerOutlined className='fault-type-icon' />,
  network: <DisconnectOutlined className='fault-type-icon' />,
  process: <StopOutlined className='fault-type-icon' />,
  io: <PauseCircleOutlined className='fault-type-icon' />,
  time: <ClockCircleOutlined className='fault-type-icon' />,
  default: <QuestionCircleOutlined className='fault-type-icon' />,
};

const faultTypeColors: Record<string, string> = {
  cpu: 'red',
  memory: 'orange',
  disk: 'blue',
  network: 'green',
  process: 'purple',
  io: 'cyan',
  time: 'gold',
  default: 'default',
};

export const FaultTypePanel: React.FC<FaultTypePanelProps> = ({
  onFaultSelect,
}) => {
  const [selectedFault, setSelectedFault] = useState<FaultType | null>(null);

  // Fetch fault types
  const {
    data: faultTypes = [],
    isLoading,
    error,
  } = useQuery({
    queryKey: ['faultTypes'],
    queryFn: () => injectionApi.getFaultTypes(),
  });

  const handleFaultClick = (fault: FaultType) => {
    setSelectedFault(fault);
    onFaultSelect(fault);
  };

  const handleDragStart = (e: React.DragEvent, fault: FaultType) => {
    e.dataTransfer.setData('application/reactflow', JSON.stringify(fault));
    e.dataTransfer.effectAllowed = 'move';
  };

  const getFaultIcon = (faultType: FaultType) => {
    const key = faultType.category?.toLowerCase() || 'default';
    return faultTypeIcons[key] || faultTypeIcons.default;
  };

  const getFaultColor = (faultType: FaultType) => {
    const key = faultType.category?.toLowerCase() || 'default';
    return faultTypeColors[key] || faultTypeColors.default;
  };

  const groupFaultTypesByCategory = (faultTypes: FaultType[]) => {
    return faultTypes.reduce(
      (acc, fault) => {
        const category = fault.category || 'Other';
        if (!acc[category]) {
          acc[category] = [];
        }
        acc[category].push(fault);
        return acc;
      },
      {} as Record<string, FaultType[]>
    );
  };

  const groupedFaultTypes = groupFaultTypesByCategory(faultTypes);

  if (error) {
    return (
      <Card title='Fault Types' className='fault-type-panel'>
        <Empty description='Failed to load fault types' />
      </Card>
    );
  }

  return (
    <Card
      title='Fault Types'
      className='fault-type-panel'
      bodyStyle={{ padding: 0 }}
    >
      {isLoading ? (
        <div style={{ padding: '24px', textAlign: 'center' }}>
          <Spin />
        </div>
      ) : (
        <div className='fault-type-list'>
          {Object.entries(groupedFaultTypes).map(([category, faults]) => (
            <div key={category} className='fault-category'>
              <div className='category-header'>
                <Tag color='blue' className='category-tag'>
                  {category}
                </Tag>
              </div>
              <List
                dataSource={faults}
                renderItem={(fault) => (
                  <List.Item
                    className={`fault-type-item ${selectedFault?.id === fault.id ? 'selected' : ''}`}
                    onClick={() => handleFaultClick(fault)}
                    draggable
                    onDragStart={(e) => handleDragStart(e, fault)}
                    style={{ cursor: 'grab' }}
                  >
                    <div className='fault-type-content'>
                      <div className='fault-type-header'>
                        {getFaultIcon(fault)}
                        <span className='fault-type-name'>{fault.name}</span>
                        <Tag
                          color={getFaultColor(fault)}
                          className='fault-type-tag'
                        >
                          {fault.type}
                        </Tag>
                      </div>
                      <div className='fault-type-description'>
                        {fault.description}
                      </div>
                      {fault.parameters && (
                        <div className='fault-type-params'>
                          <Tooltip
                            title={`${fault.parameters.length} parameters`}
                          >
                            <Tag size='small'>
                              {fault.parameters.length} params
                            </Tag>
                          </Tooltip>
                        </div>
                      )}
                    </div>
                  </List.Item>
                )}
              />
            </div>
          ))}
          {faultTypes.length === 0 && (
            <Empty description='No fault types available' />
          )}
        </div>
      )}
    </Card>
  );
};
