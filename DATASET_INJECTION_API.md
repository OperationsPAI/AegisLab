# Dataset Injection API 更新

## 概述

现在支持通过 injection name 或 injection ID 来创建和更新数据集。这解决了用户只知道 injection name 而不知道 ID 的问题。

## 新的 API 字段

### InjectionRef 结构体

```go
type InjectionRef struct {
    ID   *int    `json:"id,omitempty"`   // Injection ID
    Name *string `json:"name,omitempty"` // Injection name
}
```

### 更新后的请求结构体

#### DatasetV2CreateReq
```go
type DatasetV2CreateReq struct {
    // ... 其他字段 ...
    InjectionRefs []InjectionRef           `json:"injection_refs"`   // 通过 ID 或 name 引用 injection
    // ... 其他字段 ...
}
```

#### DatasetV2UpdateReq
```go
type DatasetV2UpdateReq struct {
    // ... 其他字段 ...
    InjectionRefs []InjectionRef           `json:"injection_refs"`   // 通过 ID 或 name 引用 injection
    // ... 其他字段 ...
}
```

## 使用示例

### 1. 通过 ID 创建数据集

```json
{
  "name": "my-dataset",
  "type": "test",
  "description": "A test dataset",
  "injection_refs": [
    {
      "id": 123
    },
    {
      "id": 456
    }
  ]
}
```

### 2. 通过 name 创建数据集

```json
{
  "name": "my-dataset",
  "type": "test",
  "description": "A test dataset",
  "injection_refs": [
    {
      "name": "injection-001"
    },
    {
      "name": "injection-002"
    }
  ]
}
```

### 3. 混合使用 ID 和 name

```json
{
  "name": "my-dataset",
  "type": "test",
  "description": "A test dataset",
  "injection_refs": [
    {
      "id": 123
    },
    {
      "name": "injection-002"
    }
  ]
}
```



## 错误处理

- 如果指定的 injection name 不存在，API 会返回 400 错误
- 如果指定的 injection ID 不存在，API 会返回 400 错误
- 如果 InjectionRef 既没有指定 ID 也没有指定 name，API 会返回 400 错误

## 实现细节

1. **创建数据集时**：系统会处理 `injection_refs`，将 name 转换为 ID
2. **更新数据集时**：如果提供了 `injection_refs`，会完全替换现有的 injection 关联
3. **统一接口**：使用 `InjectionRef` 结构体统一处理 ID 和 name 引用
4. **性能优化**：使用批量查询替代循环查询，大幅提升性能

## 优势

1. **用户友好**：用户可以通过 name 来指定 injection，而不需要知道 ID
2. **统一接口**：使用单一的结构体处理 ID 和 name 引用
3. **灵活性**：支持混合使用 ID 和 name
4. **错误处理**：提供清晰的错误信息，帮助用户快速定位问题
5. **高性能**：使用批量查询，避免 N+1 查询问题，大幅提升性能 