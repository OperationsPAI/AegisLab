# CreateContainerVersionReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**command** | **string** |  | [optional] [default to undefined]
**env_vars** | [**Array&lt;CreateParameterConfigReq&gt;**](CreateParameterConfigReq.md) |  | [optional] [default to undefined]
**github_link** | **string** |  | [optional] [default to undefined]
**helm_config** | [**CreateHelmConfigReq**](CreateHelmConfigReq.md) |  | [optional] [default to undefined]
**image_ref** | **string** |  | [default to undefined]
**name** | **string** |  | [default to undefined]

## Example

```typescript
import { CreateContainerVersionReq } from 'rcabench-client';

const instance: CreateContainerVersionReq = {
    command,
    env_vars,
    github_link,
    helm_config,
    image_ref,
    name,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
