import {
  ClockCircleOutlined,
  CloudServerOutlined,
  DatabaseOutlined,
  DeleteOutlined,
  DisconnectOutlined,
  PauseCircleOutlined,
  QuestionCircleOutlined,
  SettingOutlined,
  StopOutlined,
  ThunderboltOutlined,
} from '@ant-design/icons';
import { Button, Tag, Tooltip } from 'antd';
import { memo } from 'react';
import { Handle, type NodeProps, Position } from 'reactflow';


import type { FaultType } from '../../../types/api';

const faultTypeIcons: Record<string, React.ReactNode> = {
  cpu: <ThunderboltOutlined />,
  memory: <DatabaseOutlined />,
  disk: <CloudServerOutlined />,
  network: <DisconnectOutlined />,
  process: <StopOutlined />,
  io: <PauseCircleOutlined />,
  time: <ClockCircleOutlined />,
  default: <QuestionCircleOutlined />,
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

interface FaultNodeData {
  fault: FaultType;
  onDelete: (nodeId: string) => void;
  onConfigure: (fault: FaultType) => void;
}

export const FaultNode = memo<NodeProps<FaultNodeData>>(
  ({ data, id, selected }) => {
    const { fault, onDelete, onConfigure } = data;

    const handleDelete = (e: React.MouseEvent) => {
      e.stopPropagation();
      onDelete(id);
    };

    const handleConfigure = (e: React.MouseEvent) => {
      e.stopPropagation();
      onConfigure(fault);
    };

    const getFaultIcon = () => {
      const key = fault.category?.toLowerCase() || 'default';
      return faultTypeIcons[key] || faultTypeIcons.default;
    };

    const getFaultColor = () => {
      const key = fault.category?.toLowerCase() || 'default';
      return faultTypeColors[key] || faultTypeColors.default;
    };

    return (
      <div className={`fault-node ${selected ? 'selected' : ''}`}>
        <Handle type='target' position={Position.Top} />

        <div className='fault-node-header'>
          <div className='fault-node-title-wrapper'>
            <span className='fault-node-icon'>{getFaultIcon()}</span>
            <span className='fault-node-title'>{fault.name}</span>
          </div>
          <div className='fault-node-actions'>
            <Tooltip title='Configure'>
              <Button
                type='text'
                size='small'
                icon={<SettingOutlined />}
                onClick={handleConfigure}
                className='fault-node-action'
              />
            </Tooltip>
            <Tooltip title='Delete'>
              <Button
                type='text'
                size='small'
                danger
                icon={<DeleteOutlined />}
                onClick={handleDelete}
                className='fault-node-action'
              />
            </Tooltip>
          </div>
        </div>

        {fault.description && (
          <div className='fault-node-description'>{fault.description}</div>
        )}

        <div className='fault-node-footer'>
          <Tag color={getFaultColor()}>
            {fault.type}
          </Tag>
          {fault.parameters && (
            <span className='fault-node-params'>
              {fault.parameters.length} params
            </span>
          )}
        </div>

        <Handle type='source' position={Position.Bottom} />
      </div>
    );
  }
);

FaultNode.displayName = 'FaultNode';
