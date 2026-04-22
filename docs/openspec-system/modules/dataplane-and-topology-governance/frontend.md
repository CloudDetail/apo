# Dataplane And Topology Governance Frontend

## Pages

| Page | Route | Description |
|---|---|---|
| `AgentDataplanePage` | `/dataplane` | 本模块主页面，通过 `Tabs` 切换 `dataplane` 和 `serviceMatcher`。 |
| `DataplanePage` | `/dataplane?tab=dataplane` | 展示可接入的数据平面类型和已配置数据源。 |
| `ServiceMatcherPage` | `/dataplane?tab=serviceMatcher` | 展示监控服务与已匹配实例，并支持为单个 service 配规则。 |
| `DataplaneStep` | probe 安装流程内部页面 | 在安装向导中复用数据平面类型选择和配置表单。 |

## Routes

| File | Route | Component | Notes |
|---|---|---|---|
| `frontend/src/pro/routes.js` | `/dataplane` | `AgentDataplanePage` | 当前前端正式路由。 |
| `frontend/src/routes.js` | `/demo/service-config` | `ConfigPageAuth` | 名称容易混淆，但当前为 mock/comingSoon 页面。 |
| Shared Entry | `frontend/src/pro/components/AGUIChat/components/DataStatus.tsx` | `window.open('/#/dataplane?...')` | Agent 页会跳转到本模块。 |

## Menus

| Source | Key | Route | Notes |
|---|---|---|---|
| `backend/pkg/repository/database/init_router.go` | `/dataplane` | 路由元数据 | 当前前端主路由与数据库初始化路由已对齐。 |
| `backend/pkg/services/permission/service_check_router.go` | `/agent-dataplane` | 历史权限白名单 | `allowedRouters` 中仍硬编码旧路由名。 |
| Agent UI | `DataStatus.tsx` | `/dataplane` | 通过状态提示把用户引导回配置页面。 |

## Key Components

| Component | File | Implemented Behavior |
|---|---|---|
| `AgentDataplanePage` | `frontend/src/pro/views/agentDataplane/index.tsx` | 读取 URL query `tab`，切换 `dataplane` 与 `serviceMatcher`，并在 `serviceMatcher` 时显示时间选择器。 |
| `DataplanePage` | `frontend/src/pro/views/agentDataplane/DataplanePage.tsx` | 把左侧 `DataplaneList`、右侧 `ConfiguredSources` 和抽屉 `ConfigDrawer` 组合成主工作台。 |
| `DataplaneList` | `frontend/src/pro/components/Dataplane/DataplaneList.tsx` | 调 `/api/dataplane/type/list` 获取类型，按 capability 过滤，并读取 `/api/dataplane/capabilities` 展示已接入能力状态。 |
| `ConfiguredSources` | `frontend/src/pro/views/agentDataplane/components/ConfiguredSources.tsx` | 调 `/api/dataplane/list` 展示已连接源，并支持删除和编辑。 |
| `ConfigDrawer` | `frontend/src/pro/views/agentDataplane/ConfigDrawer/index.tsx` | 为新增和编辑统一承载 `DataplaneConfigForm`。 |
| `DataplaneConfigForm` | `frontend/src/pro/components/Dataplane/DataplaneConfigForm/index.tsx` | 动态根据 `paramSpec` 渲染表单字段，支持“测试连接”和“连接”两步。 |
| `SetupGuide` | `frontend/src/pro/components/Dataplane/DataplaneGuide` | 根据类型展示接入图文指引。 |
| `ServiceMatcherPage` | `frontend/src/pro/views/agentDataplane/ServiceMatcherPage.tsx` | 调 `/api/dataplane/servicesWithInstances` 分页展示服务和匹配实例。 |
| `ServiceRuleList` | `frontend/src/pro/views/agentDataplane/ServiceMatcher/index.tsx` | 针对单个 service 展示规则列表，支持新增、编辑、删除。 |
| `ConfigModal` | `frontend/src/pro/views/agentDataplane/ServiceMatcher/ConfigModal/index.tsx` | 表单保存前可用 `/api/dataplane/servicename/checkRule` 试跑匹配结果。 |
| `ConditionsFormList` | `frontend/src/pro/views/agentDataplane/ServiceMatcher/ConfigModal/ConditionsFormList.tsx` | 为每条规则条件提供 key/value 与匹配方式编辑能力。 |
| `DataStatus` | `frontend/src/pro/components/AGUIChat/components/DataStatus.tsx` | 显示数据源能力与服务监控状态，并把配置操作导向 `/dataplane`。 |
| `DataCoverageOverview` | `frontend/src/pro/components/Dataplane/DataCoverageOverview.tsx` | 以抽屉形式展示 capability 覆盖细节。 |

## State Management

| State | File | Notes |
|---|---|---|
| Active Tab | `frontend/src/pro/views/agentDataplane/index.tsx` | `tab` query 决定当前页签，默认 `dataplane`。 |
| Time Range | `AgentDataplanePage` + `timeRangeReducer` | `serviceMatcher` 使用 `last7Days` 默认时间范围，并支持刷新。 |
| Drawer State | `DataplanePage.tsx` | `drawerOpen` 和 `dataplaneConfigInfo` 控制新增/编辑抽屉。 |
| Capability Filter | `DataplaneList.tsx` | 通过 URL query `filter` 存储 capability 过滤条件。 |
| Settings Drawer | `DataplaneList.tsx` | 通过 query `showSettings=true` 打开 `DataCoverageOverview`。 |
| Rule Form State | `ConfigModal/index.tsx` | 使用 Ant Design `Form` 管理 service、ruleId 与 conditions。 |
| Agent Badge State | `DataStatus.tsx` | 本地保存 `dataplane`、`serviceCount`、`capabilities`。 |

