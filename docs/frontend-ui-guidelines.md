# AegisLab Frontend UI/UX Guidelines

> Version: 1.1 | Date: 2026-04-13 (updated)
> This document defines UI/UX conventions independent of business logic.
> All new components (REQ-812~817) follow these guidelines.

## 1. Design Philosophy

- **克制 (Restrained)**: 不加不必要的装饰。每个元素必须有明确的功能目的。
- **一致 (Consistent)**: 相同模式用相同组件，不在不同页面发明不同方案。
- **高信息密度**: 研究人员需要快速获取信息，避免大面积留白和低密度卡片。

## 2. Technology Stack

| Layer | Choice | Notes |
|-------|--------|-------|
| UI Framework | Ant Design 5.x | 不混用其他 UI 库 |
| Styling | Ant Design tokens + CSS Modules | 优先 theme tokens，CSS 仅做 Ant Design 无法覆盖的布局 |
| State | Zustand (global) + TanStack Query (server) | 不引入 Redux、MobX 等 |
| Routing | React Router v6 | 保持 lazy load |
| Charts | ECharts (via LabChart) | 暂不引入 D3、Cytoscape 等 |

## 3. Color System

### Primary Palette

| Token | Value | Usage |
|-------|-------|-------|
| Primary | `#0ea5e9` (sky-500) | 主按钮、链接、选中态 |
| Success | `#52c41a` | 成功状态、完成 |
| Warning | `#faad14` | 警告、进行中 |
| Error | `#ff4d4f` | 失败、错误 |
| Neutral | Ant Design 默认灰阶 | 文本、边框、背景 |

### Theme

- 支持 Light / Dark 切换
- 通过 `data-theme` attribute on `<html>` 控制
- 所有自定义颜色必须通过 CSS variables 定义，不得硬编码 hex 值
- Dark mode 下不使用纯黑 (`#000`)，使用 Ant Design 的 dark token

## 4. Layout Patterns

### 4.1 Two-Layout System

| Layout | Usage | Sidebar Width |
|--------|-------|---------------|
| MainLayout | 全局页面 (Home, Projects, Teams, Admin) | 240px, 可折叠至 0 |
| WorkspaceLayout | 项目内页面 (Datapacks, Executions, ...) | 200px, 可折叠至 64px |

两个 Layout 不共存。MainLayout 页面内无 WorkspaceSidebar，反之亦然。

### 4.2 Page Content Area

```
┌──────────────────────────────────────────┐
│ Page Header (title + actions)      24px  │
├──────────────────────────────────────────┤
│                                          │
│ Content                            24px  │
│                                   padding│
│                                          │
└──────────────────────────────────────────┘
```

- Content area padding: `24px`
- 页面最大宽度: 不限制 (fluid width)
- Card 间距: `16px` (Row gutter)

### 4.3 Responsive Breakpoints

| Breakpoint | Width | Behavior |
|------------|-------|----------|
| Mobile | < 768px | Sidebar 自动折叠 |
| Tablet | 768-1200px | 正常显示 |
| Desktop | > 1200px | 正常显示 |

暂不做移动端适配，保证 > 1024px 可用即可。

## 5. Component Patterns

### 5.1 Page Header

每个页面顶部统一使用 Page Header 模式：

```
┌──────────────────────────────────────────────────┐
│ Page Title                    [Action Button(s)] │
│ Optional description text                        │
└──────────────────────────────────────────────────┘
```

- Title: `<Typography.Title level={4}>`
- Actions: 右对齐，primary action 用 `type="primary"` 按钮
- Description: `<Typography.Text type="secondary">`

### 5.2 List Page (Table)

所有列表页统一结构：

```
Page Header (title + "Create" button)
────────────────────────────────────
Toolbar (search + filters)
────────────────────────────────────
Table (sortable columns, pagination)
```

组件:
- 使用 `WorkspaceTable` 或 Ant Design `<Table>` (视复杂度决定)
- 分页: 默认 `pageSize=20`, 可选 `[10, 20, 50]`
- 空状态: `<Empty>` + 引导 action
- 加载状态: `<Skeleton>` (首次) 或 `<Spin>` (刷新)

