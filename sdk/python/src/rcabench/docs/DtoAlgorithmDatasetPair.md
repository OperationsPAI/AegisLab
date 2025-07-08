# DtoAlgorithmDatasetPair


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithm** | **str** |  | [optional] 
**dataset** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_algorithm_dataset_pair import DtoAlgorithmDatasetPair

# TODO update the JSON string below
json = "{}"
# create an instance of DtoAlgorithmDatasetPair from a JSON string
dto_algorithm_dataset_pair_instance = DtoAlgorithmDatasetPair.from_json(json)
# print the JSON string representation of the object
print(DtoAlgorithmDatasetPair.to_json())

# convert the object into a dict
dto_algorithm_dataset_pair_dict = dto_algorithm_dataset_pair_instance.to_dict()
# create an instance of DtoAlgorithmDatasetPair from a dict
dto_algorithm_dataset_pair_from_dict = DtoAlgorithmDatasetPair.from_dict(dto_algorithm_dataset_pair_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


