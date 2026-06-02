# 自定义大盘需求补充说明

> 本文补充 `docs/superpowers/specs/2026-04-29-custom-dashboard-prd.md` 中未明确或已与当前实现、OpenSpec 决策发生漂移的内容。原 PRD 作为需求背景和初版范围保留；当前实现口径、后续验收和新变更拆分应优先参考本文与 OpenSpec 规格。

## 1. 补充原则

- 不直接重写 2026-04-29 PRD，避免丢失原始需求背景。
- 本文只补充已能从现有代码、OpenSpec 和实现状态推断的约束。
- 与当前实现冲突的原 PRD 条目，在本文中明确采用当前实现和最新 OpenSpec 决策。
- 需要新增能力但当前代码没有支持的内容，标记为“后续需求”，不混入当前 MVP 验收口径。

## 2. 当前优先口径

### 2.1 文档优先级

当前自定义大盘需求和实现的解释顺序为：

1. 最新 OpenSpec change 中已完成的设计决策。
2. 当前代码中已经落地的接口、模型、前端交互和限制。
3. 本补充说明。
4. 2026-04-29 原始 PRD。

如果以上内容存在冲突，应先按 1 和 2 判断当前系统行为，再决定是否新建 OpenSpec change 修改实现。

### 2.2 与原 PRD 已发生变化的关键点

- 新建组件类型当前只支持折线图、面积图、柱状图和表格。
- TopN 仅作为历史兼容类型读取，当前按表格渲染。
- 单值指标作为历史不支持类型保留提示，不再作为新建组件提供。
- 编辑器布局为固定两列、无限行、固定单元格，不支持用户调整组件宽高。
- 时间范围使用 APO 全局时间选择器，不作为大盘变量保存。
- 不提供大盘级自动刷新间隔配置；支持手动刷新和全局时间范围刷新。
- 大盘配置当前存储为 `custom_dashboard.config` JSON，不拆成独立 panel/query 多表。

## 3. 团队可见

### 3.1 团队来源

“团队可见”中的团队应定义为 APO 现有 Team 体系，即系统中的团队管理能力及 `user_team` 关系，不是部门、角色、用户组或数据组。

团队可见只控制大盘入口和详情页访问权限；数据组权限只控制变量候选值和指标查询结果。两者必须保持隔离：

```text
Dashboard Visibility
  everyone / team / personal
  -> controls whether user can open the dashboard

Data Group Permission
  user's allowed data scopes
  -> controls variable candidates and PromQL query scope
```

### 3.2 团队选择和成员变化

- 团队可见应支持选择多个团队，保存为 `teamIds`。
- 用户是否可见团队大盘，应按打开列表或详情时的实时团队成员关系判断。
- 创建者无论是否仍属于已选团队，都应始终能查看、编辑和删除自己创建的大盘。
- 非创建者即使属于可见团队，也只能查看，不能编辑或删除。

### 3.3 管理员可见性

当前实现没有定义“管理员无条件查看所有大盘”的规则。因此当前 MVP 不应把“管理员可查看所有团队大盘”作为验收项。

如果产品需要管理员全局可见，应作为后续权限需求补充，明确：

- 哪些角色或功能权限代表全局查看权。
- 全局查看是否包含编辑和删除。
- 是否仍受具体大盘菜单、路由权限限制。

## 4. 变量模型

### 4.1 支持变量

当前 MVP 变量口径应收敛为：

- `dataGroup`
- `cluster`
- `namespace`
- `service`
- `pod`
- `label`：用于从 Prometheus/VictoriaMetrics label values 查询候选值的普通 label 变量，例如 Grafana 的 `job_name`、`cadvisor_job_name`
- `custom`：用于固定选项变量，例如 Grafana 的 `resolution=1m/3m/5m`

时间范围不是大盘变量，由 APO 全局时间选择器控制。

### 4.2 变量创建方式

变量不是完整 Grafana 变量系统。APO 当前支持语义变量、普通 label 变量、固定选项变量，以及必要的 Grafana 导入兼容子集；不支持 datasource 变量、ad hoc filter 或完整 Grafana 变量编辑器。

当前推荐口径：

