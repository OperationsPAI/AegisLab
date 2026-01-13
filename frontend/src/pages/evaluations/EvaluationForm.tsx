import {
  PlayCircleOutlined,
  CloseOutlined,
  FunctionOutlined,
  DatabaseOutlined,
  InfoCircleOutlined,
  BarChartOutlined,
  CheckCircleOutlined,
} from '@ant-design/icons'
import { useMutation, useQuery } from '@tanstack/react-query'
import {
  Form,
  Input,
  Select,
  Button,
  Space,
  Card,
  Typography,
  Row,
  Col,
  Alert,
  Descriptions,
  Empty,
  message,
  Divider,
  Progress,
  Statistic,
  Tag,
  Switch,
} from 'antd'
import { useState } from 'react'
import { useNavigate } from 'react-router-dom'

import { containerApi } from '@/api/containers'
import { datasetApi } from '@/api/datasets'
import { evaluationApi } from '@/api/evaluations'
import { executionApi } from '@/api/executions'
import type { DatapackEvaluationSpec } from '@/types/api'

const { Title, Text } = Typography
const { TextArea } = Input
const { Option } = Select

interface EvaluationFormData {
  algorithm_name: string
  algorithm_version: string
  datapack_id: string
  dataset_id?: string
  groundtruth_dataset_id?: string
  notes?: string
}

