# CreateDatasetReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**description** | **str** |  | [optional] 
**is_public** | **bool** |  | [optional] 
**name** | **str** |  | 
**type** | **str** |  | 
**version** | [**CreateDatasetVersionReq**](CreateDatasetVersionReq.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.create_dataset_req import CreateDatasetReq

# TODO update the JSON string below
json = "{}"
# create an instance of CreateDatasetReq from a JSON string
create_dataset_req_instance = CreateDatasetReq.from_json(json)
# print the JSON string representation of the object
print(CreateDatasetReq.to_json())

# convert the object into a dict
create_dataset_req_dict = create_dataset_req_instance.to_dict()
# create an instance of CreateDatasetReq from a dict
create_dataset_req_from_dict = CreateDatasetReq.from_dict(create_dataset_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


