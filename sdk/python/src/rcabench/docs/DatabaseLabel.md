# DatabaseLabel


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**category** | **str** | Label category (dataset, fault_injection, algorithm, container, etc.) | [optional] 
**color** | **str** | Label color (hex format) | [optional] 
**created_at** | **str** | Creation time | [optional] 
**description** | **str** | Label description | [optional] 
**id** | **int** | Unique identifier | [optional] 
**is_system** | **bool** | Whether system label | [optional] 
**key** | **str** | Label key | [optional] 
**updated_at** | **str** | Update time | [optional] 
**usage** | **int** | Usage count | [optional] 
**value** | **str** | Label value | [optional] 

## Example

```python
from rcabench.openapi.models.database_label import DatabaseLabel

# TODO update the JSON string below
json = "{}"
# create an instance of DatabaseLabel from a JSON string
database_label_instance = DatabaseLabel.from_json(json)
# print the JSON string representation of the object
print(DatabaseLabel.to_json())

# convert the object into a dict
database_label_dict = database_label_instance.to_dict()
# create an instance of DatabaseLabel from a dict
database_label_from_dict = DatabaseLabel.from_dict(database_label_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


