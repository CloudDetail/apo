# System Health Abnormal Threshold Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Make system health abnormal-system judgment depend only on a globally configurable health score threshold.

**Architecture:** Add a global `system_health_config` database model and expose it through `/api/system-health/config`. Load that config in the overview API, calculate abnormal system/service summary counts using the score threshold only, and return the active config with overview responses so the frontend filter and help text use the same value. Keep metric lamps, severe/general metric counts, and pending issue generation based on existing red/yellow metric status.

**Tech Stack:** Go/Gin/Gorm backend, existing `systemhealth` API package, React/Ant Design frontend, existing `get`/`post` request helpers.

---

## File Structure

- Modify `backend/pkg/model/response/system_health.go`
  - Add `SystemHealthConfigResp`.
  - Add `Config` to `SystemHealthOverviewResp`.
- Modify `backend/pkg/model/request/system_health.go`
  - Add `UpdateSystemHealthConfigRequest`.
- Modify `backend/pkg/repository/database/dao.go`
  - Add config repo methods to `Repo`.
  - Add `SystemHealthConfig` to `AutoMigrate`.
- Create `backend/pkg/repository/database/dao_system_health_config.go`
  - Own the global config table, default, get-or-create, and update logic.
- Modify `backend/pkg/api/systemhealth/handler.go`
  - Add `GetConfig` and `UpdateConfig` to the handler interface.
- Create `backend/pkg/api/systemhealth/func_config.go`
  - Own config handlers and response conversion.
- Modify `backend/pkg/router/router_api.go`
  - Register `GET /api/system-health/config` and `POST /api/system-health/config`.
- Modify `backend/pkg/api/systemhealth/func_getoverview.go`
  - Load active threshold and pass it to summary construction.
  - Replace abnormal system/service summary judgment with score-only threshold logic.
- Modify `backend/pkg/api/systemhealth/func_getoverview_test.go`
  - Add score-threshold-only summary tests.
  - Update existing summary test call signature.
- Modify `frontend/src/pro/api/serviceHealthy.js`
  - Add config API helpers.
- Modify `frontend/src/pro/views/systemHealthOverview/metricVisuals.mjs`
  - Separate visual health-band threshold from abnormal-system threshold help text.
  - Allow summary card help text to include the active threshold.
- Modify `frontend/src/pro/views/systemHealthOverview/index.jsx`
  - Add threshold state and save handler.
  - Use threshold-only abnormal filtering.
  - Render compact threshold control next to "只看异常".
- Modify `frontend/src/pro/views/systemHealthOverview/index.css`
  - Add minimal sizing/alignment rules for the threshold control.

---

### Task 1: Backend Config Persistence And API

**Files:**
- Modify: `backend/pkg/model/response/system_health.go`
- Modify: `backend/pkg/model/request/system_health.go`
- Modify: `backend/pkg/repository/database/dao.go`
- Create: `backend/pkg/repository/database/dao_system_health_config.go`
- Modify: `backend/pkg/api/systemhealth/handler.go`
- Create: `backend/pkg/api/systemhealth/func_config.go`
- Modify: `backend/pkg/router/router_api.go`

- [ ] **Step 1: Add response and request models**

In `backend/pkg/model/response/system_health.go`, change `SystemHealthOverviewResp` to include `Config`:

```go
type SystemHealthOverviewResp struct {
	Summary       SystemHealthOverviewSummary `json:"summary"`
	Systems       []SystemHealthSystem        `json:"systems"`
	PendingIssues []SystemHealthPendingIssue  `json:"pendingIssues"`
	Config        SystemHealthConfigResp      `json:"config"`
	Window        string                      `json:"window"`
	GeneratedAt   string                      `json:"generatedAt"`
	Source        string                      `json:"source"`
}
```

Add this response type near the overview response types:

```go
type SystemHealthConfigResp struct {
	AbnormalScoreThreshold int `json:"abnormalScoreThreshold"`
}
```

In `backend/pkg/model/request/system_health.go`, add:

```go
type UpdateSystemHealthConfigRequest struct {
	AbnormalScoreThreshold int `json:"abnormalScoreThreshold" binding:"required,min=1,max=100"`
}
```

