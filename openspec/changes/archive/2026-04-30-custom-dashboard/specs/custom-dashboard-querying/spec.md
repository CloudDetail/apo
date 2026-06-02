## ADDED Requirements

### Requirement: Metric-Only Query Scope
The system SHALL support metric data queries for custom dashboards and MUST NOT support log, trace, alert event, SQL, or mixed data source queries in MVP.

#### Scenario: User configures metric query
- **WHEN** a user configures a widget query
- **THEN** the system allows only metric query modes and rejects unsupported query types

### Requirement: Preset Query Mode
The system SHALL provide preset queries grouped by scenario, with name, description, compatible chart types, parameters, and parameter descriptions.

#### Scenario: User selects preset query
- **WHEN** a user selects a preset query for a compatible widget type
- **THEN** the system displays only the parameters exposed by that preset query

#### Scenario: User attempts to edit preset definition
- **WHEN** a user configures a preset query
- **THEN** the system allows parameter values but does not allow changing the preset query definition

### Requirement: Custom Query Mode
The system SHALL allow custom PromQL/VictoriaMetrics-compatible query expressions with variable references and legend configuration.

#### Scenario: User enters custom query
- **WHEN** a user enters a custom metric query expression
- **THEN** the system stores the expression, variable references, and legend configuration with the widget

### Requirement: Query Preview
The system SHALL provide query preview using the current time range and current variable values.

#### Scenario: Preview succeeds
- **WHEN** a user previews a valid query that returns data
- **THEN** the system displays a chart or table preview matching the widget type

#### Scenario: Preview has no data
- **WHEN** a user previews a valid query that returns no data
- **THEN** the system displays a no-data state rather than a query failure

### Requirement: Query Validation Errors
The system SHALL validate query syntax and display understandable error reasons, including error location when available.

#### Scenario: Query syntax is invalid
- **WHEN** a user validates or previews an invalid query expression
- **THEN** the system displays the syntax error reason and location when available

### Requirement: Backend Query Enforcement
The system MUST execute dashboard metric queries through the backend and MUST enforce timeout, maximum returned points, concurrency, resource limits, and audit logging.

#### Scenario: Query exceeds limits
- **WHEN** a dashboard query exceeds timeout, returned point, concurrency, or resource limits
- **THEN** the backend stops or rejects the query and returns a user-understandable limit error

#### Scenario: Query executes
- **WHEN** the backend executes a dashboard query
- **THEN** the system records the user, time, query expression, and constrained data scope in audit data
