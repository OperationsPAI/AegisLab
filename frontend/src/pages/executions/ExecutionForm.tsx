
import {
  CloseOutlined,
  DatabaseOutlined,
  FunctionOutlined,
  InfoCircleOutlined,
  PlayCircleOutlined,
  TagsOutlined,
} from '@ant-design/icons';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Alert,
  Button,
  Card,
  Col,
  Descriptions,
  Divider,
  Empty,
  Form,
  Input,
  message,
  Row,
  Select,
  Space,
  Switch,
  Tag,
  Typography,
} from 'antd';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';

import { containerApi } from '@/api/containers';
import { executionApi } from '@/api/executions';
import type { Container, Label } from '@/types/api';

const { Title, Text } = Typography;
const { Option } = Select;

interface ExecutionFormData {
  algorithm_name: string;
  algorithm_version: string;
  datapack_id: string;
  labels?: Label[];
}

const ExecutionForm = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [form] = Form.useForm<ExecutionFormData>();
  const [selectedAlgorithm, setSelectedAlgorithm] = useState<Container | null>(
    null
  );
  const [selectedVersion, setSelectedVersion] = useState<string>('');
  const [labelInput, setLabelInput] = useState('');
  const [labels, setLabels] = useState<Label[]>([]);

  // Fetch algorithms
  const { data: algorithmsData } = useQuery({
    queryKey: ['algorithms'],
    queryFn: () => containerApi.getContainers({ type: 'Algorithm' }),
  });

  // Fetch datapacks (simulated - in real app would fetch from datapack API)
  const { data: datapacksData } = useQuery({
    queryKey: ['datapacks'],
    queryFn: async () => {
      // Mock datapacks for now
      return {
        data: {
          data: [
            {
              id: 'dp-12345678',
              name: 'Test Datapack 1',
              created_at: '2024-01-01',
            },
            {
              id: 'dp-87654321',
              name: 'Test Datapack 2',
              created_at: '2024-01-02',
            },
            {
              id: 'dp-11223344',
              name: 'Production Datapack',
              created_at: '2024-01-03',
            },
          ],
        },
      };
    },
  });

  // Create execution mutation
  const createMutation = useMutation({
    mutationFn: (data: ExecutionFormData) =>
      executionApi.executeAlgorithm(data),
    onSuccess: (response) => {
      message.success('Execution started successfully');
      queryClient.invalidateQueries({ queryKey: ['executions'] });
      navigate(`/executions/${response.id}`);
    },
    onError: (error) => {
      message.error('Failed to start execution');
      console.error('Create execution error:', error);
    },
  });

  const handleAlgorithmChange = (algorithmId: string) => {
    const algorithm = algorithmsData?.data.find(
      (a) => a.id === Number(algorithmId)
    );
    setSelectedAlgorithm(algorithm || null);
    if (algorithm?.versions?.[0]) {
      setSelectedVersion(algorithm.versions[0].version);
      form.setFieldsValue({ algorithm_version: algorithm.versions[0].version });
    }
  };

  const handleVersionChange = (version: string) => {
    setSelectedVersion(version);
  };

  const handleSubmit = async (values: ExecutionFormData) => {
    const data = {
      ...values,
      labels,
    };

    createMutation.mutate(data);
  };

  const handleCancel = () => {
    navigate('/executions');
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

  if (!algorithmsData?.data.length) {
    return (
      <div style={{ padding: 24 }}>
        <Card>
          <Empty
            description='No algorithms available. Please create an algorithm container first.'
            image={Empty.PRESENTED_IMAGE_SIMPLE}
          >
            <Button type='primary' onClick={() => navigate('/containers/new')}>
              Create Algorithm
            </Button>
          </Empty>
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
            New Algorithm Execution
          </Title>
        </Space>
      </div>

      <Row gutter={[24, 24]}>
        <Col xs={24} lg={16}>
          <Card
            title={
              <Space>
                <PlayCircleOutlined />
                <span>Execution Configuration</span>
              </Space>
            }
          >
            <Form
              form={form}
              layout='vertical'
              onFinish={handleSubmit}
              initialValues={{
                algorithm_version: selectedVersion,
              }}
            >
              <Alert
                message='Execution Setup'
                description='Configure the algorithm execution by selecting an algorithm, datapack, and optional parameters.'
                type='info'
                showIcon
                icon={<InfoCircleOutlined />}
                style={{ marginBottom: 24 }}
              />

              <Form.Item
                label='Algorithm'
                name='algorithm_name'
                rules={[
                  { required: true, message: 'Please select an algorithm' },
                ]}
              >
                <Select
                  placeholder='Select algorithm'
                  size='large'
                  onChange={handleAlgorithmChange}
                >
                  {algorithmsData.data.map((algorithm) => (
                    <Option key={algorithm.id} value={algorithm.name}>
                      <Space>
                        <FunctionOutlined style={{ color: '#f59e0b' }} />
                        <div>
                          <div>{algorithm.name}</div>
                          <Text
                            type='secondary'
                            style={{ fontSize: '0.75rem' }}
                          >
                            {algorithm.versions?.length || 0} versions available
                          </Text>
                        </div>
                      </Space>
                    </Option>
                  ))}
                </Select>
              </Form.Item>

              {selectedAlgorithm && (
                <>
                  <Form.Item
                    label='Algorithm Version'
                    name='algorithm_version'
                    rules={[
                      {
                        required: true,
                        message: 'Please select algorithm version',
                      },
                    ]}
                  >
                    <Select
                      placeholder='Select version'
                      size='large'
                      onChange={handleVersionChange}
                      value={selectedVersion}
                    >
                      {selectedAlgorithm.versions?.map((version) => (
                        <Option key={version.id} value={version.version}>
                          <Space>
                            <Text>{version.version}</Text>
                            <Text
                              type='secondary'
                              style={{ fontSize: '0.75rem' }}
                            >
                              ({version.registry}/{version.repository}:
                              {version.tag})
                            </Text>
                          </Space>
                        </Option>
                      ))}
                    </Select>
                  </Form.Item>

                  <Card size='small' style={{ marginBottom: 24 }}>
                    <Descriptions column={2} size='small'>
                      <Descriptions.Item label='Type'>
                        {selectedAlgorithm.type}
                      </Descriptions.Item>
                      <Descriptions.Item label='Public'>
                        <Switch
                          checked={selectedAlgorithm.is_public}
                          disabled
                          size='small'
                        />
                      </Descriptions.Item>
                      <Descriptions.Item label='Versions'>
                        {selectedAlgorithm.versions?.length || 0}
                      </Descriptions.Item>
                      <Descriptions.Item label='Created'>
                        {new Date(
                          selectedAlgorithm.created_at
                        ).toLocaleDateString()}
                      </Descriptions.Item>
                    </Descriptions>
                  </Card>
                </>
              )}

              <Form.Item
                label='Datapack'
                name='datapack_id'
                rules={[
                  { required: true, message: 'Please select a datapack' },
                ]}
              >
                <Select placeholder='Select datapack' size='large'>
                  {datapacksData?.data.map((datapack) => (
                    <Option key={datapack.id} value={datapack.id}>
                      <Space>
                        <DatabaseOutlined style={{ color: '#3b82f6' }} />
                        <div>
                          <div>{datapack.name}</div>
                          <Text
                            type='secondary'
                            style={{ fontSize: '0.75rem' }}
                          >
                            ID: {datapack.id}
                          </Text>
                        </div>
                      </Space>
                    </Option>
                  ))}
                </Select>
              </Form.Item>

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
                        onClose={() => removeLabel(label.key)}
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
                    icon={<PlayCircleOutlined />}
                    loading={createMutation.isPending}
                  >
                    Start Execution
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
                <FunctionOutlined />
                <span>Execution Guide</span>
              </Space>
            }
          >
            <Space direction='vertical' style={{ width: '100%' }}>
              <div>
                <Text strong>Algorithm Selection:</Text>
                <ul style={{ marginTop: 8, marginBottom: 16 }}>
                  <li>
                    <Text>
                      Choose an RCA algorithm from your available containers
                    </Text>
                  </li>
                  <li>
                    <Text>
                      Select the appropriate version for your experiment
                    </Text>
                  </li>
                  <li>
                    <Text>
                      Ensure the algorithm is compatible with your datapack
                      format
                    </Text>
                  </li>
                </ul>
              </div>

              <Divider />

              <div>
                <Text strong>Datapack Selection:</Text>
                <ul style={{ marginTop: 8, marginBottom: 16 }}>
                  <li>
                    <Text>
                      Select a datapack that contains the data to analyze
                    </Text>
                  </li>
                  <li>
                    <Text>
                      Datapacks are generated from fault injection experiments
                    </Text>
                  </li>
                  <li>
                    <Text>
                      Ensure the datapack contains relevant traces, logs, or
                      metrics
                    </Text>
                  </li>
                </ul>
              </div>

              <Divider />

              <div>
                <Text strong>Execution Process:</Text>
                <ol style={{ marginTop: 8 }}>
                  <li>
                    <Text>Algorithm loads the selected datapack</Text>
                  </li>
                  <li>
                    <Text>RCA analysis is performed on the data</Text>
                  </li>
                  <li>
                    <Text>Results are generated and stored</Text>
                  </li>
                  <li>
                    <Text>Execution completes with detailed metrics</Text>
                  </li>
                </ol>
              </div>

              <Divider />

              <div>
                <Text strong>Labels:</Text>
                <Text
                  type='secondary'
                  style={{ display: 'block', marginTop: 4 }}
                >
                  Use labels to organize and categorize your executions. Format:
                  key:value
                </Text>
              </div>
            </Space>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default ExecutionForm;
