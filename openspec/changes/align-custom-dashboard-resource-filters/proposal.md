## Why

Custom dashboard filters currently mix two concepts: APO's top-level data group selector and dashboard-local variables. This creates duplicated data group controls, inconsistent candidate sources, and a custom variable endpoint that reimplements resource options already provided by the system.

## What Changes

- Treat the APO top-level data group selector as the only data group selection for custom dashboard viewing, editing previews, variable candidates, and backend query scope enforcement.
- Stop offering, rendering, saving, or writing `dataGroup` as a configurable custom dashboard variable while remaining tolerant of legacy dashboard definitions that already contain it.
- Reuse existing APO resource filter data for standard resource variables such as cluster, namespace, service, pod, instance, and endpoint when those variables represent system resource filters.
- Keep `/api/custom-dashboard/variables` for custom PromQL/Grafana label-value variables, imported static option variables that require no backend call, label discovery, and query-context-dependent label candidates.
- Pass the current top-level `groupId` through custom dashboard candidate and query-preview requests so the backend can validate permissions and inject the correct data scope.
- Preserve backend permission enforcement as the security boundary; frontend candidate filtering remains a usability feature only.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `custom-dashboard-variables`: Data group is no longer a dashboard variable, standard resource candidates use the existing APO data group/resource option source, and the custom variable endpoint is scoped to PromQL/Grafana label-variable behavior.
- `custom-dashboard-querying`: Query preview and execution use the current top-level `groupId` for data scope enforcement instead of relying on a dashboard variable named `dataGroup`.
- `custom-dashboard-viewer`: Viewer controls show dashboard variables but not data group, and widget refreshes use the APO top-level data group.
- `custom-dashboard-editor`: Editor variable configuration excludes data group and preview requests use the APO top-level data group.

## Impact

- Frontend custom dashboard editor, viewer, behavior helpers, and API request builders.
- Backend custom dashboard request models, variable candidate handling, query preview handling, and data scope enforcement.
- Existing data group/resource filter APIs, especially `/api/v2/data/group/filter`, become the preferred source for standard resource candidate values.
- Existing dashboards containing legacy `dataGroup` variable definitions must keep loading without exposing that variable in the editor or viewer.
- Tests for custom dashboard variable normalization, candidate routing, request payloads, and backend data scope enforcement need updates.
