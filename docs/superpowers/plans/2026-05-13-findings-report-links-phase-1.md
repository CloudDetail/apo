# Findings Report Links Phase 1 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add reliable “related fault reports” links to Agent reasoning tree node details, with APO consuming structured links and AgentSRE producing only links backed by concrete trace/span evidence.

**Architecture:** AgentSRE remains the producer of report eligibility and returns canonical `reportLinks` on findings nodes. APO only normalizes, strictly validates, and renders links already supplied by AgentSRE; APO does not query report candidates or infer links from node status. Producer scope includes slow reports sourced from anomaly span data and verified error reports from current-node trace evidence or dependency RED error-rate checks constrained by downstream service and runtime window.

**Tech Stack:** React 18, TypeScript, Ant Design, lucide-react, i18next JSON resources, Node test runner through `tsx`, Python 3.12, pytest.

---

## Execution Status

As of 2026-05-15, the implementation has moved beyond the original slow-only producer slice. The checklist below is the original task-by-task implementation plan; current completion should be judged by the verified code, tests, and contract text:

- APO requires producer-supplied `url` values and filters incomplete or non-exact report routes.
- AgentSRE emits slow links from anomaly span data.
- AgentSRE may emit error links from concrete current-node trace evidence or dependency RED error-rate checks constrained by downstream service and runtime window, after OriginX `/originx/api/query/errortraces/{traceId}` validation succeeds.
- There is no production bypass for unverified error report links.

## Scope

This plan covers a first shippable slice:

- APO frontend accepts and renders structured `reportLinks` for `slow` and `error`.
- APO frontend filters invalid, external, incomplete, or non-exact report routes.
- AgentSRE emits structured slow-report links for abnormal evidence nodes where anomaly-span data provides `traceId + spanId`.
- AgentSRE emits structured error-report links only where the producer can tie the error `traceId` to current node evidence or a dependency RED error-rate signal in the same runtime window, and OriginX confirms the error report is available.

## File Structure

APO repo: `/Users/keting/workspace/product/commercial/apo`

- Create: `frontend/src/pro/api/findingsReportLinks.ts`
  - Owns `FindingsReportLink` type, URL validation, link normalization, and default ordering.
- Create: `frontend/src/pro/api/findingsReportLinks.test.ts`
  - Unit tests for strict route validation and field filtering.
- Modify: `frontend/src/pro/api/conversation.ts`
  - Adds `reportLinks` to `FindingsNodePayload` and preserves normalized links in `normalizeFindingsNode`.
- Modify: `frontend/src/pro/components/AGUIChat/components/FindingsTree/findingsContract.ts`
  - Adds `reportLinks` to `FindingsNodeData` and maps payload to selected-node data.
- Create: `frontend/src/pro/components/AGUIChat/components/FindingsTree/findingsContract.test.ts`
  - Confirms `toFindingsNodeData` preserves normalized links.
- Create: `frontend/src/pro/components/AGUIChat/components/EvidencePanel/RelatedReports.tsx`
  - Renders the report link section and opens links in a new tab.
- Modify: `frontend/src/pro/components/AGUIChat/components/EvidencePanel/NodeDetail.tsx`
  - Places `RelatedReports` after node conclusion/details and before metrics/table/log/trace evidence.
- Modify: `frontend/public/locales/zh/pro/agui.json`
  - Adds Chinese labels.
- Modify: `frontend/public/locales/en/pro/agui.json`
  - Adds English labels.

AgentSRE repo: `/Users/keting/workspace/gitlabProject/AgentSRE`

- Modify: `api/tools/apo/apo_dataplane.py`
  - Adds structured slow-report link generation while preserving existing string-link API.
- Modify: `tests/core/tools/test_apo_dataplane.py`
  - Tests structured slow-link generation and backward-compatible string links.
- Modify: `api/controllers/findings_data.py`
  - Copies `reportLinks` through builder projection and attaches report links only to eligible evidence nodes with valid evidence.
- Modify: `tests/core/test_findings_runtime_contract.py`
  - Tests that abnormal evidence nodes can carry slow-report links and unsupported nodes do not.
- Modify: `docs/openspec-backend/modules/investigation_findings/contracts.md`
  - Documents `reportLinks` as a canonical node extension.

---

### Task 1: APO Report Link Normalizer

**Files:**
- Create: `frontend/src/pro/api/findingsReportLinks.ts`
- Create: `frontend/src/pro/api/findingsReportLinks.test.ts`

- [ ] **Step 1: Write failing report-link validation tests**

Create `frontend/src/pro/api/findingsReportLinks.test.ts`:

