## Context

The custom dashboard feature has an APO-native editor and viewer after the previous optimization change. Manual testing found that several editor behaviors still do not meet dashboard-builder expectations: palette previews are not meaningful, drag placement is constrained incorrectly, new widgets are too large, variables expose label-level complexity to users, and the unit field is only partially useful.

The product intent is for APO custom dashboards to act as a Grafana dashboard replacement for metric dashboards. Grafana's useful concepts for this change are dashboard-level template variables and display units that consistently affect panels. APO will keep those concepts but hide low-level variable query and label-binding complexity from normal users.

## Goals / Non-Goals

**Goals:**

- Make the widget palette visually useful by showing type-specific thumbnails.
- Keep dashboard layout fixed to two columns while supporting unlimited rows, drag placement, and automatic compaction.
- Simplify variable setup so users select suggested semantic variables instead of configuring variable names, labels, or aliases.
- Preserve Grafana-like dashboard variable behavior: one dashboard variable selection can filter multiple widgets even when their metric labels differ.
- Remove unsupported widget types from new configuration and provide predictable behavior for existing dashboards that contain them.
- Make unit display effective across charts and tables.

**Non-Goals:**

- Implement the full Grafana variable system, including query variables, custom variables, constant variables, ad hoc filters, chained variable editing, or user-authored variable queries.
- Add widget resizing, multiple cell spans, or freeform grid layout.
- Add non-metric data sources.
- Add new widget types beyond line, area, bar, and table.

## Decisions

### Variables use system-suggested semantic types

Dashboard variables will be represented as enabled semantic variable types, such as service, namespace, pod, cluster, and dataGroup. The editor will show suggestions derived from the dashboard's configured metric queries and preset query metadata. Users choose which suggestions are enabled and can delete enabled variables, but they do not configure variable names, labels, or aliases.

Alternative considered: keep manual label binding. That is flexible but makes users understand metric label details and conflicts with the desired Grafana-replacement workflow. Alternative considered: implement Grafana's full variable editor. That is too broad for this change and would introduce unnecessary query and permission complexity.

### Label aliases remain internal

The system will maintain internal alias mappings from semantic variables to metric labels, for example service mapping to `svc`, `svc_name`, `service`, and `service_name`. During query execution, each widget query receives filters only for labels that exist in that query or preset definition.

Alternative considered: require every widget to reference a variable token explicitly. That is closer to Grafana but makes the current APO editor harder to use and would not satisfy the requirement that users only choose which variables to add.

### Layout is compacted after placement

The editor grid remains a fixed two-column grid with unlimited rows. Widgets occupy one fixed cell. After add, copy, delete, or drag placement, the layout is normalized in row-major order and compacted so there are no empty cells between the first and last widget.

Alternative considered: preserve holes to let users reserve space. That behavior contradicts the tested expectation that upper and lower widgets must not have empty cells between them.

### Removed widget types are handled explicitly

New dashboards and new widgets will only support line, area, bar, and table. Existing TopN widgets will render as tables because their data shape is already table-like. Existing single-value widgets will be marked as removed or unsupported and remain editable only for deletion or changing to a supported type.

Alternative considered: convert single-value widgets to line charts automatically. That can hide a product-level removal and may change the user's dashboard meaning without consent.

### Units are display configuration

The unit field remains part of widget display configuration. It will be applied consistently to chart y-axis labels, chart tooltip values, and table numeric values. Decimal places continue to control numeric precision.

Alternative considered: remove unit configuration. That would simplify the form but make common metric dashboards worse for CPU, latency, throughput, and size metrics.

## Risks / Trade-offs

- Semantic variable detection can miss project-specific labels -> Keep alias mappings centralized and easy to extend; only show suggestions when at least one widget query can use the variable.
- Automatic compaction changes layout after drag/delete -> Make compaction deterministic and cover row-major behavior with tests.
- Existing single-value widgets become unsupported -> Render a clear unsupported state with an edit path to delete or change type instead of silently changing chart semantics.
- Unit formatting may produce duplicated units if a query already returns formatted strings -> Apply units only to numeric values and leave non-numeric values unchanged.
- Grafana parity can be interpreted too broadly -> Keep this change scoped to dashboard-level filters, panel type selection, fixed layout, and unit display.

## Migration Plan

1. Normalize loaded dashboard definitions so historical `topn` widgets render through table behavior.
2. Preserve historical `single` widgets but mark them as removed/unsupported in editor and viewer until the user changes or deletes them.
3. Stop emitting `single` and `topn` for newly created or saved widgets.
4. Convert existing manual variable definitions to enabled semantic variables when their labels match known alias mappings; otherwise ignore unsupported manual bindings during save while keeping legacy data tolerated on read.
5. Roll back by keeping the compatibility reader tolerant of both semantic variables and older label-bound variables.

## Open Questions

None.
