
import {
  InfoCircleOutlined,
  ReloadOutlined,
  SaveOutlined,
} from '@ant-design/icons';
import {
  Button,
  Card,
  Col,
  Divider,
  Form,
  Input,
  InputNumber,
  Row,
  Select,
  Slider,
  Space,
  Switch,
  Tag,
  Tooltip,
} from 'antd';
import { useEffect, useState } from 'react';

import type { FaultType } from '../../../types/api';

import './FaultConfigPanel.css';

const { Option } = Select;

interface FaultParameter {
  name: string;
  type: 'string' | 'number' | 'boolean' | 'select' | 'range';
  label: string;
  description?: string;
  required?: boolean;
  default?: string | number | boolean;
  options?: string[];
  min?: number;
  max?: number;
  step?: number;
}

interface FaultConfigPanelProps {
  fault: FaultType;
  onConfigChange: (config: Record<string, string | number | boolean>) => void;
}

export const FaultConfigPanel: React.FC<FaultConfigPanelProps> = ({
  fault,
  onConfigChange,
}: FaultConfigPanelProps) => {
  const [form] = Form.useForm();
  const [config, setConfig] = useState<
    Record<string, string | number | boolean>
  >({});
  const [presets, setPresets] = useState<string[]>([]);

  // Initialize form with fault parameters
  useEffect(() => {
    if (fault && fault.parameters) {
      const initialValues: Record<string, string | number | boolean> = {};
      fault.parameters.forEach((param: FaultParameter) => {
        initialValues[param.name] =
          param.default ?? getDefaultValue(param.type);
      });
      form.setFieldsValue(initialValues);
      setConfig(initialValues);
      onConfigChange(initialValues);
    }
  }, [fault, form, onConfigChange]);

  const getDefaultValue = (type: string): string | number | boolean => {
    switch (type) {
      case 'boolean':
        return false;
      case 'number':
      case 'range':
        return 0;
      case 'select':
        return '';
      default:
        return '';
    }
  };

  const handleFormChange = (
    _: Record<string, unknown>,
    values: Record<string, unknown>
  ) => {
    setConfig(values as Record<string, string | number | boolean>);
    onConfigChange(values as Record<string, string | number | boolean>);
  };

  const handleSavePreset = () => {
    const presetName = prompt('Enter preset name:');
    if (presetName) {
      setPresets([...presets, presetName]);
      localStorage.setItem(
        `fault-preset-${fault.id}-${presetName}`,
        JSON.stringify(config)
      );
    }
  };

  const handleLoadPreset = (presetName: string) => {
    const saved = localStorage.getItem(
      `fault-preset-${fault.id}-${presetName}`
    );
    if (saved) {
      const preset = JSON.parse(saved);
      form.setFieldsValue(preset);
      setConfig(preset);
      onConfigChange(preset);
    }
  };

  const handleReset = () => {
    form.resetFields();
    const initialValues: Record<string, string | number | boolean> = {};
    fault.parameters?.forEach((param: FaultParameter) => {
      initialValues[param.name] = param.default ?? getDefaultValue(param.type);
    });
    form.setFieldsValue(initialValues);
    setConfig(initialValues);
    onConfigChange(initialValues);
  };

  const renderParameterField = (param: FaultParameter) => {
    const commonProps = {
      name: param.name,
      label: (
        <span>
          {param.label}
          {param.description && (
            <Tooltip title={param.description}>
              <InfoCircleOutlined style={{ marginLeft: 8, color: '#8c8c8c' }} />
            </Tooltip>
          )}
        </span>
      ),
      rules: param.required
        ? [{ required: true, message: `${param.label} is required` }]
        : [],
    };

    switch (param.type) {
      case 'string':
        return (
          <Form.Item key={param.name} {...commonProps}>
            <Input placeholder={`Enter ${param.label.toLowerCase()}`} />
          </Form.Item>
        );

      case 'number':
        return (
          <Form.Item key={param.name} {...commonProps}>
            <InputNumber
              style={{ width: '100%' }}
              min={param.min}
              max={param.max}
              step={param.step}
              placeholder={`Enter ${param.label.toLowerCase()}`}
            />
          </Form.Item>
        );

      case 'boolean':
        return (
          <Form.Item key={param.name} {...commonProps} valuePropName='checked'>
            <Switch />
          </Form.Item>
        );

      case 'select':
        return (
          <Form.Item key={param.name} {...commonProps}>
            <Select placeholder={`Select ${param.label.toLowerCase()}`}>
              {param.options?.map((option) => (
                <Option key={option} value={option}>
                  {option}
                </Option>
              ))}
            </Select>
          </Form.Item>
        );

      case 'range':
        return (
          <Form.Item key={param.name} {...commonProps}>
            <Row gutter={16}>
              <Col span={18}>
                <Slider
                  min={param.min ?? 0}
                  max={param.max ?? 100}
                  step={param.step ?? 1}
                  marks={{
                    [param.min ?? 0]: param.min?.toString() ?? '0',
                    [param.max ?? 100]: param.max?.toString() ?? '100',
                  }}
                />
              </Col>
              <Col span={6}>
                <Form.Item noStyle>
                  <InputNumber
                    min={param.min}
                    max={param.max}
                    step={param.step}
                    style={{ width: '100%' }}
                  />
                </Form.Item>
              </Col>
            </Row>
          </Form.Item>
        );

      default:
        return null;
    }
  };

  if (!fault) {
    return (
      <Card title='Fault Configuration' className='fault-config-panel'>
        <div className='empty-config'>
          Select a fault from the canvas to configure its parameters
        </div>
      </Card>
    );
  }

  return (
    <Card
      title={
        <div className='fault-config-header'>
          <span>Configure: {fault.name}</span>
          <Tag color='blue'>{fault.type}</Tag>
        </div>
      }
      className='fault-config-panel'
      extra={
        <Space>
          <Button
            icon={<SaveOutlined />}
            size='small'
            onClick={handleSavePreset}
          >
            Save Preset
          </Button>
          <Button icon={<ReloadOutlined />} size='small' onClick={handleReset}>
            Reset
          </Button>
        </Space>
      }
    >
      {presets.length > 0 && (
        <div className='preset-section'>
          <Divider orientation='left'>Presets</Divider>
          <Space wrap>
            {presets.map((preset) => (
              <Button
                key={preset}
                size='small'
                onClick={() => handleLoadPreset(preset)}
              >
                {preset}
              </Button>
            ))}
          </Space>
        </div>
      )}

      <Form
        form={form}
        layout='vertical'
        onValuesChange={handleFormChange}
        className='fault-config-form'
      >
        {fault.parameters && fault.parameters.length > 0 ? (
          <Row gutter={24}>
            {fault.parameters.map((param: FaultParameter) => (
              <Col span={12} key={param.name}>
                {renderParameterField(param)}
              </Col>
            ))}
          </Row>
        ) : (
          <div className='no-parameters'>
            This fault type has no configurable parameters
          </div>
        )}
      </Form>

      {fault.description && (
        <div className='fault-description'>
          <Divider orientation='left'>Description</Divider>
          <p>{fault.description}</p>
        </div>
      )}
    </Card>
  );
};
