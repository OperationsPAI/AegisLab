# InjectionDetailResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**benchmark_id** | **number** |  | [optional] [default to undefined]
**benchmark_name** | **string** |  | [optional] [default to undefined]
**category** | **string** |  | [optional] [default to undefined]
**created_at** | **string** |  | [optional] [default to undefined]
**description** | **string** |  | [optional] [default to undefined]
**display_config** | **object** |  | [optional] [default to undefined]
**end_time** | **string** |  | [optional] [default to undefined]
**engine_config** | **Array&lt;object&gt;** |  | [optional] [default to undefined]
**fault_type** | **string** |  | [optional] [default to undefined]
**ground_truth** | [**Array&lt;ChaosGroundtruth&gt;**](ChaosGroundtruth.md) |  | [optional] [default to undefined]
**id** | **number** |  | [optional] [default to undefined]
**labels** | [**Array&lt;LabelItem&gt;**](LabelItem.md) |  | [optional] [default to undefined]
**name** | **string** |  | [optional] [default to undefined]
**pedestal_id** | **number** |  | [optional] [default to undefined]
**pedestal_name** | **string** |  | [optional] [default to undefined]
**pre_duration** | **number** |  | [optional] [default to undefined]
**start_time** | **string** |  | [optional] [default to undefined]
**state** | **string** |  | [optional] [default to undefined]
**status** | **string** |  | [optional] [default to undefined]
**task_id** | **string** |  | [optional] [default to undefined]
**updated_at** | **string** |  | [optional] [default to undefined]

## Example

```typescript
import { InjectionDetailResp } from 'rcabench-client';

const instance: InjectionDetailResp = {
    benchmark_id,
    benchmark_name,
    category,
    created_at,
    description,
    display_config,
    end_time,
    engine_config,
    fault_type,
    ground_truth,
    id,
    labels,
    name,
    pedestal_id,
    pedestal_name,
    pre_duration,
    start_time,
    state,
    status,
    task_id,
    updated_at,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
