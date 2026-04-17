# Swagger Audience Marking Report

> Source of truth: Swagger annotations in `src/module/*/handler.go` and `src/httpapi/docs.go`.
> Route position column uses the `@Router` line, then `@x-api-type`, then function line when available.

## Summary

- Total operations scanned: **173**
- Marked operations: **100**
- Empty `@x-api-type {}` operations: **73**
- Missing `@x-api-type` operations: **0**
- Audience counts among marked operations: `sdk=5` `portal=43` `admin=58`

## Marked Operations

| Method | Path | Audience | Summary | Location |
| --- | --- | --- | --- | --- |
| GET | `/api/v2/access-keys` | `portal` | List access keys | `src/module/auth/handler.go:288` / `src/module/auth/handler.go:289` |
| POST | `/api/v2/access-keys` | `portal` | Create access key | `src/module/auth/handler.go:247` / `src/module/auth/handler.go:248` |
| DELETE | `/api/v2/access-keys/{access_key_id}` | `portal` | Delete access key | `src/module/auth/handler.go:357` / `src/module/auth/handler.go:358` |
| GET | `/api/v2/access-keys/{access_key_id}` | `portal` | Get access key detail | `src/module/auth/handler.go:328` / `src/module/auth/handler.go:329` |
| POST | `/api/v2/access-keys/{access_key_id}/disable` | `portal` | Disable access key | `src/module/auth/handler.go:385` / `src/module/auth/handler.go:386` |
| POST | `/api/v2/access-keys/{access_key_id}/enable` | `portal` | Enable access key | `src/module/auth/handler.go:413` / `src/module/auth/handler.go:414` |
| POST | `/api/v2/access-keys/{access_key_id}/rotate` | `portal` | Rotate access key secret | `src/module/auth/handler.go:441` / `src/module/auth/handler.go:442` |
| POST | `/api/v2/auth/access-key/token` | `sdk` | Exchange access key for token | `src/module/auth/handler.go:472` / `src/module/auth/handler.go:473` |
| POST | `/api/v2/auth/change-password` | `portal, admin` | Change user password | `src/module/auth/handler.go:177` / `src/module/auth/handler.go:178` |
| POST | `/api/v2/auth/login` | `portal, admin` | User login | `src/module/auth/handler.go:36` / `src/module/auth/handler.go:37` |
| POST | `/api/v2/auth/logout` | `portal, admin` | User logout | `src/module/auth/handler.go:139` / `src/module/auth/handler.go:140` |
| GET | `/api/v2/auth/profile` | `portal, admin` | Get current user profile | `src/module/auth/handler.go:216` / `src/module/auth/handler.go:217` |
| POST | `/api/v2/auth/refresh` | `portal, admin` | Refresh JWT token | `src/module/auth/handler.go:106` / `src/module/auth/handler.go:107` |
| POST | `/api/v2/auth/register` | `portal, admin` | User registration | `src/module/auth/handler.go:71` / `src/module/auth/handler.go:72` |
| GET | `/api/v2/labels` | `portal` | List labels | `src/module/label/handler.go:165` / `src/module/label/handler.go:166` |
| POST | `/api/v2/labels` | `portal` | Create label | `src/module/label/handler.go:69` / `src/module/label/handler.go:70` |
| POST | `/api/v2/labels/batch-delete` | `portal` | Batch delete labels | `src/module/label/handler.go:35` / `src/module/label/handler.go:36` |
| DELETE | `/api/v2/labels/{label_id}` | `portal` | Delete label | `src/module/label/handler.go:103` / `src/module/label/handler.go:104` |
| GET | `/api/v2/labels/{label_id}` | `portal` | Get label by ID | `src/module/label/handler.go:131` / `src/module/label/handler.go:132` |
| PATCH | `/api/v2/labels/{label_id}` | `portal` | Update label | `src/module/label/handler.go:201` / `src/module/label/handler.go:202` |
| GET | `/api/v2/permissions` | `admin` | List permissions | `src/module/rbac/handler.go:328` / `src/module/rbac/handler.go:329` |
| GET | `/api/v2/permissions/{id}` | `admin` | Get permission by ID | `src/module/rbac/handler.go:296` / `src/module/rbac/handler.go:297` |
| GET | `/api/v2/permissions/{permission_id}/roles` | `admin` | List roles from permission | `src/module/rbac/handler.go:362` / `src/module/rbac/handler.go:363` |
| GET | `/api/v2/projects` | `portal` | List projects | `src/module/project/handler.go:146` / `src/module/project/handler.go:147` |
| POST | `/api/v2/projects` | `portal` | Create a new project | `src/module/project/handler.go:39` / `src/module/project/handler.go:40` |
| DELETE | `/api/v2/projects/{project_id}` | `portal` | Delete project | `src/module/project/handler.go:82` / `src/module/project/handler.go:83` |
| GET | `/api/v2/projects/{project_id}` | `portal` | Get project by ID | `src/module/project/handler.go:113` / `src/module/project/handler.go:114` |
| PATCH | `/api/v2/projects/{project_id}` | `portal` | Update project | `src/module/project/handler.go:185` / `src/module/project/handler.go:186` |
| GET | `/api/v2/projects/{project_id}/executions` | `portal` | List project executions | `src/module/execution/handler.go:44` / `src/module/execution/handler.go:45` |
| POST | `/api/v2/projects/{project_id}/executions/execute` | `portal` | Submit batch algorithm execution | `src/module/execution/handler.go:88` / `src/module/execution/handler.go:89` |
| GET | `/api/v2/projects/{project_id}/injections` | `portal` | List project fault injections | `src/module/injection/handler.go:51` / `src/module/injection/handler.go:52` |
| GET | `/api/v2/projects/{project_id}/injections/analysis/no-issues` | `portal` | List project fault injections without issues | `src/module/injection/handler.go:125` / `src/module/injection/handler.go:126` |
| GET | `/api/v2/projects/{project_id}/injections/analysis/with-issues` | `portal` | List project fault injections with issues | `src/module/injection/handler.go:155` / `src/module/injection/handler.go:156` |
| POST | `/api/v2/projects/{project_id}/injections/build` | `portal` | Submit project datapack buildings | `src/module/injection/handler.go:211` / `src/module/injection/handler.go:212` |
| POST | `/api/v2/projects/{project_id}/injections/inject` | `portal` | Submit project fault injections | `src/module/injection/handler.go:183` / `src/module/injection/handler.go:184` |
| POST | `/api/v2/projects/{project_id}/injections/search` | `portal` | Search project fault injections | `src/module/injection/handler.go:95` / `src/module/injection/handler.go:96` |
| PATCH | `/api/v2/projects/{project_id}/labels` | `portal` | Manage project custom labels | `src/module/project/handler.go:229` / `src/module/project/handler.go:230` |
| GET | `/api/v2/resources` | `admin` | List resources | `src/module/rbac/handler.go:422` / `src/module/rbac/handler.go:423` |
| GET | `/api/v2/resources/{id}` | `admin` | Get resource by ID | `src/module/rbac/handler.go:391` / `src/module/rbac/handler.go:392` |
| GET | `/api/v2/resources/{id}/permissions` | `admin` | List permissions from resource | `src/module/rbac/handler.go:456` / `src/module/rbac/handler.go:457` / `src/module/rbac/handler.go:470` |
| GET | `/api/v2/roles` | `admin` | List roles | `src/module/rbac/handler.go:127` / `src/module/rbac/handler.go:128` |
| POST | `/api/v2/roles` | `admin` | Create a new role | `src/module/rbac/handler.go:38` / `src/module/rbac/handler.go:39` |
| DELETE | `/api/v2/roles/{id}` | `admin` | Delete role | `src/module/rbac/handler.go:68` / `src/module/rbac/handler.go:69` |
| GET | `/api/v2/roles/{id}` | `admin` | Get role by ID | `src/module/rbac/handler.go:96` / `src/module/rbac/handler.go:97` |
| PATCH | `/api/v2/roles/{id}` | `admin` | Update role | `src/module/rbac/handler.go:159` / `src/module/rbac/handler.go:160` |
| POST | `/api/v2/roles/{role_id}/permissions/assign` | `admin` | Assign permissions to role | `src/module/rbac/handler.go:199` / `src/module/rbac/handler.go:200` |
| POST | `/api/v2/roles/{role_id}/permissions/remove` | `admin` | Remove permissions from role | `src/module/rbac/handler.go:234` / `src/module/rbac/handler.go:235` |
| GET | `/api/v2/roles/{role_id}/users` | `admin` | List users from role | `src/module/rbac/handler.go:267` / `src/module/rbac/handler.go:268` |
| GET | `/api/v2/sdk/datasets` | `sdk` | List SDK dataset samples | `src/module/sdk/handler.go:116` / `src/module/sdk/handler.go:117` |
| GET | `/api/v2/sdk/evaluations` | `sdk` | List SDK evaluation samples | `src/module/sdk/handler.go:36` / `src/module/sdk/handler.go:37` |
| GET | `/api/v2/sdk/evaluations/experiments` | `sdk` | List SDK experiment IDs | `src/module/sdk/handler.go:92` / `src/module/sdk/handler.go:93` |
| GET | `/api/v2/sdk/evaluations/{id}` | `sdk` | Get SDK evaluation sample by ID | `src/module/sdk/handler.go:68` / `src/module/sdk/handler.go:69` |
| GET | `/api/v2/system/metrics` | `admin` | Get current system metrics | `src/module/systemmetric/handler.go:30` / `src/module/systemmetric/handler.go:31` |
| GET | `/api/v2/system/metrics/history` | `admin` | Get historical system metrics | `src/module/systemmetric/handler.go:53` / `src/module/systemmetric/handler.go:54` |
| GET | `/api/v2/systems` | `admin` | List chaos systems | `src/module/chaossystem/handler.go:35` / `src/module/chaossystem/handler.go:36` |
| POST | `/api/v2/systems` | `admin` | Create chaos system | `src/module/chaossystem/handler.go:96` / `src/module/chaossystem/handler.go:97` |
| DELETE | `/api/v2/systems/{id}` | `admin` | Delete chaos system | `src/module/chaossystem/handler.go:158` / `src/module/chaossystem/handler.go:159` |
| GET | `/api/v2/systems/{id}` | `admin` | Get chaos system by ID | `src/module/chaossystem/handler.go:67` / `src/module/chaossystem/handler.go:68` |
| PUT | `/api/v2/systems/{id}` | `admin` | Update chaos system | `src/module/chaossystem/handler.go:126` / `src/module/chaossystem/handler.go:127` |
| GET | `/api/v2/systems/{id}/metadata` | `admin` | List chaos system metadata | `src/module/chaossystem/handler.go:218` / `src/module/chaossystem/handler.go:219` |
| POST | `/api/v2/systems/{id}/metadata` | `admin` | Upsert chaos system metadata | `src/module/chaossystem/handler.go:186` / `src/module/chaossystem/handler.go:187` |
| GET | `/api/v2/teams` | `portal` | List teams | `src/module/team/handler.go:137` / `src/module/team/handler.go:138` |
| POST | `/api/v2/teams` | `portal` | Create a new team | `src/module/team/handler.go:39` / `src/module/team/handler.go:40` |
| DELETE | `/api/v2/teams/{team_id}` | `portal` | Delete team | `src/module/team/handler.go:78` / `src/module/team/handler.go:79` |
| GET | `/api/v2/teams/{team_id}` | `portal` | Get team by ID | `src/module/team/handler.go:106` / `src/module/team/handler.go:107` |
| PATCH | `/api/v2/teams/{team_id}` | `portal` | Update team | `src/module/team/handler.go:178` / `src/module/team/handler.go:179` |
| GET | `/api/v2/teams/{team_id}/members` | `portal` | List team members | `src/module/team/handler.go:391` / `src/module/team/handler.go:392` |
| POST | `/api/v2/teams/{team_id}/members` | `portal` | Add member to team | `src/module/team/handler.go:261` / `src/module/team/handler.go:262` |
| DELETE | `/api/v2/teams/{team_id}/members/{user_id}` | `portal` | Remove member from team | `src/module/team/handler.go:299` / `src/module/team/handler.go:300` |
| PATCH | `/api/v2/teams/{team_id}/members/{user_id}/role` | `portal` | Update team member role | `src/module/team/handler.go:343` / `src/module/team/handler.go:344` |
| GET | `/api/v2/teams/{team_id}/projects` | `portal` | List team projects | `src/module/team/handler.go:220` / `src/module/team/handler.go:221` |
| GET | `/api/v2/users` | `admin` | List users | `src/module/user/handler.go:134` / `src/module/user/handler.go:135` |
| POST | `/api/v2/users` | `admin` | Create a new user | `src/module/user/handler.go:36` / `src/module/user/handler.go:37` |
| DELETE | `/api/v2/users/{id}` | `admin` | Delete user | `src/module/user/handler.go:73` / `src/module/user/handler.go:74` |
| PATCH | `/api/v2/users/{id}` | `admin` | Update user | `src/module/user/handler.go:170` / `src/module/user/handler.go:171` |
| GET | `/api/v2/users/{id}/detail` | `admin` | Get user by ID | `src/module/user/handler.go:101` / `src/module/user/handler.go:102` |
| DELETE | `/api/v2/users/{user_id}/containers/{container_id}` | `admin` | Remove user from container | `src/module/user/handler.go:379` / `src/module/user/handler.go:380` |
| POST | `/api/v2/users/{user_id}/containers/{container_id}/roles/{role_id}` | `admin` | Assign user to container | `src/module/user/handler.go:342` / `src/module/user/handler.go:343` |
| DELETE | `/api/v2/users/{user_id}/datasets/{dataset_id}` | `admin` | Remove user from dataset | `src/module/user/handler.go:450` / `src/module/user/handler.go:451` |
| POST | `/api/v2/users/{user_id}/datasets/{dataset_id}/roles/{role_id}` | `admin` | Assign user to dataset | `src/module/user/handler.go:413` / `src/module/user/handler.go:414` |
| POST | `/api/v2/users/{user_id}/permissions/assign` | `admin` | Assign permission to user | `src/module/user/handler.go:264` / `src/module/user/handler.go:265` |
| POST | `/api/v2/users/{user_id}/permissions/remove` | `admin` | Remove permission from user | `src/module/user/handler.go:303` / `src/module/user/handler.go:304` |
| DELETE | `/api/v2/users/{user_id}/projects/{project_id}` | `admin` | Remove user from project | `src/module/user/handler.go:521` / `src/module/user/handler.go:522` / `src/module/user/handler.go:538` |
| POST | `/api/v2/users/{user_id}/projects/{project_id}/roles/{role_id}` | `admin` | Assign user to project | `src/module/user/handler.go:484` / `src/module/user/handler.go:485` |
| POST | `/api/v2/users/{user_id}/role/{role_id}` | `admin` | Assign global role to user | `src/module/user/handler.go:205` / `src/module/user/handler.go:206` |
| DELETE | `/api/v2/users/{user_id}/roles/{role_id}` | `admin` | Remove role from user | `src/module/user/handler.go:234` / `src/module/user/handler.go:235` |
| GET | `/system/audit` | `admin` | List audit logs | `src/module/system/handler.go:184` / `src/module/system/handler.go:185` |
| GET | `/system/audit/{id}` | `admin` | Get audit log by ID | `src/module/system/handler.go:148` / `src/module/system/handler.go:149` |
| GET | `/system/configs` | `admin` | List configurations | `src/module/system/handler.go:253` / `src/module/system/handler.go:254` |
| GET | `/system/configs/{config_id}` | `admin` | Get configuration | `src/module/system/handler.go:219` / `src/module/system/handler.go:220` |
| PATCH | `/system/configs/{config_id}` | `admin` | Update configuration value | `src/module/system/handler.go:376` / `src/module/system/handler.go:377` |
| GET | `/system/configs/{config_id}/histories` | `admin` | List configuration histories | `src/module/system/handler.go:467` / `src/module/system/handler.go:468` |
| PUT | `/system/configs/{config_id}/metadata` | `admin` | Update configuration metadata | `src/module/system/handler.go:420` / `src/module/system/handler.go:421` |
| POST | `/system/configs/{config_id}/metadata/rollback` | `admin` | Rollback configuration metadata | `src/module/system/handler.go:333` / `src/module/system/handler.go:334` |
| POST | `/system/configs/{config_id}/value/rollback` | `admin` | Rollback configuration value | `src/module/system/handler.go:289` / `src/module/system/handler.go:290` |
| GET | `/system/health` | `admin` | System health check | `src/module/system/handler.go:32` / `src/module/system/handler.go:33` |
| GET | `/system/monitor/info` | `admin` | Get system information | `src/module/system/handler.go:83` / `src/module/system/handler.go:84` |
| POST | `/system/monitor/metrics` | `admin` | Get monitoring metrics | `src/module/system/handler.go:58` / `src/module/system/handler.go:59` |
| GET | `/system/monitor/namespaces/locks` | `admin` | List namespace locks | `src/module/system/handler.go:102` / `src/module/system/handler.go:103` |
| POST | `/system/monitor/tasks/queue` | `admin` | List queued tasks | `src/module/system/handler.go:124` / `src/module/system/handler.go:125` |

