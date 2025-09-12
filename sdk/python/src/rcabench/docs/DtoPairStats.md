# DtoPairStats


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**in_degree** | **int** |  | [optional] 
**name** | **str** |  | [optional] 
**out_degree** | **int** |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_pair_stats import DtoPairStats

# TODO update the JSON string below
json = "{}"
# create an instance of DtoPairStats from a JSON string
dto_pair_stats_instance = DtoPairStats.from_json(json)
# print the JSON string representation of the object
print(DtoPairStats.to_json())

# convert the object into a dict
dto_pair_stats_dict = dto_pair_stats_instance.to_dict()
# create an instance of DtoPairStats from a dict
dto_pair_stats_from_dict = DtoPairStats.from_dict(dto_pair_stats_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