```ts
import test from 'node:test'
import assert from 'node:assert/strict'

import { normalizeFindingsReportLinks } from './findingsReportLinks'

test('accepts exact slow report hash route with traceId spanId and mutatedType', () => {
  const links = normalizeFindingsReportLinks([
    {
      type: 'slow',
      title: 'CPU slow report',
      traceId: 'trace-1',
      spanId: 'span-1',
      mutatedType: 'cpu_time',
      url: '/#/cause/report/trace-1/span-1?mutatedType=cpu_time',
      source: 'anomaly-span-list',
    },
  ])

  assert.deepEqual(links, [
    {
      type: 'slow',
      title: 'CPU slow report',
      traceId: 'trace-1',
      spanId: 'span-1',
      mutatedType: 'cpu_time',
      url: '/#/cause/report/trace-1/span-1?mutatedType=cpu_time',
      source: 'anomaly-span-list',
    },
  ])
})

test('accepts exact error report hash route with traceId', () => {
  const links = normalizeFindingsReportLinks([
    {
      type: 'error',
      traceId: 'trace-2',
      url: '/#/cause/error-report/trace-2',
    },
  ])

  assert.deepEqual(links, [
    {
      type: 'error',
      traceId: 'trace-2',
      url: '/#/cause/error-report/trace-2',
    },
  ])
})

test('rejects slow report links without spanId', () => {
  const links = normalizeFindingsReportLinks([
    {
      type: 'slow',
      traceId: 'trace-1',
      url: '/#/cause/report/trace-1/',
    },
  ])

  assert.equal(links, undefined)
})

test('rejects external URLs and routes that only contain cause report text', () => {
  const links = normalizeFindingsReportLinks([
    {
      type: 'error',
      traceId: 'trace-1',
      url: 'https://example.com/#/cause/error-report/trace-1',
    },
    {
      type: 'slow',
      traceId: 'trace-1',
      spanId: 'span-1',
      url: '/#/foo/cause/report/trace-1/span-1',
    },
  ])

  assert.equal(links, undefined)
})

test('orders slow report links before error report links', () => {
  const links = normalizeFindingsReportLinks([
    {
      type: 'error',
      traceId: 'trace-2',
      url: '/#/cause/error-report/trace-2',
    },
    {
      type: 'slow',
      traceId: 'trace-1',
      spanId: 'span-1',
      url: '/#/cause/report/trace-1/span-1',
    },
  ])

  assert.deepEqual(
    links?.map((link) => link.type),
    ['slow', 'error'],
  )
})
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
cd /Users/keting/workspace/product/commercial/apo/frontend && npx --yes tsx --test src/pro/api/findingsReportLinks.test.ts
```

Expected: FAIL with module-not-found for `./findingsReportLinks`.

- [ ] **Step 3: Implement the normalizer**

Create `frontend/src/pro/api/findingsReportLinks.ts`:

```ts
export type FindingsReportLinkType = 'slow' | 'error'

export interface FindingsReportLink {
  type: FindingsReportLinkType
  title?: string
  traceId: string
  spanId?: string
  mutatedType?: string
  url: string
  source?: string
}

const readString = (value: unknown): string | undefined => {
  if (typeof value !== 'string') {
    return undefined
  }

  const trimmed = value.trim()
  return trimmed.length > 0 ? trimmed : undefined
}

const isObject = (value: unknown): value is Record<string, unknown> =>
  value !== null && typeof value === 'object' && !Array.isArray(value)

const parseHashRoute = (url: string) => {
  if (!url.startsWith('/#/')) {
    return null
  }

  const hashRoute = url.slice(2)
  const queryStart = hashRoute.indexOf('?')
  const path = queryStart >= 0 ? hashRoute.slice(0, queryStart) : hashRoute
  const query = queryStart >= 0 ? hashRoute.slice(queryStart + 1) : ''

  if (!path.startsWith('/')) {
    return null
  }

  return {
    path,
    params: new URLSearchParams(query),
  }
}

const isExactSlowReportUrl = ({
  traceId,
  spanId,
  mutatedType,
  url,
}: {
  traceId: string
  spanId: string
  mutatedType?: string
  url: string
}) => {
  const parsed = parseHashRoute(url)
  if (!parsed) {
    return false
  }

  const expectedPath = `/cause/report/${encodeURIComponent(traceId)}/${encodeURIComponent(spanId)}`
  if (parsed.path !== expectedPath) {
    return false
  }

  if (mutatedType) {
    return parsed.params.get('mutatedType') === mutatedType
  }

  return true
}

const isExactErrorReportUrl = ({ traceId, url }: { traceId: string; url: string }) => {
  const parsed = parseHashRoute(url)
  if (!parsed) {
    return false
  }

  return parsed.path === `/cause/error-report/${encodeURIComponent(traceId)}`
}

const normalizeOneFindingsReportLink = (value: unknown): FindingsReportLink | null => {
  if (!isObject(value)) {
    return null
  }

  const type = readString(value.type)
  const traceId = readString(value.traceId)
  const url = readString(value.url)
  const title = readString(value.title)
  const source = readString(value.source)

  if ((type !== 'slow' && type !== 'error') || !traceId || !url) {
    return null
  }

  if (type === 'slow') {
    const spanId = readString(value.spanId)
    const mutatedType = readString(value.mutatedType)

    if (!spanId || !isExactSlowReportUrl({ traceId, spanId, mutatedType, url })) {
      return null
    }

    return {
      type,
      ...(title ? { title } : {}),
      traceId,
      spanId,
      ...(mutatedType ? { mutatedType } : {}),
      url,
      ...(source ? { source } : {}),
    }
  }

  if (!isExactErrorReportUrl({ traceId, url })) {
    return null
  }

  return {
    type,
    ...(title ? { title } : {}),
    traceId,
    url,
    ...(source ? { source } : {}),
  }
}

export const normalizeFindingsReportLinks = (
  value: unknown,
): FindingsReportLink[] | undefined => {
  if (!Array.isArray(value)) {
    return undefined
  }

  const normalized = value
    .map((item) => normalizeOneFindingsReportLink(item))
    .filter((item): item is FindingsReportLink => item !== null)
    .sort((a, b) => {
      if (a.type === b.type) {
        return 0
      }

      return a.type === 'slow' ? -1 : 1
    })

  return normalized.length > 0 ? normalized : undefined
}
```

