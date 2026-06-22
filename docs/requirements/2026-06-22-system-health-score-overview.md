# 系统健康评分需求目标与实现方案

日期：2026-06-22

## 1. 背景

当前 `/healthy-service` 已被最新提交切换为新的“系统健康评分”原型页，相关原型文件包括：

- `frontend/design-previews/system-health-overview.html`
- `frontend/design-previews/system-health-overview-preview.png`
- `frontend/design-previews/system-health-detail-preview.png`
- `frontend/src/pro/views/systemHealthOverview/index.jsx`
- `frontend/src/pro/views/systemHealthOverview/index.css`

现状问题：

- 前端页面仍是 mock 数据，系统名、服务、分数和指标均硬编码。
- 旧健康报告接口 `/api/system-health/report` 是报告/问题导向，不适合作为新概览表格的直接数据源。
- 产品口径已经确认：本版本不做独立“业务系统配置”，也不让用户在页面上筛选数据组。

## 2. 需求目标

在 `/healthy-service` 实现新的系统健康评分页面，用户打开页面后能直接看到当前可见系统的健康评分概览，并能进入某个系统查看服务级健康详情。

核心目标：

1. 用真实后端数据替换前端 mock。
2. 每个系统来自一个数据组，系统边界稳定、可解释。
3. `/healthy-service` 不展示数据组筛选器。
4. 前端和后端都实现，不只做静态页面。
5. 保留旧 `/api/system-health/report` 报告接口，不影响历史健康报告能力。

## 3. 产品口径

### 3.1 系统来源

本版本定义：

```text
系统 = All 下一级数据组
```

规则：

- `All` 根数据组本身不作为系统展示。
- `All.subGroups` 中每个一级数据组展示为一个系统。
- `systemId = groupId`
- `systemName = groupName`
- `source = data_group`
- 本版本默认用户不会配置复杂父子数据组关系，因此系统列表只取 `All` 直属一级数据组。

如果未来出现二级数据组：

- 二级数据组不单独展示为系统。
- 二级数据组只作为一级数据组范围的一部分参与该系统的数据范围。
- 后端实现需要防御嵌套数据组，但不在 UI 上暴露层级系统能力。

### 3.2 数据组筛选器

`/healthy-service` 页面不展示 Header 上的数据组筛选器。

规则：

- 前端路由不设置 `showDataGroup`。
- 页面请求不依赖 URL 上的 `groupId`。
- 页面请求不读取 `dataGroupReducer.dataGroupId` 作为当前筛选条件。
- 后端仍使用数据组树来生成系统列表，并用权限机制过滤当前用户可见的数据组。

### 3.3 权限口径

用户只能看到自己有权限的数据组系统。

规则：

- 如果用户有 `All` 权限，则展示 `All` 下全部一级数据组。
- 如果用户只被授权某些一级数据组，则只展示这些一级数据组。
- 如果用户只被授权某个一级数据组下的子组，本版本建议后端仍将其归并到对应一级数据组展示；如果无法可靠归并，则该系统不展示。
- 不允许用户通过详情接口访问自己无权限的数据组系统。

## 4. 页面目标

### 4.1 概览页

用户进入 `/healthy-service` 后看到：

- 系统健康概览标题。
- 汇总信息：
  - 系统总数
  - 异常系统数
  - 严重异常指标数
  - 异常服务数或无数据系统数
- 系统健康评分表：
  - 系统名称
  - 健康度分数
  - 资源水位
  - 资源波动
  - 错误率
  - 响应延时
  - 错误日志
  - 异常请求
  - 锁等待
- 只看异常开关。
- 系统搜索。
- 红/黄/绿/灰图例。

交互：

- 默认优先展示异常系统。
- 点击系统行进入该系统详情。
- 无数据时展示空态，不展示 mock 系统。

### 4.2 详情页

点击某个系统后展示：

- 返回概览按钮。
- 系统名称。
- 系统健康评分。
- 主要异常指标。
- 关注指标。
- 关键服务或服务列表摘要。
- 系统级指标卡片。
- 服务列表：
  - 服务名
  - 健康度
  - 七个指标状态
  - 主要异常
