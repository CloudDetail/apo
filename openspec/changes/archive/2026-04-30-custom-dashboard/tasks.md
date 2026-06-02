## 1. Data Model And Seeds

- [x] 1.1 Add dashboard persistence schema for metadata, owner, visibility, layout, variables, widgets, query definitions, and timestamps
- [x] 1.2 Add query audit persistence for user, query time, expression, resolved variables, constrained data scope, result status, and limit failures
- [x] 1.3 Add create-dashboard role permission seed data and permission API metadata
- [x] 1.4 Add built-in template and preset query definition files or seed loaders for service health, infrastructure resources, and blank templates

## 2. Backend Dashboard APIs

- [x] 2.1 Implement dashboard list API filtered by current user's visibility permissions
- [x] 2.2 Implement dashboard get API with not-found and no-visibility handling
- [x] 2.3 Implement create-from-template and create-blank APIs gated by create-dashboard permission
- [x] 2.4 Implement update API that only allows the dashboard creator to save metadata, visibility, variables, widgets, layout, queries, and display configuration
- [x] 2.5 Implement copy API that requires view access plus create-dashboard permission and defaults copied dashboards to personal visibility
- [x] 2.6 Implement delete API that requires dashboard creator ownership and invalidates future opens

## 3. Backend Query, Variables, And Permissions

- [x] 3.1 Implement template and preset query catalog APIs with i18n-ready names and descriptions
- [x] 3.2 Implement variable candidate API for data group, cluster, namespace, service, instance/Pod, and time range inputs
- [x] 3.3 Enforce current-user data group filtering for variable candidates in the backend
- [x] 3.4 Implement custom PromQL/VictoriaMetrics query validation with syntax error reason and location when available
- [x] 3.5 Implement preset and custom query resolution with variable substitution
- [x] 3.6 Inject or enforce current-user data group constraints before metric query execution
- [x] 3.7 Implement metric preview/execution API with timeout, maximum returned points, concurrency limits, and resource-limit errors
- [x] 3.8 Record audit data for every dashboard metric query execution attempt

## 4. Frontend Navigation And Dashboard Management

- [x] 4.1 Add the "My Dashboards" first-level menu entry and routes without removing existing system dashboards
- [x] 4.2 Build dashboard list page with name, description, creator, visibility, updated time, last visited time, and permitted actions
- [x] 4.3 Build dashboard create flow for template selection and blank creation, hiding creation actions when permission is missing
- [x] 4.4 Build dashboard copy and delete actions with permission-aware UI, default copied name, default personal visibility, and delete confirmation
- [x] 4.5 Add not-found and no-permission states for dashboard open failures

## 5. Frontend Editor

- [x] 5.1 Build native editor shell with toolbar, widget entry controls, central grid, and right-side configuration panel
- [x] 5.2 Implement grid drag, resize, snapping, collision avoidance, minimum widget dimensions, and responsive desktop behavior
- [x] 5.3 Implement widget add, copy, delete, and selection behavior
- [x] 5.4 Implement line, area, bar, single value, table, and TopN ranking widget renderers
- [x] 5.5 Implement widget configuration controls for title, description, chart type, query, unit, decimals, legend, thresholds, empty state, and error state
- [x] 5.6 Implement dashboard metadata, visibility, and variable configuration panels
- [x] 5.7 Implement save, save-failure retention, cancel, preview/view, and unsaved-change prompts

## 6. Frontend Query And Viewer

- [x] 6.1 Build preset query configuration UI grouped by service, instance/Pod, and infrastructure scenarios
- [x] 6.2 Build custom query editor with variable insertion, legend configuration, validation, and preview actions
- [x] 6.3 Render query preview success, syntax error, timeout, too-large, no-data, no-permission, and metric-missing states
- [x] 6.4 Build native dashboard viewer with variable selectors, global time range, manual refresh, auto refresh, and widget fullscreen view
- [x] 6.5 Link viewer and editor themes to the APO system light/dark theme
- [x] 6.6 Ensure widget title, legend, axis, table content, and controls remain readable at common desktop resolutions

## 7. Internationalization And Quality

- [x] 7.1 Add Chinese and English locale text for dashboard navigation, management, editor, query configuration, viewer states, and errors
- [x] 7.2 Add backend unit tests for permission checks, dashboard CRUD ownership rules, copy defaults, variable filtering, query validation, limit errors, and audit logging
- [ ] 7.3 Add frontend tests for dashboard list action visibility, create/copy/delete flows, editor unsaved prompts, query preview states, and viewer controls
- [ ] 7.4 Add integration tests proving dashboard visibility does not expand data group metric access
- [x] 7.5 Run backend, frontend, and OpenSpec validation commands and fix any failures before marking implementation complete
