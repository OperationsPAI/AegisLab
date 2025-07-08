# DtoSuccessfulExecutionItem


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**algorithm** | **str** | 算法名称 | [optional] 
**created_at** | **str** | 创建时间 | [optional] 
**dataset** | **str** | 数据集名称 | [optional] 
**id** | **int** | 执行ID | [optional] 

## Example

```python
from rcabench.openapi.models.dto_successful_execution_item import DtoSuccessfulExecutionItem

# TODO update the JSON string below
json = "{}"
# create an instance of DtoSuccessfulExecutionItem from a JSON string
dto_successful_execution_item_instance = DtoSuccessfulExecutionItem.from_json(json)
# print the JSON string representation of the object
print(DtoSuccessfulExecutionItem.to_json())

# convert the object into a dict
dto_successful_execution_item_dict = dto_successful_execution_item_instance.to_dict()
# create an instance of DtoSuccessfulExecutionItem from a dict
dto_successful_execution_item_from_dict = DtoSuccessfulExecutionItem.from_dict(dto_successful_execution_item_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


