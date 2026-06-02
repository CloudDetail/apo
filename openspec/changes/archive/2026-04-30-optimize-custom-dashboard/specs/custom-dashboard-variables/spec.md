## MODIFIED Requirements

### Requirement: System Variables
The system SHALL support dashboard variables configured from labels used by the dashboard's widget metric queries. The system MUST NOT expose time range as a dashboard variable because time is controlled by the APO global time selector.

#### Scenario: Creator configures variables
- **WHEN** a creator configures dashboard variables
- **THEN** the system allows selecting labels discovered from or declared by the dashboard's widget metric queries

#### Scenario: Creator configures equivalent labels
- **WHEN** different dashboard metrics use different labels for the same filter concept
- **THEN** the system allows the creator to bind those labels, such as `svc` and `svc_name`, to one unified dashboard variable

#### Scenario: Creator looks for time variable
- **WHEN** a creator configures dashboard variables
- **THEN** the system does not offer time range as a configurable dashboard variable

### Requirement: Variable Substitution
The system SHALL resolve dashboard variables into metric-label matchers for custom query expressions and preset query parameters, including multi-select values and configured label aliases.

#### Scenario: Query references variable
- **WHEN** a widget query references a dashboard variable
- **THEN** the system resolves the variable value before backend query execution

#### Scenario: Variable has multiple selected values
- **WHEN** a viewer selects multiple values for a dashboard variable
- **THEN** the system resolves the selection into a query matcher that includes all selected values for each bound label alias

#### Scenario: Variable maps multiple label aliases
- **WHEN** a dashboard variable is bound to multiple metric labels with the same meaning
- **THEN** the system applies the selected variable values to each relevant metric query using that query's matching label name

### Requirement: Viewer Variable Controls
The system SHALL display multi-select dashboard variable selectors on the viewing page and refresh affected widgets when variable values change.

#### Scenario: Viewer changes variable value
- **WHEN** a viewer selects a different variable value
- **THEN** the system refreshes widgets that reference that variable using the new value

#### Scenario: Viewer selects multiple values
- **WHEN** a viewer selects multiple values in a dashboard variable selector
- **THEN** the system refreshes affected widgets using all selected values

### Requirement: Permission-Filtered Candidate Values
The system MUST filter variable candidate values by the current user's data group permissions, the dashboard's widget metric labels, and already selected upstream variable values.

#### Scenario: User views variable candidates
- **WHEN** a user opens variable candidate values
- **THEN** the system returns only candidates allowed by the user's data group permissions

#### Scenario: User has no candidates
- **WHEN** a user has no permitted candidate values for a dashboard variable
- **THEN** the system displays a no-available-data or no-permission state

#### Scenario: Namespace filters service candidates
- **WHEN** a viewer selects one or more namespace values and then opens service variable candidates
- **THEN** the system returns only service candidates associated with the selected namespace values and permitted by the user's data permissions

#### Scenario: Candidates are limited to dashboard labels
- **WHEN** a viewer opens candidates for a dashboard variable
- **THEN** the system returns values from labels used by the current dashboard's widget metric queries, not unrelated global labels

## ADDED Requirements

### Requirement: Variable Label Discovery
The system SHALL expose label metadata needed to configure dashboard variables from the dashboard's widget metric queries.

#### Scenario: Creator opens variable configuration
- **WHEN** the dashboard contains widget metric queries
- **THEN** the system shows selectable label keys grouped or identified by the metric queries where they appear

#### Scenario: Dashboard has no metric labels
- **WHEN** no widget metric query can provide label metadata
- **THEN** the system explains that variables become available after configuring widget metric queries
