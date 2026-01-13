import { create } from 'zustand';
import { persist } from 'zustand/middleware';

interface ThemeState {
  theme: 'light' | 'dark';
  sidebarCollapsed: boolean;
  setTheme: (theme: 'light' | 'dark') => void;
  toggleTheme: () => void;
  toggleSidebar: () => void;
}

export const useThemeStore = create<ThemeState>()(
  persist(
    (set, get) => ({
      theme: 'light',
      sidebarCollapsed: false,
      setTheme: (theme) => {
        set({ theme });
        document.documentElement.setAttribute('data-theme', theme);
      },
      toggleTheme: () => {
        const newTheme = get().theme === 'light' ? 'dark' : 'light';
        set({ theme: newTheme });
        document.documentElement.setAttribute('data-theme', newTheme);
      },
      toggleSidebar: () => {
        set((state) => ({ sidebarCollapsed: !state.sidebarCollapsed }));
      },
    }),
    {
      name: 'aegis-theme',
      partialize: (state) => ({
        theme: state.theme,
        sidebarCollapsed: state.sidebarCollapsed,
      }),
    }
  )
);

// Initialize theme on app load
export const initializeTheme = () => {
  const theme = useThemeStore.getState().theme;
  document.documentElement.setAttribute('data-theme', theme);
};
