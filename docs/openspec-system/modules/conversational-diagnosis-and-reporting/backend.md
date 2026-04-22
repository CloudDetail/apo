# Conversational Diagnosis And Reporting Backend

## API Summary

| API Path | Method | Role |
|---|---|---|
| `/api/custom/get` | `GET` | 获取自定义数据视图模板列表。 |
| `/api/custom/create` | `POST` | 创建自定义数据视图模板。 |
| `/api/custom/update` | `POST` | 更新自定义数据视图模板。 |
| `/api/custom/delete` | `DELETE` | 删除自定义数据视图模板。 |
| `/api/alerts/events/report/get` | `POST` | 读取已生成报告快照。 |
| `/api/alerts/events/report/add` | `POST` | 写入报告快照。 |

## Controllers Or Routers

| Controller | Files | Notes |
|---|---|---|
| `custom` | `backend/pkg/api/custom/handler.go` | 本模块最核心的仓库内控制器。 |
| `GetCustomQuery` | `func_getcustomquery.go` | `ShouldBindQuery` 后返回 `response.GetCustomResponse`。 |
| `CreateCustomQuery` | `func_createcustomquery.go` | `ShouldBindJSON` 后返回 `"ok"`。 |
| `UpdateCustomQuery` | `func_updatecustomquery.go` | 更新模板。 |
| `DeleteCustomQuery` | `func_deletecustomquery.go` | 通过 `customName` 删除。 |
| `GetAlertEventReport` | `backend/pkg/api/alerts/func_getalerteventreport.go` | 读取报告快照。 |
| `AddAlertEventReport` | `backend/pkg/api/alerts/func_addalerteventreport.go` | 保存报告快照。 |

## Services

| Service | Files | Notes |
|---|---|---|
| `custom.Service` | `backend/pkg/services/custom/service.go` | 负责把 request/response 与 DAO 连接起来。 |
| `GetCustomQuery` | `service_get_custom.go` | 读取 `custom_query_service`，把逗号拼接的 `querytype` 拆成 `[]string`。 |
| `CreateCustomQuery` | `service_creat_custom.go` | 把 `Type []string` 拼接成 `QueryType string`。 |
| `UpdateCustomQuery` | `service_update_custom.go` | 更新 `querytype`。 |
| `DeleteCustomQuery` | `service_delete_custom.go` | 按 `customName` 删除。 |
| `alerts.GetAlertEventReport` | `service_alert_event_report.go` | 在查询窗口前后各扩一小时读取 `alert_report`。 |
| `alerts.AddAlertEventReport` | `service_add_alert_event_report.go` | 先查重，再写入 ClickHouse。 |

## Models And Schemas

| Type | File | Notes |
|---|---|---|
| Custom Query Table | `backend/pkg/repository/database/dao_custom_query_type.go` | `TableName()` 为 `custom_query_service`。 |
| Custom Request | `backend/pkg/model/request/custom_req.go` | `CustomName` + `Type []string`。 |
| Custom Response | `backend/pkg/model/response/custom_resp.go` | `GetCustomResponse.Customs`。 |
| Alert Report Request | `backend/pkg/model/request/alert_event.go` | `GetAlertEventReportRequest`、`AddAlertEventReportRequest`。 |
| Alert Report Response | `backend/pkg/model/response/alert_event.go` | `GetAlertEventReportResponse`、`ReportData`。 |
| License/AGUI data | 无本模块专属后端实现 | `/ai/api/*` 与 `/originx/api/*` 不在当前主后端中。 |

## Database Mapping

| Storage | File | Mapping |
|---|---|---|
| `custom_query_service` | `dao_custom_query_type.go` | `customname` 保存模板名，`querytype` 保存逗号拼接的数据视图类型。 |
| `alert_report` | `backend/pkg/repository/clickhouse/dao_alert_event_report.go` | 存储 `report_type`、`overview_*`、`topology`、`rootcause`、`suggest`、`data`。 |

## Validation And Rules

| Rule | Evidence |
|---|---|
| 自定义模板名不能重复 | `CreateCustomQuery` 先按 `customname` 计数。 |
| 更新自定义模板前必须存在记录 | `UpdateCustomQuery` 先查 `customname`。 |
| 删除自定义模板前必须存在记录 | `DeleteCustomQuery` 先查 `customname`。 |
| 报告写入前同标签窗口只允许一份 | `AddAlertEventReport` 先调用 `GetAlertEventReport`，已有结果则报错。 |
| 报告读取窗口会扩展到 `startTime-1h` 和 `endTime+1h` | `service_alert_event_report.go`。 |

## Error Handling

| Area | Behavior |
|---|---|
| Custom handlers | 参数绑定失败返回 `code.ParamBindError`；业务失败返回 `code.GetCustomTypeError`、`code.CreateCustomTypeError`、`code.DeleteCustomTypeError`。 |
| Report handlers | 参数绑定失败返回 `code.ParamBindError`；查询/新增失败分别返回 `code.GetAlertEventReportError`、`code.AddAlertEventReportError`。 |
| ClickHouse report query | 如果没有命中记录，`dao_alert_event_report.go` 返回 `no event found with tags ...`。 |

## Notes

1. 当前仓库后端只覆盖自定义数据视图与报告快照，不覆盖 ChatAI/Agent/OriginX 主对话与根因分析接口。
2. `func_addalerteventreport.go` 的 swagger 注释写成 `[get]`，但实际 router 注册是 `POST /api/alerts/events/report/add`。
3. `custom.Service` 中多个错误文案和方法名保留现有拼写，不应在文档中自动修正为别的契约。

