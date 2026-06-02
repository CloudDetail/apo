## Context

APO custom dashboards support two different kinds of variables:

- Built-in APO resource variables, which represent platform resources and have an existing resource candidate source.
- Custom PromQL label variables, which are user-authored or imported from Grafana and need query-context-specific label value lookup.

The current implementation does not express that boundary strongly enough. Imported Grafana variables can arrive as `type: "label"` even when their `name` or `labels` identify a built-in resource such as `namespace`, `pod`, or `service_name`. The editor also discovers variable options mainly from widget query labels, so a blank dashboard or a dashboard whose query does not mention `namespace` can fail to offer the built-in namespace variable.

## Goals / Non-Goals

**Goals:**

- Define a stable contract for built-in resource variables and custom label variables.
- Show built-in resource variables in the add-variable list independent of query label discovery.
- Normalize imported or saved semantic label variables by inspecting `type`, `name`, and `labels`.
- Route candidate value requests by normalized category.
- Make All behavior consistent for built-in resources and semantically correct in PromQL.
- Keep arbitrary custom label variables available for imported and manually authored PromQL dashboards.

**Non-Goals:**

- Replace the APO top-level data group selector.
- Implement a full Grafana variable editor.
- Add complex custom label variable authoring beyond the current query/import behavior.
- Treat frontend candidate filtering as a permission boundary.

## Decisions

### Built-in resources are a fixed semantic set

The built-in resource variables are exactly `cluster`, `namespace`, `service`, `endpoint`, `pod`, and `instance`. They are product-level APO variables, so the editor should offer them even when no widget query currently contains the matching label.

The canonical alias map is:

| Built-in variable | Accepted aliases |
| --- | --- |
| `cluster` | `cluster`, `cluster_id`, `clusterId`, `clusterID`, `cluster_name` |
| `namespace` | `namespace`, `namespace_name`, `namespaceName`, `src_namespace`, `dst_namespace` |
| `service` | `svc`, `svc_name`, `service`, `service_name` |
| `endpoint` | `endpoint`, `endpoint_name`, `content_key`, `contentKey` |
| `pod` | `pod`, `pod_name`, `src_pod`, `dst_pod` |
| `instance` | `instance`, `instance_name`, `node_name`, `nodeName`, `host_name`, `hostName` |

`related_nodes` is not a built-in resource alias. It remains a custom label variable unless a future product requirement explicitly adds a separate mapping.

### Variable classification must inspect type, name, and labels

Runtime classification cannot rely on `type` alone. A variable like `{ "name": "namespace", "type": "label", "labels": ["namespace"] }` must be treated as the built-in `namespace` variable. A variable like `{ "name": "job_name", "type": "label", "labels": ["job_name"] }` must remain a custom label variable.

Frontend behavior helpers should expose one canonical resolver and route all checks through it: add-variable option filtering, candidate request planning, resource tree adaptation, All defaults, and label de-duplication.

### Built-in resource controls use a custom dashboard filter

Viewer and editor toolbars should render built-in resource variables through a custom-dashboard-specific resource filter, not the Service Overview `DataSourceFilter`. The visual behavior should match Service Overview closely enough for users to recognize the pattern: compact labels beside Select controls, multi-select tags capped to a small count, clearable selections, and parent-child pruning when upstream filters change.

The component boundary remains custom-dashboard-specific because dashboard variables can be imported with alias names such as `service_name`, can coexist with dynamic custom label variables, and must write selections back to the original dashboard variable name. Reusing the Service Overview component would blur that boundary and make alias/custom-label behavior harder to reason about.

The resource filter should render only enabled built-in variables, in canonical resource order when possible. Custom label variables continue to render with the existing variable Select controls and continue to use query-context candidate lookup.

### Candidate sources are selected by normalized category

Built-in resource variables use `/api/v2/data/group/filter` with the current global data group, time range, and resource extra where required:

| Built-in variable | Resource request extra |
| --- | --- |
| `cluster` | none |
| `namespace` | none |
| `service` | none |
| `endpoint` | `endpoint` |
| `pod` | `instance` |
| `instance` | `instance` |

Custom label variables use `/api/custom-dashboard/variables/query`, including upstream variable values and query-specific scope. Static option variables use their local options without a backend candidate request.

### The add-variable list is grouped and deterministic

The dashboard information modal should show:

- Built-in resource variables: all unenabled built-ins in the canonical order.
- Labels discovered from queries: only non-built-in labels that are not already enabled or represented by an enabled variable's `labels`.

This makes blank dashboards and dashboards without resource labels still configurable while preventing duplicate options such as `namespace` appearing as both a resource variable and a custom label.

### All means no effective restriction

Built-in resource variables default to `includeAll: true` and `allValue: ".*"`. The All candidate is shown first. Custom label variables do not automatically receive All unless the imported or saved definition explicitly enables it.

Built-in resource variables also default to `multi: true`. Multi-select is the safest default because these parameters represent resource scopes and users often need to compare or aggregate multiple clusters, namespaces, services, endpoints, pods, or instances. All remains mutually exclusive with concrete selections.

Selecting All must not turn an exact matcher into a literal dot-star comparison. For example, `namespace="$namespace"` with namespace All selected must execute as no effective namespace restriction, such as a regex wildcard matcher, not as `namespace=".*"`. The same rule applies to Grafana formatted token variants such as `${namespace:raw}` and `${namespace:regex}` when they appear inside exact matchers.

### Import and runtime normalization must agree

The Grafana import script should preserve arbitrary label variables and only normalize variables that match the fixed built-in alias map. Backend query substitution must be tolerant of imported Grafana variable syntax and preserve All semantics during execution.

## Risks / Trade-offs

- Some historical dashboards may have relied on importer-specific mappings such as `related_nodes -> instance`; removing that implicit resource mapping may change candidate source behavior, but it matches the clarified product boundary.
- Resource APIs and PromQL labels can use different value shapes. The adapter must keep using the value that matches the label being filtered.
- All handling touches both frontend expression rewriting and backend token substitution, so tests must cover both injected matchers and explicit Grafana-style variable tokens.
- Query-discovered label de-duplication by both variable name and label aliases is required to avoid regressions when imported variables use different display names.

## Migration Plan

1. Add tests that capture the clarified built-in/custom boundary, grouped add-variable list, candidate routing, All defaults, and All substitution semantics.
2. Centralize frontend variable classification and route existing helpers through the canonical resolver.
3. Update the editor modal to consume grouped variable options and i18n labels.
4. Add a custom-dashboard-specific built-in resource filter for viewer/editor toolbars.
5. Normalize built-in All and multi-select defaults while preserving custom label All behavior.
6. Align backend query substitution and Grafana import normalization with the same alias map.
7. Verify the existing affected dashboard in the browser and run targeted frontend, backend, and import tests.

## Open Questions

None.
