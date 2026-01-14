import { LockOutlined, UserOutlined } from '@ant-design/icons';
import { Button, Card, Form, Input, message, Typography } from 'antd';
import { useState } from 'react';
import { useNavigate } from 'react-router-dom';


import { useAuthStore } from '@/store/auth';

const { Title, Text } = Typography;

const Login = () => {
  const navigate = useNavigate();
  const { login } = useAuthStore();
  const [loading, setLoading] = useState(false);

  const onFinish = async (values: { username: string; password: string }) => {
    setLoading(true);
    try {
      await login(values.username, values.password);
      message.success('登录成功');
      // Add a small delay to ensure state is updated
      setTimeout(() => {
        navigate('/dashboard');
      }, 100);
    } catch (error) {
      message.error('登录失败，请检查用户名和密码');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div
      style={{
        minHeight: '100vh',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        background: 'linear-gradient(135deg, #667eea 0%, #764ba2 100%)',
      }}
    >
      <Card
        style={{
          width: 420,
          boxShadow: '0 20px 25px -5px rgb(0 0 0 / 0.1)',
        }}
      >
        <div style={{ textAlign: 'center', marginBottom: '32px' }}>
          <div
            style={{
              fontSize: '48px',
              marginBottom: '16px',
            }}
          >
            🔬
          </div>
          <Title level={2} style={{ marginBottom: '8px' }}>
            RCABench
          </Title>
          <Text type='secondary'>微服务根因分析基准测试平台</Text>
        </div>

        <Form name='login' size='large' onFinish={onFinish} autoComplete='off'>
          <Form.Item
            name='username'
            rules={[{ required: true, message: '请输入用户名' }]}
          >
            <Input prefix={<UserOutlined />} placeholder='用户名' />
          </Form.Item>

          <Form.Item
            name='password'
            rules={[{ required: true, message: '请输入密码' }]}
          >
            <Input.Password prefix={<LockOutlined />} placeholder='密码' />
          </Form.Item>

          <Form.Item>
            <Button
              type='primary'
              htmlType='submit'
              block
              loading={loading}
              size='large'
            >
              登录
            </Button>
          </Form.Item>

          <div style={{ textAlign: 'center' }}>
            <Text type='secondary' style={{ fontSize: '12px' }}>
              AegisLab - Root Cause Analysis Benchmark Platform
            </Text>
          </div>
        </Form>
      </Card>
    </div>
  );
};

export default Login;
