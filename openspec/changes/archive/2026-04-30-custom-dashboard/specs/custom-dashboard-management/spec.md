## ADDED Requirements

### Requirement: Dashboard Navigation And Listing
The system SHALL provide a first-level "My Dashboards" navigation entry that opens a list of dashboards the current user has permission to view.

#### Scenario: User opens dashboard list
- **WHEN** a signed-in user opens "My Dashboards"
- **THEN** the system displays only dashboards whose visibility allows that user to view them

#### Scenario: Existing dashboards remain available
- **WHEN** the custom dashboard capability is enabled
- **THEN** existing system dashboard entries remain available in their current locations

### Requirement: Dashboard List Metadata
The system SHALL display dashboard name, description, creator, visibility scope, updated time, last visited time, and permitted actions in the dashboard list.

#### Scenario: List shows dashboard metadata
- **WHEN** a user views the dashboard list
- **THEN** each row includes the dashboard metadata and only the actions allowed for that user

### Requirement: Dashboard Creation
The system SHALL allow users with create-dashboard permission to create a dashboard from a built-in template or from a blank dashboard.

#### Scenario: Authorized user creates from template
- **WHEN** a user with create-dashboard permission selects a built-in template and confirms creation
- **THEN** the system creates a new user-owned dashboard copied from the selected template

#### Scenario: Authorized user creates blank dashboard
- **WHEN** a user with create-dashboard permission selects blank creation
- **THEN** the system creates a new dashboard with an empty grid, default time range, and default system variables

### Requirement: Dashboard Copy
The system SHALL allow users with create-dashboard permission to copy any dashboard they can view.

#### Scenario: User copies a visible dashboard
- **WHEN** a user with create-dashboard permission copies a dashboard they can view
- **THEN** the system creates a new dashboard owned by that user with name "<original name> - Copy" and personal visibility by default

#### Scenario: User without create permission cannot copy
- **WHEN** a user without create-dashboard permission views a dashboard list row
- **THEN** the system does not allow copying that dashboard

### Requirement: Dashboard Deletion
The system SHALL allow only the dashboard creator to delete their dashboard after confirmation.

#### Scenario: Creator deletes dashboard
- **WHEN** the dashboard creator confirms deletion
- **THEN** the system deletes the dashboard and removes its view entry immediately

#### Scenario: Deleted dashboard is opened
- **WHEN** a user refreshes or opens a deleted dashboard URL
- **THEN** the system displays a dashboard-not-found state

### Requirement: Owner Editing Rules
The system SHALL allow only the dashboard creator to edit dashboard metadata, visibility, variables, widgets, layout, query configuration, and display configuration.

#### Scenario: Non-owner views dashboard
- **WHEN** a user who is not the creator opens a visible dashboard
- **THEN** the system allows viewing but does not allow editing or deleting
