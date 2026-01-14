import {
  CloseOutlined,
  DatabaseOutlined,
  FileTextOutlined,
  FormOutlined,
  GlobalOutlined,
  LineChartOutlined,
  SaveOutlined,
  TagsOutlined,
  UploadOutlined,
} from '@ant-design/icons';
import type { LabelItem } from '@rcabench/client';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Button,
  Card,
  Col,
  Divider,
  Form,
  Input,
  message,
  Row,
  Select,
  Space,
  Switch,
  Tag,
  Typography,
  Upload,
} from 'antd';
import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';


import { datasetApi } from '@/api/datasets';

// DatasetType enum for internal use
type DatasetType = 'Trace' | 'Log' | 'Metric';

const { Title, Text } = Typography;
const { TextArea } = Input;
const { Option } = Select;
const { Dragger } = Upload;

interface DatasetFormData {
  name: string;
  type: DatasetType;
  description?: string;
  is_public: boolean;
  labels?: LabelItem[];
}

const DatasetForm = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [form] = Form.useForm<DatasetFormData>();
  const { id } = useParams<{ id: string }>();
  const datasetId = id ? Number(id) : undefined;
  const [labelInput, setLabelInput] = useState('');
  const [labels, setLabels] = useState<LabelItem[]>([]);
  const [uploadFile, setUploadFile] = useState<File | null>(null);

  // Fetch dataset data if editing
  const { data: datasetData, isLoading } = useQuery({
    queryKey: ['dataset', datasetId],
    queryFn: () => datasetApi.getDataset(datasetId as number),
    enabled: !!datasetId,
  });

  // Set form data when editing
  useEffect(() => {
    if (datasetData) {
      form.setFieldsValue({
        name: datasetData.name,
        type: datasetData.type as DatasetType,
        description: datasetData.description,
        is_public: datasetData.is_public,
      });
      setLabels((datasetData.labels as LabelItem[]) || []);
    }
  }, [datasetData, form]);

  // Create or update mutation
  const createMutation = useMutation({
    mutationFn: (data: DatasetFormData) => datasetApi.createDataset(data),
    onSuccess: () => {
      message.success('Dataset created successfully');
      queryClient.invalidateQueries({ queryKey: ['datasets'] });
      navigate('/datasets');
    },
    onError: (error) => {
      message.error('Failed to create dataset');
      console.error('Create dataset error:', error);
    },
  });

  const updateMutation = useMutation({
    mutationFn: (data: Partial<DatasetFormData>) =>
      datasetApi.updateDataset(datasetId as number, data),
    onSuccess: () => {
      message.success('Dataset updated successfully');
      queryClient.invalidateQueries({ queryKey: ['datasets'] });
      queryClient.invalidateQueries({ queryKey: ['dataset', datasetId] });
      navigate('/datasets');
    },
    onError: (error) => {
      message.error('Failed to update dataset');
      console.error('Update dataset error:', error);
    },
  });

  const handleSubmit = async (values: DatasetFormData) => {
    const data = {
      ...values,
      labels,
    };

    if (datasetId) {
      updateMutation.mutate(data);
    } else {
      createMutation.mutate(data);
    }
  };

  const handleCancel = () => {
    navigate('/datasets');
  };

  const addLabel = () => {
    if (!labelInput.trim()) return;

    const [key, value] = labelInput.split(':').map((s) => s.trim());
    if (!key || !value) {
      message.warning('Please enter label in format: key:value');
      return;
    }

    if (labels.some((l) => l.key === key)) {
      message.warning('Label key already exists');
      return;
    }

    setLabels([...labels, { key, value }]);
    setLabelInput('');
  };

  const removeLabel = (key: string) => {
    setLabels(labels.filter((l) => l.key !== key));
  };

  const handleFileUpload = (file: File) => {
    setUploadFile(file);
    return false; // Prevent auto upload
  };

  if (isLoading && datasetId) {
    return (
      <div style={{ padding: 24 }}>
        <Card loading>
          <div style={{ minHeight: 400 }} />
        </Card>
      </div>
    );
  }

  return (
    <div style={{ padding: 24 }}>
      {/* Header */}
      <div style={{ marginBottom: 24 }}>
        <Space>
          <Button icon={<CloseOutlined />} onClick={handleCancel}>
            Back to List
          </Button>
          <Title level={2} style={{ margin: 0 }}>
            {datasetId ? 'Edit Dataset' : 'Create Dataset'}
          </Title>
        </Space>
      </div>

      <Row gutter={[24, 24]}>
        <Col xs={24} lg={16}>
          <Card
            title={
              <Space>
                <FormOutlined />
                <span>Dataset Information</span>
              </Space>
            }
          >
            <Form
              form={form}
              layout='vertical'
              onFinish={handleSubmit}
              initialValues={{
                is_public: false,
              }}
            >
              <Form.Item
                label='Dataset Name'
                name='name'
                rules={[
                  { required: true, message: 'Please enter dataset name' },
                  { min: 3, message: 'Name must be at least 3 characters' },
                  { max: 50, message: 'Name must be less than 50 characters' },
                  {
                    pattern: /^[a-zA-Z0-9-_ ]+$/,
                    message:
                      'Name can only contain letters, numbers, spaces, hyphens, and underscores',
                  },
                ]}
              >
                <Input
                  placeholder='Enter dataset name'
                  size='large'
                  disabled={!!datasetId}
                />
              </Form.Item>

              <Form.Item
                label='Dataset Type'
                name='type'
                rules={[
                  { required: true, message: 'Please select dataset type' },
                ]}
              >
                <Select
                  placeholder='Select dataset type'
                  size='large'
                  onChange={() => {
                    // Clear any existing validation errors
                    form.validateFields(['type']);
                  }}
                >
                  <Option value='Trace'>
                    <Space>
                      <DatabaseOutlined style={{ color: '#3b82f6' }} />
                      <div>
                        <div>Trace</div>
                        <Text type='secondary' style={{ fontSize: '0.75rem' }}>
                          Distributed tracing data with spans and timing
                          information
                        </Text>
                      </div>
                    </Space>
                  </Option>
                  <Option value='Log'>
                    <Space>
                      <FileTextOutlined style={{ color: '#10b981' }} />
                      <div>
                        <div>Log</div>
                        <Text type='secondary' style={{ fontSize: '0.75rem' }}>
                          Application and system logs for analysis
                        </Text>
                      </div>
                    </Space>
                  </Option>
                  <Option value='Metric'>
                    <Space>
                      <LineChartOutlined style={{ color: '#f59e0b' }} />
                      <div>
                        <div>Metric</div>
                        <Text type='secondary' style={{ fontSize: '0.75rem' }}>
                          Time-series metrics and monitoring data
                        </Text>
                      </div>
                    </Space>
                  </Option>
                </Select>
              </Form.Item>

              <Form.Item
                label='Description'
                name='description'
                rules={[
                  {
                    max: 1000,
                    message: 'Description must be less than 1000 characters',
                  },
                ]}
              >
                <TextArea rows={4} placeholder='Enter dataset description...' />
              </Form.Item>

              <Form.Item
                label='Visibility'
                name='is_public'
                valuePropName='checked'
                help='Public datasets can be used by other users in their projects'
              >
                <Switch
                  checkedChildren={<GlobalOutlined />}
                  unCheckedChildren={<GlobalOutlined />}
                />
              </Form.Item>

              {!datasetId && (
                <>
                  <Divider />
                  <Form.Item label='Upload Dataset File (Optional)'>
                    <Dragger
                      accept='.csv,.json,.parquet,.zip'
                      maxCount={1}
                      beforeUpload={handleFileUpload}
                      showUploadList={false}
                    >
                      <p className='ant-upload-drag-icon'>
                        <UploadOutlined
                          style={{ fontSize: 48, color: '#3b82f6' }}
                        />
                      </p>
                      <p className='ant-upload-text'>
                        Click or drag dataset file to this area
                      </p>
                      <p className='ant-upload-hint'>
                        Support for single file upload. File types: .csv, .json,
                        .parquet, .zip
                      </p>
                    </Dragger>
                    {uploadFile && (
                      <div style={{ marginTop: 8 }}>
                        <Text>Selected file: </Text>
                        <Text strong>{uploadFile.name}</Text>
                        <br />
                        <Text type='secondary'>
                          Size: {(uploadFile.size / 1024 / 1024).toFixed(2)} MB
                        </Text>
                      </div>
                    )}
                  </Form.Item>
                </>
              )}

              <Divider />

              <Form.Item label='Labels'>
                <Space direction='vertical' style={{ width: '100%' }}>
                  <Space.Compact style={{ width: '100%' }}>
                    <Input
                      placeholder='Enter label (key:value)'
                      value={labelInput}
                      onChange={(e) => setLabelInput(e.target.value)}
                      onPressEnter={addLabel}
                    />
                    <Button
                      type='primary'
                      onClick={addLabel}
                      icon={<TagsOutlined />}
                    >
                      Add
                    </Button>
                  </Space.Compact>
                  <div>
                    {labels.map((label) => (
                      <Tag
                        key={label.key}
                        closable
                        onClose={() => label.key && removeLabel(label.key)}
                        icon={<TagsOutlined />}
                        style={{ marginBottom: 8 }}
                      >
                        {label.key}: {label.value}
                      </Tag>
                    ))}
                  </div>
                </Space>
              </Form.Item>

              <Form.Item>
                <Space>
                  <Button
                    type='primary'
                    htmlType='submit'
                    icon={<SaveOutlined />}
                    loading={
                      createMutation.isPending || updateMutation.isPending
                    }
                  >
                    {datasetId ? 'Update Dataset' : 'Create Dataset'}
                  </Button>
                  <Button icon={<CloseOutlined />} onClick={handleCancel}>
                    Cancel
                  </Button>
                </Space>
              </Form.Item>
            </Form>
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card
            title={
              <Space>
                <DatabaseOutlined />
                <span>Dataset Guide</span>
              </Space>
            }
          >
            <Space direction='vertical' style={{ width: '100%' }}>
              <div>
                <Text strong>Dataset Types:</Text>
                <ul style={{ marginTop: 8, marginBottom: 16 }}>
                  <li>
                    <Text>
                      <strong>Trace:</strong> Distributed tracing data with
                      spans and timing information
                    </Text>
                  </li>
                  <li>
                    <Text>
                      <strong>Log:</strong> Application and system logs for
                      analysis
                    </Text>
                  </li>
                  <li>
                    <Text>
                      <strong>Metric:</strong> Time-series metrics and
                      monitoring data
                    </Text>
                  </li>
                </ul>
              </div>

              <Divider />

              <div>
                <Text strong>Best Practices:</Text>
                <ul style={{ marginTop: 8 }}>
                  <li>
                    <Text>Use descriptive names</Text>
                  </li>
                  <li>
                    <Text>Write clear descriptions</Text>
                  </li>
                  <li>
                    <Text>Tag datasets appropriately</Text>
                  </li>
                  <li>
                    <Text>Keep datasets versioned</Text>
                  </li>
                  <li>
                    <Text>Validate data formats</Text>
                  </li>
                </ul>
              </div>

              <Divider />

              <div>
                <Text strong>Labels:</Text>
                <Text
                  type='secondary'
                  style={{ display: 'block', marginTop: 4 }}
                >
                  Use labels to organize and categorize your datasets. Format:
                  key:value
                </Text>
              </div>

              <Divider />

              <div>
                <Text strong>File Upload:</Text>
                <Text
                  type='secondary'
                  style={{ display: 'block', marginTop: 4 }}
                >
                  You can upload dataset files in various formats. The system
                  will automatically detect the format and create a version.
                </Text>
              </div>
            </Space>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default DatasetForm;