- [ ] **Step 2: Add database repo contract**

In `backend/pkg/repository/database/dao.go`, add these methods near the existing threshold methods in `Repo`:

```go
	GetOrCreateSystemHealthConfig(ctx core.Context) (SystemHealthConfig, error)
	UpdateSystemHealthConfig(ctx core.Context, abnormalScoreThreshold int) (SystemHealthConfig, error)
```

In the `AutoMigrate` list, add `&SystemHealthConfig{},` near `&Threshold{},`.

- [ ] **Step 3: Add config DAO implementation**

Create `backend/pkg/repository/database/dao_system_health_config.go`:

```go
// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"errors"
	"time"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"gorm.io/gorm"
)

const DefaultSystemHealthAbnormalScoreThreshold = 80

type SystemHealthConfig struct {
	ID                     uint      `gorm:"primaryKey;autoIncrement"`
	AbnormalScoreThreshold int       `gorm:"column:abnormal_score_threshold;not null;default:80"`
	UpdatedAt              time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (SystemHealthConfig) TableName() string {
	return "system_health_config"
}

func (repo *daoRepo) GetOrCreateSystemHealthConfig(ctx core.Context) (SystemHealthConfig, error) {
	var cfg SystemHealthConfig
	result := repo.GetContextDB(ctx).Order("id ASC").First(&cfg)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		cfg = SystemHealthConfig{AbnormalScoreThreshold: DefaultSystemHealthAbnormalScoreThreshold}
		return cfg, repo.GetContextDB(ctx).Create(&cfg).Error
	}
	if result.Error != nil {
		return SystemHealthConfig{}, result.Error
	}
	if cfg.AbnormalScoreThreshold < 1 || cfg.AbnormalScoreThreshold > 100 {
		cfg.AbnormalScoreThreshold = DefaultSystemHealthAbnormalScoreThreshold
	}
	return cfg, nil
}

func (repo *daoRepo) UpdateSystemHealthConfig(ctx core.Context, abnormalScoreThreshold int) (SystemHealthConfig, error) {
	cfg, err := repo.GetOrCreateSystemHealthConfig(ctx)
	if err != nil {
		return SystemHealthConfig{}, err
	}
	cfg.AbnormalScoreThreshold = abnormalScoreThreshold
	err = repo.GetContextDB(ctx).Model(&cfg).Updates(SystemHealthConfig{
		AbnormalScoreThreshold: abnormalScoreThreshold,
	}).Error
	if err != nil {
		return SystemHealthConfig{}, err
	}
	return cfg, nil
}
```

- [ ] **Step 4: Add config handlers**

In `backend/pkg/api/systemhealth/handler.go`, extend `Handler`:

```go
	GetConfig() core.HandlerFunc
	UpdateConfig() core.HandlerFunc
```

Create `backend/pkg/api/systemhealth/func_config.go`:

```go
// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package systemhealth

import (
	"net/http"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

func (h *handler) GetConfig() core.HandlerFunc {
	return func(c core.Context) {
		resp, err := h.systemHealthConfig(c)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.GetDataGroupError, err)
			return
		}
		c.Payload(resp)
	}
}

func (h *handler) UpdateConfig() core.HandlerFunc {
	return func(c core.Context) {
		req := new(request.UpdateSystemHealthConfigRequest)
		if err := c.ShouldBindJSON(req); err != nil {
			c.AbortWithError(http.StatusBadRequest, code.ParamBindError, err)
			return
		}
		cfg, err := h.dbRepo.UpdateSystemHealthConfig(c, req.AbnormalScoreThreshold)
		if err != nil {
			c.AbortWithError(http.StatusBadRequest, code.GetDataGroupError, err)
			return
		}
		c.Payload(systemHealthConfigResponse(cfg))
	}
}

func (h *handler) systemHealthConfig(ctx core.Context) (response.SystemHealthConfigResp, error) {
	cfg, err := h.dbRepo.GetOrCreateSystemHealthConfig(ctx)
	if err != nil {
		return response.SystemHealthConfigResp{
			AbnormalScoreThreshold: database.DefaultSystemHealthAbnormalScoreThreshold,
		}, err
	}
	return systemHealthConfigResponse(cfg), nil
}

func systemHealthConfigResponse(cfg database.SystemHealthConfig) response.SystemHealthConfigResp {
	threshold := normalizeSystemHealthAbnormalScoreThreshold(cfg.AbnormalScoreThreshold)
	return response.SystemHealthConfigResp{AbnormalScoreThreshold: threshold}
}

func normalizeSystemHealthAbnormalScoreThreshold(threshold int) int {
	if threshold < 1 || threshold > 100 {
		return database.DefaultSystemHealthAbnormalScoreThreshold
	}
	return threshold
}
```

