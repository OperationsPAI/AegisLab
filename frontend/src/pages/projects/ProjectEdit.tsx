import {
  ArrowLeftOutlined,
  DeleteOutlined,
  ProjectOutlined,
  SaveOutlined,
} from '@ant-design/icons';
import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  Alert,
  Button,
  Card,
  Col,
  Form,
  Input,
  message,
  Modal,
  Row,
  Skeleton,
  Space,
  Switch,
  Typography,
} from 'antd';
import { useEffect } from 'react';
import { useNavigate, useParams } from 'react-router-dom';


import { projectApi } from '@/api/projects';

const { Title, Text } = Typography;
const { TextArea } = Input;

interface ProjectFormData {
  name: string;
  description?: string;
  is_public: boolean;
}

const ProjectEdit = () => {
  const { id } = useParams<{ id: string }>();
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [form] = Form.useForm<ProjectFormData>();
  const projectId = Number(id);

  // Fetch project details
  const {
    data: project,
    isLoading,
    error,
  } = useQuery({
    queryKey: ['project', projectId],
    queryFn: () => projectApi.getProject(projectId),
    enabled: !!projectId && !isNaN(projectId),
  });

  // Update project mutation
  const updateMutation = useMutation({
    mutationFn: (data: ProjectFormData) =>
      projectApi.updateProject(projectId, data),
    onSuccess: () => {
      message.success('Project updated successfully');
      queryClient.invalidateQueries({ queryKey: ['project', projectId] });
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      navigate(`/projects/${projectId}`);
    },
    onError: () => {
      // Error handled by apiClient interceptor
    },
  });

  // Delete project mutation
  const deleteMutation = useMutation({
    mutationFn: () => projectApi.deleteProject(projectId),
    onSuccess: () => {
      message.success('Project deleted successfully');
      queryClient.invalidateQueries({ queryKey: ['projects'] });
      navigate('/projects');
    },
    onError: () => {
      // Error handled by apiClient interceptor
    },
  });

  // Populate form when project data is loaded
  useEffect(() => {
    if (project) {
      form.setFieldsValue({
        name: project.name,
        description: project.description,
        is_public: project.is_public ?? false,
      });
    }
  }, [project, form]);

  const handleSubmit = (values: ProjectFormData) => {
    updateMutation.mutate(values);
  };

  const handleDelete = () => {
    Modal.confirm({
      title: 'Delete Project',
      content:
        'Are you sure you want to delete this project? This action cannot be undone and will delete all associated data.',
      okText: 'Yes, Delete',
      okButtonProps: { danger: true },
      cancelText: 'Cancel',
      onOk: () => deleteMutation.mutate(),
    });
  };

  const handleCancel = () => {
    navigate(`/projects/${projectId}`);
  };

  if (isLoading) {
    return (
      <div style={{ padding: 24 }}>
        <Skeleton active paragraph={{ rows: 10 }} />
      </div>
    );
  }

  if (error || !project) {
    return (
      <div style={{ padding: 24 }}>
        <Card>
          <Alert
            message='Error'
            description='Failed to load project. Please try again.'
            type='error'
            showIcon
          />
          <Button
            style={{ marginTop: 16 }}
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate('/projects')}
          >
            Back to Projects
          </Button>
        </Card>
      </div>
    );
  }

  return (
    <div style={{ padding: 24 }}>
      {/* Header */}
      <div style={{ marginBottom: 24 }}>
        <Space>
          <Button
            icon={<ArrowLeftOutlined />}
            onClick={() => navigate(`/projects/${projectId}`)}
          >
            Back to Project
          </Button>
        </Space>
        <Title level={2} style={{ marginTop: 16, marginBottom: 0 }}>
          <ProjectOutlined style={{ marginRight: 8 }} />
          Edit Project
        </Title>
        <Text type='secondary'>Update project settings and information</Text>
      </div>

      <Row gutter={[24, 24]}>
        <Col xs={24} lg={16}>
          <Card
            title='Project Information'
            extra={
              <Space>
                <Button onClick={handleCancel}>Cancel</Button>
                <Button
                  type='primary'
                  icon={<SaveOutlined />}
                  loading={updateMutation.isPending}
                  onClick={() => form.submit()}
                >
                  Save Changes
                </Button>
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
                label='Project Name'
                name='name'
                rules={[
                  { required: true, message: 'Please enter project name' },
                  { min: 3, message: 'Name must be at least 3 characters' },
                  { max: 100, message: 'Name cannot exceed 100 characters' },
                ]}
              >
                <Input placeholder='Enter project name' size='large' />
              </Form.Item>

              <Form.Item
                label='Description'
                name='description'
                rules={[
                  {
                    max: 500,
                    message: 'Description cannot exceed 500 characters',
                  },
                ]}
              >
                <TextArea
                  placeholder='Describe your project (optional)'
                  rows={4}
                  showCount
                  maxLength={500}
                />
              </Form.Item>

              <Form.Item
                label='Public Project'
                name='is_public'
                valuePropName='checked'
                extra='Public projects can be viewed by all users'
              >
                <Switch checkedChildren='Public' unCheckedChildren='Private' />
              </Form.Item>
            </Form>
          </Card>
        </Col>

        <Col xs={24} lg={8}>
          <Card title='Danger Zone'>
            <Alert
              message='Delete Project'
              description='Once you delete a project, there is no going back. This will permanently delete the project and all its associated data including experiments, injections, and results.'
              type='error'
              showIcon
              style={{ marginBottom: 16 }}
            />
            <Button
              danger
              icon={<DeleteOutlined />}
              loading={deleteMutation.isPending}
              onClick={handleDelete}
              block
            >
              Delete This Project
            </Button>
          </Card>
        </Col>
      </Row>
    </div>
  );
};

export default ProjectEdit;
