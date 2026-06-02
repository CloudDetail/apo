## Context

Custom dashboards currently have two overlapping data-scope mechanisms. APO already exposes a top-level data group selector on custom dashboard routes, while dashboard variables can also include `dataGroup`. At the same time, standard resource candidates such as cluster, namespace, service, pod, instance, and endpoint are already available through the data group/resource filter flow used by service overview, but custom dashboards also fetch many of those values through `/api/custom-dashboard/variables`.

The desired product model is simpler: data group is global page context, not a dashboard parameter. Dashboard variables should represent filters specific to the dashboard's metric queries, while standard APO resource candidates should come from the same resource source as the rest of the product.

## Goals / Non-Goals

**Goals:**

- Make the top-level APO data group selector the only data group selector for custom dashboards.
- Keep backend data group permission enforcement authoritative for all custom dashboard candidate and query requests.
- Reuse the existing data group/resource filter source for standard resource variables where it can represent the variable.
- Keep `/api/custom-dashboard/variables` for custom label-value behavior that depends on PromQL/Grafana query context.
- Preserve legacy dashboard definitions that contain `dataGroup` variables without showing or saving them as active variables.

**Non-Goals:**

- Replace the global APO data group selector or change its route-level behavior.
- Implement the full Grafana variable editor.
- Add non-metric data sources to custom dashboards.
- Remove support for imported static option variables or PromQL/Grafana label-value variables.

## Decisions

### Data group is request context, not a dashboard variable

The frontend will read the current `dataGroupId` from APO's existing top-level data group state and include it in custom dashboard candidate and query-preview requests. The backend will accept that `groupId`, validate current-user view permission when it is non-zero, and inject the corresponding data-scope matcher before executing metric queries.

Alternative considered: keep `dataGroup` as a hidden dashboard variable. That keeps the current backend path smaller but preserves the wrong model and makes saved dashboard definitions depend on page context.

### Legacy `dataGroup` variables are tolerated but not active

Dashboard normalization will filter `dataGroup` out of active editor/viewer variable lists, similar to the existing time-range variable exclusion. Save serialization will not emit `dataGroup` variables. Backend query enforcement can keep tolerating old `variables["dataGroup"]` during migration, but explicit `groupId` request context takes precedence.

Alternative considered: reject dashboards that contain `dataGroup`. That would break dashboards created by previous iterations and is unnecessary because the variable can be ignored safely once request context exists.

### Standard resource candidates reuse existing resource APIs

For standard resource variables, the frontend should use the existing data group/resource filter source, especially `/api/v2/data/group/filter`, because it already receives `groupId`, category, time range, and extra resource type, and returns a tree used by service overview filtering. The custom dashboard frontend can adapt that tree into variable candidates for cluster, namespace, service, endpoint, pod, and instance.

Alternative considered: move all resource candidate calls behind `/api/custom-dashboard/variables` and have that service call the shared data service internally. That keeps a single custom dashboard API surface but still duplicates routing and response adaptation in the backend.

### Custom variable endpoint remains query-context-specific

`/api/custom-dashboard/variables` remains responsible for behavior that the resource API cannot represent: `label_values(...)`, arbitrary label-value variables, label discovery from PromQL selectors, imported Grafana query variables, static option compatibility, upstream variable substitution, and query-specific label scoping.

Alternative considered: remove the custom endpoint entirely. That would lose the Grafana replacement behavior required for imported or manually authored PromQL variables.

### Backend enforcement remains the security boundary

The frontend may use the resource filter tree to improve candidate UX, but every query preview and execution must still pass through backend data scope enforcement. Candidate filtering cannot be treated as permission enforcement.

Alternative considered: rely on frontend resource candidates to prevent out-of-scope selections. That is insufficient because users can edit requests and custom PromQL directly.

## Risks / Trade-offs

- Existing dashboards may contain `dataGroup` variables in saved config -> Normalize them out of active variables and keep backend tolerance during migration.
- Resource tree values may use ids while metric labels use names -> Candidate adapter must preserve the value shape required by the matching label and avoid changing query semantics silently.
- `/api/v2/data/group/filter` is tree-shaped, while dashboard variables expect flat candidates -> Keep the adapter small and local to custom dashboard behavior helpers.
- Some custom PromQL labels may look like standard resources but not match the resource tree -> Route only known standard semantic variables through the resource source; keep arbitrary label variables on the custom endpoint.
- Query context can change when global data group changes -> Include `groupId` in request signatures so stale candidate and preview responses are ignored.

## Migration Plan

1. Add `groupId` to custom dashboard candidate and query preview request contracts.
2. Update backend data scope enforcement to prefer explicit `groupId` and fall back to legacy `variables["dataGroup"]` only for compatibility.
3. Update frontend normalization and serialization so `dataGroup` variables are not displayed or written.
4. Route standard resource variable candidates to the existing resource filter source.
5. Keep `/api/custom-dashboard/variables` for custom label-value variables and label discovery.
6. Verify legacy dashboards still load and queries remain permission-scoped after data group changes.

## Open Questions

None.
