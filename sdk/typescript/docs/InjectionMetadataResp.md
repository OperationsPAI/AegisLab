# InjectionMetadataResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**config** | [**ChaosNode**](ChaosNode.md) |  | [optional] [default to undefined]
**fault_resource_map** | [**{ [key: string]: ChaosChaosResourceMapping; }**](ChaosChaosResourceMapping.md) |  | [optional] [default to undefined]
**fault_type_map** | **{ [key: string]: string; }** |  | [optional] [default to undefined]
**ns_resources** | [**ChaosSystemResource**](ChaosSystemResource.md) |  | [optional] [default to undefined]

## Example

```typescript
import { InjectionMetadataResp } from 'rcabench-client';

const instance: InjectionMetadataResp = {
    config,
    fault_resource_map,
    fault_type_map,
    ns_resources,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