### 5.3 Detail Page (Tabbed)

所有详情页统一结构：

```
Detail Header (title + status badge + back button + actions)
────────────────────────────────────────────────────────────
Tab Bar (Overview | Files | Config | ...)
────────────────────────────────────────────────────────────
Tab Content
```

组件:
- Header: 使用 `DetailViewHeader`
- Tabs: Ant Design `<Tabs>` with `type="line"`
- 每个 Tab content 有统一的 `padding: 16px 0`

### 5.4 Form Page (Multi-step)

多步表单统一模式：

```
┌──────────────────────────────────────────┐
│ Step 1: xxx   ●──●──○──○   Step 4: xxx  │
├──────────────────────────────────────────┤
│                                          │
│ Form Fields for current step             │
│                                          │
├──────────────────────────────────────────┤
│              [Previous]  [Next / Submit] │
└──────────────────────────────────────────┘
```

- 使用 Ant Design `<Steps>` 组件
- 每步表单用独立的 `<Form>` 或同一 `<Form>` 分段渲染
- 最后一步显示配置摘要 (Review)
- Previous/Next 按钮固定在底部

### 5.5 Modal / Dialog

- 确认操作 (删除、批量操作): 使用 `Modal.confirm()`
- 简单创建 (如 Create Team): 使用 `<Modal>` + 内嵌 `<Form>`
- 复杂创建: 独立页面，不用 Modal

### 5.6 Status Badge

统一 status 显示方式：

| State | Color | Badge |
|-------|-------|-------|
| Pending / Initial | default (grey) | `<Badge status="default">` |
| Running / Building | processing (blue) | `<Badge status="processing">` |
| Completed / Success / Ready | success (green) | `<Badge status="success">` |
| Failed / Error | error (red) | `<Badge status="error">` |
| Disabled | warning (yellow) | `<Badge status="warning">` |

封装为 `<StatusBadge state={...}>` 组件 (已有 `src/components/ui/StatusBadge.tsx`)。

## 6. Loading States

### 6.1 First Load (no data yet)

```tsx
<Skeleton active paragraph={{ rows: 6 }} />
```

### 6.2 Refresh / Mutation in progress

```tsx
<Spin spinning={isLoading}>
  {content}
</Spin>
```

### 6.3 Button loading

```tsx
<Button loading={isPending}>Submit</Button>
```

### 6.4 禁止使用

- 纯文字 "Loading..." (当前 `LoadingFallback` 组件需改为 `<Spin>`)
- 自定义 spinner / skeleton

## 7. Empty States

### 7.1 No data (initial)

```tsx
<Empty
  image={Empty.PRESENTED_IMAGE_SIMPLE}
  description="No datapacks yet"
>
  <Button type="primary" onClick={...}>
    Create Your First Experiment
  </Button>
</Empty>
```

### 7.2 Search/Filter no results

```tsx
<Empty description="No results matching your filters" />
```

### 7.3 禁止使用

- 空白页面 (无任何提示)
- 自定义空状态插画 (统一用 Ant Design Empty)

## 8. Error States

### 8.1 Page-level error (404, permission denied)

```tsx
<Result
  status="404"
  title="Not Found"
  subTitle="The resource could not be found."
  extra={<Button onClick={goBack}>Go Back</Button>}
/>
```

### 8.2 Inline error (API failure)

```tsx
<Alert type="error" message="Failed to load data" showIcon />
```

### 8.3 Form validation

使用 Ant Design Form 内置的 `rules` 校验，不自定义 error 样式。

### 8.4 API error notification

```tsx
message.error('Failed to create execution');
```

使用 `message.error()` / `message.success()` 全局提示，不用 `notification`。

## 9. Typography

