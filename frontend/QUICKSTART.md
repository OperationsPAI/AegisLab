# RCABench 前端快速开始

## 第一步:安装 Node.js

确保已安装 Node.js 18 或更高版本:

```bash
node --version  # 应显示 v18.x.x 或更高
npm --version   # 应显示 9.x.x 或更高
```

如果未安装,请访问 https://nodejs.org/ 下载安装。

## 第二步:安装项目依赖

在项目根目录的 `frontend` 文件夹下执行:

```bash
cd /Users/lincyaw/workspace/rcabench/frontend
npm install
```

等待依赖安装完成(首次安装可能需要几分钟)。

## 第三步:启动开发服务器

```bash
npm run dev
```

看到如下输出表示启动成功:

```
  VITE v5.1.5  ready in 500 ms

  ➜  Local:   http://localhost:3000/
  ➜  Network: use --host to expose
  ➜  press h + enter to show help
```

## 第四步:确保后端服务运行

前端需要后端 API 支持,请确保后端服务已启动:

```bash
# 在项目根目录执行
cd /Users/lincyaw/workspace/rcabench

# 启动基础设施
docker compose up redis mysql jaeger buildkitd -d

# 运行后端服务
make local-debug
```

后端应该在 `http://localhost:8082` 运行。

## 第五步:访问应用

打开浏览器访问: http://localhost:3000

你将看到登录页面!

## 登录测试

使用后端配置的测试用户登录:

```
用户名: admin (或你的测试用户)
密码: password (或你设置的密码)
```

登录成功后,你将看到仪表盘页面。

## 项目目录结构

```
frontend/
├── src/
│   ├── api/          # API 客户端
│   ├── pages/        # 页面组件
│   ├── components/   # 通用组件
│   ├── store/        # 状态管理
│   ├── types/        # 类型定义
│   ├── utils/        # 工具函数
│   ├── App.tsx       # 根组件
│   └── main.tsx      # 入口文件
├── package.json      # 依赖配置
├── vite.config.ts    # Vite 配置
└── tsconfig.json     # TypeScript 配置
```

## 可用的 npm 命令

```bash
npm run dev          # 启动开发服务器
npm run build        # 构建生产版本
npm run preview      # 预览生产构建
npm run lint         # 运行 ESLint 检查
npm run type-check   # 运行 TypeScript 类型检查
```

## 页面导航

已实现的页面:

- 🔐 `/login` - 登录页面
- 📊 `/dashboard` - 仪表盘(关键指标、任务状态、最近活动)
- 📁 `/projects` - 项目管理(列表、搜索、删除)
- 📦 `/containers` - 容器管理(列表、类型筛选、删除)
- 🧪 `/injections` - 故障注入(待完善)
- ⚙️ `/executions` - 算法执行(待完善)

## 开发指南

### 修改 API 地址

如果后端运行在不同的端口,修改 `vite.config.ts`:

```typescript
server: {
  proxy: {
    '/api': {
      target: 'http://localhost:YOUR_PORT',  // 修改这里
      changeOrigin: true,
    },
  },
}
```

### 添加新页面

1. 在 `src/pages/{module}/` 创建组件
2. 在 `src/App.tsx` 添加路由
3. 在 `src/components/layout/MainLayout.tsx` 添加菜单项

### 调试技巧

1. **查看网络请求**: 打开浏览器开发者工具 → Network 标签
2. **查看 React 组件**: 安装 React DevTools 扩展
3. **查看 API 缓存**: 安装 `@tanstack/react-query-devtools`

## 常见问题

### Q: npm install 失败怎么办?

尝试清除缓存后重新安装:

```bash
rm -rf node_modules package-lock.json
npm cache clean --force
npm install
```

### Q: 启动后看不到界面?

检查:
1. 是否成功启动(查看终端输出)
2. 浏览器地址是否正确(http://localhost:3000)
3. 是否有端口冲突(尝试修改 vite.config.ts 中的 port)

### Q: 登录后显示 API 错误?

检查:
1. 后端服务是否正常运行(http://localhost:8082)
2. 浏览器开发者工具 Network 标签查看具体错误
3. 是否有 CORS 错误(应该由 Vite 代理自动处理)

### Q: 如何查看完整的 API 文档?

后端运行时访问: http://localhost:8082/swagger/index.html

## 下一步

1. 熟悉现有页面的代码结构
2. 阅读 `DEVELOPMENT.md` 了解开发指南
3. 查看 `PROJECT_SUMMARY.md` 了解项目全貌
4. 开始实现待开发的功能模块

## 需要帮助?

- 查看 `README.md` - 详细的项目说明
- 查看 `DEVELOPMENT.md` - 开发指南和最佳实践
- 查看 `PROJECT_SUMMARY.md` - 项目完成情况和下一步计划

祝你开发顺利! 🚀
