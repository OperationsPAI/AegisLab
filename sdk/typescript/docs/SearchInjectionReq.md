# SearchInjectionReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**benchmarks** | **Array&lt;string&gt;** |  | [optional] [default to undefined]
**categories** | [**Array&lt;ChaosSystemType&gt;**](ChaosSystemType.md) |  | [optional] [default to undefined]
**created_at** | [**DateRange**](DateRange.md) |  | [optional] [default to undefined]
**end_time** | [**DateRange**](DateRange.md) |  | [optional] [default to undefined]
**fault_types** | [**Array&lt;ChaosChaosType&gt;**](ChaosChaosType.md) |  | [optional] [default to undefined]
**include_labels** | **boolean** | Whether to include labels in the response | [optional] [default to undefined]
**include_task** | **boolean** | Whether to include task details in the response | [optional] [default to undefined]
**labels** | [**Array&lt;LabelItem&gt;**](LabelItem.md) | Custom labels to filter by | [optional] [default to undefined]
**name_pattern** | **string** |  | [optional] [default to undefined]
**names** | **Array&lt;string&gt;** |  | [optional] [default to undefined]
**page** | **number** |  | [optional] [default to undefined]
**size** | [**PageSize**](PageSize.md) |  | [optional] [default to undefined]
**sort** | [**Array&lt;SortOption&gt;**](SortOption.md) |  | [optional] [default to undefined]
**start_time** | [**DateRange**](DateRange.md) |  | [optional] [default to undefined]
**states** | [**Array&lt;DatapackState&gt;**](DatapackState.md) |  | [optional] [default to undefined]
**status** | [**Array&lt;StatusType&gt;**](StatusType.md) |  | [optional] [default to undefined]
**task_ids** | **Array&lt;string&gt;** |  | [optional] [default to undefined]
**updated_at** | [**DateRange**](DateRange.md) |  | [optional] [default to undefined]

## Example

```typescript
import { SearchInjectionReq } from 'rcabench-client';

const instance: SearchInjectionReq = {
    benchmarks,
    categories,
    created_at,
    end_time,
    fault_types,
    include_labels,
    include_task,
    labels,
    name_pattern,
    names,
    page,
    size,
    sort,
    start_time,
    states,
    status,
    task_ids,
    updated_at,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
