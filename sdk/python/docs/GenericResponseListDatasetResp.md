# GenericResponseListDatasetResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**code** | **int** | Status code | [optional] 
**data** | [**ListDatasetResp**](ListDatasetResp.md) | Generic type data | [optional] 
**message** | **str** | Response message | [optional] 
**timestamp** | **int** | Response generation time | [optional] 

## Example

```python
from rcabench.openapi.models.generic_response_list_dataset_resp import GenericResponseListDatasetResp

# TODO update the JSON string below
json = "{}"
# create an instance of GenericResponseListDatasetResp from a JSON string
generic_response_list_dataset_resp_instance = GenericResponseListDatasetResp.from_json(json)
# print the JSON string representation of the object
print(GenericResponseListDatasetResp.to_json())

# convert the object into a dict
generic_response_list_dataset_resp_dict = generic_response_list_dataset_resp_instance.to_dict()
# create an instance of GenericResponseListDatasetResp from a dict
generic_response_list_dataset_resp_from_dict = GenericResponseListDatasetResp.from_dict(generic_response_list_dataset_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


