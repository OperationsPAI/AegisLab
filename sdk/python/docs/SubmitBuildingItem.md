# SubmitBuildingItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**index** | **int** |  | [optional] 
**task_id** | **str** |  | [optional] 
**trace_id** | **str** |  | [optional] 

## Example

```python
from openapi.models.submit_building_item import SubmitBuildingItem

# TODO update the JSON string below
json = "{}"
# create an instance of SubmitBuildingItem from a JSON string
submit_building_item_instance = SubmitBuildingItem.from_json(json)
# print the JSON string representation of the object
print(SubmitBuildingItem.to_json())

# convert the object into a dict
submit_building_item_dict = submit_building_item_instance.to_dict()
# create an instance of SubmitBuildingItem from a dict
submit_building_item_from_dict = SubmitBuildingItem.from_dict(submit_building_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


