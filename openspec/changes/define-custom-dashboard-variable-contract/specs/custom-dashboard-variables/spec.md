## MODIFIED Requirements

### Requirement: System Variables
The system SHALL support dashboard-local variables as either APO built-in resource variables or custom PromQL label variables. The built-in resource variables SHALL be exactly `cluster`, `namespace`, `service`, `endpoint`, `pod`, and `instance`. The system MUST classify a variable by its `type`, `name`, and `labels` rather than by `type` alone.

#### Scenario: Creator configures variables on a blank dashboard
- **WHEN** a creator opens dashboard variable configuration before any widget query labels are available
- **THEN** the system offers the unenabled built-in resource variables

#### Scenario: Blank dashboard starts with built-in resource filters
- **WHEN** a blank dashboard is created
- **THEN** its default variables include `cluster`, `namespace`, `service`, `endpoint`, `pod`, and `instance`
- **AND** each default built-in resource variable supports multi-select

#### Scenario: Imported label variable matches a built-in resource
- **WHEN** a dashboard variable is saved as `{ "name": "namespace", "type": "label", "labels": ["namespace"] }`
- **THEN** the system classifies it as the built-in `namespace` variable

#### Scenario: Imported service label matches a built-in resource
- **WHEN** a dashboard variable is saved as `{ "name": "service_name", "type": "label", "labels": ["service_name"] }`
- **THEN** the system classifies it as the built-in `service` variable

#### Scenario: Custom label does not match a built-in resource
- **WHEN** a dashboard variable is saved as `{ "name": "job_name", "type": "label", "labels": ["job_name"] }`
- **THEN** the system classifies it as a custom label variable

#### Scenario: Related nodes remains a custom label
- **WHEN** a dashboard variable is named `related_nodes`
- **THEN** the system classifies it as a custom label variable rather than the built-in `instance` variable

### Requirement: Permission-Filtered Candidate Values
The system MUST filter variable candidate values by the current top-level APO data group, the current user's data group permissions, the dashboard's widget metric labels, and already selected upstream variable values. Candidate source routing SHALL be selected by the normalized variable category.

#### Scenario: Built-in resource variable uses system resource source
- **WHEN** a dashboard variable is classified as `cluster`, `namespace`, `service`, `endpoint`, `pod`, or `instance`
- **THEN** the system obtains candidate values from `/api/v2/data/group/filter`

#### Scenario: Built-in resource variable is imported as a label
- **WHEN** a dashboard variable is saved as `type: "label"` but its `name` or `labels` match a built-in resource alias
- **THEN** the system obtains candidate values from `/api/v2/data/group/filter`

#### Scenario: Built-in aliases participate in linked filtering
- **WHEN** a built-in resource variable is saved with an alias name such as `namespace_name`, `dst_namespace`, `cluster_name`, or `service_name`
- **THEN** downstream built-in resource candidates are filtered by the selected alias values as the corresponding canonical resource selections

#### Scenario: Custom label variable uses dashboard query endpoint
- **WHEN** a dashboard variable does not match a built-in resource alias and does not have static local options
- **THEN** the system obtains candidate values from `/api/custom-dashboard/variables/query`

#### Scenario: Static option variable stays local
- **WHEN** a dashboard variable has static candidate options
- **THEN** the system uses the local options without requesting candidate values from a backend endpoint

### Requirement: Variable Label Discovery
The system SHALL expose label metadata needed to configure dashboard variables from built-in resource definitions and the dashboard's widget metric queries.

#### Scenario: Query contains a built-in resource label
- **WHEN** widget query labels contain `namespace`, `service_name`, `pod`, or another built-in resource alias
- **THEN** the system exposes the corresponding built-in resource variable and does not also expose that label as a custom label variable

#### Scenario: Query contains custom labels
- **WHEN** widget query labels contain labels that do not match built-in resource aliases
- **THEN** the system exposes those labels as query-discovered custom label variable options

#### Scenario: Enabled variable has custom label aliases
- **WHEN** an enabled variable has `labels: ["job_name"]` and another option is discovered for `job_name`
- **THEN** the system does not show a duplicate add-variable option for `job_name`

### Requirement: Resource Parameter Controls
The system SHALL render enabled built-in resource variables through a custom-dashboard-specific resource filter in viewer and editor toolbars.

#### Scenario: Built-in resources use dedicated controls
- **WHEN** a dashboard has enabled built-in resource variables
- **THEN** the toolbar renders them with the custom dashboard resource filter rather than the Service Overview filter component

#### Scenario: Custom labels keep query controls
- **WHEN** a dashboard has custom label variables
- **THEN** the toolbar continues to render them with the dashboard variable Select controls backed by query-context candidates

### Requirement: Variable All Values
The system SHALL apply Include All defaults consistently by variable category.

#### Scenario: Built-in resource variable defaults to All
- **WHEN** a built-in resource variable is created or normalized
- **THEN** it defaults to `includeAll: true` and `allValue: ".*"` unless an explicit compatible value is already configured
- **AND** it defaults to `multi: true`

#### Scenario: Built-in resource candidates include All first
- **WHEN** candidate values are displayed for a built-in resource variable
- **THEN** the All option appears before concrete candidate values

#### Scenario: Custom label does not default to All
- **WHEN** a custom label variable is created without explicit Include All configuration
- **THEN** the system does not add an All option automatically

#### Scenario: Imported custom label preserves All
- **WHEN** an imported custom label variable has Grafana Include All enabled
- **THEN** the system preserves the All option for that variable
