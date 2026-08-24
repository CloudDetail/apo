# System Health Trend Complete Buckets Design

## Problem

The system health trend selector says "过去15天", but the chart only renders dates returned by `/api/system-health/trend`. When earlier dates in the selected range have no snapshot rows, `buildSystemHealthTrendPoints` skips those intervals, so a 15 day selection can visually start at 6/24.

This is technically consistent with the current data aggregation, but it is misleading in the UI. Users expect the selected range to control the full x-axis.

## Decision

Trend APIs should preserve the complete requested time buckets and explicitly mark buckets that have no snapshots. The frontend should render those buckets on the x-axis without treating them as "无数据服务" service status.

## Backend Behavior

- `systemHealthTrendWindowFor` continues to define the requested range and granularity.
- `buildSystemHealthTrendPoints` returns one point per interval.
- Intervals with snapshot samples keep the existing aggregate behavior.
- Intervals without snapshot samples return a point with a missing-snapshot marker.
- Missing snapshots are not counted as abnormal, attention, healthy, or no-data services.

## Frontend Behavior

- Summary, system, and detail trend charts use all returned points for x-axis labels.
- Missing-snapshot points render as gaps or empty bars rather than zero-valued service status bars.
- Tooltips for missing-snapshot points say "当日无快照数据" or the time-granularity equivalent.
- Existing "暂无趋势快照" empty states remain for cases where the API returns no points at all.

## Testing

- Backend test: a multi-interval trend with an empty first interval returns both intervals, with the first marked as missing snapshot.
- Frontend trend row tests: missing-snapshot points preserve timestamps and are distinguishable from real zero counts.
- Frontend chart option tests: missing-snapshot service ratios render as gaps and tooltip text uses the missing-snapshot copy.

## Non-Goals

- Do not backfill historical snapshot data.
- Do not change snapshot collection cadence.
- Do not redefine service `gray` / "无数据" status.
- Do not make the global page time range drive the trend selector in this change.
