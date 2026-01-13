# ContainerVersionItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**command** | **string** |  | [optional] [default to undefined]
**container_id** | **number** |  | [optional] [default to undefined]
**container_name** | **string** |  | [optional] [default to undefined]
**env_vars** | [**Array&lt;ParameterItem&gt;**](ParameterItem.md) |  | [optional] [default to undefined]
**extra** | [**HelmConfigItem**](HelmConfigItem.md) |  | [optional] [default to undefined]
**id** | **number** |  | [optional] [default to undefined]
**image_ref** | **string** |  | [optional] [default to undefined]
**name** | **string** |  | [optional] [default to undefined]

## Example

```typescript
import { ContainerVersionItem } from 'rcabench-client';

const instance: ContainerVersionItem = {
    command,
    container_id,
    container_name,
    env_vars,
    extra,
    id,
    image_ref,
    name,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
