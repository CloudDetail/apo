# Alert Sources And Enrichment Frontend

## Pages

| Page | Route | Description |
|---|---|---|
| `AlertsIntegrationPage` | `/integration/alerts` | 告警接入配置页，使用 `AlertIntegrationProvider` 包裹页面，左侧为来源类型入口，右侧为已配置 source 表格，抽屉用于新增和编辑。 |

## Routes

| File | Route | Component | Notes |
|---|---|---|---|
| `frontend/src/routes.js` | `/integration/alerts` | `AlertsIntegrationPage` | 通过 `AuthRouter(React.lazy(...))` 包装，route name 使用 `core/routes:alertsIntegration`。 |

## Menus

| Source | Key | Route | Notes |
|---|---|---|---|
| `backend/pkg/repository/database/init_menu_item.go` | `alertsIntegration` | `/integration/alerts` | 父级菜单为 `integration`。 |
| `backend/pkg/repository/database/init_feature_mapping.go` | `告警接入` | `alertsIntegration` | feature 与菜单项存在初始化映射。 |
| `backend/pkg/repository/database/init_router.go` | `/integration/alerts` | `HideTimeSelector: true` | 页面默认隐藏全局时间选择器。 |

## Key Components

| Component | File | Implemented Behavior |
|---|---|---|
| `AlertsIntegrationPage` | `frontend/src/core/views/IntegrationCenter/AlertsIntegration/index.tsx` | 组合 `IntegrationDrawer`、`AlertsDatasourceList`、`AlertsIntegrationTable`。 |
| `AlertsDatasourceList` | `AlertsDatasourceList/index.tsx` | 点击来源类型后设置 query `{ sourceId: 'add', sourceType }` 并打开抽屉。 |
| `AlertsIntegrationTable` | `AlertsIntegrationTable/index.tsx` | 调用 `getAlertInputSourceListApi` 渲染 `alertSources`，支持编辑和删除。 |
| `IntegrationDrawer` | `ConfigDrawer/index.tsx` | 根据 URL query 区分新增和编辑；编辑时调用 `getAlertInputBaseInfoApi`。 |
| `BaseInfoContent` | `ConfigDrawer/BaseInfoContent/index.tsx` | 表单保存 `sourceName`、`sourceType`、`clusters`；新建时前端生成 `uuid`；展示 push URL。 |
| `BaseInfoDescriptions` | `ConfigDrawer/BaseInfoContent/BaseInfoDescriptions.tsx` | 查看态展示 `sourceName` 和 push URL。 |
| `ClusterSelector` | `ConfigDrawer/BaseInfoContent/ClusterSelector.tsx` | 已实现 cluster 多选和 `/api/alertinput/cluster/list` 调用，但当前在 `BaseInfoContent` 中被注释。 |
| `TagContent` | `ConfigDrawer/TagContent/index.tsx` | source 已存在时展示 enrich rule 卡片。 |
| `TagContactRule` | `ConfigDrawer/TagContent/TagContactRule/index.tsx` | 调用 `getAlertEnrichApi` 读取规则，调用 `saveAlertEnrichApi` 保存规则。 |
| `TagRuleFormListCom` | `TagContactRule/TagRuleFormListCom.tsx` | 单条 enrich rule 表单，支持 rule 类型、条件、来源字段、目标 tag 或 schema mapping。 |
| `ConditionsFormList` | `TagContactRule/ConditionsFormList.tsx` | 条件列表当前 UI 支持 `match` 和 `notMatch`。 |
| `SchemaSource` | `TagContactRule/SchemaSource.tsx` | 从 context `schemas` 构造 schema/column cascader。 |
| `SchemaFormList` | `TagContactRule/SchemaFormList.tsx` | 选择 schema target fields，并对每个 field 绑定 target tag。 |
| `TargetTagSelector` | `TagContactRule/TargetTagSelector.tsx` | 使用 context `targetTags`，并允许新增 id 为 `0` 的自定义 tag。 |
| `SourceTypeInfo` | `ConfigDrawer/SourceTypeInfo/index.tsx` | 根据 `sourceType` 展示 `ZabbixInfo`、`PrometheusInfo` 或 `JsonInfo`。 |

## State Management

| State | File | Notes |
|---|---|---|
| `configDrawerVisible` | `frontend/src/core/contexts/AlertIntegrationContext.tsx` | 控制配置抽屉是否打开，关闭后表格会刷新 source 列表。 |
| `schemas` | `AlertIntegrationContext.tsx` | 抽屉打开时调用 `getAllSchemaApi`，保存 schema 到 context。 |
| `targetTags` | `AlertIntegrationContext.tsx` | 抽屉打开时调用 `getTargetTagsListApi`，过滤 `id !== 0` 后保存。 |
| URL Query | `ConfigDrawer/index.tsx` | `sourceId`、`sourceType` 决定新增、编辑或关闭抽屉。 |
| Local Form State | `BaseInfoContent`, `TagContactRule` | 通过 Ant Design `Form` 管理 source 基础信息和 enrich rule 配置。 |

