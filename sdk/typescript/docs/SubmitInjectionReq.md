# SubmitInjectionReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithms** | [**Array&lt;ContainerSpec&gt;**](ContainerSpec.md) | RCA algorithms to execute (optional) | [optional] [default to undefined]
**benchmark** | [**ContainerSpec**](ContainerSpec.md) | Benchmark (detector) configuration | [default to undefined]
**interval** | **number** | Total experiment interval in minutes | [default to undefined]
**labels** | [**Array&lt;LabelItem&gt;**](LabelItem.md) | Labels to attach to the injection | [optional] [default to undefined]
**pedestal** | [**ContainerSpec**](ContainerSpec.md) | Pedestal (workload) configuration | [default to undefined]
**pre_duration** | **number** | Normal data collection duration before fault injection | [default to undefined]
**project_name** | **string** | Project name | [default to undefined]
**specs** | **Array&lt;Array&lt;ChaosNode&gt;&gt;** | Fault injection specs - 2D array where each sub-array is a batch of parallel faults | [default to undefined]

## Example

```typescript
import { SubmitInjectionReq } from 'rcabench-client';

const instance: SubmitInjectionReq = {
    algorithms,
    benchmark,
    interval,
    labels,
    pedestal,
    pre_duration,
    project_name,
    specs,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
