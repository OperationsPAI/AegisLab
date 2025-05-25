# DtoDatasetDeleteResp


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**failed_names** | **List[str]** |  | [optional] 
**success_count** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_dataset_delete_resp import DtoDatasetDeleteResp

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatasetDeleteResp from a JSON string
dto_dataset_delete_resp_instance = DtoDatasetDeleteResp.from_json(json)
# print the JSON string representation of the object
print(DtoDatasetDeleteResp.to_json())

# convert the object into a dict
dto_dataset_delete_resp_dict = dto_dataset_delete_resp_instance.to_dict()
# create an instance of DtoDatasetDeleteResp from a dict
dto_dataset_delete_resp_from_dict = DtoDatasetDeleteResp.from_dict(dto_dataset_delete_resp_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