- 系统根据组件查询和预置查询发现可用 label。
- 对常见语义变量使用内置 alias 映射。
- 用户只选择启用哪些变量，可删除已启用变量。
- 对无法识别为语义变量的 label，可作为普通 label 变量暴露，但不应要求普通用户理解全部底层 label 体系。
- 从 Grafana 导入或转换时，可以保留 `label_values(metric{...},label)` 形式的 label 变量查询、静态选项变量、`includeAll` 和 `allValue`；查询中的上游变量会在候选值请求前替换。
- 导入后的 PromQL 兼容 `$name`、`${name}`、`${name:regex}`、`${name:raw}` 和 `{{ name }}` 变量格式；`${name:regex}` 用于 PromQL regex matcher，普通 `$name`/`${name}` 保持原始值。
- Grafana 的 datasource 变量不迁移为 APO 大盘变量，因为 APO 自定义大盘统一走后端配置的 Prometheus/VictoriaMetrics 数据源。

内置语义 alias 建议记录为：

| 变量 | Label aliases |
|---|---|
| `dataGroup` | `data_group`, `dataGroup`, `data_group_id`, `group_id` |
| `cluster` | `cluster`, `cluster_id`, `clusterId` |
| `namespace` | `namespace`, `namespace_name`, `namespaceName` |
| `service` | `svc`, `svc_name`, `service`, `service_name` |
| `pod` | `pod`, `pod_name`, `instance`, `instance_name` |

### 4.3 级联关系

变量候选值应按当前用户数据组权限和已选上游变量过滤。

建议定义顺序：

```text
dataGroup -> cluster -> namespace -> service -> pod
```

具体规则：

- 选择 `dataGroup` 后，后续变量候选值仅来自该数据组允许的数据范围。
- 选择 `cluster` 后，`namespace/service/pod` 候选值按集群过滤。
- 选择 `namespace` 后，`service/pod` 候选值按命名空间过滤。
- 选择 `service` 后，`pod` 候选值按服务过滤。
- 当前变量查询候选值时，应排除当前变量自身选择，避免自己过滤自己。

### 4.4 单选、多选和空值

- 变量控件按变量定义的 `multi` 字段决定单选或多选；未声明 `multi: true` 的 Grafana 导入变量按单选处理。
- 当变量声明 `includeAll: true` 时，查看页在候选列表前插入 `All`；其值使用变量定义的 `allValue`，未声明时默认为 `.*`。
- `dataGroup` 可以是单选，具备全局数据权限的用户可出现 `All Data`。
- 空选择建议定义为“不过滤该变量”，即查询当前用户有权限范围内的全部值。
- 如果某变量被标记为 required，查看页应在无默认值时提示用户选择，避免把空值替换进 PromQL。

当前代码对空值处理仍偏宽松，后续如要严格 required 行为，需要新建需求补充实现。

### 4.5 变量变更刷新范围

变量变化后，不需要刷新所有组件。应只刷新引用该变量或命中该变量 label alias 的组件。

首次打开大盘、手动刷新、全局时间变化时，可以刷新所有组件。

## 5. PromQL 变量和权限注入

### 5.1 变量语法

当前应区分两类机制：

1. 预置查询参数兼容 `$name`、`${name}`、`${name:regex}`、`${name:raw}` 和 `{{ name }}`。
2. 自定义查询优先使用 label alias 自动注入 matcher。

不建议把 `$service` 定义成用户必须掌握的唯一语法。更稳妥的产品口径是：

- 用户在 PromQL 中写正常 label selector。
- 系统识别 selector 中的 label 是否匹配已启用变量。
- 如果匹配，则把对应 label matcher 改写为当前变量选择。

示例：

```promql
sum(rate(http_requests_total{svc_name="checkout"}[5m]))
```

当 `service = ["checkout", "cart"]` 时，可改写为：

```promql
sum(rate(http_requests_total{svc_name=~"checkout|cart"}[5m]))
```

### 5.2 多选变量展开

多选变量统一展开为 PromQL regex matcher：

```text
label=~"value1|value2|value3"
```

变量值必须进行 PromQL regex 转义，避免 `.`、`+`、`|` 等字符改变匹配语义。

### 5.3 不匹配查询的变量

如果某组件查询不包含变量对应的任何 label alias，则该变量不应注入该组件查询。

这能避免一个 dashboard-level 变量破坏无关指标查询。

### 5.4 后端数据权限注入

前端变量过滤不能作为权限边界。所有查询必须经过后端，并由后端按当前用户数据组权限追加数据范围过滤。

如果用户显式选择了 `dataGroup`：

- 后端必须校验当前用户是否有该数据组查看权限。
- 校验通过后按该数据组 scope 生成 PromQL matcher。
- 校验失败时返回无权限状态，不执行 Prometheus 查询。

