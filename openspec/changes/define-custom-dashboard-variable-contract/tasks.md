## 1. Frontend Contract Tests

- [x] 1.1 Add tests for grouped add-variable options always exposing unenabled built-in resources before query-discovered custom labels.
- [x] 1.2 Add tests for blank dashboards still offering built-in resource variables.
- [x] 1.3 Add tests for imported label-shaped variables such as `namespace`, `pod`, and `service_name` using built-in resource candidate routing.
- [x] 1.4 Add tests that `related_nodes`, `related_pid`, `jvm_instance`, `job_name`, and `cadvisor_job_name` remain custom label variables.
- [x] 1.5 Add tests for de-duplicating query-discovered custom labels by enabled variable name and enabled variable `labels`.
- [x] 1.6 Add tests for built-in resource All defaults and custom label All preservation.

## 2. Frontend Classification And Candidate Routing

- [x] 2.1 Move the canonical built-in resource names and alias resolver before any helper that needs built-in defaults, avoiding initialization-order errors.
- [x] 2.2 Implement one canonical resolver that classifies variables from `type`, `name`, and `labels`.
- [x] 2.3 Route `isStandardResourceVariable`, resource request extra selection, resource tree candidate adaptation, and candidate request planning through the canonical resolver.
- [x] 2.4 Keep `related_nodes` out of the built-in alias map so it routes through `/api/custom-dashboard/variables/query`.
- [x] 2.5 Preserve local candidate handling for static option variables.

## 3. Editor Add-Variable List

- [x] 3.1 Add a grouped option helper that returns unenabled built-in resource variables in canonical order plus non-built-in query labels.
- [x] 3.2 Update the dashboard information modal to render grouped add-variable options.
- [x] 3.3 Add Chinese and English i18n keys for the built-in resource group and query-discovered label group.
- [x] 3.4 Ensure the add-variable list remains useful when there are no query-discovered custom labels.

## 4. All Semantics

- [x] 4.1 Default built-in resource variables to `includeAll: true` and `allValue: ".*"`.
- [x] 4.2 Show All first in built-in resource candidate selectors.
- [x] 4.3 Leave custom label variables without All unless the saved or imported definition explicitly enables it.
- [x] 4.4 Update frontend matcher application so All-injected matchers are wildcard regex matchers and not exact literal comparisons.
- [x] 4.5 Update backend variable substitution so exact matchers containing an All-selected variable token do not execute as literal `label=".*"`.

## 5. Backend And Grafana Import Alignment

- [x] 5.1 Add backend tests for All substitution in exact and regex PromQL matchers.
- [x] 5.2 Add or update backend normalization tests for imported built-in resource aliases.
- [x] 5.3 Remove `related_nodes -> instance` from the Grafana import built-in resource mapping.
- [x] 5.4 Add Python unittest coverage for built-in aliases, custom labels, Include All preservation, and `related_nodes` remaining custom.

## 6. Verification

- [x] 6.1 Run `node --test frontend/src/pro/views/CustomDashboard/behavior.test.mjs`.
- [x] 6.2 Run targeted Go tests for `backend/pkg/services/customdashboard`.
- [x] 6.3 Run `python3 -m unittest scripts/test_import_grafana_dashboard.py`.
- [x] 6.4 Manually verify the affected dashboard as `admin / APO2024@admin`: built-in resource variables use `/api/v2/data/group/filter`, custom labels use `/api/custom-dashboard/variables/query`, and the add-variable list is populated on blank or query-light dashboards.

## 7. Dedicated Built-in Resource Filter

- [x] 7.1 Add tests for a custom-dashboard-specific built-in resource filter component and viewer/editor wiring.
- [x] 7.2 Render enabled built-in resource variables through the dedicated component while leaving custom label variables on the existing dashboard Select controls.
- [x] 7.3 Default all built-in resource variables to multi-select and include `endpoint` and `instance` in blank dashboard defaults.
- [x] 7.4 Align endpoint translation with existing Service Overview wording.
- [x] 7.5 Add tests and fixes for full alias participation in built-in resource linked filtering.
- [x] 7.6 Add tests and fixes for All exact matcher substitution with Grafana formatted tokens such as `${var:raw}` and `${var:regex}`.
