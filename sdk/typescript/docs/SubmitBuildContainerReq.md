# SubmitBuildContainerReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**build_options** | [**BuildOptions**](BuildOptions.md) |  | [optional] [default to undefined]
**github_branch** | **string** |  | [optional] [default to undefined]
**github_commit** | **string** |  | [optional] [default to undefined]
**github_repository** | **string** | GitHub repository information | [default to undefined]
**github_token** | **string** |  | [optional] [default to undefined]
**image_name** | **string** | Container Meta | [default to undefined]
**sub_path** | **string** |  | [optional] [default to undefined]
**tag** | **string** |  | [optional] [default to undefined]

## Example

```typescript
import { SubmitBuildContainerReq } from 'rcabench-client';

const instance: SubmitBuildContainerReq = {
    build_options,
    github_branch,
    github_commit,
    github_repository,
    github_token,
    image_name,
    sub_path,
    tag,
};
```

[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)
