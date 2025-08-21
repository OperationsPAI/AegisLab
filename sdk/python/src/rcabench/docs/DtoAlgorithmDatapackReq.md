# DtoAlgorithmDatapackReq


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithm** | **str** |  | 
**datapack** | **str** |  | 
**tag** | **str** | Tag filter for filtering execution results | [optional] 

## Example

```python
from rcabench.openapi.models.dto_algorithm_datapack_req import DtoAlgorithmDatapackReq

# TODO update the JSON string below
json = "{}"
# create an instance of DtoAlgorithmDatapackReq from a JSON string
dto_algorithm_datapack_req_instance = DtoAlgorithmDatapackReq.from_json(json)
# print the JSON string representation of the object
print(DtoAlgorithmDatapackReq.to_json())

# convert the object into a dict
dto_algorithm_datapack_req_dict = dto_algorithm_datapack_req_instance.to_dict()
# create an instance of DtoAlgorithmDatapackReq from a dict
dto_algorithm_datapack_req_from_dict = DtoAlgorithmDatapackReq.from_dict(dto_algorithm_datapack_req_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