## API Usage

| UI Action | Wrapper | API Path | Notes |
|---|---|---|---|
| 加载 source 表格 | `getAlertInputSourceListApi` | `/api/alertinput/source/list` | 返回 `alertSources`。 |
| 打开编辑抽屉 | `getAlertInputBaseInfoApi` | `/api/alertinput/source/get` | 请求体包含 `sourceId`。 |
| 新增 source | `creatAlertInputSourceApi` | `/api/alertinput/source/create` | 前端会传 `sourceId`、`sourceName`、`sourceType`、`clusters`。 |
| 更新 source | `updateAlertsIntegrationApi` | `/api/alertinput/source/update` | 只观察到基础信息表单调用。 |
| 删除 source | `deleteAlertIntegrationApi` | `/api/alertinput/source/delete` | 请求体为 `{ sourceId }`。 |
| 抽屉打开加载 target tags | `getTargetTagsListApi` | `/api/alertinput/enrich/tags/list` | context 过滤掉 id 为 `0` 的 tag。 |
| 抽屉打开加载 schema | `getAllSchemaApi` | `/api/alertinput/schema/listwithcolumns` | 返回 `schemas` map。 |
| 读取 enrich rule | `getAlertEnrichApi` | `/api/alertinput/source/enrich/get` | 当前 wrapper 使用 `post`。 |
| 保存 enrich rule | `saveAlertEnrichApi` | `/api/alertinput/source/enrich/update` | 保存 `sourceId` 和 `enrichRuleConfigs`。 |
| cluster 列表 | `getClusterListApi` | `/api/alertinput/cluster/list` | 只在已注释的 `ClusterSelector` 中调用。 |
| schema 列表 | `getSchemaListApi` | `/api/alertinput/schema/list` | wrapper 存在，当前主页面未观察到直接调用。 |
| schema columns | `getSchemaColumnsApi` | `/api/alertinput/schema/column/get` | wrapper 存在，当前主页面未观察到直接调用。 |

## User Interaction Flow

1. 用户进入 `/integration/alerts`，页面渲染左侧 source type 入口和右侧 source 表格。
2. 页面表格通过 `/api/alertinput/source/list` 读取已配置 `alertSources`。
3. 用户点击 `json`、`zabbix` 或 `prometheus`，前端设置 `sourceId=add` 和 `sourceType` query，并打开抽屉。
4. 新增抽屉中用户填写 `sourceName`，前端生成 `uuid` 并展示 `/api/alertinput/event/source?sourceId=<uuid>`。
5. 用户保存基础信息后调用 `/api/alertinput/source/create`，成功后把返回的 `sourceId` 写入 URL query。
6. 已存在 source 的抽屉展示 `setting` 和 `documentation` tabs；`setting` 中包含基础信息和 enrich rule。
7. 用户进入 enrich rule 编辑，前端读取 `/api/alertinput/source/enrich/get`，再通过表单保存到 `/api/alertinput/source/enrich/update`。
8. 用户删除 source 时，表格调用 `/api/alertinput/source/delete`，完成后刷新列表。

## Permissions

| Area | Evidence | Notes |
|---|---|---|
| Route Guard | `AuthRouter` | `/integration/alerts` 被包装，但 `checkRoutePermission` 的 effect 当前被注释，`isAllowed` 初始为 `true`。 |
| Menu Visibility | `init_menu_item.go`, `init_feature_mapping.go` | 后端初始化 `alertsIntegration` 菜单项和 `告警接入` feature 映射。 |
| Backend Middleware | `router_api.go` | `/api/alertinput` 使用 `LicMiddleware()`，未观察到 `AuthMiddleware()`。 |

## Notes

- 当前前端 `AlertKey` 类型只允许 `json`、`zabbix`、`prometheus`，不能把 `datadog` 和 `pagerduty` 写成当前 UI 可新增项。
- 当前前端未观察到直接编辑 `AlertSource.Params`、`EnabledPull` 或 `LastPullMillTS` 的 UI。
- `SourceTypeInfo` 只是接入说明展示，不能推断为自动配置能力。
- `ClusterSelector` 已实现但注释，cluster 关联目前不应写成可见主流程。
- `TagRuleFormListCom` 的 UI 选项值包含 `staticEnrich`，保存逻辑中非 `tagMapping` 分支会被当作 schema mapping 处理；该命名差异需要后续确认。