- [ ] **Step 4: Run test to verify it passes**

Run:

```bash
cd /Users/keting/workspace/product/commercial/apo/frontend && npx --yes tsx --test src/pro/api/findingsReportLinks.test.ts
```

Expected: PASS.

- [ ] **Step 5: Commit APO normalizer**

```bash
cd /Users/keting/workspace/product/commercial/apo
git add frontend/src/pro/api/findingsReportLinks.ts frontend/src/pro/api/findingsReportLinks.test.ts
git commit -m "feat: normalize findings report links"
```

---

### Task 2: APO Findings Contract Wiring

**Files:**
- Modify: `frontend/src/pro/api/conversation.ts`
- Modify: `frontend/src/pro/components/AGUIChat/components/FindingsTree/findingsContract.ts`
- Create: `frontend/src/pro/components/AGUIChat/components/FindingsTree/findingsContract.test.ts`

- [ ] **Step 1: Write failing data-mapping test**

Create `frontend/src/pro/components/AGUIChat/components/FindingsTree/findingsContract.test.ts`:

```ts
import test from 'node:test'
import assert from 'node:assert/strict'

import { toFindingsNodeData } from './findingsContract'
import type { FindingsNodePayload } from '../../../../api/conversation'

test('maps reportLinks from payload to selected node data', () => {
  const payload: FindingsNodePayload = {
    nodeId: 'code_cpu_check',
    title: 'Code CPU',
    nodeKind: 'evidence',
    status: 'abnormal',
    parentId: null,
    reportLinks: [
      {
        type: 'slow',
        traceId: 'trace-1',
        spanId: 'span-1',
        url: '/#/cause/report/trace-1/span-1',
      },
    ],
  }

  assert.deepEqual(toFindingsNodeData(payload).reportLinks, payload.reportLinks)
})
```

- [ ] **Step 2: Run test to verify it fails**

Run:

```bash
cd /Users/keting/workspace/product/commercial/apo/frontend && npx --yes tsx --test src/pro/components/AGUIChat/components/FindingsTree/findingsContract.test.ts
```

Expected: FAIL because `FindingsNodePayload` and `FindingsNodeData` do not include `reportLinks`.

- [ ] **Step 3: Add `reportLinks` to API payload normalization**

Modify `frontend/src/pro/api/conversation.ts`:

```ts
import {
  normalizeFindingsReportLinks,
  type FindingsReportLink,
} from './findingsReportLinks'
```

Add this field to `FindingsNodePayload`:

```ts
  reportLinks?: FindingsReportLink[]
```

Add this field to the object returned by `normalizeFindingsNode`:

```ts
    reportLinks: normalizeFindingsReportLinks(value.reportLinks),
```

- [ ] **Step 4: Add `reportLinks` to tree node data mapping**

Modify `frontend/src/pro/components/AGUIChat/components/FindingsTree/findingsContract.ts`:

```ts
import type {
  FindingsLogItem,
  FindingsMetrics,
  FindingsHydrationStatus,
  FindingsNodeKind,
  FindingsNodePayload,
  FindingsReportLink,
  FindingsStatus,
  FindingsTable,
  FindingsTraceData,
} from '../../../../api/conversation'
```

Add this field to `FindingsNodeData`:

```ts
  reportLinks?: FindingsReportLink[]
```

