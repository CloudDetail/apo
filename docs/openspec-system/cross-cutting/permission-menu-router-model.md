# Permission, Menu, And Router Model

## Scope

本文档说明当前 APO 的菜单、前端路由、后端路由元数据、功能权限和初始化数据如何串联。它的直接用途是：维护 `/#/knowledge-bases` 这类页面的菜单权限时，按本文档核对插入点，不只照截图改单个文件。

当前状态：

- 前端 `/#/knowledge-bases` 对应 React route path 是 `/knowledge-bases`，已经注册在 `frontend/src/pro/routes.js`。
- 后端数据库初始化已为 `/knowledge-bases` 补齐 `router`、`menu_item`、`feature_mapping`、admin 默认 `auth_permission` 和 `i18n_translation`。
- 菜单 key 为 `knowledgeBases`，作为顶层菜单显示，不挂到其他父级下。
- 本次按用户确认只落当前商业仓；这些文件仍属于共享后端候选 / 共享前端路径，后续跨仓同步需要单独处理。

## Governance

后续真正修改菜单、路由、权限初始化时会触及共享路径，需要先走仓库治理判断：

| Area | Files | Governance |
|---|---|---|
| 后端初始化 | `backend/pkg/repository/database/init_feature.go`, `init_router.go`, `init_menu_item.go`, `init_feature_mapping.go`, `init_permission.go`, `init_insert_page.go`, `init_router_page.go`, `backend/sqlscripts/i18n.yml` | 这些同路径文件在 OSS 仓也存在，属于共享后端候选。改代码前必须先确认是否回 OSS，以及使用哪个 OSS 分支。 |
| 前端路由 | `frontend/src/pro/routes.js` | Pro 商业前端路径；当前 `/knowledge-bases` 已在此注册，后续可直接在商业分支维护。 |
| 前端菜单图标 | `frontend/src/_nav.js` | 共享前端路径。如果要新增顶层菜单图标 key，应先在 OSS `feat/agent-ce` 落地，再同步回商业分支，除非用户明确要求公开到 OSS `main`。 |

## Runtime Flow

```text
DB startup
  -> seed feature / router / menu / page / mapping / permission / i18n
  -> user login
  -> UserContext calls GET /api/permission/config
  -> backend resolves user feature IDs from role + team + user grants
  -> feature_mapping(feature -> menu) selects visible menu items
  -> menu_item.router_id attaches router metadata
  -> i18n_translation fills menu labels
  -> AppSidebar renders backend menu tree
  -> AppContent renders local React routes from frontend/src/routes.js
```

一个页面要成为“用户可见菜单”，通常需要同时满足：

1. 前端本地 route 已注册。
2. `router` 表存在对应 path。
3. `menu_item` 表存在菜单 key，并通过 `router_id` 指到该 router。
4. `feature` 表存在功能名。
5. `feature_mapping` 里存在 feature -> menu 映射。
6. 当前用户、角色或团队在 `auth_permission` 里拥有该 feature。
7. `i18n_translation` 有菜单 label，避免返回空 label。
8. 顶层菜单如果需要图标，`frontend/src/_nav.js` 的 `navIcon` 需要同 key 图标。

## Initialization Order

`backend/pkg/repository/database/dao.go` 的 `New()` 里当前顺序是：

```text
initRole
initFeature
initRouterData
initMenuItems
initInsertPages
initRouterPage
initFeatureMenuItems
initFeatureRouter
initPermissions
initI18nTranslation
createAdmin
createAnonymousUser
```

顺序很重要：

- `initMenuItems` 会按 `RouterKey` 查询 `router.router_id`，因此 router 必须先初始化。
- `initRouterPage` 会按 router 和 insert page 建关系，因此 router/page 必须先初始化。
- `initFeatureMenuItems` 和 `initFeatureRouter` 依赖 feature、menu、router 都已存在。
- `initPermissions` 依赖 feature 已存在。
- `initI18nTranslation` 依赖 feature/menu 已存在，才能把配置里的 target key 解析成实体 ID。

