
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ConfigProvider } from 'antd';
import zhCN from 'antd/locale/zh_CN';
import dayjs from 'dayjs';
import React from 'react';
import ReactDOM from 'react-dom/client';
import { BrowserRouter } from 'react-router-dom';
import 'dayjs/locale/zh-cn';


import App from './App';
import { initializeTheme } from './store/theme';

import './index.css';
import './styles/theme.css';
import './styles/responsive.css';

dayjs.locale('zh-cn');

const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      refetchOnWindowFocus: false,
      retry: 1,
      staleTime: 5 * 60 * 1000, // 5 minutes
    },
  },
});

// Initialize theme on app load
initializeTheme();

const rootElement = document.getElementById('root');
if (!rootElement) {
  throw new Error('Failed to find the root element');
}

ReactDOM.createRoot(rootElement).render(
  <React.StrictMode>
    <BrowserRouter
      future={{
        v7_startTransition: true,
        v7_relativeSplatPath: true,
      }}
    >
      <QueryClientProvider client={queryClient}>
        <ConfigProvider
          locale={zhCN}
          theme={{
            token: {
              colorPrimary: '#0ea5e9',
              colorSuccess: '#10b981',
              colorWarning: '#f59e0b',
              colorError: '#ef4444',
              colorInfo: '#06b6d4',
              borderRadius: 8,
              fontFamily:
                'Inter, -apple-system, BlinkMacSystemFont, sans-serif',
              fontSize: 14,
              controlHeight: 40,
            },
            components: {
              Layout: {
                headerBg: 'transparent',
                headerHeight: 64,
                siderBg: 'transparent',
              },
              Menu: {
                itemBg: 'transparent',
                itemSelectedBg: '#0ea5e920',
                itemSelectedColor: '#0ea5e9',
                itemHoverBg: '#f1f5f9',
              },
              Card: {
                borderRadiusLG: 12,
              },
              Button: {
                borderRadius: 8,
                controlHeight: 40,
                controlHeightLG: 48,
              },
              Input: {
                borderRadius: 8,
                controlHeight: 40,
              },
              Table: {
                borderRadius: 12,
                headerBg: '#f8fafc',
                headerColor: '#334155',
              },
            },
          }}
        >
          <App />
        </ConfigProvider>
      </QueryClientProvider>
    </BrowserRouter>
  </React.StrictMode>
);
