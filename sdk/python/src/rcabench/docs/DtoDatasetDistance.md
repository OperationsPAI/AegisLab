# DtoDatasetDistance


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**dataset1** | **str** |  | [optional] 
**dataset2** | **str** |  | [optional] 
**distance** | **float** |  | [optional] 
**figma** | **float** |  | [optional] 
**pi** | **float** |  | [optional] 
**tao** | **float** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_dataset_distance import DtoDatasetDistance

# TODO update the JSON string below
json = "{}"
# create an instance of DtoDatasetDistance from a JSON string
dto_dataset_distance_instance = DtoDatasetDistance.from_json(json)
# print the JSON string representation of the object
print(DtoDatasetDistance.to_json())

# convert the object into a dict
dto_dataset_distance_dict = dto_dataset_distance_instance.to_dict()
# create an instance of DtoDatasetDistance from a dict
dto_dataset_distance_from_dict = DtoDatasetDistance.from_dict(dto_dataset_distance_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


