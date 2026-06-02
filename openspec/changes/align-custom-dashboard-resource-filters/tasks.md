## 1. Backend Request Contract And Scope Enforcement

- [x] 1.1 Add top-level `groupId` to custom dashboard variable candidate and query preview request models.
- [x] 1.2 Update backend query scope enforcement to prefer request `groupId`, validate view permission for non-zero groups, and keep legacy `variables["dataGroup"]` as compatibility fallback only.
- [x] 1.3 Update backend variable candidate handling so custom label-value candidates receive and enforce the current request `groupId`.
- [x] 1.4 Add or update backend tests covering selected `groupId`, no selected group, all-data permission, invalid group permission, and legacy `variables["dataGroup"]` fallback.

## 2. Frontend Variable Normalization And Request Payloads

- [x] 2.1 Remove `dataGroup` from active semantic dashboard variable definitions and filter legacy `dataGroup` variables during dashboard normalization.
- [x] 2.2 Update dashboard serialization so saved definitions do not emit `dataGroup` or time range variables.
- [x] 2.3 Include the current top-level `dataGroupId` in custom dashboard variable candidate requests from both editor and viewer.
- [x] 2.4 Include the current top-level `dataGroupId` in widget preview requests from both editor and viewer.
- [x] 2.5 Ensure request signatures include `groupId` so stale candidate and preview responses are ignored after data group changes.

## 3. Standard Resource Candidate Reuse

- [x] 3.1 Add a frontend adapter that converts the existing APO data group/resource filter tree into flat dashboard variable candidates for cluster, namespace, service, endpoint, pod, and instance.
- [x] 3.2 Route standard resource variables to the existing data group/resource filter API using current `groupId`, category, time range, and resource type.
- [x] 3.3 Keep static option variables local and route custom label-value or imported Grafana label variables through `/api/custom-dashboard/variables`.
- [x] 3.4 Preserve upstream variable filtering when standard resource candidates cascade from cluster to namespace to service to endpoint, pod, or instance.

## 4. Editor And Viewer Behavior

- [x] 4.1 Update editor variable suggestions so data group is never offered as an addable dashboard variable.
- [x] 4.2 Update editor and viewer toolbar rendering so legacy data group variables are not displayed.
- [x] 4.3 Refresh candidates and affected widgets when the top-level APO data group changes.
- [x] 4.4 Confirm legacy dashboards containing `dataGroup` still load, render, preview, and save without reintroducing that variable.

## 5. Verification

- [x] 5.1 Update frontend behavior tests for data group filtering, serialization, request payloads, candidate routing, and stale response signatures.
- [x] 5.2 Run `cd frontend && node --test src/pro/views/CustomDashboard/behavior.test.mjs`.
- [x] 5.3 Run targeted backend tests for custom dashboard service and API handlers.
- [x] 5.4 Manually verify editor and viewer flows with top-level data group changes, standard resource variables, custom label-value variables, and legacy `dataGroup` dashboard definitions.