## Core Tables

| Table | Model | Purpose |
|---|---|---|
| `feature` | `profile.Feature` | 权限功能点。角色、用户、团队最终都授予 feature。 |
| `router` | `database.Router` | 前端 path 元数据，包含 `router_to` 和 `hide_time_selector`。Hash 路由中的 `/#/knowledge-bases` 在这里写成 `/knowledge-bases`。 |
| `menu_item` | `database.MenuItem` | 侧边栏菜单项，含 `key`、`router_id`、`parent_id`、`sort_order`。 |
| `insert_page` | `database.InsertPage` | Grafana / Jaeger 等嵌入页 URL。普通 React 页面一般不需要。 |
| `router_insert_page` | `database.RouterInsertPage` | 路由到嵌入页的按语言映射。普通 React 页面一般不需要。 |
| `feature_mapping` | `database.FeatureMapping` | feature 到 menu/router/api 的映射，`mapped_type` 为 `menu`、`router` 或 `api`。 |
| `auth_permission` | `database.AuthPermission` | subject(role/user/team) 到 feature 的授权。 |
| `i18n_translation` | `database.I18nTranslation` | feature/menu 的 label、abbreviation 翻译。 |

## Source Files

### `init_feature.go`

`validFeatures` 是默认产品模式的功能白名单；`sreFeatures` 是 `PRODUCT_MODE=syncause` 时的功能白名单。`initFeature()` 会：

- 添加白名单里缺失的非自定义 feature。
- 删除已经不在当前产品白名单里的非自定义 feature。
- 用 `parentChildMapping` 写入 feature 的父子关系。

注意：

- `isValidFeature()` 只检查 `validFeatures`。如果后续新功能需要参与 menu/router/api mapping，必须先进入 `validFeatures`。
- 如果该功能也要在 `PRODUCT_MODE=syncause` 下出现，还要进入 `sreFeatures`。
- 如果是子功能，需要在 `parentChildMapping` 里把它挂到父 feature 下。

### `init_router.go`

`validRouters` 和 `sreRouters` 分别定义不同产品模式下的 router 白名单。`initRouterData()` 会添加缺失 router、更新 `hide_time_selector`，并删除不在当前产品 router 白名单中的非自定义 router。

约定：

- 浏览器地址 `/#/knowledge-bases` 在 `RouterTo` 中写 `/knowledge-bases`。
- `HideTimeSelector` 控制系统时间选择器是否隐藏。知识库管理这类管理页面通常应为 `true`，但最终以产品要求为准。
- 如果新路由也要在 `PRODUCT_MODE=syncause` 下可见，需要同时加入 `sreRouters`。

### `init_insert_page.go` And `init_router_page.go`

这两类只服务嵌入式页面：

- `validPages` 初始化 `insert_page`。
- `routerPageMap` 初始化 `router_insert_page`，把 router 关联到按语言区分的 Grafana / Jaeger 页面。

普通 React route，例如 `/knowledge-bases`，不需要添加 `InsertPage` 或 `routerPageMap`。只有路由目标是大盘或嵌入页时才需要。

### `init_menu_item.go`

`validMenuItemMappings` 和 `sreMenuItemMappings` 定义菜单 key、排序和可选 router。`initMenuItems()` 会：

- 根据 `RouterKey` 查找 router ID 并写入 `menu_item.router_id`。
- 以 `key` 为唯一键 upsert 菜单。
- 用 `relations` 写入父子菜单关系。

注意：

- 顶层菜单不写 `relations`；子菜单必须在 `relations` 中映射 `childKey -> parentKey`。
- `AppSidebar` 顶层菜单图标来自 `navIcon[menuItem.key]`，因此新增顶层 key 时还要考虑 `frontend/src/_nav.js`。
- 子菜单目前不显示单独 icon，通常只需要 label 和 router。

### `init_feature_mapping.go`

这里有三类映射：