## API Usage

| UI Action | Wrapper | API Path | Notes |
|---|---|---|---|
| 读取数据平面类型 | `getDataplaneTypeListApi` | `/api/dataplane/type/list` | `DataplaneList`、`DataplaneRadioList` 使用。 |
| 读取已配置数据源 | `getDataplaneListApi` | `/api/dataplane/list` | `ConfiguredSources` 使用。 |
| 新建数据平面 | `createDataplaneApi` | `/api/dataplane/create` | `DataplaneConfigForm` 保存新记录。 |
| 更新数据平面 | `updateDataplaneApi` | `/api/dataplane/update` | `DataplaneConfigForm` 编辑时使用。 |
| 删除数据平面 | `deleteDataplaneApi` | `/api/dataplane/delete` | 已配置列表删除动作。 |
| 测试数据源 | `checkDataplaneDatasourceApi` | `/api/dataplane/checkDataSource` | “Test” 按钮调用。 |
| 读取服务分页 | `getServiceWithIstanceApi` | `/api/dataplane/servicesWithInstances` | `ServiceMatcherPage` 主表。 |
| 读取服务列表 | `getServicesApi` | `/api/dataplane/services` | 规则配置时调用。 |
| 查询标签 key | `getAppinfoTagsApi` | `/api/dataplane/appinfo/tags` | 条件编辑辅助。 |
| 查询标签 value | `getAppinfoTagValuesApi` | `/api/dataplane/appinfo/tags/values` | 条件编辑辅助。 |
| 规则试跑 | `checkServiceNameRuleApi` | `/api/dataplane/servicename/checkRule` | 返回当前匹配实例列表。 |
| 保存规则 | `saveServiceRuleApi` | `/api/dataplane/servicename/upsertRule` | 新增和编辑共用。 |
| 读取规则 | `listServiceNameRuleApi` | `/api/dataplane/servicename/listRule` | 展开某个 service 行时获取。 |
| 删除规则 | `deleteServiceNameRuleApi` | `/api/dataplane/servicename/deleteRule` | 删除规则。 |
| Agent 前置检查 | `checkDataplaneDataApi` | `/api/dataplane/data/check` | Agent 页面判断 `agentReady` 与 `dataPlaneExist`。 |
| Agent 数据状态 | `getDataplaneDatasourcesApi` | `/api/dataplane/datasources` | Badge 和 Tooltip 使用。 |
| 能力覆盖状态 | `getCapabilitiesApi` | `/api/dataplane/capabilities` | 多处使用；后端当前按已配置 dataplane capability 聚合摘要。 |
| 拓扑查看预留 | `getTopologyApi` | `/api/dataplane/topology` | wrapper 已存在，当前主 UI 未见消费。 |
| 自定义拓扑预留 | `createCustomTopologyApi` 等 | `/api/dataplane/customtopology/*` | wrapper 已存在，当前主 UI 未见消费。 |

## User Interaction Flow

1. 用户进入 `/dataplane`，默认看到 `dataplane` tab。
2. `DataplaneList` 读取 `/api/dataplane/type/list`，渲染可接入类型卡片。
3. 用户点击某个类型卡片，页面打开 `ConfigDrawer`。
4. `DataplaneConfigForm` 按 `paramSpec` 生成字段，用户可先点 “Test” 调 `/api/dataplane/checkDataSource`。
5. 测试通过或忽略后，用户点击 “Connect”，前端调用 `/api/dataplane/create` 或 `/api/dataplane/update`。
6. 保存后右侧 `ConfiguredSources` 刷新 `/api/dataplane/list`，展示当前已配置源。
7. 用户切到 `serviceMatcher` tab，页面按时间范围调用 `/api/dataplane/servicesWithInstances`。
8. 用户展开某个 service 行，前端读取 `/api/dataplane/servicename/listRule`，并可继续新增/编辑规则。
9. 在规则弹窗中，用户点击 “Test” 调 `/api/dataplane/servicename/checkRule` 预览当前会匹配到的实例。
10. 用户点击保存后，前端调 `/api/dataplane/servicename/upsertRule`，刷新规则列表和 service 匹配结果。

## Permissions

| Area | Evidence | Notes |
|---|---|---|
| Route Guard | `AuthRouter` | 主路由会被包装，但实际权限检查 `useEffect` 已注释。 |
| Hardcoded Allowlist | `backend/pkg/services/permission/service_check_router.go` | `/agent-dataplane` 被硬编码放行，而不是 `/dataplane`。 |
| Backend Middleware | `router_api.go` | `/api/dataplane` 整组受 `LicMiddleware()` 控制。 |

## Notes

- `DataplaneList`、`DataCoverageOverview`、`DataStatus` 都依赖 `/api/dataplane/capabilities`；该接口已实现，但语义更接近“配置覆盖摘要”而不是实时健康探测。
- `DataplaneConfigForm` 对 `signoz` 会额外显示 `signozMissingFdMetric` warning，这是当前唯一明显的类型特例。
- 前端 API 文件中存在多组重复 wrapper，如 `upsertServiceNameRuleApi`/`saveServiceRuleApi`、`getAppinfoTagsApi`/`getServiceNameRuleApi`，说明这一层有历史兼容痕迹。
- `customtopology` 相关 wrapper 已存在，但主页面没有稳定入口，不能写成已交付前台能力。
