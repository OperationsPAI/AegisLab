# ChaosSystemResource


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**app_labels** | **Array&lt;string&gt;** |  | [optional] [default to undefined]
**container_names** | **Array&lt;string&gt;** |  | [optional] [default to undefined]
**database_app_names** | **Array&lt;string&gt;** |  | [optional] [default to undefined]
**dns_app_names** | **Array&lt;string&gt;** |  | [optional] [default to undefined]
**http_app_names** | **Array&lt;string&gt;** |  | [optional] [default to undefined]
**jvm_app_names** | **Array&lt;string&gt;** |  | [optional] [default to undefined]
**network_pairs** | [**Array&lt;ChaosPair&gt;**](ChaosPair.md) |  | [optional] [default to undefined]

## Example

```typescript
import { ChaosSystemResource } from 'rcabench-client';

const instance: ChaosSystemResource = {
    app_labels,
    container_names,
    database_app_names,
    dns_app_names,
    http_app_names,
    jvm_app_names,
    network_pairs,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
