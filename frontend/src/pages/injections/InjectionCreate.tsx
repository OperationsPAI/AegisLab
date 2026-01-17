import type {
  ChaosNode,
  ContainerResp,
  ContainerSpec,
  LabelItem,
  ProjectResp,
  SubmitInjectionReq,
} from '@rcabench/client';
import { useQuery } from '@tanstack/react-query';
import {
  Button,
  Card,
  Col,
  Form,
  Input,
  InputNumber,
  message,
  Row,
  Select,
  Space,
} from 'antd';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';


import { containerApi } from '../../api/containers';
import { injectionApi } from '../../api/injections';
import { projectApi } from '../../api/projects';
import type { FaultType } from '../../types/api';

import { AlgorithmSelector } from './components/AlgorithmSelector';
import { FaultConfigPanel } from './components/FaultConfigPanel';
import { FaultTypePanel } from './components/FaultTypePanel';
import { TagManager } from './components/TagManager';
import { VisualCanvas } from './components/VisualCanvas';


import './InjectionCreate.css';

const { Option } = Select;
const { TextArea } = Input;

interface InjectionFormData {
  project_id: number;
  name: string;
  description?: string;
  container_config: {
    pedestal_container_id: number;
    benchmark_container_id: number;
    algorithm_container_ids: number[];
  };
  fault_matrix: FaultType[][];
  experiment_params: {
    duration: number;
    interval: number;
    parallel: boolean;
  };
  tags?: string[];
}

