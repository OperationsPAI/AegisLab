# DtoAlgorithmResponse


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**command** | **str** |  | [optional] 
**created_at** | **str** |  | [optional] 
**env_vars** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**image** | **str** |  | [optional] 
**is_public** | **bool** |  | [optional] 
**name** | **str** |  | [optional] 
**project** | [**DtoProjectResponse**](DtoProjectResponse.md) | Related entities (only included when specifically requested) | [optional] 
**project_id** | **int** |  | [optional] 
**status** | **int** |  | [optional] 
**tag** | **str** |  | [optional] 
**type** | **str** |  | [optional] 
**updated_at** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_algorithm_response import DtoAlgorithmResponse

# TODO update the JSON string below
json = "{}"
# create an instance of DtoAlgorithmResponse from a JSON string
dto_algorithm_response_instance = DtoAlgorithmResponse.from_json(json)
# print the JSON string representation of the object
print(DtoAlgorithmResponse.to_json())

# convert the object into a dict
dto_algorithm_response_dict = dto_algorithm_response_instance.to_dict()
# create an instance of DtoAlgorithmResponse from a dict
dto_algorithm_response_from_dict = DtoAlgorithmResponse.from_dict(dto_algorithm_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


