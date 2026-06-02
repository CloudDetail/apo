## MODIFIED Requirements

### Requirement: Dashboard List Metadata
The system SHALL display dashboard name, description, creator, visibility scope, and permitted actions in the dashboard list, and MUST NOT display updated time or last visited time columns.

#### Scenario: List shows dashboard metadata
- **WHEN** a user views the dashboard list
- **THEN** each row includes the dashboard metadata and only the actions allowed for that user

#### Scenario: List hides removed time columns
- **WHEN** a user views the dashboard list
- **THEN** the dashboard list does not include updated time or last visited time columns

#### Scenario: Long description is available
- **WHEN** a dashboard description is too long to fit in the list cell
- **THEN** the system provides a hover tooltip, popover, or equivalent detail display that shows the complete description

### Requirement: Dashboard Creation
The system SHALL allow users with create-dashboard permission to create a dashboard from a built-in template or from a blank dashboard, and each newly created dashboard SHALL become an independent menu item governed by the existing menu, route, and role permission control system.

#### Scenario: Authorized user creates from template
- **WHEN** a user with create-dashboard permission selects a built-in template and confirms creation
- **THEN** the system creates a new user-owned dashboard copied from the selected template

#### Scenario: Authorized user creates blank dashboard
- **WHEN** a user with create-dashboard permission selects blank creation
- **THEN** the system creates a new dashboard with an empty fixed-cell grid and default label-based variable configuration

#### Scenario: Created dashboard appears in menu
- **WHEN** dashboard creation succeeds
- **THEN** the system registers the dashboard as a menu item in the current permission control system and makes it visible only to users whose menu, route, role, and dashboard visibility permissions allow it

#### Scenario: Dashboard menu item is opened
- **WHEN** a user opens a dashboard through its menu item
- **THEN** the system enforces the same route permission, dashboard visibility, and metric data permission checks used by the dashboard list and viewer

## ADDED Requirements

### Requirement: Dashboard Menu Lifecycle
The system SHALL keep custom dashboard menu items synchronized with dashboard lifecycle changes and with the existing permission control records that determine menu visibility and route access.

#### Scenario: Dashboard metadata changes
- **WHEN** the dashboard creator saves a changed dashboard name or visibility
- **THEN** the corresponding menu item and permission-control metadata update their label, route, and visibility rules consistently

#### Scenario: Dashboard is deleted
- **WHEN** the dashboard creator deletes a dashboard
- **THEN** the system removes the corresponding menu item or makes it unavailable immediately

#### Scenario: Dashboard is copied
- **WHEN** a user copies a dashboard successfully
- **THEN** the copied dashboard receives its own menu item according to the copied dashboard visibility

#### Scenario: User lacks menu or route permission
- **WHEN** a user lacks permission for a custom dashboard menu item or route
- **THEN** the system does not show the menu item and rejects direct route access even if the dashboard URL is known
