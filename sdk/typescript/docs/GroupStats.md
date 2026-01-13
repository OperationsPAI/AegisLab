# GroupStats


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**avg_duration** | **number** |  | [optional] [default to undefined]
**max_duration** | **number** |  | [optional] [default to undefined]
**min_duration** | **number** |  | [optional] [default to undefined]
**total_traces** | **number** |  | [optional] [default to undefined]
**trace_state_map** | **{ [key: string]: Array&lt;TraceStatsItem&gt;; }** |  | [optional] [default to undefined]

## Example

```typescript
import { GroupStats } from 'rcabench-client';

const instance: GroupStats = {
    avg_duration,
    max_duration,
    min_duration,
    total_traces,
    trace_state_map,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
