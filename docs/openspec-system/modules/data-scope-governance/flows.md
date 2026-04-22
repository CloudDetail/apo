# Data Scope Governance Flows

## Main Flow 1

### DataGroup Tree And Edit Flow

1. 用户进入 `/system/data-group`。
2. `DataGroupContext` 调 `/api/v2/data/group` 获取当前用户可见树。
3. 页面左侧渲染 `DataGroupTree`，右侧渲染 `DataGroupInfoPanel`。
4. 用户点击新增或编辑时，`InfoModal` 打开。
5. `DatasourceSelector` 调 `/api/v2/data/group/datasource/list?groupId=...` 读取父组允许选择的 scope tree。
6. 用户勾选 scope 并保存。
7. 新增时前端调 `/api/v2/data/group/add`；编辑时调 `/api/v2/data/group/update`。
8. 后端写入 `data_group` 和 `data_group2scope`，并刷新 `DataGroupStorage.DataGroupTreeNode`。
9. 前端刷新树和详情视图。

## Main Flow 2

### Group To Subject Authorization Flow

1. 用户在数据组页面打开 `PermissionModal`。
2. 前端先调 `/api/data/subs` 读取该组当前已授权 user/team。
3. 前端同时调 `/api/user/list` 与 `/api/team` 获取可选主体。
4. 用户选择 user/team 后点击保存。
5. 前端调用 `/api/data/subs/operation`。
6. 后端 `GroupSubsOperation` 比对已有授权与新授权，分别生成 `toAdd` 和 `toDelete`。
7. 后端事务执行 `AssignDataGroup` 与 `RevokeDataGroupBySub`。

## Main Flow 3

### Subject To Group Authorization Flow

1. 用户或团队管理页面打开 `DataGroupAuthorizeModal`。
2. 前端调用 `/api/data/sub/group?subjectId=...&subjectType=user|team`。
3. 如果主体类型是 user，后端会把 user 自己和 team 继承的组一起返回，并标记 `source`。
4. 前端对 user 仅允许编辑 `source === user` 的那部分组，把 team 继承部分显示在只读区域。
5. 用户保存后，前端调用 `/api/data/group/operation`。
6. 后端 `DataGroupOperation` 计算 `toModify` 和 `toDelete`，更新 `auth_data_group`。

## Main Flow 4

### Expired Scope Refresh Flow

1. 用户在 `DatasourceSelector` 中点击 refresh。
2. 前端调用 `/api/v2/data/group/datasource/refresh?groupId=<id>&clean=true`。
3. 后端实时扫描 Prometheus/ClickHouse 可见 scope，与当前 group 范围做差集。
4. 后端把无权限删除的 scope 放入 `protected`，可删的放入 `toBeDeleted`。
5. 当 `clean=true` 时，后端删除 group2scope 关系和对应 scope 记录，并刷新 `DataGroupStorage`。
6. 前端把 `protected` 通过 Tooltip 展示给用户，并重新拉取 scope tree。

## Edge Cases

| Case | Implemented Behavior |
|---|---|
| 用户没有任何授权组 | 仍会得到 root 节点，但子树权限可能都是 `known`。 |
| 新增组未传父组 | 前端直接报错 `addGroupError`，不发请求。 |
| 删除仍有子组的组 | 前端和后端都会阻止。 |
| 清理 scope 时存在无权删除项 | 返回到 `protected` 列表，不会被删。 |
| user 继承 team 数据组 | 在 `DataGroupAuthorizeModal` 中单独显示为只读。 |
| root group 不存在 | `InitRootGroup` 启动时会自动补齐。 |

## Failure Paths

| Path | Failure Behavior |
|---|---|
| 无权限读取 group scope | `GetDGScopeList` 返回 permission error 和空结构。 |
| 新增/更新时选到父组外 scope | 返回权限不足错误。 |
| 删除不存在的数据组 | 返回 `DataGroupNotExistError`。 |
| scope 刷新失败 | 前端不会更新树，并收到 error。 |
| 旧版 wrapper 调用旧增删改接口 | 当前主页面没有走这条链，存在接口漂移风险。 |

## Permission Constraints

| Layer | Constraint |
|---|---|
| Frontend Tree | `permissionType === 'known'` 的节点禁用。 |
| Backend Group Read | `CheckGroupPermission` 限制 group scope 查询。 |
| Cleanup Protection | 无权操作的 scope 即使过期也不会被当前用户清掉。 |
| Auth Route | `/api/data` 和 `/api/v2/data` 整组都挂 `AuthMiddleware()` 与 `LicMiddleware()`。 |

## To Be Confirmed

- 旧版 `/api/data/group/create|update|delete` 在当前版本是否还存在其他调用方，需要确认。
- `DataGroupContext` 中 `know`/`known` 的拼写差异是否只是前端小 bug，需要确认。
- scope 来源当前主要依赖 Prometheus 查询，ClickHouse 扫描路径被注释，是否是策略调整需要确认。
