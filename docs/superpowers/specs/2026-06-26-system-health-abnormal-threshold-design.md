# System Health Abnormal Threshold Design

## Context

The system health overview page currently treats a system as abnormal when it
has service data and either:

```text
system.score < 80
```

or any system metric is yellow/red.

That logic is implemented in both places:

- Backend summary calculation in
  `backend/pkg/api/systemhealth/func_getoverview.go`
- Frontend "只看异常" filtering in
  `frontend/src/pro/views/systemHealthOverview/index.jsx`

This makes the abnormal system count too strict: a system with acceptable
overall health can be counted as abnormal because one indicator is yellow.

## Goal

Make abnormal system judgment depend only on health score, and allow users to
configure the global abnormal score threshold.

The threshold is global. All users see the same abnormal system count and the
same "只看异常" result.

Default behavior remains compatible:

```text
default abnormalScoreThreshold = 80
```

## Non-Goals

- Do not change how metric lamps are calculated.
- Do not change the severe/general abnormal metric counts.
- Do not change pending issue generation; it continues to list yellow/red
  service metrics.
- Do not redesign the whole system health page.
- Do not make the threshold per-user, per-team, or per-system.
- Do not reuse service RED threshold settings for system health score settings.

## Product Semantics

A system is abnormal when:

```text
len(system.services) > 0 && system.score < abnormalScoreThreshold
```

A service is abnormal for summary purposes when:

```text
service.score < abnormalScoreThreshold
```

No-data systems remain no-data and are not counted as abnormal.

Metric status remains independent:

```text
red/yellow metric lamps still show red/yellow
severeMetrics still counts red metrics
generalMetrics still counts yellow metrics
pending issues still use yellow/red service metrics
```

This keeps the distinction clear:

- "异常系统" answers: Which systems have low health score?
- "严重异常 / 一般异常" answers: Which indicators are currently bad?
- "待处理问题" answers: Which service-level metric issues should be reviewed?

## Backend Design

Add a small global configuration model in the database layer:

```go
type SystemHealthConfig struct {
    ID                     uint      `gorm:"primaryKey;autoIncrement"`
    AbnormalScoreThreshold int       `gorm:"column:abnormal_score_threshold"`
    UpdatedAt              time.Time `gorm:"column:updated_at"`
}
```

Use one row for the global setting. The repository should expose get-or-create
and update operations. If the row is missing, create or return the default
threshold of 80.

Add the model to existing Gorm `AutoMigrate` in
`backend/pkg/repository/database/dao.go`.

### API

Add two endpoints under the existing authenticated system health group:

```text
GET  /api/system-health/config
POST /api/system-health/config
```

Response:

```json
{
  "abnormalScoreThreshold": 80
}
```

POST request:

```json
{
  "abnormalScoreThreshold": 75
}
```

Validation:

```text
abnormalScoreThreshold must be an integer from 1 to 100
```

If validation fails, return the existing parameter binding/validation error
pattern used by the project.

### Overview Flow

`GetOverview` should load config before building the summary and pass the
threshold into summary construction:

```go
summary := buildSystemHealthOverviewSummary(systems, abnormalThreshold)
```

The existing optional previous-system argument for new abnormal metrics can
remain, but it should not be confused with abnormal system judgment. New
abnormal metric counting remains based on metric status transitions.

The helper signature should make the threshold explicit before the existing
optional previous-system data:

```go
func buildSystemHealthOverviewSummary(
    systems []response.SystemHealthSystem,
    abnormalThreshold int,
    previousSystems ...[]response.SystemHealthSystem,
) response.SystemHealthOverviewSummary
```

The response should include the threshold so frontend views can stay in sync
without always making a separate config request:

```go
type SystemHealthOverviewResp struct {
    Summary  SystemHealthOverviewSummary `json:"summary"`
    Systems  []SystemHealthSystem        `json:"systems"`
    Config   SystemHealthConfigResp      `json:"config"`
    ...
}
```

