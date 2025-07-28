# DatabaseUser


## Properties

Name | Type | Description | Notes
------------ | ------------- | ------------- | -------------
**avatar** | **str** | 头像URL | [optional] 
**created_at** | **str** | 创建时间 | [optional] 
**email** | **str** | 邮箱（唯一） | [optional] 
**full_name** | **str** | 全名 | [optional] 
**id** | **int** | 唯一标识 | [optional] 
**is_active** | **bool** | 是否激活 | [optional] 
**last_login_at** | **str** | 最后登录时间 | [optional] 
**phone** | **str** | 电话号码 | [optional] 
**status** | **int** | 0:禁用 1:启用 -1:删除 | [optional] 
**updated_at** | **str** | 更新时间 | [optional] 
**username** | **str** | 用户名（唯一） | [optional] 

## Example

```python
from rcabench.openapi.models.database_user import DatabaseUser

# TODO update the JSON string below
json = "{}"
# create an instance of DatabaseUser from a JSON string
database_user_instance = DatabaseUser.from_json(json)
# print the JSON string representation of the object
print(DatabaseUser.to_json())

# convert the object into a dict
database_user_dict = database_user_instance.to_dict()
# create an instance of DatabaseUser from a dict
database_user_from_dict = DatabaseUser.from_dict(database_user_dict)
```
[[Back to Model list]](../README.md#documentation-for-models) [[Back to API list]](../README.md#documentation-for-api-endpoints) [[Back to README]](../README.md)


