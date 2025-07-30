# DatabaseProject


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**created_at** | **str** |  | [optional] 
**description** | **str** |  | [optional] 
**id** | **int** |  | [optional] 
**is_public** | **bool** | Whether publicly visible | [optional] 
**name** | **str** |  | [optional] 
**status** | **int** | 0:disabled 1:enabled -1:deleted | [optional] 
**updated_at** | **str** |  | [optional] 

## Example

```python
from rcabench.openapi.models.database_project import DatabaseProject

# TODO update the JSON string below
json = "{}"
# create an instance of DatabaseProject from a JSON string
database_project_instance = DatabaseProject.from_json(json)
# print the JSON string representation of the object
print(DatabaseProject.to_json())

# convert the object into a dict
database_project_dict = database_project_instance.to_dict()
# create an instance of DatabaseProject from a dict
database_project_from_dict = DatabaseProject.from_dict(database_project_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