Add this field to `toFindingsNodeData`:

```ts
  reportLinks: node.reportLinks,
```

- [ ] **Step 5: Run data-mapping test**

Run:

```bash
cd /Users/keting/workspace/product/commercial/apo/frontend && npx --yes tsx --test src/pro/components/AGUIChat/components/FindingsTree/findingsContract.test.ts
```

Expected: PASS.

- [ ] **Step 6: Commit APO contract wiring**

```bash
cd /Users/keting/workspace/product/commercial/apo
git add frontend/src/pro/api/conversation.ts frontend/src/pro/components/AGUIChat/components/FindingsTree/findingsContract.ts frontend/src/pro/components/AGUIChat/components/FindingsTree/findingsContract.test.ts
git commit -m "feat: carry report links through findings contract"
```

---

### Task 3: APO Node Detail Rendering

**Files:**
- Create: `frontend/src/pro/components/AGUIChat/components/EvidencePanel/RelatedReports.tsx`
- Modify: `frontend/src/pro/components/AGUIChat/components/EvidencePanel/NodeDetail.tsx`
- Modify: `frontend/public/locales/zh/pro/agui.json`
- Modify: `frontend/public/locales/en/pro/agui.json`

- [ ] **Step 1: Create the related reports component**

Create `frontend/src/pro/components/AGUIChat/components/EvidencePanel/RelatedReports.tsx`:

```tsx
import React from 'react'
import { Button, Space, theme } from 'antd'
import { ExternalLink, FileSearch } from 'lucide-react'
import { useTranslation } from 'react-i18next'

import type { FindingsReportLink } from '../../../../api/conversation'

const { useToken } = theme

interface RelatedReportsProps {
  reportLinks?: FindingsReportLink[]
}

const getDefaultLabelKey = (type: FindingsReportLink['type']) =>
  type === 'slow' ? 'evidencePanel.viewSlowReport' : 'evidencePanel.viewErrorReport'

const RelatedReports: React.FC<RelatedReportsProps> = ({ reportLinks }) => {
  const { token } = useToken()
  const { t } = useTranslation('pro/agui')

  if (!reportLinks || reportLinks.length === 0) {
    return null
  }

  return (
    <div className="space-y-2">
      <div className="flex items-center gap-2">
        <FileSearch size={16} color={token.colorTextSecondary} />
        <span className="text-base font-medium" style={{ color: token.colorTextBase }}>
          {t('evidencePanel.relatedReports')}
        </span>
      </div>
      <Space wrap size={[8, 8]}>
        {reportLinks.map((link, index) => (
          <Button
            key={`${link.type}-${link.traceId}-${link.spanId ?? 'trace'}-${index}`}
            href={link.url}
            target="_blank"
            rel="noopener noreferrer"
            size="small"
            icon={<ExternalLink size={14} />}
          >
            {link.title || t(getDefaultLabelKey(link.type))}
          </Button>
        ))}
      </Space>
    </div>
  )
}

export default RelatedReports
```

- [ ] **Step 2: Render related reports in node details**

Modify `frontend/src/pro/components/AGUIChat/components/EvidencePanel/NodeDetail.tsx`:

```ts
import RelatedReports from './RelatedReports'
```

Insert this block after the `detailSections.content` block and before `detailSections.abnormalMetrics`:

```tsx
        <RelatedReports reportLinks={selectedNode.reportLinks} />
```

- [ ] **Step 3: Add Chinese i18n labels**

Modify `frontend/public/locales/zh/pro/agui.json` inside `evidencePanel`:

```json
    "relatedReports": "相关故障报告",
    "viewSlowReport": "查看慢调用报告",
    "viewErrorReport": "查看错误报告",
```

Place the keys near existing node detail labels such as `details`, `logs`, and `traces`.

- [ ] **Step 4: Add English i18n labels**

Modify `frontend/public/locales/en/pro/agui.json` inside `evidencePanel`:

```json
    "relatedReports": "Related reports",
    "viewSlowReport": "View slow report",
    "viewErrorReport": "View error report",
```

Place the keys near existing node detail labels such as `details`, `logs`, and `traces`.

- [ ] **Step 5: Run APO targeted tests**

Run:

```bash
cd /Users/keting/workspace/product/commercial/apo/frontend && npx --yes tsx --test src/pro/api/findingsReportLinks.test.ts src/pro/components/AGUIChat/components/FindingsTree/findingsContract.test.ts
```

Expected: PASS.

- [ ] **Step 6: Run APO build**

Run:

```bash
cd /Users/keting/workspace/product/commercial/apo/frontend && npm run build
```

Expected: PASS. If the build fails because of unrelated existing issues, record the exact failure and still keep the targeted tests as the feature verification evidence.

- [ ] **Step 7: Commit APO rendering**

