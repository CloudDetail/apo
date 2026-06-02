## ADDED Requirements

### Requirement: Create Permission
The system MUST require a role-bound create-dashboard permission for creating blank dashboards, creating from templates, and copying dashboards.

#### Scenario: User lacks create permission
- **WHEN** a user without create-dashboard permission attempts to create or copy a dashboard
- **THEN** the system denies the operation

### Requirement: Dashboard Visibility
The system SHALL support dashboard visibility scopes of everyone, team, and personal, with new dashboards defaulting to everyone.

#### Scenario: Creator creates dashboard
- **WHEN** a creator creates a new dashboard
- **THEN** the dashboard visibility defaults to everyone unless the creator changes it

#### Scenario: Viewer lacks visibility
- **WHEN** a user without matching dashboard visibility attempts to open the dashboard
- **THEN** the system denies access to the dashboard

### Requirement: Visibility Editing
The system SHALL allow only the dashboard creator to modify dashboard visibility.

#### Scenario: Non-creator attempts visibility change
- **WHEN** a user who is not the creator attempts to change dashboard visibility
- **THEN** the system denies the change

### Requirement: Data Permission Enforcement
The system MUST constrain query results by the current user's data group permissions regardless of dashboard visibility.

#### Scenario: Shared dashboard is viewed
- **WHEN** a user opens an everyone-visible dashboard
- **THEN** the system returns only metric data allowed by that user's data group permissions

#### Scenario: User query omits data filter
- **WHEN** a custom query does not include required data group constraints
- **THEN** the backend applies the current user's data group constraints before executing the query

### Requirement: Dashboard Sharing Does Not Expand Data Access
The system MUST NOT grant additional metric data access because a dashboard is visible or shared.

#### Scenario: Users with different data permissions view same dashboard
- **WHEN** users from different data groups view the same visible dashboard
- **THEN** each user sees only their own permitted metric data
