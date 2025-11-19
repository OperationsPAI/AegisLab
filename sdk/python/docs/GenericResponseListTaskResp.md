# GenericResponseListTaskResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**ListTaskResp**](ListTaskResp.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from openapi.models.generic_response_list_task_resp import GenericResponseListTaskResp

# TODO update the JSON string below
json = "{}"
# create an instance of GenericResponseListTaskResp from a JSON string
generic_response_list_task_resp_instance = GenericResponseListTaskResp.from_json(json)
# print the JSON string representation of the object
print(GenericResponseListTaskResp.to_json())

# convert the object into a dict
generic_response_list_task_resp_dict = generic_response_list_task_resp_instance.to_dict()
# create an instance of GenericResponseListTaskResp from a dict
generic_response_list_task_resp_from_dict = GenericResponseListTaskResp.from_dict(generic_response_list_task_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