```bash
cd /Users/keting/workspace/product/commercial/apo
git add frontend/src/pro/components/AGUIChat/components/EvidencePanel/RelatedReports.tsx frontend/src/pro/components/AGUIChat/components/EvidencePanel/NodeDetail.tsx frontend/public/locales/zh/pro/agui.json frontend/public/locales/en/pro/agui.json
git commit -m "feat: show related reports in findings node details"
```

---

### Task 4: AgentSRE Structured Slow Report Links

**Files:**
- Modify: `/Users/keting/workspace/gitlabProject/AgentSRE/api/tools/apo/apo_dataplane.py`
- Modify: `/Users/keting/workspace/gitlabProject/AgentSRE/tests/core/tools/test_apo_dataplane.py`

- [ ] **Step 1: Write failing dataplane tests**

Append these tests to `/Users/keting/workspace/gitlabProject/AgentSRE/tests/core/tools/test_apo_dataplane.py`:

```py
def test_get_slow_report_links_returns_structured_report_links(monkeypatch):
    payload = {
        "list": [
            {"traceId": "trace-1", "spanId": "span-1"},
            {"traceId": "trace-2", "spanId": "span-2"},
        ]
    }

    def _post(url, json=None, **kwargs):
        assert url.endswith("/api/alerts/anomaly-span/list")
        assert json["reason"] == "cpu_time"
        return _FakeResponse(payload)

    monkeypatch.setattr(apo_dataplane.requests, "post", _post)

    links = apo_dataplane.TracesDataFetcher.get_slow_report_links(
        1,
        2,
        "svc-a",
        "cpu_time",
    )

    assert links == [
        {
            "type": "slow",
            "title": "CPU slow report",
            "traceId": "trace-1",
            "spanId": "span-1",
            "mutatedType": "cpu_time",
            "url": "/#/cause/report/trace-1/span-1?mutatedType=cpu_time",
            "source": "anomaly-span-list",
        },
        {
            "type": "slow",
            "title": "CPU slow report",
            "traceId": "trace-2",
            "spanId": "span-2",
            "mutatedType": "cpu_time",
            "url": "/#/cause/report/trace-2/span-2?mutatedType=cpu_time",
            "source": "anomaly-span-list",
        },
    ]


def test_get_report_list_keeps_existing_string_link_shape(monkeypatch):
    payload = {"list": [{"traceId": "trace-1", "spanId": "span-1"}]}

    def _post(*_args, **_kwargs):
        return _FakeResponse(payload)

    monkeypatch.setattr(apo_dataplane.requests, "post", _post)

    assert apo_dataplane.TracesDataFetcher.get_report_list(
        1,
        2,
        "svc-a",
        "network_time",
    ) == ["/#/cause/report/trace-1/span-1?mutatedType=network_time"]
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
cd /Users/keting/workspace/gitlabProject/AgentSRE && .venv/bin/python -m pytest tests/core/tools/test_apo_dataplane.py::test_get_slow_report_links_returns_structured_report_links tests/core/tools/test_apo_dataplane.py::test_get_report_list_keeps_existing_string_link_shape -q
```

Expected: FAIL because `TracesDataFetcher.get_slow_report_links` does not exist.

- [ ] **Step 3: Implement structured link helper**

Modify `/Users/keting/workspace/gitlabProject/AgentSRE/api/tools/apo/apo_dataplane.py`:

```py
from urllib.parse import quote, urlencode
```

Add helpers near `_get_report`:

```py
SLOW_REPORT_TITLES = {
    "cpu_time": "CPU slow report",
    "network_time": "Network slow report",
    "scheduling_time": "Scheduling slow report",
}


def _build_slow_report_link(trace_id: str, span_id: str, reason: str) -> dict | None:
    trace_id = str(trace_id or "").strip()
    span_id = str(span_id or "").strip()
    reason = str(reason or "").strip()

    if not trace_id or not span_id or not reason:
        return None

    query = urlencode({"mutatedType": reason})
    url = (
        f"/#/cause/report/{quote(trace_id, safe='')}/"
        f"{quote(span_id, safe='')}?{query}"
    )

    return {
        "type": "slow",
        "title": SLOW_REPORT_TITLES.get(reason, "Slow report"),
        "traceId": trace_id,
        "spanId": span_id,
        "mutatedType": reason,
        "url": url,
        "source": "anomaly-span-list",
    }


def _get_slow_report_links(params: dict, reason: str) -> list:
    url = APO_BACKEND + "/api/alerts/anomaly-span/list"
    resp = requests.post(url, json=params, timeout=BACKEND_QUERY_TIMEOUT_SECONDS)
    res_list = []

    if resp.status_code == 200:
        for item in resp.json()["list"] or []:
            link = _build_slow_report_link(item.get("traceId"), item.get("spanId"), reason)
            if link:
                res_list.append(link)
            if len(res_list) == 3:
                break
    else:
        print(f"Get Traces Report Error: {resp.text}")

    return res_list
```

