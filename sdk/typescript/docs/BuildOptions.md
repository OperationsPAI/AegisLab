# BuildOptions


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**build_args** | **object** |  | [optional] [default to undefined]
**context_dir** | **string** |  | [optional] [default to '.']
**dockerfile_path** | **string** |  | [optional] [default to 'Dockerfile']
**force_rebuild** | **boolean** |  | [optional] [default to undefined]
**target** | **string** |  | [optional] [default to undefined]

## Example

```typescript
import { BuildOptions } from 'rcabench-client';

const instance: BuildOptions = {
    build_args,
    context_dir,
    dockerfile_path,
    force_rebuild,
    target,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
