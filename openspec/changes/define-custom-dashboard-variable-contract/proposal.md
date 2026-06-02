## Why

Custom dashboard variables currently blur APO built-in resource filters with arbitrary PromQL label variables. This causes imported variables such as `namespace` or `service_name` to be treated as custom labels, makes the editor add-variable list empty when no query labels are discovered, and gives inconsistent All behavior across equivalent resource variables.

## What Changes

- Define a fixed built-in resource variable set: `cluster`, `namespace`, `service`, `endpoint`, `pod`, and `instance`.
- Normalize built-in resource variables from `type`, `name`, and `labels`, so imported Grafana variables saved as `type: "label"` still use APO resource behavior when their semantic label matches a built-in resource.
- Keep non-resource PromQL labels, including `job_name`, `cadvisor_job_name`, `related_nodes`, `related_pid`, `jvm_instance`, and business labels, as custom label variables.
- Make the dashboard information modal's add-variable list always show unenabled built-in resource variables, then supplement them with non-built-in labels discovered from widget queries.
- Route built-in resource candidates to `/api/v2/data/group/filter`, custom label candidates to `/api/custom-dashboard/variables/query`, and static option variables to local options.
- Make built-in resource variables default to Include All with `allValue: ".*"`, while custom label variables only show All when explicitly configured or imported with Include All.
- Make built-in resource variables multi-select by default and render them through a custom-dashboard-specific resource filter aligned with the Service Overview interaction pattern, without reusing the Service Overview `DataSourceFilter` component boundary.
- Ensure All means no effective restriction for both injected matchers and explicit variable tokens in PromQL label matchers.
- Align frontend runtime normalization, backend query substitution, and Grafana import normalization to the same contract.

## Capabilities

### New Capabilities

None.

### Modified Capabilities

- `custom-dashboard-variables`: Defines built-in resource versus custom label variable boundaries, alias normalization, candidate source routing, de-duplication, and All defaults.
- `custom-dashboard-editor`: Defines the grouped add-variable list shown in the dashboard information modal and the dedicated built-in resource parameter filter used in dashboard toolbars.
- `custom-dashboard-querying`: Defines PromQL variable substitution semantics when All is selected, especially exact matchers containing variable tokens.

## Impact

- Frontend custom dashboard behavior helpers, editor modal option rendering, viewer/editor resource parameter rendering, candidate loading, and variable substitution tests.
- Backend custom dashboard query substitution and variable normalization tests.
- Grafana import script normalization and its unit tests.
- APIs remain the same, but candidate requests must be routed to the correct existing endpoint based on the normalized variable category.
