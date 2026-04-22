# Data Scope Governance Dependencies

## Upstream Dependencies

| Dependency | Evidence | Why It Matters |
|---|---|---|
| `UserContext` | `DataGroupContext.tsx` | 当前登录用户决定可见数据组树。 |
| `request.js` | `dataGroup.ts` | 统一认证和错误处理。 |
| Prometheus Repo | `service_data_scope.go`, `data_group_store_v2.go` | 实时 scope 扫描与过滤。 |
| ClickHouse Repo | `data_group_store_v2.go` | 虽然部分扫描逻辑已注释，但仍是数据源之一。 |
| DB Repo | `dao_data_group_new.go`, `dao_data_scope.go`, `dao_auth_data_group.go` | 组、scope、授权关系持久化。 |

## Downstream Dependencies

| Consumer | Evidence | Dependency Relationship |
|---|---|---|
| `identity-access-and-organization` | `UserManage/index.tsx`, `Team/index.tsx` | 用户和团队页面复用本模块授权弹窗。 |
| 观测过滤能力 | `GetFilterByGroupID` | 依赖本模块生成可用的 endpoint / instance 范围树。 |

## Shared Components

| Shared Component | File | Role In This Module |
|---|---|---|
| `BasicCard` | `DataGroup/index.tsx` | 主页面容器。 |
| `DataGroupAuthorizeModal` | `core/components/PermissionAuthorize/DataGroupAuthorizeModal.tsx` | 跨模块复用的授权 UI。 |
| `notify` | 多个组件 | 保存、删除、刷新提示。 |

## Shared Utilities

| Utility | File | Role |
|---|---|---|
| `DataGroupStorage` | `data_group_store_v2.go` | 共享缓存与树克隆。 |
| `CloneWithPermission` | `data_group.go` | 基于授权生成前端可见树。 |
| `CloneScopeWithPermission` | `data_scope.go` / store | 构造 scope tree。 |
| `ConvertScopeNodeToPQLFilter` | `services/common` | 把 scope tree 转换为 PromQL 过滤条件。 |

## External Services

| Service | Evidence | Current Role |
|---|---|---|
| Prometheus | `QueryMetricsWithPQLFilter` | 提供 service、namespace、instance、endpoint 级 scope 观测。 |
| ClickHouse | `DataGroupStorage.Refresh` | scope 扫描预留的数据源之一。 |

## Notes

- 本模块不是纯权限表 CRUD，而是把“观测侧现存 scope”与“治理侧授权关系”持续对齐。
- 因为授权既能从 group 视角维护，也能从 subject 视角维护，后续写更细文档时必须始终注明操作方向。
