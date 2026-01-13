import { useState } from 'react';
import { useNavigate } from 'react-router-dom';

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

import { AlgorithmSelector } from './components/AlgorithmSelector';
import { FaultConfigPanel } from './components/FaultConfigPanel';
import { FaultTypePanel } from './components/FaultTypePanel';
import { TagManager } from './components/TagManager';
import { VisualCanvas } from './components/VisualCanvas';

import { containerApi } from '../../api/containers';
import { injectionApi } from '../../api/injections';
import { projectApi } from '../../api/projects';
import MainLayout from '../../components/layout/MainLayout';
import type { ContainerResp, ProjectResp } from '@rcabench/client';

import type { FaultType } from '../../types/api';

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
    queryFn: () => projectApi.getProjects({ page: 1, size: 100 }),
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
        size: 100,
      });
    },
    enabled: !!selectedProject,
    select: (data: any) => data.items || [],
  });

  // Group containers by type
  const groupedContainers = containers.reduce(
    (
      acc: {
        pedestals: any[];
        benchmarks: any[];
        algorithms: any[];
      },
      container: any
    ) => {
      if (container.type === 2) {
        // Pedestal
        acc.pedestals.push(container);
      } else if (container.type === 1) {
        // Benchmark
        acc.benchmarks.push(container);
      } else if (container.type === 0) {
        // Algorithm
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

  const handleSubmit = async (values: InjectionFormData) => {
    try {
      const payload = {
        ...values,
        fault_matrix: faultMatrix,
        tags,
      };

      await injectionApi.createInjection(payload);
      message.success('Fault injection created successfully');
      navigate('/injections');
    } catch (error) {
      message.error('Failed to create fault injection');
      console.error('Create injection error:', error);
    }
  };

  return (
    <MainLayout>
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
    </MainLayout>
  );
};

export default InjectionCreate;
