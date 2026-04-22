# Identity Access And Organization Dependencies

## Upstream Dependencies

| Dependency | Evidence | Why It Matters |
|---|---|---|
| `request.js` | `frontend/src/core/api/*.js|ts` | 统一注入 token、处理登录过期。 |
| `jwt.GenerateTokens` | `backend/pkg/services/user/service_login.go` | 登录 token 生成基础。 |
| `DeployModeContext` | `Login.jsx` | 决定品牌展示和是否登录 Dify。 |
| init metadata | `init_feature.go`, `init_menu_item.go`, `init_router.go` | 没有这些预置元数据，权限树和菜单就无法稳定生成。 |

## Downstream Dependencies

| Consumer | Evidence | Dependency Relationship |
|---|---|---|
| 全站导航 | `UserContext.js`, `AppSidebar.js` | 侧边栏菜单和 routes 直接依赖本模块的 `/api/permission/config`。 |
| `data-scope-governance` | `DataGroupAuthorizeModal` | 用户和团队页面复用数据组授权能力。 |
| `platform-administration` | `DefaultLayout`, `DeployModeContext` | 品牌与 license 告警会影响用户进入系统后的整体体验。 |

## Shared Components

| Shared Component | File | Role In This Module |
|---|---|---|
| `AuthRouter` | `frontend/src/core/components/AuthRouter.tsx` | 统一系统页包装。 |
| `PermissionTree` | `frontend/src/core/components/PermissionTree/index.tsx` | 角色/权限配置复用。 |
| `BasicCard` | 多个系统管理页面 | 标准系统页容器。 |
| `DataGroupAuthorizeModal` | `frontend/src/core/components/PermissionAuthorize/DataGroupAuthorizeModal.tsx` | 用户/团队数据组授权。 |

## Shared Utilities

| Utility | File | Role |
|---|---|---|
| `redirectToLogin` | `frontend/src/core/utils/redirectToLogin.js` | 退出登录后统一跳转登录页。 |
| `notify` | `src/core/utils/notify` | 登录、保存、删除、授权提示。 |
| `getUserFeatureIDs` | `service_get_user_permission.go` | 聚合 role/team/user 三类 feature。 |
| `checkRouterMatch` | `service_check_router.go` | 路由模式匹配。 |

## External Services

| Service | Evidence | Current Role |
|---|---|---|
| Dify Console | `frontend/src/core/api/workflows.ts` | 登录后获取 `difyToken`。 |
| Dify User Repo | `backend/pkg/repository/dify` | 创建用户、删除用户、修改密码、重置密码同步。 |

## Notes

- 本模块是全站“谁能看到什么、谁能做什么”的根部能力，因此它与菜单、路由、数据组、品牌和外部 workflow 平台都有直接连接。
- 由于存在 Dify 同步，本模块并不是纯粹的本地权限子系统；文档必须明确这是外部边界而非本仓库完整实现。
