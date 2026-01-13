# ExecutionDetailResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithm_id** | **number** |  | [optional] [default to undefined]
**algorithm_name** | **string** |  | [optional] [default to undefined]
**algorithm_version** | **string** |  | [optional] [default to undefined]
**algorithm_version_id** | **number** |  | [optional] [default to undefined]
**created_at** | **string** |  | [optional] [default to undefined]
**datapack_id** | **number** |  | [optional] [default to undefined]
**datapack_name** | **string** |  | [optional] [default to undefined]
**detector_results** | [**Array&lt;DetectorResultItem&gt;**](DetectorResultItem.md) |  | [optional] [default to undefined]
**duration** | **number** |  | [optional] [default to undefined]
**granularity_results** | [**Array&lt;GranularityResultItem&gt;**](GranularityResultItem.md) |  | [optional] [default to undefined]
**id** | **number** |  | [optional] [default to undefined]
**labels** | [**Array&lt;LabelItem&gt;**](LabelItem.md) |  | [optional] [default to undefined]
**state** | **string** |  | [optional] [default to undefined]
**status** | **string** |  | [optional] [default to undefined]
**task_id** | **string** |  | [optional] [default to undefined]
**updated_at** | **string** |  | [optional] [default to undefined]

## Example

```typescript
import { ExecutionDetailResp } from 'rcabench-client';

const instance: ExecutionDetailResp = {
    algorithm_id,
    algorithm_name,
    algorithm_version,
    algorithm_version_id,
    created_at,
    datapack_id,
    datapack_name,
    detector_results,
    duration,
    granularity_results,
    id,
    labels,
    state,
    status,
    task_id,
    updated_at,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