| Element | Component | Size |
|---------|-----------|------|
| Page title | `<Title level={4}>` | 20px |
| Section title | `<Title level={5}>` | 16px |
| Body text | `<Text>` | 14px (default) |
| Secondary text | `<Text type="secondary">` | 14px, muted |
| Small text | `<Text style={{ fontSize: 12 }}>` | 12px |
| Code | `<Text code>` | 14px, monospace |

- 语言: 英文为主 (面向科研人员)
- 不使用 emoji (保持克制)
- 不使用全大写标题

## 10. Spacing & Sizing

| Token | Value | Usage |
|-------|-------|-------|
| Page padding | 24px | Content area 内边距 |
| Card gap | 16px | Cards 之间的间距 |
| Form item gap | 24px (Ant default) | 表单项之间 |
| Button gap | 8px | 相邻按钮之间 |
| Table row height | 48px (Ant default) | 表格行高 |

## 11. Sidebar Navigation

### MainSidebar Items

```
Home
─────────
Projects (section header)
  Project 1
  Project 2
  View all →
─────────
Teams (section header)
  Team 1
  + Create a team
─────────
Admin (if admin)
  Users
  Containers
  Datasets
  System
─────────
Tasks (bottom)
```

### WorkspaceSidebar Items

```
Overview
Datapacks
Executions
Evaluations
Algorithms
─────────
Settings
```

- Active item: 背景高亮 + 左侧 border
- Icon + text，collapsed 时只显示 icon
- 不加 badge/count 数字（避免噪音）

## 12. Table Conventions

### Column Types

| Type | Alignment | Width |
|------|-----------|-------|
| ID / Name | left | auto |
| Status | center | 100px |
| Date/Time | left | 160px |
| Number | right | 80-120px |
| Actions | right | 按按钮数量 |

### Sort

- 默认按 `created_at` 降序 (最新在前)
- 允许用户切换排序列

### Selection & Bulk Actions

- Checkbox 选择 (全选 / 单选)
- 选中后显示 Bulk Action Bar: "N selected — [Delete] [Add Label]"
- Bulk action 需二次确认 (Modal.confirm)

## 13. Form Conventions

### Field Layout

- 使用 `layout="vertical"` (label 在上方)
- 必填字段标 `*`
- 可选字段不加额外标注
- 帮助文本用 `<Form.Item help="...">`

### Submit

- 主操作按钮: `type="primary"`, 右对齐
- Cancel 按钮: `type="default"`, 在 Submit 左边
- Submit 时 loading state: `<Button loading={isPending}>`

### Validation

- 实时校验 (onChange)
- 提交时再次校验
- Error 显示在 field 下方 (Ant Design 默认行为)

## 14. Animation & Transition

- **不使用自定义动画** — Ant Design 内置过渡已够用
- Tab 切换、Modal 打开/关闭: 使用 Ant Design 默认
- 路由切换: 无过渡动画 (避免拖慢感知速度)
- Loading spinner: Ant Design `<Spin>` 默认动画

## 15. Accessibility

基础要求：
- 所有 interactive 元素可 Tab 聚焦
- 按钮有明确的 aria-label (当只有 icon 时)
- 颜色不作为唯一信息传达方式 (status 同时有文字 + 颜色)
- 不自定义 focus ring 样式 (保持浏览器默认或 Ant Design 默认)

## 16. Anti-Patterns (禁止清单)

| 禁止 | 原因 | 替代 |
|------|------|------|
| Hardcoded mock data in components | 零 Mock 原则 | 接入真实 API |
| `console.log` in production code | 噪音 | 删除或使用 dev-only logger |
| Inline style objects in JSX (大量使用) | 不可复用 | CSS Module 或 Ant Design token |
| `any` type in TypeScript | 类型安全 | 使用 SDK 生成类型 |
| 嵌套超过 3 层的 ternary | 可读性 | 提取为 helper 或 early return |
| 在 render 中定义组件 | 性能 | 提取为独立组件 |
| 自定义 HTTP client wrapper | 一致性 | 统一使用 `src/api/client.ts` |
| `localStorage` 直接操作 (分散) | 一致性 | 通过 Zustand persist 或集中的 util |
