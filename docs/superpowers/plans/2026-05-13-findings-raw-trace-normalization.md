# Findings Raw Trace Normalization

Date: 2026-05-13

## Context

The `/ai/api/v1/conversations/findings-data/stream` response for:

- `threadId=thread_1778645246829_6h1zqb2zf`
- `serviceName=ts-price-service`
- node `db_traces_check`

returned `node.traces` as raw trace rows:

```json
{
  "serviceName": "ts-order-service",
  "operationName": "SELECT Order",
  "duration": 2021,
  "isError": false
}
```

The frontend `NodeTraces` component expects canonical grouped trace data:

```json
{
  "service": "ts-order-service",
  "calls": [
    {
      "target": "SELECT Order",
      "type": "database",
      "count": 1,
      "avgDuration": 2021,
      "maxDuration": 2021,
      "errorCount": 0
    }
  ]
}
```

## Problem

Because the stream `node_patch` path merged `event.node` directly into UI state, raw trace rows bypassed the findings node normalizer. `NodeTraces` then read `trace.service` and `trace.calls` from rows that only had `serviceName`, `operationName`, `duration`, and `isError`, producing empty service pills with only the trace icon visible.

## Frontend Patch

- Extract findings node normalization into `frontend/src/pro/api/findingsDataNormalizer.ts`.
- Keep `frontend/src/pro/api/conversation.ts` as the API module and re-export the normalizer for existing consumers.
- Normalize raw trace rows by grouping on `serviceName + operationName`.
- Convert SQL-like operations into `type: "database"` calls.
- Make stream `node_patch` use `normalizeFindingsNode()` before merging nodes into the evidence tree.
- Add regression tests for raw trace aggregation and stream node patch normalization.

## Contract Follow-Up

The canonical producer contract should still prefer returning grouped `service + calls` trace data directly. The frontend compatibility layer is a rollout guard for raw trace rows, not a replacement for the canonical findings evidence contract.

## UI Follow-Up

- Compact the `NodeTraces` rendering for the narrow evidence detail panel.
- Replace the nested trace card + collapse layout with a single-level service group and slim call rows.
- Keep trace metrics in a two-column grid so labels such as `调用` and `平均` do not wrap vertically.
- Suppress operation detail rows when a normalized raw trace only contains one operation whose name is already the call target.