Change existing `_get_report` to preserve the legacy string-list API:

```py
def _get_report(params: dict, reason: str) -> list:
    return [link["url"] for link in _get_slow_report_links(params, reason)]
```

Add this method to `TracesDataFetcher`:

```py
    @staticmethod
    def get_slow_report_links(start_time, end_time, service, reason) -> list:
        params = {
            "currentPage": 1,
            "contentKey": "",
            "endTime": end_time,
            "pageSize": 10,
            "reason": reason,
            "service": service,
            "startTime": start_time,
        }
        return _get_slow_report_links(params, reason)
```

- [ ] **Step 4: Run dataplane tests**

Run:

```bash
cd /Users/keting/workspace/gitlabProject/AgentSRE && .venv/bin/python -m pytest tests/core/tools/test_apo_dataplane.py::test_get_slow_report_links_returns_structured_report_links tests/core/tools/test_apo_dataplane.py::test_get_report_list_keeps_existing_string_link_shape -q
```

Expected: PASS.

- [ ] **Step 5: Commit AgentSRE dataplane helper**

```bash
cd /Users/keting/workspace/gitlabProject/AgentSRE
git add api/tools/apo/apo_dataplane.py tests/core/tools/test_apo_dataplane.py
git commit -m "feat: return structured slow report links"
```

---

### Task 5: AgentSRE Findings Runtime Report Links

**Files:**
- Modify: `/Users/keting/workspace/gitlabProject/AgentSRE/api/controllers/findings_data.py`
- Modify: `/Users/keting/workspace/gitlabProject/AgentSRE/tests/core/test_findings_runtime_contract.py`

- [ ] **Step 1: Write failing runtime contract tests**

Modify `_install_metric_modules` in `/Users/keting/workspace/gitlabProject/AgentSRE/tests/core/test_findings_runtime_contract.py` to accept a report-link map:

```py
def _install_metric_modules(
    *,
    topology=None,
    instances_by_service=None,
    fallback_instances_by_service=None,
    service_list=None,
    metric_payloads=None,
    slow_report_links_by_reason=None,
    code_lock_metric="Thread Polaris Metrics - 北极星指标（线程） - 各类型耗时折线图 - Futex",
):
```

Inside the helper, add:

```py
    slow_report_links_by_reason = slow_report_links_by_reason or {}
```

Inside `_FakeTracesDataFetcher`, add:

```py
        @staticmethod
        def get_slow_report_links(_start_time, _end_time, _service, reason):
            return slow_report_links_by_reason.get(reason, [])
```

Append this test:

```py
def test_code_cpu_check_attaches_slow_report_links_to_abnormal_evidence(monkeypatch):
    original_nodes = [
        {
            "id": "code_cpu_check",
            "title": "代码CPU检查",
            "status": "normal",
            "statusColor": "normal",
            "summaryConclusion": "检查代码CPU占用",
            "parentId": None,
        }
    ]
    oncpu_metric = "Thread Polaris Metrics - 北极星指标（线程） - 各类型耗时折线图 - OnCPU"
    metric_payloads = {
        (
            oncpu_metric,
            (("namespace", ".*"), ("pod", "svc-0")),
        ): _timeseries_payload(unit="ns", series_name="oncpu")
    }
    _install_metric_modules(
        instances_by_service={"svc": []},
        fallback_instances_by_service={
            "svc": [{"podName": "svc-0", "namespace": "default"}]
        },
        metric_payloads=metric_payloads,
        slow_report_links_by_reason={
            "cpu_time": [
                {
                    "type": "slow",
                    "title": "CPU slow report",
                    "traceId": "trace-1",
                    "spanId": "span-1",
                    "mutatedType": "cpu_time",
                    "url": "/#/cause/report/trace-1/span-1?mutatedType=cpu_time",
                    "source": "anomaly-span-list",
                }
            ]
        },
    )

    monkeypatch.setattr(fd, "_create_service_info", _service_info)
    monkeypatch.setattr(fd, "_create_node_structure", lambda: original_nodes)
    monkeypatch.setattr(
        fd,
        "_detect_metric_anomalies",
        lambda *_args, **_kwargs: ("cpu spike", True),
    )

    payload = fd.build_findings_payload("svc")
    node = next(item for item in payload["nodes"] if item["nodeId"] == "code_cpu_check")

    assert node["nodeKind"] == "evidence"
    assert node["status"] == "abnormal"
    assert node["reportLinks"] == [
        {
            "type": "slow",
            "title": "CPU slow report",
            "traceId": "trace-1",
            "spanId": "span-1",
            "mutatedType": "cpu_time",
            "url": "/#/cause/report/trace-1/span-1?mutatedType=cpu_time",
            "source": "anomaly-span-list",
        }
    ]
```

