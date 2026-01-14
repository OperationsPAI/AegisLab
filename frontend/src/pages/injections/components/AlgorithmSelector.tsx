
import { InfoCircleOutlined, SettingOutlined } from '@ant-design/icons';
import type { ContainerResp } from '@rcabench/client';
import {
  Button,
  Empty,
  Form,
  List,
  Modal,
  Select,
  Space,
  Tag,
  Tooltip,
} from 'antd';
import { useState } from 'react';

import './AlgorithmSelector.css';

const { Option } = Select;

interface AlgorithmSelectorProps {
  algorithms: ContainerResp[];
  value: number[];
  onChange: (algorithms: number[]) => void;
}

export const AlgorithmSelector: React.FC<AlgorithmSelectorProps> = ({
  algorithms,
  value,
  onChange,
}) => {
  const [selectedAlgorithms, setSelectedAlgorithms] = useState<number[]>(value);
  const [showDetails, setShowDetails] = useState(false);

  const handleChange = (newValue: number[]) => {
    setSelectedAlgorithms(newValue);
    onChange(newValue);
  };

  const getAlgorithmTags = (algorithm: ContainerResp) => {
    const tags: string[] = [];
    if (algorithm.type) tags.push(algorithm.type);
    return tags;
  };

  const tagRender = (props: {
    label: React.ReactNode;
    value: number;
    closable: boolean;
    onClose: () => void;
  }) => {
    const { label, closable, onClose } = props;

    const onPreventMouseDown = (event: React.MouseEvent) => {
      event.preventDefault();
      event.stopPropagation();
    };

    return (
      <Tag
        color='default'
        onMouseDown={onPreventMouseDown}
        closable={closable}
        onClose={onClose}
        style={{ marginRight: 3 }}
      >
        {label}
      </Tag>
    );
  };

  return (
    <>
      <Form.Item
        name={['container_config', 'algorithm_container_ids']}
        label={
          <Space>
            <span>Algorithm Containers</span>
            <Tooltip title='Select RCA algorithms to run during the experiment'>
              <InfoCircleOutlined style={{ color: '#8c8c8c' }} />
            </Tooltip>
            <Button
              type='link'
              size='small'
              icon={<SettingOutlined />}
              onClick={() => setShowDetails(true)}
            >
              Details
            </Button>
          </Space>
        }
      >
        <Select
          mode='multiple'
          placeholder='Select algorithm containers'
          value={selectedAlgorithms}
          onChange={handleChange}
          tagRender={tagRender}
          disabled={algorithms.length === 0}
          maxTagCount={3}
          maxTagPlaceholder={(omittedValues) => (
            <span>+{omittedValues.length} more</span>
          )}
        >
          {algorithms.map((algorithm) => (
            <Option key={algorithm.id} value={algorithm.id}>
              <div className='algorithm-option'>
                <div className='algorithm-option-main'>
                  <span className='algorithm-option-name'>
                    {algorithm.name}
                  </span>
                  <Space size='small'>
                    {getAlgorithmTags(algorithm).map((tag) => (
                      <Tag key={tag}>{tag}</Tag>
                    ))}
                  </Space>
                </div>
              </div>
            </Option>
          ))}
        </Select>
      </Form.Item>

      <Modal
        title='Algorithm Containers Details'
        open={showDetails}
        onCancel={() => setShowDetails(false)}
        footer={null}
        width={800}
      >
        <List
          dataSource={algorithms}
          renderItem={(algorithm) => (
            <List.Item
              key={algorithm.id}
              actions={[
                <Button
                  key='select'
                  type={
                    algorithm.id !== undefined && selectedAlgorithms.includes(algorithm.id)
                      ? 'default'
                      : 'primary'
                  }
                  size='small'
                  onClick={() => {
                    if (algorithm.id === undefined) return;
                    if (selectedAlgorithms.includes(algorithm.id)) {
                      handleChange(
                        selectedAlgorithms.filter((id) => id !== algorithm.id)
                      );
                    } else {
                      handleChange([...selectedAlgorithms, algorithm.id]);
                    }
                  }}
                >
                  {algorithm.id !== undefined && selectedAlgorithms.includes(algorithm.id)
                    ? 'Remove'
                    : 'Select'}
                </Button>,
              ]}
            >
              <List.Item.Meta
                title={
                  <Space>
                    <span>{algorithm.name}</span>
                  </Space>
                }
                description={
                  <div>
                    <div className='algorithm-meta'>
                      <Space size='large'>
                        <span>Type: {algorithm.type}</span>
                        <span>
                          Created:{' '}
                          {new Date(
                            algorithm.created_at || ''
                          ).toLocaleDateString()}
                        </span>
                      </Space>
                    </div>
                  </div>
                }
              />
            </List.Item>
          )}
          locale={{
            emptyText: (
              <Empty
                description='No algorithm containers available'
                image={Empty.PRESENTED_IMAGE_SIMPLE}
              />
            ),
          }}
        />
      </Modal>
    </>
  );
};
