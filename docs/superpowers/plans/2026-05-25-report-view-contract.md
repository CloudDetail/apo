# Report View Contract Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a non-breaking `reportView` contract for `/report` and make APO consume it without deriving UI conclusions from existing report text.

**Architecture:** Dify/Agent adds a new top-level `reportView` output while existing report fields remain unchanged. APO stores `reportView` in an internal hidden overview tag to avoid ClickHouse schema migration, restores it into the snapshot response, and lets the frontend render `reportView` with root-cause metric filtering.

**Tech Stack:** Go backend, ClickHouse DAO, React/TypeScript frontend, node:test frontend unit tests, Go unit tests.

---

### Task 1: Backend reportView persistence and response contract

**Files:**
- Modify: `backend/pkg/model/alert_report.go`
- Modify: `backend/pkg/model/request/alert_event.go`
- Modify: `backend/pkg/repository/clickhouse/dao_alert_event_report.go`
- Modify: `backend/pkg/model/response/alert_event.go`
- Test: `backend/pkg/model/response/alert_event_report_view_test.go`

- [ ] **Step 1: Write failing Go tests**

Add tests that:

- `AddAlertEventReportRequest.ReportView` is stored in `overview_tags["__report_view"]`.
- `ToAlertReportResponse` returns `reportData.reportView`.
- public `overview.tags` does not expose `__report_view`.
- `ToAlertReportViewResponse` no longer builds derived `headline`, `metricDeltas`, `businessImpact`, `faultLocation`, or `timeline` from existing text.

- [ ] **Step 2: Verify tests fail**

Run:

```bash
cd backend && go test ./pkg/model/response ./pkg/repository/clickhouse
```

Expected: tests referencing `reportView` fail before implementation.

- [ ] **Step 3: Implement minimal backend changes**

Implement:

- `const ReportViewTagKey = "__report_view"`.
- `AddAlertEventReportRequest.ReportView map[string]any`.
- helper to serialize `ReportView` into overview tags on insert.
- helper to deserialize hidden tag into `ReportData.ReportView`.
- public tag filtering for `ReportOverviewHeadlineTagKey` and `ReportViewTagKey`.
- a minimal `GetAlertEventReportViewResponse` that returns `snapshot` and current status only when the service applies it, not derived UI fields.

- [ ] **Step 4: Run backend tests**

Run:

```bash
cd backend && go test ./pkg/model/response ./pkg/repository/clickhouse ./pkg/model/integration/alert
```

Expected: selected packages pass.

### Task 2: Dify workflow input contract

**Files:**
- Modify: `backend/pkg/model/integration/alert/alert_event_processed.go`
- Modify: `backend/pkg/repository/dify/workflow_worker.go`
- Modify: `backend/pkg/services/alerts/service_alert_event_list.go`
- Modify: `backend/pkg/services/alerts/service_alert_event_detail.go`
- Test: `backend/pkg/model/integration/alert/alert_event_processed_test.go`
- Test: `backend/pkg/repository/dify/workflow_worker_test.go`

- [ ] **Step 1: Write failing tests**

Update tests so `reportOutputContract` describes the new top-level `reportView` output and explicitly says existing output fields must not be changed.

- [ ] **Step 2: Verify tests fail**

Run:

```bash
cd backend && go test ./pkg/model/integration/alert ./pkg/repository/dify
```

Expected: tests fail until the new contract text is present.

- [ ] **Step 3: Implement contract changes**

Add `ReportView` to `AlertReportOutputContract` and make `AttachReportOutputContract` describe:

- `reportView` is additive.
- `reportView.rootCauseMetricRefs` must reference real `evidence` indexes.
- no existing node output or existing field structure should be changed.
- business impact should be omitted/null when not meaningful.
- action titles must be concise agent-written titles.

- [ ] **Step 4: Run Dify contract tests**

Run:

```bash
cd backend && go test ./pkg/model/integration/alert ./pkg/repository/dify
```

Expected: tests pass.

### Task 3: Frontend report view model

**Files:**
- Modify/Create: `frontend/src/pro/views/report/reportViewModel.ts`
- Modify: `frontend/src/pro/views/report/reportViewModel.test.ts`

- [ ] **Step 1: Write failing frontend tests**

Add tests that:

- first screen does not require report ID, generation time, or analysis window.
- headline and conclusion summary prefer `reportView`.
- root-cause metrics for `cpu` include CPU/error-log/retry metrics and exclude RED.
- root-cause metrics for `downstream` may include RED.
- no matching root-cause metrics returns an empty metric list and UI fallback copy.
- business impact is hidden when `reportView.businessImpact` is null or empty.
- action title uses `reportView.actionTitles[].title`.
- raw evidence conclusion uses `evidence.description`.

- [ ] **Step 2: Verify tests fail**

Run:

```bash
cd frontend && node --test src/pro/views/report/reportViewModel.test.ts
```

Expected: new tests fail before implementation.

- [ ] **Step 3: Implement minimal view model**

Implement a focused model that:

- preserves existing snapshot fields.
- consumes `reportData.reportView`.
- selects support metrics by the root-cause mapping in the requirement document.
- never uses RED for CPU/code/log/retry/net/storage/runtime/infra root causes.
- returns empty business impact for null/empty impact.
- maps concise action titles by `suggestIndex`.

- [ ] **Step 4: Run frontend view model tests**

Run:

```bash
cd frontend && node --test src/pro/views/report/reportViewModel.test.ts
```

Expected: tests pass.

### Task 4: Frontend UI cleanup

**Files:**
- Modify: `frontend/src/pro/views/report/index.tsx`
- Modify: `frontend/src/pro/views/report/index.module.scss`
- Modify: `frontend/src/pro/views/report/components/SupportingEvidence/index.tsx`
- Modify: `frontend/src/pro/views/report/components/AnalysisProcess/index.tsx`

- [ ] **Step 1: Write failing structural tests**

Use source-level tests in `reportViewModel.test.ts` to assert:

- source does not render report ID, generated time, or analysis window.
- source includes “支撑指标” and “未返回根因证据指标”.
- source hides business impact when model has no impact groups.
- raw evidence section renders evidence descriptions.

- [ ] **Step 2: Verify tests fail**

Run:

```bash
cd frontend && node --test src/pro/views/report/reportViewModel.test.ts
```

Expected: source-level tests fail until UI is adjusted.

- [ ] **Step 3: Implement UI cleanup**

Remove meta header display, render a simple first screen, show root-cause metric cards only when present, hide business impact when empty, keep raw evidence inline with description/conclusion text.

- [ ] **Step 4: Run frontend tests**

Run:

```bash
cd frontend && node --test src/pro/views/report/reportViewModel.test.ts
```

Expected: tests pass.

### Task 5: Final verification

**Files:**
- All files above.

- [ ] **Step 1: Run focused backend tests**

```bash
cd backend && go test ./pkg/model/response ./pkg/model/integration/alert ./pkg/repository/dify
```

- [ ] **Step 2: Run focused frontend tests**

```bash
cd frontend && node --test src/pro/views/report/reportViewModel.test.ts
```

- [ ] **Step 3: Inspect git diff**

```bash
git diff --stat
git diff -- docs/requirements/2026-05-25-report-view-contract.md docs/superpowers/plans/2026-05-25-report-view-contract.md
```

Expected: no unrelated push or commit actions; changes are local only.