`GetSystemDetail` may also include the config if needed by the detail page, but
it is not required for the current UI because the threshold only affects the
overview filter and summary.

### Summary Calculation

Replace the current abnormal system logic:

```go
if system.Score < 80 || hasAbnormalMetric(system.Metrics) {
    summary.AbnormalSystems++
}
```

with:

```go
if system.Score < abnormalThreshold {
    summary.AbnormalSystems++
}
```

Replace abnormal service summary logic similarly:

```go
if service.Score < abnormalThreshold {
    summary.AbnormalServices++
}
```

Keep existing metric counting:

```go
isAbnormalMetricStatus(status) == red || yellow
```

## Frontend Design

Add config API helpers in `frontend/src/pro/api/serviceHealthy.js`:

```js
getSystemHealthConfigApi()
setSystemHealthConfigApi(data)
```

The overview page should initialize `abnormalScoreThreshold` from
`resp.config.abnormalScoreThreshold` when loading overview data. If the field is
missing, fall back to 80.

### Abnormal Filtering

Replace frontend abnormal system logic:

```js
system.score < 80 || has yellow/red metric
```

with:

```js
system.services.length > 0 && system.score < abnormalScoreThreshold
```

This affects only the overview "只看异常" filter.

### Threshold Control

Place a compact control in the system health panel header beside "只看异常":

```text
异常阈值：低于 [80] 分
```

Use Ant Design `InputNumber` with:

```text
min = 1
max = 100
precision = 0
```

Save behavior:

- Save on Enter or blur, or with a compact save button if the project style is
  clearer with explicit confirmation.
- Disable saving while request is in flight.
- On success, reload overview so summary, table filtering, and help text use
  the same threshold.
- On failure, show an Ant Design message/error and keep the previous persisted
  value.

### Help Text

Update abnormal system help text from:

```text
健康度低于 80 分，或存在关注/异常指标时计入异常系统。
```

to:

```text
健康度低于当前阈值时计入异常系统。
```

When rendering, include the actual threshold:

```text
当前阈值：低于 75 分计入异常系统。
```

Health ring colors can remain unchanged for this pass:

```text
<60 red
<80 yellow
>=80 green
```

The customer request changes abnormal judgment, not the visual score bands.

## Data Flow

1. User opens the system health overview page.
2. Frontend requests `/api/system-health/overview`.
3. Backend loads the global config, builds systems, and calculates summary with
   `abnormalScoreThreshold`.
4. Frontend stores the threshold from `resp.config`.
5. Summary cards use backend values.
6. "只看异常" filters locally using the same threshold.
7. User edits the threshold.
8. Frontend posts `/api/system-health/config`.
9. Frontend reloads overview after save succeeds.

## Error Handling

- Missing config row falls back to 80.
- Invalid threshold returns a parameter error and does not update config.
- Config save failure keeps the previous threshold in UI.
- Overview still renders if `resp.config` is absent by falling back to 80.

## Tests

Add backend tests in `backend/pkg/api/systemhealth/func_getoverview_test.go`:

1. A system with score equal to threshold is not abnormal.
2. A system with score below threshold is abnormal.
3. A system with score above threshold and yellow/red metrics is not counted as
   abnormal.
4. Abnormal services use the same threshold.
5. Severe/general metric counts still include red/yellow metrics.

Add frontend tests or focused utility tests if this codebase has an established
pattern for the page helpers:

1. `isAbnormal(system, threshold)` returns true only when score is below
   threshold and the system has service data.
2. Yellow/red metrics alone do not make a system abnormal.
3. Missing config uses the default threshold of 80.

Manual browser check:

1. Set threshold to 75.
2. Confirm the abnormal system card changes according to score only.
3. Confirm "只看异常" shows the same systems as the card count.
4. Confirm yellow/red lamps and pending issues still show metric problems.

## Assumptions

- The threshold is shared globally across all users.
- A threshold of 80 preserves the existing score cutoff while removing the
  metric-lamp part of abnormal system judgment.
- Current health score color bands remain useful and should not be redefined in
  this change.