Append this test for the unsupported boundary:

```py
def test_slow_report_links_are_not_added_to_unsupported_nodes(monkeypatch):
    original_nodes = [
        {
            "id": "code_cpu_check",
            "title": "代码CPU检查",
            "status": "normal",
            "statusColor": "normal",
            "summaryConclusion": "检查代码CPU占用",
            "parentId": None,
        }
    ]
    _install_metric_modules(
        instances_by_service={"svc": []},
        fallback_instances_by_service={"svc": []},
        slow_report_links_by_reason={
            "cpu_time": [
                {
                    "type": "slow",
                    "traceId": "trace-1",
                    "spanId": "span-1",
                    "mutatedType": "cpu_time",
                    "url": "/#/cause/report/trace-1/span-1?mutatedType=cpu_time",
                    "source": "anomaly-span-list",
                }
            ]
        },
    )

    monkeypatch.setattr(fd, "_create_service_info", _service_info)
    monkeypatch.setattr(fd, "_create_node_structure", lambda: original_nodes)

    payload = fd.build_findings_payload("svc")
    node = next(item for item in payload["nodes"] if item["nodeId"] == "code_cpu_check")

    assert node["nodeKind"] == "unsupported"
    assert "reportLinks" not in node
```

- [ ] **Step 2: Run tests to verify they fail**

Run:

```bash
cd /Users/keting/workspace/gitlabProject/AgentSRE && .venv/bin/python -m pytest tests/core/test_findings_runtime_contract.py::test_code_cpu_check_attaches_slow_report_links_to_abnormal_evidence tests/core/test_findings_runtime_contract.py::test_slow_report_links_are_not_added_to_unsupported_nodes -q
```

Expected: FAIL because findings runtime does not attach `reportLinks`.

- [ ] **Step 3: Add runtime helper and projection support**

Modify `/Users/keting/workspace/gitlabProject/AgentSRE/api/controllers/findings_data.py`.

Add `reportLinks` to `_apply_builder_projection` copied keys:

```py
        "reportLinks",
```

Add these helpers near `_node_has_valid_evidence`:

```py
SLOW_REPORT_REASON_BY_NODE_ID = {
    "code_cpu_check": "cpu_time",
    "ebpf_check_cpu": "cpu_time",
    "ebpf_check_network": "network_time",
    "code_lock_check": "scheduling_time",
}


def _attach_slow_report_links_if_available(
    node: Dict[str, Any],
    node_id: str,
) -> Dict[str, Any]:
    reason = SLOW_REPORT_REASON_BY_NODE_ID.get(node_id)
    if not reason:
        return node

    if _normalize_runtime_status(node.get("status")) == "normal":
        node.pop("reportLinks", None)
        return node

    if _has_meaningful_value(node.get("unsupportedReason")):
        node.pop("reportLinks", None)
        return node

    if not _node_has_valid_evidence(node):
        node.pop("reportLinks", None)
        return node

    try:
        from api.tools.apo.apo_dataplane import TracesDataFetcher

        service = _current_runtime_service_name()
        start_ts, end_ts = _default_runtime_window_us()
        links = TracesDataFetcher.get_slow_report_links(
            start_ts,
            end_ts,
            service,
            reason,
        )
    except Exception as exc:
        print(f"Failed to get slow report links for {node_id}: {exc}")
        links = []

    if links:
        node["reportLinks"] = links
    else:
        node.pop("reportLinks", None)

    return node
```

Add this call at the end of each phase-1 slow-report builder after metrics/status are set and before returning:

```py
    _attach_slow_report_links_if_available(node, "code_cpu_check")
```

For builders that return a separate `merged` object, call before `return merged`:

```py
            _attach_slow_report_links_if_available(merged, "ebpf_check_cpu")
            return merged
```

Apply the helper to these nodes in phase 1:

- `build_code_cpu_check`
- `build_code_lock_check`
- `build_ebpf_check_cpu`
- `build_ebpf_check_network`

- [ ] **Step 4: Prevent finalized unsupported nodes from carrying links**

In `_finalize_runtime_nodes`, when the node falls into the `unsupported` branch, add:

```py
            node.pop("reportLinks", None)
```

This keeps the requirement explicit: unsupported nodes do not expose report links.

- [ ] **Step 5: Run runtime contract tests**

Run:

```bash
cd /Users/keting/workspace/gitlabProject/AgentSRE && .venv/bin/python -m pytest tests/core/test_findings_runtime_contract.py::test_code_cpu_check_attaches_slow_report_links_to_abnormal_evidence tests/core/test_findings_runtime_contract.py::test_slow_report_links_are_not_added_to_unsupported_nodes -q
```

