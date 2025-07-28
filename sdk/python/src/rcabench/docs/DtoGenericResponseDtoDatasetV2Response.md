# DtoGenericResponseDtoDatasetV2Response


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | 状态码 | [optional] 
**data** | [**DtoDatasetV2Response**](DtoDatasetV2Response.md) | 泛型类型的数据 | [optional] 
**message** | **str** | 响应消息 | [optional] 
**timestamp** | **int** | 响应生成时间 | [optional] 

## Example

```python
from rcabench.openapi.models.dto_generic_response_dto_dataset_v2_response import DtoGenericResponseDtoDatasetV2Response

# TODO update the JSON string below
json = "{}"
# create an instance of DtoGenericResponseDtoDatasetV2Response from a JSON string
dto_generic_response_dto_dataset_v2_response_instance = DtoGenericResponseDtoDatasetV2Response.from_json(json)
# print the JSON string representation of the object
print(DtoGenericResponseDtoDatasetV2Response.to_json())

# convert the object into a dict
dto_generic_response_dto_dataset_v2_response_dict = dto_generic_response_dto_dataset_v2_response_instance.to_dict()
# create an instance of DtoGenericResponseDtoDatasetV2Response from a dict
dto_generic_response_dto_dataset_v2_response_from_dict = DtoGenericResponseDtoDatasetV2Response.from_dict(dto_generic_response_dto_dataset_v2_response_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


