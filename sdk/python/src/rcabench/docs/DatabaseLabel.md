# DatabaseLabel


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**category** | **str** | 标签分类 (dataset, fault_injection, algorithm, container等) | [optional] 
**color** | **str** | 标签颜色 (hex格式) | [optional] 
**created_at** | **str** | 创建时间 | [optional] 
**description** | **str** | 标签描述 | [optional] 
**id** | **int** | 唯一标识 | [optional] 
**is_system** | **bool** | 是否为系统标签 | [optional] 
**key** | **str** | 标签键 | [optional] 
**updated_at** | **str** | 更新时间 | [optional] 
**usage** | **int** | 使用次数 | [optional] 
**value** | **str** | 标签值 | [optional] 

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


