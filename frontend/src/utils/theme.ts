// Global theme colors for academic research style
export const theme = {
  colors: {
    // Primary - Deep blue for professionalism
    primary: '#2563eb',
    primaryHover: '#1d4ed8',
    primaryLight: '#dbeafe',

    // Secondary - Slate gray for neutrality
    secondary: '#475569',
    secondaryLight: '#e2e8f0',

    // Success - Green for completed states
    success: '#10b981',
    successLight: '#d1fae5',

    // Warning - Amber for attention
    warning: '#f59e0b',
    warningLight: '#fef3c7',

    // Error - Red for failures
    error: '#ef4444',
    errorLight: '#fee2e2',

    // Info - Cyan for informational
    info: '#06b6d4',
    infoLight: '#cffafe',

    // Neutrals
    gray50: '#f9fafb',
    gray100: '#f3f4f6',
    gray200: '#e5e7eb',
    gray300: '#d1d5db',
    gray400: '#9ca3af',
    gray500: '#6b7280',
    gray600: '#4b5563',
    gray700: '#374151',
    gray800: '#1f2937',
    gray900: '#111827',

    // Background
    background: '#ffffff',
    backgroundSecondary: '#f9fafb',

    // Text
    text: '#111827',
    textSecondary: '#6b7280',
    textTertiary: '#9ca3af',

    // Borders
    border: '#e5e7eb',
    borderLight: '#f3f4f6',

    // Data visualization colors (academic palette)
    chart: {
      blue: '#2563eb',
      cyan: '#06b6d4',
      emerald: '#10b981',
      amber: '#f59e0b',
      rose: '#f43f5e',
      violet: '#8b5cf6',
      orange: '#f97316',
      teal: '#14b8a6',
    },
  },

  // Task state colors
  taskStates: {
    pending: '#9ca3af', // Gray
    running: '#3b82f6', // Blue (with animation)
    completed: '#10b981', // Green
    error: '#ef4444', // Red
    cancelled: '#f59e0b', // Orange
  },

  // Spacing
  spacing: {
    xs: '4px',
    sm: '8px',
    md: '16px',
    lg: '24px',
    xl: '32px',
    xxl: '48px',
  },

  // Border radius
  radius: {
    sm: '4px',
    md: '8px',
    lg: '12px',
    xl: '16px',
    full: '9999px',
  },

  // Shadows
  shadows: {
    sm: '0 1px 2px 0 rgb(0 0 0 / 0.05)',
    md: '0 4px 6px -1px rgb(0 0 0 / 0.1)',
    lg: '0 10px 15px -3px rgb(0 0 0 / 0.1)',
    xl: '0 20px 25px -5px rgb(0 0 0 / 0.1)',
  },

  // Typography
  fonts: {
    sans: '-apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, "Helvetica Neue", Arial, sans-serif',
    mono: 'ui-monospace, SFMono-Regular, "SF Mono", Menlo, Consolas, "Liberation Mono", monospace',
  },
};