- [ ] **Step 5: Register routes**

In `backend/pkg/router/router_api.go`, in the existing `systemHealthApi` group, add the config routes before `/overview/:groupId` so `/config` cannot be interpreted as `:groupId`:

```go
systemHealthApi.GET("/overview", systemHealthHandler.GetOverview())
systemHealthApi.GET("/new-abnormal-metrics", systemHealthHandler.GetNewAbnormalMetrics())
systemHealthApi.GET("/config", systemHealthHandler.GetConfig())
systemHealthApi.POST("/config", systemHealthHandler.UpdateConfig())
systemHealthApi.GET("/overview/:groupId", systemHealthHandler.GetSystemDetail())
```

- [ ] **Step 6: Run formatting for backend config files**

Run:

```bash
cd /Users/keting/workspace/product/commercial/apo/backend && gofmt -w pkg/model/response/system_health.go pkg/model/request/system_health.go pkg/repository/database/dao.go pkg/repository/database/dao_system_health_config.go pkg/api/systemhealth/handler.go pkg/api/systemhealth/func_config.go pkg/router/router_api.go
```

Expected: command exits 0.

---

### Task 2: Backend Summary Threshold Logic

**Files:**
- Modify: `backend/pkg/api/systemhealth/func_getoverview_test.go`
- Modify: `backend/pkg/api/systemhealth/func_getoverview.go`
- Modify: `backend/pkg/api/systemhealth/func_config.go`

- [ ] **Step 1: Write failing summary tests**

In `backend/pkg/api/systemhealth/func_getoverview_test.go`, replace `TestBuildSystemHealthOverviewSummaryCountsNewAbnormalMetrics` with:

```go
func TestBuildSystemHealthOverviewSummaryUsesScoreThresholdOnly(t *testing.T) {
	systems := []response.SystemHealthSystem{
		{
			GroupID: 101,
			Name:    "订单系统",
			Score:   79,
			Services: []response.SystemHealthService{
				{Name: "order-api", Score: 79},
			},
			Metrics: map[string]string{
				healthMetricErrorRate: healthStatusGreen,
			},
		},
		{
			GroupID: 102,
			Name:    "支付系统",
			Score:   80,
			Services: []response.SystemHealthService{
				{Name: "payment-api", Score: 80},
			},
			Metrics: map[string]string{
				healthMetricLatency: healthStatusRed,
			},
		},
		{
			GroupID: 103,
			Name:    "库存系统",
			Score:   95,
			Services: []response.SystemHealthService{
				{Name: "inventory-api", Score: 95},
			},
			Metrics: map[string]string{
				healthMetricErrorLog: healthStatusYellow,
			},
		},
		{
			GroupID: 104,
			Name:    "空系统",
			Score:   0,
			Metrics: map[string]string{
				healthMetricLatency: healthStatusRed,
			},
		},
	}

	summary := buildSystemHealthOverviewSummary(systems, 80)

	if summary.AbnormalSystems != 1 {
		t.Fatalf("expected only the below-threshold system to be abnormal, got %d", summary.AbnormalSystems)
	}
	if summary.AbnormalServices != 1 {
		t.Fatalf("expected only the below-threshold service to be abnormal, got %d", summary.AbnormalServices)
	}
	if summary.AbnormalMetrics != 3 {
		t.Fatalf("expected red/yellow metric counting to stay unchanged, got %d", summary.AbnormalMetrics)
	}
	if summary.SevereMetrics != 2 {
		t.Fatalf("expected red metric counting to stay unchanged, got %d", summary.SevereMetrics)
	}
	if summary.NoDataSystems != 1 {
		t.Fatalf("expected no-data system count to stay unchanged, got %d", summary.NoDataSystems)
	}
}

func TestBuildSystemHealthOverviewSummaryCountsNewAbnormalMetrics(t *testing.T) {
	current := []response.SystemHealthSystem{
		{
			GroupID: 101,
			Name:    "订单系统",
			Score:   72,
			Services: []response.SystemHealthService{
				{Name: "order-api"},
			},
			Metrics: map[string]string{
				healthMetricErrorRate: healthStatusRed,
				healthMetricLatency:   healthStatusYellow,
				healthMetricErrorLog:  healthStatusGreen,
			},
		},
		{
			GroupID: 102,
			Name:    "支付系统",
			Score:   92,
			Services: []response.SystemHealthService{
				{Name: "payment-api"},
			},
			Metrics: map[string]string{
				healthMetricErrorRate: healthStatusGreen,
				healthMetricLatency:   healthStatusYellow,
			},
		},
	}
	previous := []response.SystemHealthSystem{
		{
			GroupID: 101,
			Metrics: map[string]string{
				healthMetricErrorRate: healthStatusGreen,
				healthMetricLatency:   healthStatusYellow,
				healthMetricErrorLog:  healthStatusGreen,
			},
		},
		{
			GroupID: 102,
			Metrics: map[string]string{
				healthMetricErrorRate: healthStatusGreen,
				healthMetricLatency:   healthStatusGray,
			},
		},
	}

	summary := buildSystemHealthOverviewSummary(current, 80, previous)

	if summary.AbnormalSystems != 1 {
		t.Fatalf("expected one abnormal system by score only, got %d", summary.AbnormalSystems)
	}
	if summary.AbnormalMetrics != 3 {
		t.Fatalf("expected three abnormal metrics, got %d", summary.AbnormalMetrics)
	}
	if summary.SevereMetrics != 1 {
		t.Fatalf("expected one severe metric, got %d", summary.SevereMetrics)
	}
	if summary.NewAbnormalMetrics != 2 {
		t.Fatalf("expected two newly abnormal metrics, got %d", summary.NewAbnormalMetrics)
	}
}
```

Add this test near the summary tests:

```go
func TestNormalizeSystemHealthAbnormalScoreThresholdFallsBackToDefault(t *testing.T) {
	if got := normalizeSystemHealthAbnormalScoreThreshold(75); got != 75 {
		t.Fatalf("expected valid threshold to be kept, got %d", got)
	}
	if got := normalizeSystemHealthAbnormalScoreThreshold(0); got != 80 {
		t.Fatalf("expected low threshold to fall back to 80, got %d", got)
	}
	if got := normalizeSystemHealthAbnormalScoreThreshold(101); got != 80 {
		t.Fatalf("expected high threshold to fall back to 80, got %d", got)
	}
}
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```bash
cd /Users/keting/workspace/product/commercial/apo/backend && go test ./pkg/api/systemhealth -run "TestBuildSystemHealthOverviewSummaryUsesScoreThresholdOnly|TestBuildSystemHealthOverviewSummaryCountsNewAbnormalMetrics|TestNormalizeSystemHealthAbnormalScoreThresholdFallsBackToDefault" -count=1
```

Expected before implementation: FAIL because `buildSystemHealthOverviewSummary` still has the old signature and/or still counts yellow/red metric lamps as abnormal systems.

- [ ] **Step 3: Implement summary threshold logic**

In `backend/pkg/api/systemhealth/func_getoverview.go`, load config before summary:

```go
	configResp, err := h.systemHealthConfig(ctx)
	if err != nil {
		systemHealthWarn("system-health overview config failed, using default threshold", systemHealthStageFields(ctx, buildStartedAt, "04_config",
			zap.Error(err),
		)...)
	}
	abnormalThreshold := normalizeSystemHealthAbnormalScoreThreshold(configResp.AbnormalScoreThreshold)

	summary := buildSystemHealthOverviewSummary(systems, abnormalThreshold)
