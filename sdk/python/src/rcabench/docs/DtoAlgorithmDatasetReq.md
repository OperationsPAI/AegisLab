# DtoAlgorithmDatasetReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithm** | **str** |  | 
**dataset** | **str** |  | 
**dataset_version** | **str** | Dataset version (optional, defaults to \&quot;v1.0\&quot;) | [optional] 
**tag** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_algorithm_dataset_req import DtoAlgorithmDatasetReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoAlgorithmDatasetReq from a JSON string
dto_algorithm_dataset_req_instance = DtoAlgorithmDatasetReq.from_json(json)
# print the JSON string representation of the object
print(DtoAlgorithmDatasetReq.to_json())

# convert the object into a dict
dto_algorithm_dataset_req_dict = dto_algorithm_dataset_req_instance.to_dict()
# create an instance of DtoAlgorithmDatasetReq from a dict
dto_algorithm_dataset_req_from_dict = DtoAlgorithmDatasetReq.from_dict(dto_algorithm_dataset_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


