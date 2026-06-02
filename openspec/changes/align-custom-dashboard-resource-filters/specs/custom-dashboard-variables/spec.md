## MODIFIED Requirements

### Requirement: System Variables
The system SHALL support dashboard variables as dashboard-local filters derived from labels used by the dashboard's widget metric queries and imported variable definitions. The system MUST NOT expose time range as a dashboard variable because time is controlled by the APO global time selector. The system MUST NOT expose data group as a dashboard variable because data group is controlled by the APO top-level data group selector.

#### Scenario: Creator configures variables
- **WHEN** a creator configures dashboard variables
- **THEN** the system allows enabling supported dashboard-local variables discovered from or declared by the dashboard's widget metric queries

#### Scenario: Creator configures equivalent labels
- **WHEN** different dashboard metrics use different labels for the same filter concept
- **THEN** the system maps those labels, such as `svc` and `svc_name`, to one unified dashboard variable without requiring creator label binding

#### Scenario: Creator looks for time variable
- **WHEN** a creator configures dashboard variables
- **THEN** the system does not offer time range as a configurable dashboard variable

#### Scenario: Creator looks for data group variable
- **WHEN** a creator configures dashboard variables
- **THEN** the system does not offer data group as a configurable dashboard variable

#### Scenario: Dashboard contains legacy data group variable
- **WHEN** a dashboard definition contains a legacy variable named `dataGroup` or typed as `dataGroup`
- **THEN** the editor and viewer ignore it as an active dashboard variable and do not write it back when saving

### Requirement: Variable Substitution
The system SHALL resolve enabled dashboard-local variables into metric-label matchers for custom query expressions and preset query parameters, including multi-select values, imported Grafana variable tokens, and system-managed label aliases. The system MUST NOT resolve the APO data group through dashboard variable substitution.

#### Scenario: Query references variable
- **WHEN** a widget query references an enabled dashboard-local variable
- **THEN** the system resolves the variable value before backend query execution

#### Scenario: Query can use semantic variable
- **WHEN** a widget query contains a label that matches an enabled semantic variable's alias mapping
- **THEN** the system resolves the selected variable values into a matcher for that query's matching label before backend query execution

#### Scenario: Query cannot use semantic variable
- **WHEN** a widget query does not contain any label matching an enabled semantic variable
- **THEN** the system does not inject that variable into the widget query

#### Scenario: Variable has multiple selected values
- **WHEN** a viewer selects multiple values for a dashboard variable
- **THEN** the system resolves the selection into a query matcher that includes all selected values for each applicable label alias

#### Scenario: Variable maps multiple label aliases
- **WHEN** a dashboard variable maps to multiple metric labels with the same meaning
- **THEN** the system applies the selected variable values to each relevant metric query using that query's matching label name

#### Scenario: Imported query uses Grafana variable syntax
- **WHEN** an imported widget query contains `$name`, `${name}`, `${name:regex}`, `${name:raw}`, or `{{ name }}`
- **THEN** the backend resolves those tokens before Prometheus-compatible query execution, with regex format escaping regular-expression metacharacters except the imported all value

### Requirement: Viewer Variable Controls
The system SHALL display dashboard variable selectors on the viewing page for enabled dashboard-local variables, static option variables, and label-value variables, and refresh affected widgets when variable values change. The system MUST NOT display data group in the dashboard variable selector area.

#### Scenario: Viewer changes variable value
- **WHEN** a viewer selects a different variable value
- **THEN** the system refreshes widgets affected by that dashboard-local variable using the new value

#### Scenario: Viewer selects multiple values
- **WHEN** a viewer selects multiple values in a dashboard variable selector
- **THEN** the system refreshes affected widgets using all selected values

#### Scenario: Viewer has disabled variable
- **WHEN** a semantic variable is not enabled for the dashboard
- **THEN** the viewer does not display a selector for that variable and does not apply it to widget queries

#### Scenario: Viewer changes data group
- **WHEN** a viewer changes the APO top-level data group selector
- **THEN** the system refreshes candidate values and widgets using the selected top-level data group rather than a dashboard variable

### Requirement: Permission-Filtered Candidate Values
The system MUST filter variable candidate values by the current top-level APO data group, the current user's data group permissions, the dashboard's widget metric labels, and already selected upstream variable values.

#### Scenario: User views variable candidates
- **WHEN** a user opens variable candidate values
- **THEN** the system returns only candidates allowed by the current top-level APO data group and the user's data group permissions

#### Scenario: User has no candidates
- **WHEN** a user has no permitted candidate values for a dashboard variable
- **THEN** the system displays a no-available-data or no-permission state

#### Scenario: Namespace filters service candidates
- **WHEN** a viewer selects one or more namespace values and then opens service variable candidates
- **THEN** the system returns only service candidates associated with the selected namespace values and permitted by the current top-level data group

#### Scenario: Standard resource variable uses system resource source
- **WHEN** a dashboard variable represents a standard APO resource such as cluster, namespace, service, endpoint, pod, or instance
- **THEN** the system obtains candidate values from the existing APO data group/resource filter source for the current data group and time range

#### Scenario: Custom label variable uses dashboard variable endpoint
- **WHEN** a dashboard variable is a custom label-value variable or imported Grafana label variable
- **THEN** the system obtains candidate values from the custom dashboard variable endpoint using the variable query, current upstream variable values, and current data group context

#### Scenario: Candidates are limited to dashboard labels
- **WHEN** a viewer opens candidates for a dashboard variable resolved through the custom dashboard variable endpoint
- **THEN** the system returns values from labels used by the current dashboard's widget metric queries or variable query, not unrelated global labels

### Requirement: Per-Viewer Variable Results
The system SHALL allow different users viewing the same dashboard to receive different variable candidates according to their permissions and current top-level data group selection.

#### Scenario: Two users view same dashboard
- **WHEN** two users with different data group permissions open the same dashboard
- **THEN** each user sees variable candidates filtered to their own data permissions and current top-level data group selection

### Requirement: Variable Label Discovery
The system SHALL expose label metadata needed to configure dashboard variables from the dashboard's widget metric queries, excluding labels that are controlled by APO global page context.

#### Scenario: Creator opens variable configuration
- **WHEN** the dashboard contains widget metric queries with labels matching known semantic variable aliases
- **THEN** the system shows selectable semantic variable suggestions for those labels except data group

#### Scenario: Dashboard has no metric labels
- **WHEN** no widget metric query can provide label metadata
- **THEN** the system explains that variables become available after configuring widget metric queries

#### Scenario: Dashboard has unknown labels
- **WHEN** widget metric queries contain labels that do not match supported semantic variable aliases
- **THEN** the system does not expose those labels as standard resource variables