## Empty `@x-api-type {}` Operations

| Method | Path | Summary | Raw | Location |
| --- | --- | --- | --- | --- |
| GET | `/api/_docs/models` | API Model Definitions | `{}` | `src/httpapi/docs.go:36` / `src/httpapi/docs.go:37` / `src/httpapi/docs.go:38` |
| GET | `/api/v2/containers` | List containers | `{}` | `src/module/container/handler.go:148` / `src/module/container/handler.go:149` |
| POST | `/api/v2/containers` | Create container | `{}` | `src/module/container/handler.go:41` / `src/module/container/handler.go:42` |
| POST | `/api/v2/containers/build` | Submit container building | `{}` | `src/module/container/handler.go:474` / `src/module/container/handler.go:475` |
| DELETE | `/api/v2/containers/{container_id}` | Delete container | `{}` | `src/module/container/handler.go:84` / `src/module/container/handler.go:85` |
| GET | `/api/v2/containers/{container_id}` | Get container by ID | `{}` | `src/module/container/handler.go:114` / `src/module/container/handler.go:115` |
| PATCH | `/api/v2/containers/{container_id}` | Update container | `{}` | `src/module/container/handler.go:187` / `src/module/container/handler.go:188` |
| PATCH | `/api/v2/containers/{container_id}/labels` | Manage container custom labels | `{}` | `src/module/container/handler.go:226` / `src/module/container/handler.go:227` |
| GET | `/api/v2/containers/{container_id}/versions` | List container versions | `{}` | `src/module/container/handler.go:387` / `src/module/container/handler.go:388` |
| POST | `/api/v2/containers/{container_id}/versions` | Create container version | `{}` | `src/module/container/handler.go:270` / `src/module/container/handler.go:271` |
| DELETE | `/api/v2/containers/{container_id}/versions/{version_id}` | Delete container version | `{}` | `src/module/container/handler.go:319` / `src/module/container/handler.go:320` |
| GET | `/api/v2/containers/{container_id}/versions/{version_id}` | Get container version by ID | `{}` | `src/module/container/handler.go:350` / `src/module/container/handler.go:351` |
| PATCH | `/api/v2/containers/{container_id}/versions/{version_id}` | Update container version | `{}` | `src/module/container/handler.go:432` / `src/module/container/handler.go:433` |
| POST | `/api/v2/containers/{container_id}/versions/{version_id}/helm-chart` | Upload Helm chart package | `{}` | `src/module/container/handler.go:521` / `src/module/container/handler.go:522` |
| POST | `/api/v2/containers/{container_id}/versions/{version_id}/helm-values` | Upload Helm values file | `{}` | `src/module/container/handler.go:577` / `src/module/container/handler.go:578` |
| GET | `/api/v2/datasets` | List datasets | `{}` | `src/module/dataset/handler.go:149` / `src/module/dataset/handler.go:150` |
| POST | `/api/v2/datasets` | Create dataset | `{}` | `src/module/dataset/handler.go:42` / `src/module/dataset/handler.go:43` |
| POST | `/api/v2/datasets/search` | Search datasets | `{}` | `src/module/dataset/handler.go:186` / `src/module/dataset/handler.go:187` |
| DELETE | `/api/v2/datasets/{dataset_id}` | Delete dataset | `{}` | `src/module/dataset/handler.go:85` / `src/module/dataset/handler.go:86` |
| GET | `/api/v2/datasets/{dataset_id}` | Get dataset by ID | `{}` | `src/module/dataset/handler.go:115` / `src/module/dataset/handler.go:116` |
| PATCH | `/api/v2/datasets/{dataset_id}` | Update dataset | `{}` | `src/module/dataset/handler.go:225` / `src/module/dataset/handler.go:226` |
| PATCH | `/api/v2/datasets/{dataset_id}/labels` | Manage dataset custom labels | `{}` | `src/module/dataset/handler.go:269` / `src/module/dataset/handler.go:270` |
| PATCH | `/api/v2/datasets/{dataset_id}/version/{version_id}/injections` | Manage dataset injections | `{}` | `src/module/dataset/handler.go:569` / `src/module/dataset/handler.go:570` |
| GET | `/api/v2/datasets/{dataset_id}/versions` | List dataset versions | `{}` | `src/module/dataset/handler.go:430` / `src/module/dataset/handler.go:431` |
| POST | `/api/v2/datasets/{dataset_id}/versions` | Create dataset version | `{}` | `src/module/dataset/handler.go:313` / `src/module/dataset/handler.go:314` |
| DELETE | `/api/v2/datasets/{dataset_id}/versions/{version_id}` | Delete dataset version | `{}` | `src/module/dataset/handler.go:362` / `src/module/dataset/handler.go:363` |
| GET | `/api/v2/datasets/{dataset_id}/versions/{version_id}` | Get dataset version by ID | `{}` | `src/module/dataset/handler.go:393` / `src/module/dataset/handler.go:394` |
| PATCH | `/api/v2/datasets/{dataset_id}/versions/{version_id}` | Update dataset version | `{}` | `src/module/dataset/handler.go:475` / `src/module/dataset/handler.go:476` |
| GET | `/api/v2/datasets/{dataset_id}/versions/{version_id}/download` | Download dataset version | `{}` | `src/module/dataset/handler.go:521` / `src/module/dataset/handler.go:522` |
| GET | `/api/v2/evaluations` | List evaluations | `{}` | `src/module/evaluation/handler.go:122` / `src/module/evaluation/handler.go:123` |
| POST | `/api/v2/evaluations/datapacks` | List Datapack Evaluation Results | `{}` | `src/module/evaluation/handler.go:37` / `src/module/evaluation/handler.go:38` |
| POST | `/api/v2/evaluations/datasets` | List Dataset Evaluation Results | `{}` | `src/module/evaluation/handler.go:80` / `src/module/evaluation/handler.go:81` |
| DELETE | `/api/v2/evaluations/{id}` | Delete evaluation by ID | `{}` | `src/module/evaluation/handler.go:186` / `src/module/evaluation/handler.go:187` |
| GET | `/api/v2/evaluations/{id}` | Get evaluation by ID | `{}` | `src/module/evaluation/handler.go:157` / `src/module/evaluation/handler.go:158` |
| GET | `/api/v2/executions` | List executions | `{}` | `src/module/execution/handler.go:146` / `src/module/execution/handler.go:147` |
| POST | `/api/v2/executions/batch-delete` | Batch delete executions | `{}` | `src/module/execution/handler.go:271` / `src/module/execution/handler.go:272` |
| GET | `/api/v2/executions/labels` | List execution labels | `{}` | `src/module/execution/handler.go:206` / `src/module/execution/handler.go:207` |
| POST | `/api/v2/executions/{execution_id}/detector_results` | Upload detector results | `{}` | `src/module/execution/handler.go:306` / `src/module/execution/handler.go:307` |
| POST | `/api/v2/executions/{execution_id}/granularity_results` | Upload granularity results | `{}` | `src/module/execution/handler.go:346` / `src/module/execution/handler.go:347` |
| GET | `/api/v2/executions/{id}` | Get execution by ID | `{}` | `src/module/execution/handler.go:180` / `src/module/execution/handler.go:181` |
| PATCH | `/api/v2/executions/{id}/labels` | Manage execution custom labels | `{}` | `src/module/execution/handler.go:233` / `src/module/execution/handler.go:234` |
| GET | `/api/v2/groups/{group_id}/stats` | Get statistics for a group of traces | `{}` | `src/module/group/handler.go:43` / `src/module/group/handler.go:44` |
| GET | `/api/v2/groups/{group_id}/stream` | Stream group trace events in real-time | `{}` | `src/module/group/handler.go:82` / `src/module/group/handler.go:84` |
| GET | `/api/v2/injections` | List injections | `{}` | `src/module/injection/handler.go:242` / `src/module/injection/handler.go:243` |
| GET | `/api/v2/injections/analysis/no-issues` | Query Fault Injection Records Without Issues | `{}` | `src/module/injection/handler.go:333` / `src/module/injection/handler.go:334` |
| GET | `/api/v2/injections/analysis/with-issues` | Query Fault Injection Records With Issues | `{}` | `src/module/injection/handler.go:351` / `src/module/injection/handler.go:352` |
| POST | `/api/v2/injections/batch-delete` | Batch delete injections | `{}` | `src/module/injection/handler.go:522` / `src/module/injection/handler.go:523` |
| POST | `/api/v2/injections/build` | Submit batch datapack buildings | `{}` | `src/module/injection/handler.go:315` / `src/module/injection/handler.go:316` |
| POST | `/api/v2/injections/inject` | Submit batch fault injections | `{}` | `src/module/injection/handler.go:296` / `src/module/injection/handler.go:297` |
| PATCH | `/api/v2/injections/labels/batch` | Batch manage injection labels | `{}` | `src/module/injection/handler.go:488` / `src/module/injection/handler.go:489` |
| GET | `/api/v2/injections/metadata` | Get Injection Metadata | `{}` | `src/module/injection/handler.go:401` / `src/module/injection/handler.go:402` |
| POST | `/api/v2/injections/search` | Search injections | `{}` | `src/module/injection/handler.go:276` / `src/module/injection/handler.go:277` |
| POST | `/api/v2/injections/upload` | Upload a manual datapack | `{}` | `src/module/injection/handler.go:828` / `src/module/injection/handler.go:829` |
| GET | `/api/v2/injections/{id}` | Get injection by ID | `{}` | `src/module/injection/handler.go:372` / `src/module/injection/handler.go:373` |
| POST | `/api/v2/injections/{id}/clone` | Clone injection | `{}` | `src/module/injection/handler.go:556` / `src/module/injection/handler.go:557` |
| GET | `/api/v2/injections/{id}/download` | Download datapack | `{}` | `src/module/injection/handler.go:617` / `src/module/injection/handler.go:618` |
| GET | `/api/v2/injections/{id}/files` | List datapack files | `{}` | `src/module/injection/handler.go:653` / `src/module/injection/handler.go:654` |
| GET | `/api/v2/injections/{id}/files/download` | Download datapack file | `{}` | `src/module/injection/handler.go:691` / `src/module/injection/handler.go:692` |
| GET | `/api/v2/injections/{id}/files/query` | Query datapack file content | `{}` | `src/module/injection/handler.go:745` / `src/module/injection/handler.go:746` |
| PUT | `/api/v2/injections/{id}/groundtruth` | Update datapack ground truth | `{}` | `src/module/injection/handler.go:787` / `src/module/injection/handler.go:788` |
| PATCH | `/api/v2/injections/{id}/labels` | Manage injection custom labels | `{}` | `src/module/injection/handler.go:450` / `src/module/injection/handler.go:451` |
| GET | `/api/v2/injections/{id}/logs` | Get injection logs | `{}` | `src/module/injection/handler.go:589` / `src/module/injection/handler.go:590` |
| GET | `/api/v2/metrics/algorithms` | Get algorithm comparison metrics | `{}` | `src/module/metric/handler.go:99` / `src/module/metric/handler.go:100` |
| GET | `/api/v2/metrics/executions` | Get execution metrics | `{}` | `src/module/metric/handler.go:67` / `src/module/metric/handler.go:68` |
| GET | `/api/v2/metrics/injections` | Get injection metrics | `{}` | `src/module/metric/handler.go:35` / `src/module/metric/handler.go:36` |
| GET | `/api/v2/notifications/stream` | Stream global notifications in real-time | `{}` | `src/module/notification/handler.go:40` / `src/module/notification/handler.go:42` |
| GET | `/api/v2/tasks` | List tasks | `{}` | `src/module/task/handler.go:124` / `src/module/task/handler.go:125` |
| POST | `/api/v2/tasks/batch-delete` | Batch delete tasks | `{}` | `src/module/task/handler.go:48` / `src/module/task/handler.go:49` |
| GET | `/api/v2/tasks/{task_id}` | Get task by ID | `{}` | `src/module/task/handler.go:85` / `src/module/task/handler.go:86` |
| GET | `/api/v2/tasks/{task_id}/logs/ws` | Stream task logs via WebSocket | `{}` | `src/module/task/handler.go:159` / `src/module/task/handler.go:160` |
| GET | `/api/v2/traces` | List traces | `{}` | `src/module/trace/handler.go:81` / `src/module/trace/handler.go:82` |
| GET | `/api/v2/traces/{trace_id}` | Get trace by ID | `{}` | `src/module/trace/handler.go:44` / `src/module/trace/handler.go:45` |
| GET | `/api/v2/traces/{trace_id}/stream` | Stream trace events in real-time | `{}` | `src/module/trace/handler.go:118` / `src/module/trace/handler.go:119` |

## Missing `@x-api-type` Operations

| Method | Path | Summary | Location |
| --- | --- | --- | --- |

