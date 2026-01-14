import {
  CloseOutlined,
  ContainerOutlined,
  FormOutlined,
  GlobalOutlined,
  SaveOutlined,
  TagsOutlined,
} from '@ant-design/icons';
import { ContainerType } from '@rcabench/client';
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
} from 'antd';
import { useEffect, useState } from 'react';
import { useNavigate, useParams } from 'react-router-dom';

import { containerApi } from '@/api/containers';

const { Title, Text } = Typography;
const { TextArea } = Input;
const { Option } = Select;

interface LabelItemWithKey {
  key: string;
  value?: string;
}

interface ContainerFormData {
  name: string;
  type: ContainerType;
  readme?: string;
  is_public: boolean;
  labels?: LabelItemWithKey[];
}

const ContainerForm = () => {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [form] = Form.useForm<ContainerFormData>();
  const { id } = useParams<{ id: string }>();
  const containerId = id ? Number(id) : undefined;
  const [labelInput, setLabelInput] = useState('');
  const [labels, setLabels] = useState<LabelItemWithKey[]>([]);

  // Fetch container data if editing
  const { data: containerData, isLoading } = useQuery({
    queryKey: ['container', containerId],
    queryFn: () => containerApi.getContainer(containerId as number),
    enabled: !!containerId,
  });

  // Set form data when editing
  useEffect(() => {
    if (containerData) {
      // Convert string type to ContainerType enum
      const typeValue =
        containerData.type !== undefined
          ? (Number(containerData.type) as ContainerType)
          : undefined;

      form.setFieldsValue({
        name: containerData.name,
        type: typeValue,
        readme: containerData.readme,
        is_public: containerData.is_public,
      });
      // Filter out labels with undefined keys
      const validLabels = (containerData.labels || []).filter(
        (l): l is { key: string; value?: string } => l.key !== undefined
      );
      setLabels(validLabels);
    }
  }, [containerData, form]);

  // Create or update mutation
  const createMutation = useMutation({
    mutationFn: (data: ContainerFormData) => containerApi.createContainer(data),
    onSuccess: () => {
      message.success('容器创建成功');
      queryClient.invalidateQueries({ queryKey: ['containers'] });
      navigate('/containers');
    },
    onError: (error) => {
      message.error('容器创建失败');
      console.error('Create container error:', error);
    },
  });

  const updateMutation = useMutation({
    mutationFn: (data: Partial<ContainerFormData>) =>
      containerApi.updateContainer(containerId as number, data),
    onSuccess: () => {
      message.success('容器更新成功');
      queryClient.invalidateQueries({ queryKey: ['containers'] });
      queryClient.invalidateQueries({ queryKey: ['container', containerId] });
      navigate('/containers');
    },
    onError: (error) => {
      message.error('容器更新失败');
      console.error('Update container error:', error);
    },
  });

  const handleSubmit = async (values: ContainerFormData) => {
    const data = {
      ...values,
      labels,
    };

    if (containerId) {
      updateMutation.mutate(data);
    } else {
      createMutation.mutate(data);
    }
  };

  const handleCancel = () => {
    navigate('/containers');
  };

  const addLabel = () => {
    if (!labelInput.trim()) return;

    const [key, value] = labelInput.split(':').map((s) => s.trim());
    if (!key || !value) {
      message.warning('请按格式输入标签: key:value');
      return;
    }

    if (labels.some((l) => l.key === key)) {
      message.warning('标签键已存在');
      return;
    }

    setLabels([...labels, { key, value }]);
    setLabelInput('');
  };

  const removeLabel = (key: string) => {
    setLabels(labels.filter((l) => l.key !== key));
  };

  if (isLoading && containerId) {
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
            返回列表
          </Button>
          <Title level={2} style={{ margin: 0 }}>
            {containerId ? '编辑容器' : '创建容器'}
          </Title>
        </Space>
      </div>

      <Row gutter={[24, 24]}>
        <Col xs={24} lg={16}>
          <Card
            title={
              <Space>
                <FormOutlined />
                <span>容器信息</span>
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
                label='容器名称'
                name='name'
                rules={[
                  { required: true, message: '请输入容器名称' },
                  { min: 3, message: '名称至少3个字符' },
                  { max: 50, message: '名称不能超过50个字符' },
                  {
                    pattern: /^[a-zA-Z0-9-_]+$/,
                    message: '名称只能包含字母、数字、连字符和下划线',
                  },
                ]}
              >
                <Input
                  placeholder='输入容器名称'
                  size='large'
                  disabled={!!containerId}
                />
              </Form.Item>

              <Form.Item
                label='容器类型'
                name='type'
                rules={[{ required: true, message: '请选择容器类型' }]}
              >
                <Select
                  placeholder='选择容器类型'
                  size='large'
                  onChange={() => {
                    form.validateFields(['type']);
                  }}
                >
                  <Option value={ContainerType.Pedestal}>
                    <Space>
                      <ContainerOutlined style={{ color: '#3b82f6' }} />
                      <div>
                        <div>Pedestal</div>
                        <Text type='secondary' style={{ fontSize: '0.75rem' }}>
                          基础微服务环境，用于故障注入和观测
                        </Text>
                      </div>
                    </Space>
                  </Option>
                  <Option value={ContainerType.Benchmark}>
                    <Space>
                      <ContainerOutlined style={{ color: '#10b981' }} />
                      <div>
                        <div>Benchmark</div>
                        <Text type='secondary' style={{ fontSize: '0.75rem' }}>
                          基准测试容器，用于生成负载和评估
                        </Text>
                      </div>
                    </Space>
                  </Option>
                  <Option value={ContainerType.Algorithm}>
                    <Space>
                      <ContainerOutlined style={{ color: '#8b5cf6' }} />
                      <div>
                        <div>Algorithm</div>
                        <Text type='secondary' style={{ fontSize: '0.75rem' }}>
                          RCA算法容器，实现根因分析逻辑
                        </Text>
                      </div>
                    </Space>
                  </Option>
                </Select>
              </Form.Item>

              <Form.Item
                label='README'
                name='readme'
                rules={[
                  {
                    max: 5000,
                    message: 'README不能超过5000个字符',
                  },
                ]}
              >
                <TextArea
                  rows={6}
                  placeholder='输入容器的使用说明和文档...'
                />
              </Form.Item>

              <Form.Item
                label='可见性'
                name='is_public'
                valuePropName='checked'
                help='公开容器可被其他用户在其项目中使用'
              >
                <Switch
                  checkedChildren={<GlobalOutlined />}
                  unCheckedChildren={<GlobalOutlined />}
                />
              </Form.Item>

              <Divider />

              <Form.Item label='标签'>
                <Space direction='vertical' style={{ width: '100%' }}>
                  <Space.Compact style={{ width: '100%' }}>
                    <Input
                      placeholder='输入标签 (key:value)'
                      value={labelInput}
                      onChange={(e) => setLabelInput(e.target.value)}
                      onPressEnter={addLabel}
                    />
                    <Button
                      type='primary'
                      onClick={addLabel}
                      icon={<TagsOutlined />}
                    >
                      添加
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
                    icon={<SaveOutlined />}
                    loading={createMutation.isPending || updateMutation.isPending}
                  >
                    {containerId ? '更新容器' : '创建容器'}
                  </Button>
                  <Button icon={<CloseOutlined />} onClick={handleCancel}>
                    取消
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
                <ContainerOutlined />
                <span>容器指南</span>
              </Space>
            }
          >
            <Space direction='vertical' style={{ width: '100%' }}>
              <div>
                <Text strong>容器类型：</Text>
                <ul style={{ marginTop: 8, marginBottom: 16 }}>
                  <li>
                    <Text>
                      <strong>Pedestal:</strong> 基础微服务环境，用于故障注入和观测
                    </Text>
                  </li>
                  <li>
                    <Text>
                      <strong>Benchmark:</strong> 基准测试容器，用于生成负载和评估
                    </Text>
                  </li>
                  <li>
                    <Text>
                      <strong>Algorithm:</strong> RCA算法容器，实现根因分析逻辑
                    </Text>
                  </li>
                </ul>
              </div>

              <Divider />

              <div>
                <Text strong>最佳实践：</Text>
                <ul style={{ marginTop: 8 }}>
                  <li>
                    <Text>使用描述性名称</Text>
                  </li>
                  <li>
                    <Text>编写清晰的README文档</Text>
                  </li>
                  <li>
                    <Text>适当添加标签以便分类</Text>
                  </li>
                  <li>
                    <Text>保持容器版本管理</Text>
                  </li>
                  <li>
                    <Text>测试容器镜像的可用性</Text>
                  </li>
                </ul>
              </div>

              <Divider />

              <div>
                <Text strong>标签：</Text>
                <Text
                  type='secondary'
                  style={{ display: 'block', marginTop: 4 }}
                >
                  使用标签来组织和分类您的容器。格式: key:value
                </Text>
              </div>

              <Divider />

              <div>
                <Text strong>版本管理：</Text>
                <Text
                  type='secondary'
                  style={{ display: 'block', marginTop: 4 }}
                >
                  创建容器后，您可以添加多个版本以跟踪容器镜像的不同迭代。
                </Text>
              </div>
            </Space>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default ContainerForm;