```

When constructing `response.SystemHealthOverviewResp`, add:

```go
		Config:        response.SystemHealthConfigResp{AbnormalScoreThreshold: abnormalThreshold},
```

Change summary helper signature and body:

```go
func buildSystemHealthOverviewSummary(systems []response.SystemHealthSystem, abnormalThreshold int, previousSystems ...[]response.SystemHealthSystem) response.SystemHealthOverviewSummary {
	abnormalThreshold = normalizeSystemHealthAbnormalScoreThreshold(abnormalThreshold)
	summary := response.SystemHealthOverviewSummary{TotalSystems: len(systems)}
	for _, system := range systems {
		summary.Services += len(system.Services)
		if len(system.Services) == 0 {
			summary.NoDataSystems++
		}
		if len(system.Services) > 0 {
			if system.Score < 60 {
				summary.LowScoreSystems++
			}
			if system.Score < abnormalThreshold {
				summary.AbnormalSystems++
			}
		}
		for _, service := range system.Services {
			if service.Score < abnormalThreshold {
				summary.AbnormalServices++
			}
		}
		for _, status := range system.Metrics {
			if isAbnormalMetricStatus(status) {
				summary.AbnormalMetrics++
			}
			if status == healthStatusRed {
				summary.SevereMetrics++
			}
		}
	}
	if len(previousSystems) > 0 {
		summary.NewAbnormalMetrics = countNewAbnormalMetrics(systems, previousSystems[0])
	}
	return summary
}
```

- [ ] **Step 4: Format and run targeted backend tests**

Run:

```bash
cd /Users/keting/workspace/product/commercial/apo/backend && gofmt -w pkg/api/systemhealth/func_getoverview.go pkg/api/systemhealth/func_getoverview_test.go pkg/api/systemhealth/func_config.go
cd /Users/keting/workspace/product/commercial/apo/backend && go test ./pkg/api/systemhealth -run "TestBuildSystemHealthOverviewSummaryUsesScoreThresholdOnly|TestBuildSystemHealthOverviewSummaryCountsNewAbnormalMetrics|TestNormalizeSystemHealthAbnormalScoreThresholdFallsBackToDefault" -count=1
```

Expected after implementation: PASS.

---

### Task 3: Frontend Threshold Control And Filtering

**Files:**
- Modify: `frontend/src/pro/api/serviceHealthy.js`
- Modify: `frontend/src/pro/views/systemHealthOverview/metricVisuals.mjs`
- Modify: `frontend/src/pro/views/systemHealthOverview/index.jsx`
- Modify: `frontend/src/pro/views/systemHealthOverview/index.css`

- [ ] **Step 1: Add frontend API helpers**

In `frontend/src/pro/api/serviceHealthy.js`, add:

```js
export const getSystemHealthConfigApi = () => {
  return get('/api/system-health/config')
}

export const setSystemHealthConfigApi = (data) => {
  return post('/api/system-health/config', data)
}
```

- [ ] **Step 2: Update metric visual helpers**

In `frontend/src/pro/views/systemHealthOverview/metricVisuals.mjs`, keep `HEALTH_SCORE_ABNORMAL_SYSTEM_THRESHOLD` for visual bands and add:

```js
export const DEFAULT_ABNORMAL_SCORE_THRESHOLD = 80
export const normalizeAbnormalScoreThreshold = (value) => {
  const number = Number(value)
  if (!Number.isFinite(number)) return DEFAULT_ABNORMAL_SCORE_THRESHOLD
  const threshold = Math.trunc(number)
  if (threshold < 1 || threshold > 100) return DEFAULT_ABNORMAL_SCORE_THRESHOLD
  return threshold
}
export const abnormalSystemsHelpText = (threshold = DEFAULT_ABNORMAL_SCORE_THRESHOLD) =>
  `当前阈值：低于 ${normalizeAbnormalScoreThreshold(threshold)} 分计入异常系统。`
