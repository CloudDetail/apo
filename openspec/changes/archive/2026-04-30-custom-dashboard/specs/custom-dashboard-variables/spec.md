## ADDED Requirements

### Requirement: System Variables
The system SHALL support dashboard variables for data group, cluster, namespace, service, instance/Pod, and time range in MVP.

#### Scenario: Creator configures variables
- **WHEN** a creator configures dashboard variables
- **THEN** the system allows selecting supported system variables for use in widget queries and preset parameters

### Requirement: Variable Substitution
The system SHALL allow dashboard variables to be referenced by custom query expressions and preset query parameters.

#### Scenario: Query references variable
- **WHEN** a widget query references a dashboard variable
- **THEN** the system resolves the variable value before backend query execution

### Requirement: Viewer Variable Controls
The system SHALL display dashboard variable selectors on the viewing page and refresh affected widgets when variable values change.

#### Scenario: Viewer changes variable value
- **WHEN** a viewer selects a different variable value
- **THEN** the system refreshes widgets that reference that variable using the new value

### Requirement: Permission-Filtered Candidate Values
The system MUST filter variable candidate values by the current user's data group permissions.

#### Scenario: User views variable candidates
- **WHEN** a user opens variable candidate values
- **THEN** the system returns only candidates allowed by the user's data group permissions

#### Scenario: User has no candidates
- **WHEN** a user has no permitted candidate values for a dashboard variable
- **THEN** the system displays a no-available-data or no-permission state

### Requirement: Per-Viewer Variable Results
The system SHALL allow different users viewing the same dashboard to receive different variable candidates according to their permissions.

#### Scenario: Two users view same dashboard
- **WHEN** two users with different data group permissions open the same dashboard
- **THEN** each user sees variable candidates filtered to their own data permissions