- 优先处理列表。

交互：

- 支持服务搜索。
- 支持按状态过滤服务。
- AI 分析按钮第一版可沿用原型静态文案或隐藏；如果没有真实 AI 结果，不应伪装成真实分析。

## 5. 指标口径

后端必须返回固定七个指标 key，前端只负责展示，不在前端推导业务评分。

| key | 中文名 | 目标含义 | 无数据处理 |
| --- | --- | --- | --- |
| `resource` | 资源水位 | CPU、内存、实例/副本可用性等资源压力 | `gray` |
| `resourceShift` | 资源波动 | 服务关联容器 CPU/内存相对 24h 前基线的突增、突降或异常波动 | `gray` |
| `errorRate` | 错误率 | 请求错误率或服务错误比例 | `gray` |
| `latency` | 响应延时 | 服务请求 P95/P99 或平均延时风险 | `gray` |
| `errorLog` | 错误日志 | 错误日志数量或同比/环比增长 | `gray` |
| `badRequests` | 异常请求 | 失败请求、异常 trace、相关告警等请求风险 | `gray` |
| `lockCost` | 锁等待 | 服务关联 pod 的 futex 锁等待耗时按请求数均摊后的风险 | `gray` |

状态值固定为：

- `green`：正常
- `yellow`：关注
- `red`：异常
- `gray`：无数据

要求：

- 没有可靠数据源的指标必须返回 `gray`，不能为了让页面好看而默认成 `green`。
- 系统级指标由系统内服务级指标聚合得到，取最严重状态。
- 系统健康分数由后端计算，范围 `0-100`。
- 详情页服务分数也由后端返回，范围 `0-100`。

## 6. 后端接口

### 6.1 系统健康概览

```http
GET /api/system-health/overview
```

查询参数：

| 参数 | 类型 | 必填 | 说明 |
| --- | --- | --- | --- |
| `relativeTime` | string | 否 | 相对时间，如 `15m`、`1h`、`1d` |
| `startTime` | int64 | 否 | 自定义开始时间，微秒 |
| `endTime` | int64 | 否 | 自定义结束时间，微秒 |

默认时间：

- 如果没有传时间，默认 `relativeTime=15m`。
- 如果传了有效 `startTime/endTime`，优先使用自定义时间。

响应结构建议：

```json
{
  "summary": {
    "totalSystems": 6,
    "abnormalSystems": 3,
    "lowScoreSystems": 1,
    "abnormalMetrics": 12,
    "severeMetrics": 4,
    "noDataSystems": 0,
    "services": 28,
    "abnormalServices": 7
  },
  "systems": [
    {
      "id": 1001,
      "groupId": 1001,
      "name": "订单系统",
      "score": 72,
      "status": "yellow",
      "source": "data_group",
      "metrics": {
        "resource": "green",
        "resourceShift": "gray",
        "errorRate": "yellow",
        "latency": "red",
        "errorLog": "green",
        "badRequests": "yellow",
        "lockCost": "gray"
      },
      "metricDetails": {
        "latency": {
          "status": "red",
          "value": "960ms",
          "note": "P95 响应延时明显异常"
        }
      }
    }
  ],
  "window": "15m",
  "generatedAt": "2026-06-22T10:00:00+08:00",
  "source": "data_group"
}
```

### 6.2 系统健康详情

```http
GET /api/system-health/overview/:groupId
```

路径参数：

| 参数 | 类型 | 说明 |
| --- | --- | --- |
| `groupId` | int64 | 系统对应的一级数据组 ID |

查询参数同概览接口。

响应结构建议：

```json
{
  "system": {
    "id": 1001,
    "groupId": 1001,
    "name": "订单系统",
    "score": 72,
    "status": "yellow",
    "source": "data_group",
    "metrics": {},
    "metricDetails": {},
    "services": [
      {
        "name": "order-api",
        "desc": "服务健康指标",
        "score": 58,
        "status": "red",
        "metrics": {
          "resource": "green",
          "resourceShift": "gray",
          "errorRate": "yellow",
          "latency": "red",
          "errorLog": "green",
          "badRequests": "yellow",
          "lockCost": "gray"
        },
        "metricDetails": {}
      }
    ]
  },
  "window": "15m",
  "generatedAt": "2026-06-22T10:00:00+08:00",
  "source": "data_group"
}
```