| Function | Mapping Type | Purpose |
|---|---|---|
| `initFeatureMenuItems()` | `menu` | feature -> menu item。决定拥有某 feature 的用户能在侧边栏看到哪些菜单。 |
| `initFeatureRouter()` | `router` | feature -> router。用于授权详情页、报告页等没有菜单项但需要受同一 feature 控制的 route。 |
| `initFeatureAPI()` | `api` | feature -> API。当前函数读取 `backend/sqlscripts/feature_api.yml`，但 `dao.New()` 里未调用 `initApi()` / `initFeatureAPI()`，实际启用前需要重新核对调用链。 |

`/knowledge-bases` 如果作为菜单入口，至少需要 `initFeatureMenuItems()` 中的 feature -> menu 映射。只有新增详情页、编辑页等不在菜单里的子 route 时，才需要在 `initFeatureRouter()` 里补 feature -> router 映射。

### `init_permission.go`

`roleFeatures` 定义角色默认拥有的 feature。`initPermissions()` 只在某个 role 当前完全没有 feature 权限时初始化默认权限：

```text
if role already has any feature permission:
  skip default initialization for that role
```

这意味着：

- 新库或空权限角色会吃到 `roleFeatures` 默认值。
- 已运行环境中，已有 `admin`、`viewer`、`anonymous` 如果已经有任意 feature 权限，新增 feature 不会自动授予。
- 后续给已有环境开放新菜单时，除了改种子数据，还要考虑迁移脚本、后台“菜单管理”配置，或手动授权路径。

### `backend/sqlscripts/i18n.yml`

`initI18nTranslation()` 会读取该文件，按 `entity_type` 决定 target：

- `entity_type: "feature"` 时，`key` 应是 feature name，例如 `知识库管理`。
- `entity_type: "menu"` 时，`key` 应是 menu key，例如 `knowledgeBases`。

外层 YAML key 只作为配置项名；真正查库用的是内层 `key` 字段。

### Frontend Files

| File | Current Role |
|---|---|
| `frontend/src/pro/routes.js` | 注册 Pro React route。`/knowledge-bases` 已在此注册。 |
| `frontend/src/routes.js` | 拼接 base / oss / pro route 表，`AppContent` 只渲染这里出现的 route。 |
| `frontend/src/core/contexts/UserContext.js` | 调 `/api/permission/config`，把 `menuItem` 写入 Redux。 |
| `frontend/src/core/components/AppSidebar.js` | 渲染后端返回的菜单树，顶层 icon 读 `navIcon[menuItem.key]`。 |
| `frontend/src/core/components/AuthRouter.tsx` | 当前客户端 route 权限校验 `useEffect` 被注释，不能把前端守卫视为有效授权边界。 |

## API Flow

`GET /api/permission/config` 的主要流程在 `backend/pkg/services/permission/service_user_configs.go`：

1. `getUserFeatureIDs()` 汇总用户角色、团队、用户自身的 feature 权限。
2. `GetFeatureMappingByFeature(featureIDs, "menu")` 找到授权 feature 对应的 menu item IDs。
3. `GetMenuItems()` 读取全部 menu item。
4. `FillItemRouter()` 给 menu item 填充 router。
5. `GetRouterInsertedPage()` 给嵌入页 route 填充 page。
6. `GetMenuItemTans()` 按语言填充 label/abbreviation。
7. 按授权 menu IDs 组装父子菜单树，并按 `sort_order` 排序。

已观察到的现状：

- 菜单树会按 feature -> menu 授权过滤。
- `AuthRouter` 的前端 route 校验目前未生效。
- `CheckRouterPermission()` 后端服务会同时检查 feature -> menu 派生的 router 和 feature -> router 映射，但当前前端没有实际启用该检查。
- `GetUserConfig()` 中 `routes` 的生成路径与菜单过滤不是完全同一套逻辑，后续不要只依赖 `routes` 判断菜单可见性。

## Implemented Checklist For `/knowledge-bases`

以下是本次已落地的知识库管理菜单配置，可作为后续同类菜单的核对清单。

已采用命名：

