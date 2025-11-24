# GenericResponseContainerDetailResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**ContainerDetailResp**](ContainerDetailResp.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.generic_response_container_detail_resp import GenericResponseContainerDetailResp

# TODO update the JSON string below
json = "{}"
# create an instance of GenericResponseContainerDetailResp from a JSON string
generic_response_container_detail_resp_instance = GenericResponseContainerDetailResp.from_json(json)
# print the JSON string representation of the object
print(GenericResponseContainerDetailResp.to_json())

# convert the object into a dict
generic_response_container_detail_resp_dict = generic_response_container_detail_resp_instance.to_dict()
# create an instance of GenericResponseContainerDetailResp from a dict
generic_response_container_detail_resp_from_dict = GenericResponseContainerDetailResp.from_dict(generic_response_container_detail_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


