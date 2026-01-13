import { useEffect } from 'react';

import { MoonOutlined, SunOutlined } from '@ant-design/icons';
import { Button } from 'antd';

import { useThemeStore } from '@/store/theme';

const ThemeToggle = () => {
  const { theme, toggleTheme } = useThemeStore();

  // Initialize theme on mount
  useEffect(() => {
    document.documentElement.setAttribute('data-theme', theme);
  }, [theme]);

  return (
    <Button
      type='text'
      icon={theme === 'light' ? <MoonOutlined /> : <SunOutlined />}
      onClick={toggleTheme}
      style={{
        width: 40,
        height: 40,
        borderRadius: '50%',
        display: 'flex',
        alignItems: 'center',
        justifyContent: 'center',
        transition: 'all 0.3s ease',
      }}
    />
  );
};

export default ThemeToggle;