| Concept | Candidate |
|---|---|
| Feature name | `知识库管理` |
| Menu key | `knowledgeBases` |
| Router path | `/knowledge-bases` |
| Menu label zh/en | `知识库管理` / `Knowledge Bases` |
| Hide time selector | `true` |
| Placement | 顶层菜单 |
| Default role grant | 新库默认授予 `admin`，不默认授予 `viewer` / `anonymous` |

代码插入点：

1. `backend/pkg/repository/database/init_feature.go`
   - 已在 `validFeatures` 添加 `FeatureName: "知识库管理"`。
   - 已在 `sreFeatures` 添加同名 feature。
   - 顶层 feature，不写 `parentChildMapping`。
2. `backend/pkg/repository/database/init_router.go`
   - 已在 `validRouters` 添加 `{RouterTo: "/knowledge-bases", HideTimeSelector: true}`。
   - 已在 `sreRouters` 添加同路由。
3. `backend/pkg/repository/database/init_insert_page.go`
   - 普通 React 页面不需要改。
4. `backend/pkg/repository/database/init_router_page.go`
   - 普通 React 页面不需要改。
5. `backend/pkg/repository/database/init_menu_item.go`
   - 已在 `validMenuItemMappings` 添加 `MenuItem{Key: "knowledgeBases", Order: 2}` 和 `RouterKey: "/knowledge-bases"`。
   - 已在 `sreMenuItemMappings` 添加同 key。
   - 顶层菜单，不写 `relations`。
6. `backend/pkg/repository/database/init_feature_mapping.go`
   - 已在 `featureMenuMapping` 添加 `{"知识库管理", "knowledgeBases"}`。
   - 当前没有非菜单详情页，不写 `featureRoutes`。
7. `backend/pkg/repository/database/init_permission.go`
   - 已在 `ROLE_ADMIN` 的 feature 列表添加 `知识库管理`。
   - 不默认授予 `viewer` / `anonymous`。
   - 对已有环境，单靠这里不会自动授权已有角色，需要补迁移或操作流程。
8. `backend/sqlscripts/i18n.yml`
   - 已添加 feature 翻译：`key: "知识库管理"`, `entity_type: "feature"`, `field_name: "featureName"`。
   - 已添加 menu 翻译：`key: "knowledgeBases"`, `entity_type: "menu"`, `field_name: "label"`。
   - 已添加 menu abbreviation：`Knowledge` / `知识库`。
9. `frontend/src/_nav.js`
   - 已给 `navIcon.knowledgeBases` 添加顶层菜单图标。
10. 验证
   - 启动后确认 `router`、`menu_item`、`feature`、`feature_mapping`、`i18n_translation` 有新记录。
   - 用 admin 调 `/api/permission/config?language=zh`，确认返回 `knowledgeBases` 菜单和 `/knowledge-bases` router。
   - 用目标角色测试是否可见；已有环境重点验证授权是否已经补齐。
   - 浏览器访问 `/#/knowledge-bases`，确认菜单点击和直达都能到前端页面。

## Common Pitfalls

- 不要把 HashRouter 的 `/#/knowledge-bases` 原样写进 `router.router_to`；数据库里写 `/knowledge-bases`。
- 不要只加 `validRouters`。没有 `menu_item` 和 `feature_mapping`，侧边栏不会出现。
- 不要只加 `menu_item`。没有 feature 授权，菜单不会进入用户配置。
- 不要假设改了 `init_permission.go` 就能更新已有角色；已有角色只要已经有 feature 权限，就会跳过默认初始化。
- 不要给普通 React 页面添加 `InsertPage`；`InsertPage` 是 Grafana / Jaeger 等嵌入页使用的。
- 不要忘记 `PRODUCT_MODE=syncause` 分支；如果目标产品是 Syncause，需要同时更新 `sreFeatures`、`sreRouters`、`sreMenuItemMappings`。
- 不要在未确认前直接改共享后端候选；这些初始化文件在 OSS 仓存在同路径文件。
