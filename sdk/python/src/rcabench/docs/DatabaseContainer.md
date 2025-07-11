# DatabaseContainer


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**command** | **str** | 启动命令 | [optional] 
**created_at** | **str** | 创建时间 | [optional] 
**env_vars** | **str** | 环境变量名称列表 | [optional] 
**id** | **int** | 唯一标识 | [optional] 
**image** | **str** | 镜像名 | [optional] 
**name** | **str** | 名称 | [optional] 
**status** | **bool** | 0: 已删除 1: 活跃 | [optional] 
**tag** | **str** | 镜像标签 | [optional] 
**type** | **str** | 镜像类型 | [optional] 
**updated_at** | **str** | 更新时间 | [optional] 

## Example

```python
from rcabench.openapi.models.database_container import DatabaseContainer

# TODO update the JSON string below
json = "{}"
# create an instance of DatabaseContainer from a JSON string
database_container_instance = DatabaseContainer.from_json(json)
# print the JSON string representation of the object
print(DatabaseContainer.to_json())

# convert the object into a dict
database_container_dict = database_container_instance.to_dict()
# create an instance of DatabaseContainer from a dict
database_container_from_dict = DatabaseContainer.from_dict(database_container_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


