# TraceStatsItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**current_event** | **string** |  | [optional] [default to undefined]
**current_task** | **string** |  | [optional] [default to undefined]
**end_time** | **string** |  | [optional] [default to undefined]
**start_time** | **string** |  | [optional] [default to undefined]
**state** | **string** |  | [optional] [default to undefined]
**task_type_durations** | **object** | Average durations per task type in seconds | [optional] [default to undefined]
**trace_id** | **string** |  | [optional] [default to undefined]
**type** | **string** |  | [optional] [default to undefined]

## Example

```typescript
import { TraceStatsItem } from 'rcabench-client';

const instance: TraceStatsItem = {
    current_event,
    current_task,
    end_time,
    start_time,
    state,
    task_type_durations,
    trace_id,
    type,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