```

Change `buildOverviewSummaryCards` signature:

```js
export const buildOverviewSummaryCards = (
  summary = {},
  systems = [],
  abnormalScoreThreshold = DEFAULT_ABNORMAL_SCORE_THRESHOLD,
) => {
```

Change the abnormal system card help to:

```js
help: abnormalSystemsHelpText(abnormalScoreThreshold),
```

- [ ] **Step 3: Update overview page imports and state**

In `frontend/src/pro/views/systemHealthOverview/index.jsx`, update imports:

```js
  InputNumber,
  message,
```

Add `Save` to lucide imports:

```js
import { ArrowLeft, CircleAlert, HelpCircle, Save, Search, Sparkles, TriangleAlert } from 'lucide-react'
```

Import config API and threshold helpers:

```js
  setSystemHealthConfigApi,
```

```js
  DEFAULT_ABNORMAL_SCORE_THRESHOLD,
  normalizeAbnormalScoreThreshold,
```

Replace `isAbnormal` with:

```js
const isAbnormal = (system, threshold = DEFAULT_ABNORMAL_SCORE_THRESHOLD) =>
  system.status !== 'gray' &&
  system.services?.length > 0 &&
  system.score < normalizeAbnormalScoreThreshold(threshold)
```

Add state near other overview state:

```js
  const [abnormalScoreThreshold, setAbnormalScoreThreshold] = useState(
    DEFAULT_ABNORMAL_SCORE_THRESHOLD,
  )
  const [thresholdDraft, setThresholdDraft] = useState(DEFAULT_ABNORMAL_SCORE_THRESHOLD)
  const [thresholdSaving, setThresholdSaving] = useState(false)
```

- [ ] **Step 4: Sync threshold from overview and save updates**

In `loadOverview`, after `nextSystems` is created and before `setSummary`, add:

```js
      const nextThreshold = normalizeAbnormalScoreThreshold(
        resp?.config?.abnormalScoreThreshold,
      )
      setAbnormalScoreThreshold(nextThreshold)
      setThresholdDraft(nextThreshold)
```

Add this callback after `handlePendingIssuePageChange`:

```js
  const saveAbnormalScoreThreshold = async () => {
    const nextThreshold = normalizeAbnormalScoreThreshold(thresholdDraft)
    setThresholdDraft(nextThreshold)
    if (nextThreshold === abnormalScoreThreshold) return

    setThresholdSaving(true)
    try {
      const resp = await setSystemHealthConfigApi({
        abnormalScoreThreshold: nextThreshold,
      })
      const savedThreshold = normalizeAbnormalScoreThreshold(
        resp?.abnormalScoreThreshold ?? nextThreshold,
      )
      setAbnormalScoreThreshold(savedThreshold)
      setThresholdDraft(savedThreshold)
      message.success('异常阈值已更新')
      await loadOverview()
    } catch (error) {
      message.error(error?.message || '异常阈值保存失败')
      setThresholdDraft(abnormalScoreThreshold)
    } finally {
      setThresholdSaving(false)
    }
  }
```

Change overview filtering dependency and predicate:

```js
        .filter((system) => !onlyAbnormal || isAbnormal(system, abnormalScoreThreshold))
```

and include `abnormalScoreThreshold` in the dependency array.

Change summary cards:

```js
const summaryCards = buildOverviewSummaryCards(summary, systems, abnormalScoreThreshold)
```

- [ ] **Step 5: Render compact threshold control**

In the system health panel header, place this `Space` before the existing "只看异常" control:

```jsx
                  <Space className="sho-threshold-control" size={4}>
                    <span className="sho-muted">异常阈值</span>
                    <span className="sho-muted">低于</span>
                    <InputNumber
                      max={100}
                      min={1}
                      precision={0}
                      size="small"
                      value={thresholdDraft}
                      onChange={(value) => setThresholdDraft(value)}
                      onPressEnter={saveAbnormalScoreThreshold}
                    />
                    <span className="sho-muted">分</span>
                    <Button
                      icon={<Save size={13} />}
                      loading={thresholdSaving}
                      size="small"
                      onClick={saveAbnormalScoreThreshold}
                    >
                      保存
                    </Button>
                  </Space>
```

- [ ] **Step 6: Add minimal CSS**

In `frontend/src/pro/views/systemHealthOverview/index.css`, near `.sho-search`, add:

```css
.sho-threshold-control {
  flex-wrap: nowrap;
}

.sho-threshold-control .ant-input-number {
  width: 68px;
}
```

In the `@media (max-width: 720px)` block, add:

```css
  .sho-threshold-control {
    width: 100%;
  }
```

- [ ] **Step 7: Run frontend lint**

Run:

```bash
cd /Users/keting/workspace/product/commercial/apo/frontend && npm run lint
```

Expected: command exits 0, or reports pre-existing lint errors unrelated to the changed system health files. If unrelated errors appear, run:

```bash
cd /Users/keting/workspace/product/commercial/apo/frontend && npx eslint src/pro/api/serviceHealthy.js src/pro/views/systemHealthOverview/index.jsx
```

Expected for changed JS files: command exits 0.

---

### Task 4: Full Verification And Commit

**Files:**
- Verify all files touched by Tasks 1-3.

- [ ] **Step 1: Run backend systemhealth package tests**

Run:

```bash
cd /Users/keting/workspace/product/commercial/apo/backend && go test ./pkg/api/systemhealth
```

Expected: PASS.

- [ ] **Step 2: Run frontend lint on changed JS files**

Run:

```bash
cd /Users/keting/workspace/product/commercial/apo/frontend && npx eslint src/pro/api/serviceHealthy.js src/pro/views/systemHealthOverview/index.jsx
```

Expected: PASS.

- [ ] **Step 3: Review diff boundaries**

Run:

```bash
cd /Users/keting/workspace/product/commercial/apo && git diff -- backend/pkg/model/response/system_health.go backend/pkg/model/request/system_health.go backend/pkg/repository/database/dao.go backend/pkg/repository/database/dao_system_health_config.go backend/pkg/api/systemhealth/handler.go backend/pkg/api/systemhealth/func_config.go backend/pkg/router/router_api.go backend/pkg/api/systemhealth/func_getoverview.go backend/pkg/api/systemhealth/func_getoverview_test.go frontend/src/pro/api/serviceHealthy.js frontend/src/pro/views/systemHealthOverview/metricVisuals.mjs frontend/src/pro/views/systemHealthOverview/index.jsx frontend/src/pro/views/systemHealthOverview/index.css
```

Expected:

- Backend overview response includes `config`.
- Backend summary uses `system.Score < abnormalThreshold` and `service.Score < abnormalThreshold`.
- Frontend "只看异常" uses score threshold only.
- Red/yellow metric lamp and pending issue logic is unchanged.

- [ ] **Step 4: Stage only implementation files and commit**

Run:

```bash
cd /Users/keting/workspace/product/commercial/apo && git add backend/pkg/model/response/system_health.go backend/pkg/model/request/system_health.go backend/pkg/repository/database/dao.go backend/pkg/repository/database/dao_system_health_config.go backend/pkg/api/systemhealth/handler.go backend/pkg/api/systemhealth/func_config.go backend/pkg/router/router_api.go backend/pkg/api/systemhealth/func_getoverview.go backend/pkg/api/systemhealth/func_getoverview_test.go frontend/src/pro/api/serviceHealthy.js frontend/src/pro/views/systemHealthOverview/metricVisuals.mjs frontend/src/pro/views/systemHealthOverview/index.jsx frontend/src/pro/views/systemHealthOverview/index.css docs/superpowers/plans/2026-06-26-system-health-abnormal-threshold.md
git commit -m "feat(systemhealth): configure abnormal score threshold"
```

Expected: commit succeeds and includes only the planned files.

---

## Self-Review

Spec coverage:

- Global shared threshold: Task 1 creates a single-row config table and API.
- Score-only abnormal judgment: Task 2 changes backend summary and Task 3 changes frontend filter.
- User configurable threshold: Task 3 adds a compact setting control.
- Backend/frontend consistency: Task 2 returns config in overview; Task 3 consumes it.
- Metrics and pending issues unchanged: Task 2 tests metric counts; Task 4 diff review checks pending issue logic.

Placeholder scan: This plan contains no open-ended implementation placeholders.

Type consistency:

- Request field is `abnormalScoreThreshold`.
- Response field is `abnormalScoreThreshold`.
- Backend response type is `SystemHealthConfigResp`.
- Frontend state is `abnormalScoreThreshold`.
