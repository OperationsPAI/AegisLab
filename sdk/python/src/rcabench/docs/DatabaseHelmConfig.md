# DatabaseHelmConfig


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**chart_name** | **str** | Helm chart information | [optional] 
**chart_version** | **str** | Chart version | [optional] 
**full_chart** | **str** | Full chart reference (not stored in DB, used for display) | [optional] 
**id** | **int** | Unique identifier | [optional] 
**ns_prefix** | **str** | Deployment configuration | [optional] 
**port_template** | **str** | Port template for dynamic port assignment, e.g., \&quot;31%03d\&quot; | [optional] 
**repo_name** | **str** | Repository name | [optional] 
**repo_url** | **str** | Repository URL | [optional] 
**values** | **str** | Helm values in JSON format | [optional] 

## Example

```python
from rcabench.openapi.models.database_helm_config import DatabaseHelmConfig

# TODO update the JSON string below
json = "{}"
# create an instance of DatabaseHelmConfig from a JSON string
database_helm_config_instance = DatabaseHelmConfig.from_json(json)
# print the JSON string representation of the object
print(DatabaseHelmConfig.to_json())

# convert the object into a dict
database_helm_config_dict = database_helm_config_instance.to_dict()
# create an instance of DatabaseHelmConfig from a dict
database_helm_config_from_dict = DatabaseHelmConfig.from_dict(database_helm_config_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


