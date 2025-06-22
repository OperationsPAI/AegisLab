# DtoRawDataItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithm** | **str** |  | [optional] 
**dataset** | **str** |  | [optional] 
**entries** | [**List[DtoGranularityRecord]**](DtoGranularityRecord.md) |  | [optional] 
**groundtruth** | [**HandlerGroundtruth**](HandlerGroundtruth.md) |  | [optional] 

## Example

```python
from rcabench.openapi.models.dto_raw_data_item import DtoRawDataItem

# TODO update the JSON string below
json = "{}"
# create an instance of DtoRawDataItem from a JSON string
dto_raw_data_item_instance = DtoRawDataItem.from_json(json)
# print the JSON string representation of the object
print(DtoRawDataItem.to_json())

# convert the object into a dict
dto_raw_data_item_dict = dto_raw_data_item_instance.to_dict()
# create an instance of DtoRawDataItem from a dict
dto_raw_data_item_from_dict = DtoRawDataItem.from_dict(dto_raw_data_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


