## ADDED Requirements

### Requirement: Built-In Templates
The system SHALL provide built-in dashboard templates for service health, infrastructure resources, and blank dashboard creation in MVP.

#### Scenario: User opens template selection
- **WHEN** a user with create-dashboard permission starts dashboard creation
- **THEN** the system displays service health, infrastructure resources, and blank template options

### Requirement: Template Contents
Built-in templates SHALL include default layout, widgets, query modes, query configuration, and variables.

#### Scenario: Dashboard is created from template
- **WHEN** a user creates a dashboard from a built-in template
- **THEN** the new dashboard includes the template's default layout, widgets, queries, and variables

### Requirement: Template Editability
The system SHALL allow users to edit dashboards created from built-in templates without changing the built-in template definition.

#### Scenario: User edits template-created dashboard
- **WHEN** a creator edits a dashboard created from a template
- **THEN** only that dashboard changes and the built-in template remains unchanged

### Requirement: Template Upgrade Isolation
The system MUST NOT mutate existing user dashboards when built-in templates are upgraded in a later release.

#### Scenario: Template changes after dashboard creation
- **WHEN** a built-in template definition changes after a user created a dashboard from it
- **THEN** the user's existing dashboard configuration remains unchanged

### Requirement: Preset Query Catalog
The system SHALL ship the first preset query catalog with service, instance/Pod, and infrastructure metric scenarios.

#### Scenario: Preset query catalog is loaded
- **WHEN** a user configures a preset query
- **THEN** the system provides service metrics, instance/Pod metrics, and infrastructure metric preset categories