### 6.3 与旧接口关系

保留：

- `GET /api/system-health/report`
- `GET /api/system-health/report/records`
- `GET /api/system-health/report/records/:recordId`

新概览页面优先使用 `/api/system-health/overview` 系列接口；旧报告接口不作为新页面主数据源。

## 7. 后端实现边界

后端负责：

1. 从数据组树中解析系统列表。
2. 按当前用户权限过滤系统列表。
3. 按每个系统的 `groupId` 限定指标查询范围。
4. 计算系统级和服务级分数。
5. 返回固定七指标状态和详情。
6. 明确返回无数据状态，不伪造正常状态。

后端不负责：

- 提供独立系统配置表。
- 支持用户在健康评分页切换数据组。
- 支持多级数据组作为多级系统树展示。
- 修改旧健康报告语义。

## 8. 前端实现边界

前端负责：

1. `/healthy-service` 路由隐藏数据组筛选器。
2. 页面加载时请求 `/api/system-health/overview`。
3. 用接口返回的 `systems` 渲染概览表。
4. 点击系统后请求 `/api/system-health/overview/:groupId` 或使用已返回详情数据渲染详情。
5. 保留原型中的搜索、只看异常、服务过滤、图例和详情切换交互。
6. 增加加载态、错误态、空态。

前端不负责：

- 从 mock 数据回退展示系统。
- 在前端推导系统来源。
- 在前端计算最终业务分数。
- 读取全局 `dataGroupReducer` 作为本页筛选条件。

## 9. 验收标准

功能验收：

1. 打开 `/healthy-service`，Header 不展示数据组筛选器。
2. 页面不再展示硬编码的 `订单系统 / 支付系统 / 营销系统` mock 数据。
3. 后端返回的每个系统都对应 `All` 下一级数据组。
4. `All` 根数据组不会作为系统展示。
5. 系统详情中的 `groupId` 与概览行一致。
6. 无权限的数据组不会出现在系统列表，也不能访问详情。
7. 没有数据的指标显示灰色“无数据”。
8. 后端异常时页面展示错误态，不保留假数据。

接口验收：

1. `GET /api/system-health/overview` 返回 `summary/systems/window/generatedAt/source`。
2. `GET /api/system-health/overview/:groupId` 返回单个系统详情和服务列表。
3. 响应中所有系统和服务都包含固定七个指标 key。
4. 分数范围始终是 `0-100`。
5. 状态值只出现 `green/yellow/red/gray`。

测试验收：

1. 后端单元测试覆盖：
   - 只取 `All` 下一级数据组作为系统。
   - 不展示 `All` 根组。
   - 不展示无权限数据组。
   - 七指标 key 固定返回。
   - 服务状态聚合成系统状态。
2. 前端至少覆盖：
   - 概览 API 数据渲染。
   - 空态渲染。
   - 错误态渲染。
   - 点击系统进入详情。
   - `/healthy-service` 不依赖 `dataGroupId`。

## 10. 非目标

本版本不做：

- 系统自定义配置管理。
- 系统与服务的手工映射。
- 多级系统树。
- 健康评分阈值配置 UI。
- AI 真实诊断闭环。
- 替换旧健康报告页面和报告记录能力。

## 11. 待确认问题

1. 如果用户只拥有二级数据组权限，是否必须展示其所属一级系统，还是可以不展示。
2. `lockCost` 已接入 Polaris futex 指标；没有服务关联 Polaris 基线时按 `gray` 返回。
3. `resourceShift` 已接入 cadvisor CPU/内存指标；没有服务关联资源指标或 24h 前基线时按 `gray` 返回。
4. 详情页“AI 分析”按钮第一版是隐藏、静态建议，还是接真实分析接口。
