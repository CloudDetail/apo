# 自定义大盘单图多查询设计

## 背景

当前自定义大盘组件只有一个 `query` 对象，预览接口一次只执行一个 PromQL 或一个预置查询。新增需求要求一个图表能展示多个指标查询语句。

## 范围

本次只扩展自定义 PromQL 模式：

- 一个组件可配置多条自定义 PromQL 查询。
- 多条查询结果合并到同一个图表或表格中展示。
- 新增展示类型 `stat` 和 `barGauge`。
- 预置查询模式保持单条预置查询，不支持多个预置查询混用。
- 历史单查询配置继续兼容。

## 数据模型

在 `WidgetQuery` 中新增 `queries` 数组：

```json
{
  "mode": "custom",
  "expression": "up",
  "legend": "{{svc_name}}",
  "queries": [
    {
      "id": "q-1",
      "expression": "up{job=\"apiserver\"}",
      "legend": "API Server"
    },
    {
      "id": "q-2",
      "expression": "up{job=\"kubelet\"}",
      "legend": "Kubelet"
    }
  ]
}
```

兼容规则：

- 如果 `queries` 为空且存在旧字段 `expression`，归一化为一条查询。
- 保存时保留旧字段 `expression` 和 `legend` 为第一条查询的镜像值，降低旧代码读取风险。
- 至少保留一条自定义查询。

## 查询执行

前端预览请求新增 `queries` 字段。后端收到多条查询后：

- 逐条进行 PromQL/VictoriaMetrics 语法校验。
- 逐条执行变量替换和后端数据权限注入。
- 逐条执行 range query。
- 将所有返回 series 合并到同一个响应的 `series` 中。
- 为每条 series 增加查询 legend 元信息，图表渲染优先使用单条查询 legend。

如果任一查询语法错误或运行失败，组件预览返回失败状态，错误信息包含出错查询序号。

## 前端编辑

自定义查询模式下展示查询列表：

- 每条查询包含 PromQL 输入和 Legend 输入。
- 支持添加查询。
- 支持删除查询，但至少保留一条。
- 变量发现、变量注入、预览和查看页刷新都基于所有查询文本。

## Stat 和 Bar Gauge 展示

`stat` 和 `barGauge` 与 `line`、`area`、`bar`、`table` 同级，允许新建和切换。

展示规则：

- 每条查询返回的每个 series 都取最后一个点展示。
- `stat` 展示多个数值块，每个块显示 series 名称和值。
- `barGauge` 展示多个横向进度条，每条显示 series 名称和值。
- `barGauge` 进度按 `0-100` 计算并 clamp 到边界，文本保留真实格式化值。
- 单位、小数位和 legend 继续生效。

## 测试

需要覆盖：

- 历史 `expression/legend` 自动归一化为 `queries[0]`。
- `buildPreviewRequest` 为多查询生成 `queries` payload，并对每条表达式注入变量 matcher。
- 后端多查询预览逐条执行并合并 series。
- 任一查询校验失败时返回 `syntax_error`。
- 图表 series 名称优先使用每条查询的 legend。
- `stat` 和 `barGauge` 可作为新组件类型创建并展示最后一个点。