如果用户未选择 `dataGroup`：

- 后端按当前用户全部可见数据组生成 scope 过滤。
- 用户具备全局数据权限时，允许不追加数据范围 matcher。

## 6. 预置查询

### 6.1 发布方式

预置查询由后端 catalog 随版本发布。MVP 不支持管理员在线新增、编辑或删除预置查询。

### 6.2 字段结构

当前预置查询结构应按现有模型说明：

```json
{
  "id": "service-qps",
  "name": "Request QPS",
  "description": "Service request throughput",
  "category": "service",
  "expression": "sum(rate(span_trace_duration_count{svc_name=\"$service\"}[1m]))",
  "compatibleCharts": ["line", "area"],
  "parameters": [
    {
      "name": "service",
      "label": "Service",
      "description": "Service variable or concrete service name",
      "required": true,
      "default": "$service"
    }
  ]
}
```

### 6.3 第一批分类

当前分类应使用：

- `service`
- `instance`
- `infrastructure`

第一批预置查询应覆盖：

- 服务请求量 QPS
- 服务错误率
- P95 延迟
- P99 延迟
- Pod CPU
- Pod 内存
- Pod 重启次数
- 节点 CPU
- 节点内存
- 磁盘使用率

## 7. 图表和组件配置

### 7.1 支持组件类型

当前 MVP 新建组件只支持：

- 折线图 `line`
- 面积图 `area`
- 柱状图 `bar`
- 表格 `table`

以下类型不再作为新建组件提供：

- 单值指标 `single`
- TopN 排名 `topn`

历史兼容规则：

- 历史 `topn` 组件按 `table` 渲染。
- 历史 `single` 组件显示不支持状态，允许用户删除或改成支持类型。

### 7.2 布局

当前编辑器采用固定两列网格：

- 每个组件占一个固定单元格。
- 支持拖拽调整位置。
- 不支持拖拽调整宽高。
- 不支持配置组件宽高。
- 增加、复制、删除或拖拽后按行优先自动压缩布局，不保留中间空洞。

### 7.3 通用配置

当前有效配置包括：

- 标题
- 描述
- 图表类型
- 查询模式
- 预置查询或自定义 PromQL
- Legend
- 单位
- 小数位
- 空状态文案
- 错误状态文案

不应把未产生运行时效果的配置作为 MVP 验收项，例如阈值、多 Y 轴、堆叠、平滑曲线、null 值处理策略等。

### 7.4 各类型展示规则

折线图、面积图、柱状图：

- 使用时间序列数据。
- legend 默认从 `svc_name`、`pod`、`instance` 或序列序号生成。
- 支持 `{{label}}` 和 `$label` 形式的 legend 模板。
- 单位应用到 y 轴和 tooltip 数值。
- 小数位应用到数值格式化。

表格：

- 以 series 为行。
- 默认列为序列名和值。
- 值取该序列最后一个点。
- 数值型值应用单位和小数位；非数值不追加单位。

## 8. 查询限制和性能边界

### 8.1 当前后端默认值

当前 MVP 查询限制应记录为：

| 限制项 | 当前值 |
|---|---:|
| 单个大盘最大组件数 | 60 |
| 固定网格列数 | 2 |
| 单次 range query 最大返回点数估算 | 12000 |
| 默认 step | 60 秒 |
| 查询超时 | 30 秒 |
| 后端全局并发查询数 | 8 |

### 8.2 未实现或未明确的限制

以下限制当前没有完整产品化配置，不应作为已实现能力宣称：

- 单个组件最多几条查询。
- 单个大盘每用户并发刷新队列。
- 自动刷新最小间隔。
- 根据时间范围自动计算最佳 step 的完整策略。

如果后续需要更精细的限流和 step 规划，应新增 OpenSpec change。

## 9. 时间范围和刷新

### 9.1 时间范围来源

自定义大盘查看页和编辑预览使用 APO 全局时间选择器，不单独保存 dashboard-level time range。

当前 fallback 快捷范围为：

- 最近 15 分钟
- 最近 1 小时
- 最近 6 小时
- 最近 24 小时

如果全局时间选择器未提供有效开始和结束时间，前端 fallback 为最近 15 分钟。

### 9.2 Step 规则

当前前端请求 step 规则：

- 时间范围小于等于 6 小时：60 秒。
- 时间范围大于 6 小时：300 秒。

后端在 step 为空或非法时使用 60 秒默认值。

