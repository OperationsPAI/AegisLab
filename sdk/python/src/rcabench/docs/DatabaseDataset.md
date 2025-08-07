# DatabaseDataset


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**checksum** | **str** | File checksum | [optional] 
**created_at** | **str** | Creation time | [optional] 
**data_source** | **str** | Data source description | [optional] 
**dataset_version** | **str** | Dataset version with size limit | [optional] 
**description** | **str** | Dataset description | [optional] 
**download_url** | **str** | Download link with size limit | [optional] 
**file_count** | **int** | File count with validation | [optional] 
**format** | **str** | Data format (json, csv, parquet, etc.) | [optional] 
**id** | **int** | Unique identifier | [optional] 
**is_public** | **bool** | Whether public | [optional] 
**name** | **str** | Dataset name with size limit | [optional] 
**status** | **int** | Status: -1:deleted 0:disabled 1:enabled | [optional] 
**type** | **str** | Dataset type (e.g., \&quot;microservice\&quot;, \&quot;database\&quot;, \&quot;network\&quot;) | [optional] 
**updated_at** | **str** | Update time | [optional] 

## Example

```python
from rcabench.openapi.models.database_dataset import DatabaseDataset

# TODO update the JSON string below
json = "{}"
# create an instance of DatabaseDataset from a JSON string
database_dataset_instance = DatabaseDataset.from_json(json)
# print the JSON string representation of the object
print(DatabaseDataset.to_json())

# convert the object into a dict
database_dataset_dict = database_dataset_instance.to_dict()
# create an instance of DatabaseDataset from a dict
database_dataset_from_dict = DatabaseDataset.from_dict(database_dataset_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


