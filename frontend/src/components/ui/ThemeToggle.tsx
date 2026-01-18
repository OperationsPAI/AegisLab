
import { MoonOutlined, SunOutlined } from '@ant-design/icons';
import { Button } from 'antd';
import { useEffect } from 'react';

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
      aria-label={theme === 'light' ? 'Switch to dark mode' : 'Switch to light mode'}
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
