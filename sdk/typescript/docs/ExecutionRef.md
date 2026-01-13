# ExecutionRef


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**detector_results** | [**Array&lt;DetectorResultItem&gt;**](DetectorResultItem.md) | Detector results | [optional] [default to undefined]
**executed_at** | **string** | Execution time | [optional] [default to undefined]
**execution_duration** | **number** | Execution duration in seconds | [optional] [default to undefined]
**execution_id** | **number** | Execution ID | [optional] [default to undefined]
**predictions** | [**Array&lt;GranularityResultItem&gt;**](GranularityResultItem.md) | Algorithm predictions | [optional] [default to undefined]

## Example

```typescript
import { ExecutionRef } from 'rcabench-client';

const instance: ExecutionRef = {
    detector_results,
    executed_at,
    execution_duration,
    execution_id,
    predictions,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