Expected: PASS.

- [ ] **Step 6: Run AgentSRE targeted test suite**

Run:

```bash
cd /Users/keting/workspace/gitlabProject/AgentSRE && .venv/bin/python -m pytest tests/core/tools/test_apo_dataplane.py tests/core/test_findings_runtime_contract.py -q
```

Expected: PASS.

- [ ] **Step 7: Commit AgentSRE runtime links**

```bash
cd /Users/keting/workspace/gitlabProject/AgentSRE
git add api/controllers/findings_data.py tests/core/test_findings_runtime_contract.py
git commit -m "feat: attach slow report links to findings nodes"
```

---

### Task 6: Contract Documentation And Verification

**Files:**
- Modify: `/Users/keting/workspace/gitlabProject/AgentSRE/docs/openspec-backend/modules/investigation_findings/contracts.md`
- Verify: `/Users/keting/workspace/product/commercial/apo/docs/requirements/2026-05-13-findings-report-links-requirements.md`

- [ ] **Step 1: Document AgentSRE canonical extension**

Modify `/Users/keting/workspace/gitlabProject/AgentSRE/docs/openspec-backend/modules/investigation_findings/contracts.md`.

Add this field to the canonical node shape table:

```md
| `reportLinks` | no | Structured links to APO fault report pages. Only valid on `evidence` nodes with concrete trace/span evidence. |
```

Add this section after the evidence-node semantics:

````md
### `reportLinks`

`reportLinks` is an optional canonical extension on findings nodes.

Allowed shapes:

```ts
reportLinks?: Array<{
  type: 'slow' | 'error'
  title?: string
  traceId: string
  spanId?: string
  mutatedType?: string
  url: string
  source?: string
}>
```

Producer rules:

- `slow` links require `traceId`, `spanId`, and an APO relative hash route matching `/#/cause/report/{traceId}/{spanId}`.
- `error` links require `traceId`, an APO relative hash route matching `/#/cause/error-report/{traceId}`, and successful OriginX validation before emission.
- Producer support includes `slow` links sourced from anomaly span data and verified `error` links from current-node trace evidence or dependency RED error-rate checks constrained by downstream service and runtime window.
- For dependency RED `error` links, the producer returns every validated link in the constrained result set, with `serviceName`, `errorGroupKey`, `representative`, and `groupCount` so APO can group links without hiding any of them.
- `unsupported` nodes must not include `reportLinks`.
- A link must refer to a trace/span that is the source of, or strongly tied to, the current node evidence.
- Markdown `ReportLink` text is not a canonical contract.
````

- [ ] **Step 2: Verify APO targeted tests**

Run:

```bash
cd /Users/keting/workspace/product/commercial/apo/frontend && npx --yes tsx --test src/pro/api/findingsReportLinks.test.ts src/pro/components/AGUIChat/components/FindingsTree/findingsContract.test.ts
```

Expected: PASS.

- [ ] **Step 3: Verify APO build**

Run:

```bash
cd /Users/keting/workspace/product/commercial/apo/frontend && npm run build
```

Expected: PASS. If unrelated existing build failures occur, record the command, full failure summary, and confirm the targeted tests still pass.

- [ ] **Step 4: Verify AgentSRE targeted tests**

Run:

```bash
cd /Users/keting/workspace/gitlabProject/AgentSRE && .venv/bin/python -m pytest tests/core/tools/test_apo_dataplane.py tests/core/test_findings_runtime_contract.py -q
```

Expected: PASS.

- [ ] **Step 5: Review diffs for scope control**

Run:

```bash
cd /Users/keting/workspace/product/commercial/apo && git diff --stat
cd /Users/keting/workspace/gitlabProject/AgentSRE && git diff --stat
```

Expected:

- APO diff only touches report-link normalizer/tests, findings contract mapping, node detail rendering, i18n JSON, and this plan.
- AgentSRE diff only touches structured slow links, findings runtime propagation, tests, and contract docs.

- [ ] **Step 6: Commit documentation**

```bash
cd /Users/keting/workspace/gitlabProject/AgentSRE
git add docs/openspec-backend/modules/investigation_findings/contracts.md
git commit -m "docs: document findings report links contract"
```

---

## Further Error-Link Expansion Boundary

Broad error-report production remains outside this slice. Start a separate implementation plan before adding new error-report producers beyond the currently verified DB trace and dependency RED paths.

Expansion minimum acceptance:

- AgentSRE returns `type: "error"` only for nodes whose evidence is tied to a concrete error trace.
- AgentSRE validates `/originx/api/query/errortraces/{traceId}` successfully before emitting the link.
- Nodes with only error-rate metrics, logs aggregation, service name, endpoint, or time range still return no error report link.
