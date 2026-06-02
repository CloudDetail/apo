## MODIFIED Requirements

### Requirement: Query Preview
The system SHALL provide query preview using the current APO global time range, current APO top-level data group, current dashboard variable values, and widget display formatting.

#### Scenario: Preview succeeds
- **WHEN** a user previews a valid query that returns data
- **THEN** the system displays a chart or table preview matching the widget type and applying configured decimals and unit display

#### Scenario: Preview has no data
- **WHEN** a user previews a valid query that returns no data
- **THEN** the system displays a no-data state rather than a query failure

#### Scenario: Global time changes before preview
- **WHEN** a user changes the APO global time selector and then previews a widget query
- **THEN** the preview request uses the selected global time range as the query window

#### Scenario: Top-level data group changes before preview
- **WHEN** a user changes the APO top-level data group selector and then previews a widget query
- **THEN** the preview request uses the selected data group as the backend query scope

#### Scenario: Variable changes before preview
- **WHEN** a user changes an enabled dashboard-local variable and then previews a widget query
- **THEN** the preview request uses the variable values that apply to that widget query

### Requirement: Backend Query Enforcement
The system MUST execute dashboard metric queries through the backend and MUST enforce data group permissions, timeout, maximum returned points, concurrency, resource limits, and audit logging.

#### Scenario: Query uses selected data group
- **WHEN** a dashboard query request includes a non-zero top-level `groupId`
- **THEN** the backend validates that the current user has view permission for that data group before querying

#### Scenario: Query has no selected data group
- **WHEN** a dashboard query request has no top-level `groupId` or uses the all-data group
- **THEN** the backend constrains the query to the current user's permitted data groups unless the user has all-data permission

#### Scenario: Query contains legacy data group variable
- **WHEN** a dashboard query request includes both top-level `groupId` and legacy `variables["dataGroup"]`
- **THEN** the backend uses the top-level `groupId` as the authoritative data scope

#### Scenario: Query exceeds limits
- **WHEN** a dashboard query exceeds timeout, returned point, concurrency, or resource limits
- **THEN** the backend stops or rejects the query and returns a user-understandable limit error

#### Scenario: Query executes
- **WHEN** the backend executes a dashboard query
- **THEN** the system records the user, time, query expression, and constrained data scope in audit data