const InjectionCreate: React.FC = () => {
  const navigate = useNavigate();
  const [form] = Form.useForm<InjectionFormData>();
  const [selectedProject, setSelectedProject] = useState<number | null>(null);
  const [selectedFault, setSelectedFault] = useState<FaultType | null>(null);
  const [faultMatrix, setFaultMatrix] = useState<FaultType[][]>([]);
  const [selectedAlgorithms, setSelectedAlgorithms] = useState<number[]>([]);
  const [tags, setTags] = useState<string[]>([]);

  // Fetch projects
  const { data: projects = [], isLoading: projectsLoading } = useQuery({
    queryKey: ['projects'],
    queryFn: () => projectApi.getProjects({ page: 1, size: 50 }),
    select: (data: any) => data.items || [],
  });

  // Fetch containers when project is selected
  const { data: containers = [], isLoading: containersLoading } = useQuery({
    queryKey: ['containers', selectedProject],
    queryFn: () => {
      if (!selectedProject) {
        return Promise.resolve({ items: [] });
      }
      return containerApi.getContainers({
        page: 1,
        size: 50,
      });
    },
    enabled: !!selectedProject,
    select: (data: any) => data.items || [],
  });

  // Group containers by type (API returns type as string: "algorithm", "benchmark", "pedestal")
  const groupedContainers = containers.reduce(
    (
      acc: {
        pedestals: any[];
        benchmarks: any[];
        algorithms: any[];
      },
      container: any
    ) => {
      if (container.type === 'pedestal') {
        acc.pedestals.push(container);
      } else if (container.type === 'benchmark') {
        acc.benchmarks.push(container);
      } else if (container.type === 'algorithm') {
        acc.algorithms.push(container);
      }
      return acc;
    },
    { pedestals: [], benchmarks: [], algorithms: [] }
  );

  const handleProjectChange = (projectId: number) => {
    setSelectedProject(projectId);
    form.setFieldsValue({
      container_config: {
        pedestal_container_id: undefined,
        benchmark_container_id: undefined,
        algorithm_container_ids: [],
      },
    });
  };

  const handleFaultSelect = (fault: FaultType) => {
    setSelectedFault(fault);
  };

  const handleFaultMatrixChange = (matrix: FaultType[][]) => {
    setFaultMatrix(matrix);
  };

  const handleAlgorithmChange = (algorithms: number[]) => {
    setSelectedAlgorithms(algorithms);
  };

  const handleTagChange = (newTags: string[]) => {
    setTags(newTags);
  };

  // Helper function to find container by ID
  const findContainerById = (
    containerId: number
  ): ContainerResp | undefined => {
    return containers.find((c: ContainerResp) => c.id === containerId);
  };

  // Helper function to convert container to ContainerSpec
  // Note: ContainerResp doesn't include version, using 'latest' as default
  const toContainerSpec = (container: ContainerResp): ContainerSpec => ({
    name: container.name || '',
    version: 'latest', // Default to 'latest' as version is not available in ContainerResp
  });

  // Helper function to convert FaultType[][] to ChaosNode[][]
  const toSpecs = (matrix: FaultType[][]): ChaosNode[][] => {
    return matrix.map((batch) =>
      batch.map((fault) => ({
        name: fault.name,
        description: fault.type,
        // Convert fault parameters to ChaosNode children if needed
        children: fault.parameters?.reduce(
          (acc, param) => {
            if (param && typeof param === 'object' && 'name' in param) {
              acc[(param as { name: string }).name] = {
                name: (param as { name: string }).name,
                value: (param as { value?: number }).value,
              };
            }
            return acc;
          },
          {} as { [key: string]: ChaosNode }
        ),
      }))
    );
  };

  const handleSubmit = async (values: InjectionFormData) => {
    try {
      // Find the selected project
      const selectedProjectData = projects.find(
        (p: ProjectResp) => p.id === values.project_id
      );
      if (!selectedProjectData) {
        message.error('Please select a project');
        return;
      }

      // Find the selected containers
      const pedestalContainer = findContainerById(
        values.container_config.pedestal_container_id
      );
      const benchmarkContainer = findContainerById(
        values.container_config.benchmark_container_id
      );

      if (!pedestalContainer || !benchmarkContainer) {
        message.error('Please select pedestal and benchmark containers');
        return;
      }

      // Build algorithm specs
      const algorithmSpecs: ContainerSpec[] = selectedAlgorithms
        .map((id) => findContainerById(id))
        .filter((c): c is ContainerResp => c !== undefined)
        .map(toContainerSpec);

      // Convert tags to LabelItem format
      const labels: LabelItem[] = tags.map((tag) => ({
        key: tag,
        value: tag,
      }));

      // Build the SDK request
      const payload: SubmitInjectionReq = {
        project_name: selectedProjectData.name || '',
        pedestal: toContainerSpec(pedestalContainer),
        benchmark: toContainerSpec(benchmarkContainer),
        algorithms: algorithmSpecs.length > 0 ? algorithmSpecs : undefined,
        interval: Math.ceil(values.experiment_params.duration / 60), // Convert seconds to minutes
        pre_duration: Math.ceil(values.experiment_params.interval / 60), // Pre-injection duration in minutes
        labels: labels.length > 0 ? labels : undefined,
        specs: toSpecs(faultMatrix),
      };

      await injectionApi.submitInjection(payload);
      message.success('Fault injection submitted successfully');
      navigate('/injections');
    } catch (error) {
      message.error('Failed to submit fault injection');
      console.error('Submit injection error:', error);
    }
  };

  return (
    <div className='injection-create'>
      <Form
        form={form}
        layout='vertical'
        onFinish={handleSubmit}
        initialValues={{
          experiment_params: {
            duration: 300,
            interval: 60,
            parallel: false,
          },
        }}
      >
          <Row gutter={24}>
            {/* Left Panel - Basic Configuration */}
            <Col span={8}>
              <Card title='Basic Configuration' className='injection-card'>
                <Form.Item
                  name='project_id'
                  label='Project'
                  rules={[
                    { required: true, message: 'Please select a project' },
                  ]}
                >
                  <Select
                    placeholder='Select project'
                    loading={projectsLoading}
                    onChange={handleProjectChange}
                  >
                    {projects.map((project: ProjectResp) => (
                      <Option key={project.id} value={project.id}>
                        {project.name}
                      </Option>
                    ))}
                  </Select>
                </Form.Item>

                <Form.Item
                  name='name'
                  label='Injection Name'
                  rules={[
                    { required: true, message: 'Please input injection name' },
                  ]}
                >
                  <Input placeholder='Enter injection name' />
                </Form.Item>

                <Form.Item name='description' label='Description'>
                  <TextArea rows={3} placeholder='Enter description' />
                </Form.Item>

                <Form.Item
                  name={['container_config', 'pedestal_container_id']}
                  label='Pedestal Container'
                  rules={[
                    {
                      required: true,
                      message: 'Please select pedestal container',
                    },
                  ]}
                >
                  <Select
                    placeholder='Select pedestal container'
                    loading={containersLoading}
                    disabled={!selectedProject}
                  >
                    {groupedContainers.pedestals.map((container: ContainerResp) => (
                      <Option key={container.id} value={container.id}>
                        {container.name}
                      </Option>
                    ))}
                  </Select>
                </Form.Item>

                <Form.Item
                  name={['container_config', 'benchmark_container_id']}
                  label='Benchmark Container'
                  rules={[
                    {
                      required: true,
                      message: 'Please select benchmark container',
                    },
                  ]}
                >
                  <Select
                    placeholder='Select benchmark container'
                    loading={containersLoading}
                    disabled={!selectedProject}
                  >
                    {groupedContainers.benchmarks.map(
                      (container: ContainerResp) => (
                        <Option key={container.id} value={container.id}>
                          {container.name}
                        </Option>
                      )
                    )}
                  </Select>
                </Form.Item>

                <AlgorithmSelector
                  algorithms={groupedContainers.algorithms}
                  value={selectedAlgorithms}
                  onChange={handleAlgorithmChange}
                />

                <TagManager value={tags} onChange={handleTagChange} />
              </Card>

              <Card title='Experiment Parameters' className='injection-card'>
                <Form.Item
                  name={['experiment_params', 'duration']}
                  label='Duration (seconds)'
                  rules={[{ required: true, message: 'Please input duration' }]}
                >
                  <InputNumber
                    min={60}
                    max={3600}
                    style={{ width: '100%' }}
                    placeholder='300'
                  />
                </Form.Item>

                <Form.Item
                  name={['experiment_params', 'interval']}
                  label='Interval (seconds)'
                  rules={[{ required: true, message: 'Please input interval' }]}
                >
                  <InputNumber
                    min={10}
                    max={600}
                    style={{ width: '100%' }}
                    placeholder='60'
                  />
                </Form.Item>

                <Form.Item
                  name={['experiment_params', 'parallel']}
                  label='Parallel Execution'
                  valuePropName='checked'
                >
                  <Select placeholder='Select execution mode'>
                    <Option value={false}>Sequential</Option>
                    <Option value>Parallel</Option>
                  </Select>
                </Form.Item>
              </Card>
            </Col>

            {/* Middle Panel - Fault Types */}
            <Col span={6}>
              <FaultTypePanel onFaultSelect={handleFaultSelect} />
            </Col>

            {/* Right Panel - Visual Canvas */}
            <Col span={10}>
              <VisualCanvas
                faultMatrix={faultMatrix}
                onFaultMatrixChange={handleFaultMatrixChange}
                selectedFault={selectedFault}
              />
            </Col>
          </Row>

          {/* Bottom Panel - Fault Configuration */}
          {selectedFault && (
            <Row gutter={24} style={{ marginTop: 24 }}>
              <Col span={24}>
                <FaultConfigPanel
                  fault={selectedFault}
                  onConfigChange={(
                    config: Record<string, string | number | boolean>
                  ) => {
                    // Update fault configuration in matrix
                    console.error('Fault config updated:', config);
                  }}
                />
              </Col>
            </Row>
          )}

          {/* Submit Button */}
          <Row gutter={24} style={{ marginTop: 24 }}>
            <Col span={24}>
              <Space>
                <Button type='primary' htmlType='submit'>
                  Create Injection
                </Button>
                <Button onClick={() => navigate('/injections')}>Cancel</Button>
              </Space>
            </Col>
          </Row>
        </Form>
    </div>
  );
};

export default InjectionCreate;
