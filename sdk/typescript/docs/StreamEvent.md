# StreamEvent


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**event_name** | [**EventType**](EventType.md) |  | [optional] [default to undefined]
**payload** | **object** |  | [optional] [default to undefined]
**task_id** | **string** |  | [optional] [default to undefined]
**task_type** | [**TaskType**](TaskType.md) |  | [optional] [default to undefined]

## Example

```typescript
import { StreamEvent } from 'rcabench-client';

const instance: StreamEvent = {
    event_name,
    payload,
    task_id,
    task_type,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