### 9.3 刷新行为

- 查看页支持手动刷新。
- 全局时间范围变化或全局 refresh timestamp 变化时刷新组件。
- 当前不提供大盘级自动刷新间隔配置。
- 如需自动刷新间隔，应作为后续需求新增，明确默认值、最小间隔和权限/性能限制。

## 10. 保存并发

当前保存模型是“创建者保存即覆盖服务端配置”，没有 version 或 updateTime 乐观锁。

因此 MVP 口径应明确：

- 只有创建者可以编辑。
- 同一创建者打开多个页面时，后保存的页面会覆盖先保存的配置。
- 保存失败时前端保留当前编辑状态并展示错误。

如果要避免多页面覆盖，需要新增需求：

- 请求携带 `updatedAt` 或 `version`。
- 服务端保存前比较当前版本。
- 版本不一致时返回冲突状态，提示刷新后重试。

## 11. 删除影响

当前大盘删除应按以下口径说明：

- 只有创建者可以删除。
- 删除前由前端二次确认。
- 删除大盘记录采用数据库软删除。
- 删除时同步移除该 dashboard 对应的自定义 menu、router、feature、feature mapping、权限和 i18n 翻译。
- 删除原大盘不影响已经复制出来的新大盘。
- 已打开或浏览器历史中的详情页刷新后，应进入大盘不存在或无权限状态。
- 查询审计记录不因大盘删除而删除，可用于后续审计追踪。

## 12. 当前数据模型

### 12.1 Dashboard 存储

当前主表为 `custom_dashboard`：

| 字段 | 说明 |
|---|---|
| `id` | 大盘 ID |
| `name` | 大盘名称 |
| `description` | 描述 |
| `owner_id` | 创建者用户 ID |
| `visibility` | `everyone` / `team` / `personal` |
| `team_ids` | 团队可见范围，JSON 数组字符串 |
| `config` | 大盘配置 JSON |
| `last_visited_at` | 最近访问时间 |
| `created_at` | 创建时间 |
| `updated_at` | 更新时间 |
| `deleted_at` | 软删除时间 |

`config` 中保存：

- variables
- widgets
- widget layout
- widget query
- widget display

### 12.2 查询审计

当前查询审计表为 `custom_dashboard_query_audit`：

| 字段 | 说明 |
|---|---|
| `user_id` | 执行查询的用户 |
| `dashboard_id` | 大盘 ID；当前代码中可能为空，后续可补齐 |
| `widget_id` | 组件 ID |
| `query_mode` | `preset` / `custom` |
| `expression` | 原始表达式 |
| `resolved_query` | 变量替换和权限注入后的表达式 |
| `resolved_vars` | 查询变量 |
| `constrained_scope` | 数据权限约束 |
| `status` | 查询状态 |
| `error_message` | 错误信息 |
| `result_series` | 返回序列数 |
| `result_points` | 返回点数 |
| `created_at` | 审计时间 |

## 13. 建议同步到 OpenSpec 的变更点

后续如果要继续开发或正式验收，建议新建 OpenSpec change，至少同步以下规格差异：

- `custom-dashboard-editor`：组件类型应从 `line/area/bar/single/table/topn` 更新为 `line/area/bar/table`，并补充历史 `single/topn` 兼容规则。
- `custom-dashboard-management`：列表是否显示更新时间和最近访问时间需要重新决策；当前实现是显示。
- `custom-dashboard-variables`：把“手动 label binding”改为“语义变量 + 内置 alias + 可选普通 label 变量 + 固定选项变量”的当前实现口径，并说明 Grafana `label_values` 变量的兼容边界。
- `custom-dashboard-querying`：补充具体查询限制值、变量 matcher 改写规则和数据权限注入规则。
- `custom-dashboard-viewer`：明确不提供大盘级自动刷新间隔，刷新依赖手动刷新和全局时间范围。

## 14. 后续待决问题

以下问题当前不应默认实现，需要产品或架构确认：

- 管理员是否能无条件查看所有大盘。
- 管理员是否能编辑或删除他人创建的大盘。
- 是否需要团队可见的团队选择 UI 和团队名称展示。
- 变量 required 时空选择是否阻止查询。
- 是否支持 dashboard-level 自动刷新间隔。
- 是否需要乐观锁防止多页面保存覆盖。
- 查询审计是否必须补齐 `dashboard_id`。
- 是否需要导入、导出或迁移 dashboard 配置。