const EvaluationForm = () => {
  const navigate = useNavigate()
  const [form] = Form.useForm<EvaluationFormData>()
  const [selectedAlgorithm, setSelectedAlgorithm] = useState<string>('')
  const [selectedVersion, setSelectedVersion] = useState<string>('')
  const [selectedDatapack, setSelectedDatapack] = useState<string>('')
  const [selectedDataset, setSelectedDataset] = useState<string>('')
  const [evaluationType, setEvaluationType] = useState<'datapack' | 'dataset'>('datapack')
  const [isEvaluating, setIsEvaluating] = useState(false)
  const [evaluationProgress, setEvaluationProgress] = useState(0)

  // Fetch algorithms
  const { data: algorithmsData } = useQuery({
    queryKey: ['algorithms'],
    queryFn: () => containerApi.getContainers({ type: 'Algorithm' }),
  })

  // Fetch executions for datapacks
  const { data: executionsData } = useQuery({
    queryKey: ['executions'],
    queryFn: () => executionApi.getExecutions({ state: 2 }), // Only completed executions
  })

  // Fetch datasets
  const { data: datasetsData } = useQuery({
    queryKey: ['datasets'],
    queryFn: () => datasetApi.getDatasets(),
  })

  // Evaluate mutation
  const evaluateMutation = useMutation({
    mutationFn: (specs: DatapackEvaluationSpec[]) =>
      evaluationType === 'datapack'
        ? evaluationApi.evaluateDatapacks(specs)
        : evaluationApi.evaluateDatasets(specs),
    onSuccess: (data) => {
      message.success('Evaluation completed successfully!')
      navigate('/evaluations')
    },
    onError: (error) => {
      message.error('Failed to complete evaluation')
      console.error('Evaluation error:', error)
      setIsEvaluating(false)
      setEvaluationProgress(0)
    },
  })

  const handleAlgorithmChange = (algorithmName: string) => {
    setSelectedAlgorithm(algorithmName)
    const algorithm = algorithmsData?.data.find(a => a.name === algorithmName)
    if (algorithm?.versions?.[0]) {
      setSelectedVersion(algorithm.versions[0].version)
      form.setFieldsValue({ algorithm_version: algorithm.versions[0].version })
    }
  }

  const handleVersionChange = (version: string) => {
    setSelectedVersion(version)
  }

  const handleDatapackChange = (datapackId: string) => {
    setSelectedDatapack(datapackId)
  }

  const handleDatasetChange = (datasetId: string) => {
    setSelectedDataset(datasetId)
  }

  const handleEvaluationTypeChange = (type: 'datapack' | 'dataset') => {
    setEvaluationType(type)
    // Reset form fields when changing type
    form.setFieldsValue({
      datapack_id: undefined,
      dataset_id: undefined,
      groundtruth_dataset_id: undefined,
    })
    setSelectedDatapack('')
    setSelectedDataset('')
  }

  const handleSubmit = async (values: EvaluationFormData) => {
    if (!selectedAlgorithm || !selectedVersion) {
      message.error('Please select an algorithm and version')
      return
    }

    if (evaluationType === 'datapack' && !selectedDatapack) {
      message.error('Please select a datapack')
      return
    }

    if (evaluationType === 'dataset' && !selectedDataset) {
      message.error('Please select a dataset')
      return
    }

    const specs: DatapackEvaluationSpec[] = [{
      algorithm_name: selectedAlgorithm,
      algorithm_version: selectedVersion,
      datapack_id: evaluationType === 'datapack' ? selectedDatapack : undefined,
      dataset_id: evaluationType === 'dataset' ? selectedDataset : undefined,
      groundtruth_dataset_id: values.groundtruth_dataset_id,
    }]

    setIsEvaluating(true)
    setEvaluationProgress(0)

    // Simulate progress
    const progressInterval = setInterval(() => {
      setEvaluationProgress(prev => {
        if (prev >= 90) {
          clearInterval(progressInterval)
          return 90
        }
        return prev + 10
      })
    }, 500)

    try {
      await evaluateMutation.mutateAsync(specs)
      setEvaluationProgress(100)
    } finally {
      clearInterval(progressInterval)
      setIsEvaluating(false)
    }
  }

  const handleCancel = () => {
    navigate('/evaluations')
  }

  if (!algorithmsData?.data.length) {
    return (
      <div style={{ padding: 24 }}>
        <Card>
          <Empty
            description="No algorithms available. Please create an algorithm container first."
            image={Empty.PRESENTED_IMAGE_SIMPLE}
          >
            <Button type="primary" onClick={() => navigate('/containers/new')}>
              Create Algorithm
            </Button>
          </Empty>
        </Card>
      </div>
    )
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
            New Evaluation
          </Title>
        </Space>
      </div>

      <Row gutter={[24, 24]}>
        <Col xs={24} lg={16}>
          <Card
            title={
              <Space>
                <BarChartOutlined />
                <span>Evaluation Configuration</span>
              </Space>
            }
          >
            <Form
              form={form}
              layout="vertical"
              onFinish={handleSubmit}
              initialValues={{
                evaluation_type: 'datapack',
              }}
            >
              <Alert
                message="Evaluation Setup"
                description="Configure the evaluation by selecting an algorithm, data source, and optional parameters."
                type="info"
                showIcon
                icon={<InfoCircleOutlined />}
                style={{ marginBottom: 24 }}
              />

              <Form.Item
                label="Evaluation Type"
                name="evaluation_type"
                rules={[{ required: true, message: 'Please select evaluation type' }]}
              >
                <Select
                  placeholder="Select evaluation type"
                  size="large"
                  onChange={handleEvaluationTypeChange}
                  value={evaluationType}
                >
                  <Option value="datapack">
                    <Space>
                      <DatabaseOutlined style={{ color: '#3b82f6' }} />
                      <div>
                        <div>Datapack Evaluation</div>
                        <Text type="secondary" style={{ fontSize: '0.75rem' }}>
                          Evaluate algorithm performance on collected datapacks
                        </Text>
                      </div>
                    </Space>
                  </Option>
                  <Option value="dataset">
                    <Space>
                      <DatabaseOutlined style={{ color: '#10b981' }} />
                      <div>
                        <div>Dataset Evaluation</div>
                        <Text type="secondary" style={{ fontSize: '0.75rem' }}>
                          Evaluate algorithm performance on standard datasets
                        </Text>
                      </div>
                    </Space>
                  </Option>
                </Select>
              </Form.Item>

              <Form.Item
                label="Algorithm"
                name="algorithm_name"
                rules={[{ required: true, message: 'Please select an algorithm' }]}
              >
                <Select
                  placeholder="Select algorithm"
                  size="large"
                  onChange={handleAlgorithmChange}
                >
                  {algorithmsData.data.map(algorithm => (
                    <Option key={algorithm.id} value={algorithm.name}>
                      <Space>
                        <FunctionOutlined style={{ color: '#f59e0b' }} />
                        <div>
                          <div>{algorithm.name}</div>
                          <Text type="secondary" style={{ fontSize: '0.75rem' }}>
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
                    label="Algorithm Version"
                    name="algorithm_version"
                    rules={[{ required: true, message: 'Please select algorithm version' }]}
                  >
                    <Select
                      placeholder="Select version"
                      size="large"
                      onChange={handleVersionChange}
                      value={selectedVersion}
                    >
                      {algorithmsData.data
                        .find(a => a.name === selectedAlgorithm)
                        ?.versions?.map(version => (
                          <Option key={version.id} value={version.version}>
                            <Space>
                              <Text>{version.version}</Text>
                              <Text type="secondary" style={{ fontSize: '0.75rem' }}>
                                ({version.registry}/{version.repository}:{version.tag})
                              </Text>
                            </Space>
                          </Option>
                        ))}
                    </Select>
                  </Form.Item>

                  <Card size="small" style={{ marginBottom: 24 }}>
                    <Descriptions column={2} size="small">
                      <Descriptions.Item label="Type">Algorithm</Descriptions.Item>
                      <Descriptions.Item label="Public">
                        <Switch
                          checked={algorithmsData.data.find(a => a.name === selectedAlgorithm)?.is_public}
                          disabled
                          size="small"
                        />
                      </Descriptions.Item>
                      <Descriptions.Item label="Versions">
                        {algorithmsData.data.find(a => a.name === selectedAlgorithm)?.versions?.length || 0}
                      </Descriptions.Item>
                      <Descriptions.Item label="Created">
                        {new Date(algorithmsData.data.find(a => a.name === selectedAlgorithm)?.created_at || '').toLocaleDateString()}
                      </Descriptions.Item>
                    </Descriptions>
                  </Card>
                </>
              )}

              {evaluationType === 'datapack' && executionsData?.data && (
                <Form.Item
                  label="Datapack"
                  name="datapack_id"
                  rules={[{ required: true, message: 'Please select a datapack' }]}
                >
                  <Select
                    placeholder="Select datapack"
                    size="large"
                    onChange={handleDatapackChange}
                  >
                    {executionsData.data.map(execution => (
                      <Option key={execution.id} value={execution.datapack_id || ''}>
                        <Space>
                          <DatabaseOutlined style={{ color: '#3b82f6' }} />
                          <div>
                            <div>Datapack {execution.datapack_id?.substring(0, 8)}</div>
                            <Text type="secondary" style={{ fontSize: '0.75rem' }}>
                              From execution #{execution.id} - {execution.algorithm?.name}
                            </Text>
                          </div>
                        </Space>
                      </Option>
                    ))}
                  </Select>
                </Form.Item>
              )}

              {evaluationType === 'dataset' && datasetsData?.data.data && (
                <Form.Item
                  label="Dataset"
                  name="dataset_id"
                  rules={[{ required: true, message: 'Please select a dataset' }]}
                >
                  <Select
                    placeholder="Select dataset"
                    size="large"
                    onChange={handleDatasetChange}
                  >
                    {datasetsData.data.map(dataset => (
                      <Option key={dataset.id} value={String(dataset.id)}>
                        <Space>
                          <DatabaseOutlined style={{ color: '#10b981' }} />
                          <div>
                            <div>{dataset.name}</div>
                            <Text type="secondary" style={{ fontSize: '0.75rem' }}>
                              {dataset.type} - {dataset.versions?.length || 0} versions
                            </Text>
                          </div>
                        </Space>
                      </Option>
                    ))}
                  </Select>
                </Form.Item>
              )}

              <Form.Item
                label="Groundtruth Dataset (Optional)"
                name="groundtruth_dataset_id"
              >
                <Select
                  placeholder="Select groundtruth dataset (optional)"
                  size="large"
                  allowClear
                >
                  {datasetsData?.data.data.map(dataset => (
                    <Option key={dataset.id} value={String(dataset.id)}>
                      <Space>
                        <CheckCircleOutlined style={{ color: '#10b981' }} />
                        <div>
                          <div>{dataset.name}</div>
                          <Text type="secondary" style={{ fontSize: '0.75rem' }}>
                            {dataset.type}
                          </Text>
                        </div>
                      </Space>
                    </Option>
                  ))}
                </Select>
              </Form.Item>

              <Form.Item
                label="Notes"
                name="notes"
              >
                <TextArea
                  rows={3}
                  placeholder="Add any notes about this evaluation..."
                />
              </Form.Item>

              {isEvaluating && (
                <Card size="small" style={{ marginBottom: 24 }}>
                  <Space direction="vertical" style={{ width: '100%' }}>
                    <Text strong>Evaluation in progress...</Text>
                    <Progress
                      percent={evaluationProgress}
                      status="active"
                      strokeColor={{ '0%': '#108ee9', '100%': '#87d068' }}
                    />
                  </Space>
                </Card>
              )}

              <Form.Item>
                <Space>
                  <Button
                    type="primary"
                    htmlType="submit"
                    icon={<PlayCircleOutlined />}
                    loading={isEvaluating}
                    disabled={isEvaluating}
                  >
                    Start Evaluation
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
                <BarChartOutlined />
                <span>Evaluation Guide</span>
              </Space>
            }
          >
            <Space direction="vertical" style={{ width: '100%' }}>
              <div>
                <Text strong>Evaluation Types:</Text>
                <ul style={{ marginTop: 8, marginBottom: 16 }}>
                  <li>
                    <Text>
                      <strong>Datapack Evaluation:</strong> Test algorithm performance
                      on real experiment data collected from fault injections
                    </Text>
                  </li>
                  <li>
                    <Text>
                      <strong>Dataset Evaluation:</strong> Test algorithm performance
                      on standard benchmark datasets
                    </Text>
                  </li>
                </ul>
              </div>

              <Divider />

              <div>
                <Text strong>Metrics:</Text>
                <ul style={{ marginTop: 8, marginBottom: 16 }}>
                  <li>
                    <Text>
                      <strong>Precision:</strong> Accuracy of positive predictions
                    </Text>
                  </li>
                  <li>
                    <Text>
                      <strong>Recall:</strong> Coverage of actual positive cases
                    </Text>
                  </li>
                  <li>
                    <Text>
                      <strong>F1-Score:</strong> Harmonic mean of precision and recall
                    </Text>
                  </li>
                  <li>
                    <Text>
                      <strong>Accuracy:</strong> Overall correctness of predictions
                    </Text>
                  </li>
                </ul>
              </div>

              <Divider />

              <div>
                <Text strong>Best Practices:</Text>
                <ul style={{ marginTop: 8 }}>
                  <li>
                    <Text>Use consistent datasets for fair comparison</Text>
                  </li>
                  <li>
                    <Text>Include groundtruth data when available</Text>
                  </li>
                  <li>
                    <Text>Run multiple evaluations for statistical significance</Text>
                  </li>
                  <li>
                    <Text>Document evaluation parameters and conditions</Text>
                  </li>
                </ul>
              </div>

              <Divider />

              <div>
                <Text strong>Performance Benchmarks:</Text>
                <div style={{ marginTop: 8 }}>
                  <Tag color="green">Excellent: F1 ≥ 0.9</Tag>
                  <br />
                  <Tag color="orange">Good: 0.7 ≤ F1 &lt; 0.9</Tag>
                  <br />
                  <Tag color="red">Needs Improvement: F1 &lt; 0.7</Tag>
                </div>
              </div>
            </Space>
          </Card>

          <Card title="Quick Stats" style={{ marginTop: 16 }}>
            <Row gutter={[16, 16]}>
              <Col span={12}>
                <Statistic
                  title="Available Algorithms"
                  value={algorithmsData?.data.length || 0}
                  prefix={<FunctionOutlined />}
                  valueStyle={{ color: '#f59e0b' }}
                />
              </Col>
              <Col span={12}>
                <Statistic
                  title="Available Datapacks"
                  value={executionsData?.data.length || 0}
                  prefix={<DatabaseOutlined />}
                  valueStyle={{ color: '#3b82f6' }}
                />
              </Col>
              <Col span={12}>
                <Statistic
                  title="Available Datasets"
                  value={datasetsData?.data.length || 0}
                  prefix={<DatabaseOutlined />}
                  valueStyle={{ color: '#10b981' }}
                />
              </Col>
              <Col span={12}>
                <Statistic
                  title="Total Evaluations"
                  value="∞"
                  prefix={<BarChartOutlined />}
                  valueStyle={{ color: '#8b5cf6' }}
                />
              </Col>
            </Row>
          </Card>
        </Col>
      </Row>
    </div>
  )
}

export default EvaluationForm