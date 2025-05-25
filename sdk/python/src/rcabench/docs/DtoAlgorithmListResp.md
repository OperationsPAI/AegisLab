# DtoAlgorithmListResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithms** | **List[str]** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_algorithm_list_resp import DtoAlgorithmListResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoAlgorithmListResp from a JSON string
dto_algorithm_list_resp_instance = DtoAlgorithmListResp.from_json(json)
# print the JSON string representation of the object
print(DtoAlgorithmListResp.to_json())

# convert the object into a dict
dto_algorithm_list_resp_dict = dto_algorithm_list_resp_instance.to_dict()
# create an instance of DtoAlgorithmListResp from a dict
dto_algorithm_list_resp_from_dict = DtoAlgorithmListResp.from_dict(dto_algorithm_list_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


