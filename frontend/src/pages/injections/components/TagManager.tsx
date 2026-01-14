
import {
  InfoCircleOutlined,
  PlusOutlined,
  TagsOutlined,
} from '@ant-design/icons';
import { Button, Form, Input, message, Space, Tag, Tooltip } from 'antd';
import type { InputRef } from 'antd/es/input';
import { useEffect, useRef, useState } from 'react';

import './TagManager.css';

interface TagManagerProps {
  value: string[];
  onChange: (tags: string[]) => void;
  maxTags?: number;
  maxLength?: number;
}

export const TagManager: React.FC<TagManagerProps> = ({
  value = [],
  onChange,
  maxTags = 10,
  maxLength = 50,
}) => {
  const [inputVisible, setInputVisible] = useState(false);
  const [inputValue, setInputValue] = useState('');
  const [editInputIndex, setEditInputIndex] = useState(-1);
  const [editInputValue, setEditInputValue] = useState('');
  const inputRef = useRef<InputRef>(null);
  const editInputRef = useRef<InputRef>(null);

  useEffect(() => {
    if (inputVisible) {
      inputRef.current?.focus();
    }
  }, [inputVisible]);

  useEffect(() => {
    if (editInputIndex !== -1) {
      editInputRef.current?.focus();
    }
  }, [editInputIndex]);

  const handleClose = (removedTag: string) => {
    const newTags = value.filter((tag) => tag !== removedTag);
    onChange(newTags);
    message.success(`Tag "${removedTag}" removed`);
  };

  const showInput = () => {
    if (value.length >= maxTags) {
      message.warning(`Maximum ${maxTags} tags allowed`);
      return;
    }
    setInputVisible(true);
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setInputValue(e.target.value);
  };

  const handleInputConfirm = () => {
    if (inputValue && value.indexOf(inputValue) === -1) {
      if (inputValue.length > maxLength) {
        message.error(`Tag cannot exceed ${maxLength} characters`);
        return;
      }
      onChange([...value, inputValue]);
      message.success(`Tag "${inputValue}" added`);
    }
    setInputVisible(false);
    setInputValue('');
  };

  const handleEditInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setEditInputValue(e.target.value);
  };

  const handleEditInputConfirm = () => {
    const newTags = [...value];
    if (editInputValue && editInputValue !== value[editInputIndex]) {
      if (editInputValue.length > maxLength) {
        message.error(`Tag cannot exceed ${maxLength} characters`);
        return;
      }
      newTags[editInputIndex] = editInputValue;
      onChange(newTags);
      message.success('Tag updated');
    }
    setEditInputIndex(-1);
    setEditInputValue('');
  };

  const handleKeyDown = (
    e: React.KeyboardEvent<HTMLInputElement>,
    isEdit = false
  ) => {
    if (e.key === 'Enter') {
      if (isEdit) {
        handleEditInputConfirm();
      } else {
        handleInputConfirm();
      }
    } else if (e.key === 'Escape') {
      if (isEdit) {
        setEditInputIndex(-1);
        setEditInputValue('');
      } else {
        setInputVisible(false);
        setInputValue('');
      }
    }
  };


  const handleQuickAdd = (tag: string) => {
    if (value.indexOf(tag) === -1) {
      onChange([...value, tag]);
      message.success(`Tag "${tag}" added`);
    } else {
      message.info(`Tag "${tag}" already exists`);
    }
  };

  // Common tags for fault injection experiments
  const commonTags = [
    'performance',
    'reliability',
    'chaos',
    'network',
    'cpu',
    'memory',
    'disk',
    'io',
    'critical',
    'high-priority',
    'testing',
    'staging',
    'production',
    'baseline',
    'regression',
  ];

  return (
    <Form.Item
      label={
        <Space>
          <TagsOutlined />
          <span>Tags</span>
          <Tooltip title='Add tags to categorize and search for this injection'>
            <InfoCircleOutlined style={{ color: '#8c8c8c' }} />
          </Tooltip>
        </Space>
      }
    >
      <div className='tag-manager'>
        <Space wrap className='tag-list'>
          {value.map((tag, index) => {
            if (editInputIndex === index) {
              return (
                <Input
                  ref={editInputRef}
                  key={tag}
                  size='small'
                  className='tag-input tag-edit-input'
                  value={editInputValue}
                  onChange={handleEditInputChange}
                  onBlur={handleEditInputConfirm}
                  onKeyDown={(e) => handleKeyDown(e, true)}
                  maxLength={maxLength}
                />
              );
            }

            const isLongTag = tag.length > 20;
            const tagElem = (
              <Tag
                key={tag}
                closable
                className='tag-item'
                onClose={() => handleClose(tag)}
              >
                <span
                  onDoubleClick={(e) => {
                    if (index !== 0) {
                      setEditInputIndex(index);
                      setEditInputValue(tag);
                      e.preventDefault();
                    }
                  }}
                >
                  {isLongTag ? `${tag.slice(0, 20)}...` : tag}
                </span>
              </Tag>
            );

            return isLongTag ? (
              <Tooltip title={tag} key={tag}>
                {tagElem}
              </Tooltip>
            ) : (
              tagElem
            );
          })}

          {inputVisible && (
            <Input
              ref={inputRef}
              type='text'
              size='small'
              className='tag-input'
              value={inputValue}
              onChange={handleInputChange}
              onBlur={handleInputConfirm}
              onKeyDown={handleKeyDown}
              placeholder='Enter tag'
              maxLength={maxLength}
              style={{ width: 100 }}
            />
          )}

          {!inputVisible && (
            <Tag className='site-tag-plus' onClick={showInput}>
              <PlusOutlined /> New Tag
            </Tag>
          )}
        </Space>

        <div className='common-tags'>
          <div className='common-tags-label'>Quick add:</div>
          <Space wrap size='small'>
            {commonTags.slice(0, 8).map((tag) => (
              <Button
                key={tag}
                size='small'
                type='text'
                onClick={() => handleQuickAdd(tag)}
                disabled={value.includes(tag)}
              >
                {tag}
              </Button>
            ))}
          </Space>
        </div>

        <div className='tag-stats'>
          {value.length} / {maxTags} tags
        </div>
      </div>
    </Form.Item>
  );
};
