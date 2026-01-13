# API字段映射参考

本文件用于记录前端与后端API字段的映射关系，确保前后端数据一致性。

## 时间字段

| 后端字段 | 前端类型 | 说明 |
|---------|---------|------|
| created_at | string → Date | 创建时间 |
| updated_at | string → Date | 更新时间 |
| deleted_at | string \| null → Date \| null | 删除时间 |
| executed_at | string → Date | 执行时间 |
| started_at | string → Date | 开始时间 |
| finished_at | string → Date | 结束时间 |

## 状态字段

### 任务状态 (TaskState)
```typescript
// 后端返回数值
enum TaskState {
  PENDING = 0,
  RUNNING = 1,
  COMPLETED = 2,
  ERROR = -1,
  CANCELLED = 4
}
```

### 注入状态 (InjectionState)
```typescript
// 后端返回数值
enum InjectionState {
  BUILDING_DATAPACK = 0,
  DATAPACK_READY = 1,
  INJECTING = 2,
  INJECTION_COMPLETE = 3,
  COLLECTING_RESULT = 4,
  COMPLETED = 5,
  ERROR = -1
}
```

### 执行状态 (ExecutionState)
```typescript
// 后端返回数值
enum ExecutionState {
  PENDING = 0,
  RUNNING = 1,
  COMPLETED = 2,
  ERROR = -1
}
```

## 容器类型 (ContainerType)
```typescript
// 后端返回数值
enum ContainerType {
  PEDESTAL = 0,
  BENCHMARK = 1,
  ALGORITHM = 2
}
```

## 数据集类型 (DatasetType)
```typescript
// 后端返回数值
enum DatasetType {
  TRACE = 0,
  LOG = 1,
  METRIC = 2
}
```

## 故障类型 (FaultType)
```typescript
// 后端返回字符串
const FaultType = {
  // Network
  NETWORK_DELAY: 'NetworkDelay',
  NETWORK_LOSS: 'NetworkLoss',
  NETWORK_CORRUPT: 'NetworkCorrupt',
  NETWORK_DUPLICATE: 'NetworkDuplicate',
  NETWORK_PARTITION: 'NetworkPartition',

  // CPU
  CPU_STRESS: 'CPUStress',

  // Memory
  MEMORY_STRESS: 'MemoryStress',

  // IO
  IO_STRESS: 'IOStress',
  IO_FAULT: 'IOFault',

  // Pod
  POD_KILL: 'PodKill',
  POD_FAILURE: 'PodFailure',

  // Container
  CONTAINER_KILL: 'ContainerKill'
} as const;
```

## 粒度级别 (Granularity)
```typescript
// 后端返回数值
enum Granularity {
  SERVICE = 0,
  POD = 1,
  SPAN = 2,
  METRIC = 3
}
```

## 重要字段映射

### 项目相关
| 后端字段 | 前端使用 | 说明 |
|---------|---------|------|
| is_public | visibility | 可见性 |
| is_system | system | 是否系统级 |
| is_deleted | deleted | 是否已删除 |

### 分页相关
| 后端字段 | 前端使用 | 说明 |
|---------|---------|------|
| page | current | 当前页 |
| size | pageSize | 每页大小 |
| total | total | 总数 |
| items | data | 数据列表 |

### 标签相关
| 后端字段 | 前端使用 | 说明 |
|---------|---------|------|
| labels | tags | 标签列表 |
| key | name | 标签键 |
| value | value | 标签值 |
| category | type | 标签分类 |

## 配置字段说明

### 故障配置 (FaultConfig)
```typescript
interface FaultConfig {
  // 通用字段
  duration: number;        // 持续时间（秒）
  probability: number;     // 概率（0-100）

  // Network特有
  delay?: number;          // 延迟时间（ms）
  jitter?: number;         // 抖动（ms）
  loss?: number;           // 丢包率（%）
  corrupt?: number;        // 损坏率（%）
  duplicate?: number;      // 重复率（%）

  // CPU/Memory特有
  workers?: number;        // 工作线程数
  load?: number;           // 负载百分比

  // IO特有
  path?: string;           // 路径
  method?: string;         // 方法

  // 目标选择
  target: {
    type: 'service' | 'pod' | 'container';
    name: string;
    namespace?: string;
    labels?: Record<string, string>;
  };
}
```

## 注意事项

1. **严禁自创字段**：所有字段必须与后端API文档完全一致
2. **时间格式**：后端返回ISO 8601格式字符串，前端需要转换为Date对象
3. **状态码**：使用数值枚举，不要自创字符串状态
4. **ID字段**：统一使用字符串类型，即使后端是数字
5. **配置字段**：保持原样传递，不做任何修改

## 更新记录

- 2025-01-12: 初始版本，记录常见字段映射关系