# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Frontend Overview

RCABench frontend is a React 18 + TypeScript application using Ant Design 5, built with Vite. It serves as the web interface for the AegisLab RCA benchmarking platform.

## Essential Commands

```bash
# Development
npm run dev          # Start dev server on http://localhost:3000

# Code Quality
npm run lint         # Run ESLint checks
npm run lint:fix     # Auto-fix ESLint issues
npm run format       # Format code with Prettier
npm run type-check   # Run TypeScript type checking

# Build
npm run build        # Build for production
npm run preview      # Preview production build
```

## Architecture

### Technology Stack
- **Framework**: React 18.3.1 with TypeScript (strict mode)
- **Build Tool**: Vite 5 with React plugin
- **UI Library**: Ant Design 5.x with custom theme
- **State Management**: Zustand (client state) + TanStack Query (server state)
- **HTTP Client**: Axios with interceptors for auth
- **Routing**: React Router v6
- **Charts**: ECharts, D3.js, Cytoscape.js
- **Code Editor**: Monaco Editor

### Project Structure
```
src/
├── api/           # API clients (modular by domain)
├── components/    # Reusable components
│   ├── charts/    # Chart components
│   ├── dashboard/ # Dashboard-specific components
│   ├── layout/    # Layout components (MainLayout)
│   └── ui/        # Base UI components
├── hooks/         # Custom React hooks
├── pages/         # Page components (route-based)
├── store/         # Zustand stores (auth, theme)
├── types/         # TypeScript type definitions
├── utils/         # Utility functions
└── styles/        # Global styles and CSS variables
```

### Key Patterns

1. **API Integration**: All API calls go through `/api/v2` (proxied to backend)
2. **Authentication**: JWT-based with automatic token refresh
3. **State Management**:
   - Server state via TanStack Query (caching, refetching)
   - Client state via Zustand (auth, theme)
4. **Error Handling**: Centralized in Axios interceptors with Ant Design message notifications
5. **Type Safety**: Strict TypeScript with comprehensive type definitions matching backend

## Development Guidelines

### API Integration (CRITICAL)
- **NEVER modify backend field names** - Use exact field names from API
- **Backend uses snake_case** - Keep it in frontend types
- **All API types in `src/types/api.ts`** - Must match backend exactly
- **Use provided API clients** in `src/api/` directory
- **Handle errors consistently** - 401 triggers auto-refresh, others show message

### Component Development
- **Functional components only** with hooks
- **Use Ant Design components** as base building blocks
- **Follow existing patterns** in similar components
- **Extract reusable logic** into custom hooks
- **Keep components focused** - one component per file

### State Management
- **Server state**: Use TanStack Query for API data
  ```typescript
  const { data, isLoading, error } = useQuery({
    queryKey: ['projects', { page, size }],
    queryFn: () => projectApi.getProjects({ page, size }),
  })
  ```
- **Client state**: Use Zustand stores
  ```typescript
  const { user, login, logout } = useAuthStore()
  ```

### Styling
- **Ant Design theme** configured in `src/main.tsx`
- **Primary color**: #2563eb (deep blue)
- **CSS variables** for theming in `src/styles/`
- **Responsive design** using Ant Design Grid system

### Code Quality
- **ESLint rules enforced** - no unused vars, explicit types preferred
- **Prettier formatting** - 80 char width, single quotes, semicolons
- **Import organization** - external libs first, then internal modules
- **Naming conventions** - camelCase for variables, PascalCase for components

## Backend Integration Notes

### Current API Proxy
Vite dev server proxies `/api` to `http://10.10.10.220:32080` (change in `vite.config.ts` if needed)

### Authentication Flow
1. Login stores JWT in localStorage
2. Axios interceptor adds Authorization header
3. 401 responses trigger token refresh
4. Refresh failure redirects to login

### Key Backend Concepts
- **Projects**: Container for experiments
- **Containers**: Pedestal/Benchmark/Algorithm types
- **Injections**: Fault injection configurations
- **Executions**: Algorithm execution instances
- **Tasks**: Background job tracking
- **Datapacks**: Collected data from injections

## Common Tasks

### Adding a New Page
1. Create component in `src/pages/`
2. Add route in `src/App.tsx`
3. Add navigation item in `src/components/layout/MainLayout.tsx`
4. Create API client if needed in `src/api/`

### Creating API Integration
1. Define types in `src/types/api.ts` (match backend exactly)
2. Create API client in `src/api/` using `apiClient`
3. Use TanStack Query for data fetching
4. Handle loading/error states

### Working with Forms
- Use Ant Design Form component
- Define form types based on API requirements
- Handle validation before submission
- Show success/error feedback

### Adding Charts
- Use ECharts for standard charts
- Use D3.js for custom visualizations
- Use Cytoscape.js for network graphs
- Follow existing chart component patterns in `src/components/charts/`

## Important Considerations

1. **Backend-First Development**: Always check backend API before implementing frontend features
2. **Type Safety**: All API responses must have corresponding TypeScript types
3. **Error Handling**: Use consistent error messages and handling patterns
4. **Performance**: Use React.memo, useMemo, useCallback where appropriate
5. **Accessibility**: Follow Ant Design accessibility guidelines
6. **Mobile Support**: Ensure responsive design for all screen sizes