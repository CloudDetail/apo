## ADDED Requirements

### Requirement: Native Grid Editor
The system SHALL provide an APO-native grid-based dashboard editor with a top toolbar, widget entry controls, central grid canvas, and right-side configuration panel.

#### Scenario: Creator opens editor
- **WHEN** the dashboard creator opens edit mode
- **THEN** the system displays the native editor without embedding a Grafana iframe

### Requirement: Grid Layout Editing
The system SHALL allow creators to drag widgets, resize widgets, align widgets to the grid, and avoid overlapping widgets.

#### Scenario: Creator moves a widget
- **WHEN** the creator drags a widget to a new valid grid position
- **THEN** the system updates the widget layout coordinates without overlapping other widgets

#### Scenario: Creator resizes a widget
- **WHEN** the creator resizes a widget above its minimum dimensions
- **THEN** the system updates the widget width and height and keeps chart content readable

### Requirement: Widget Lifecycle
The system SHALL allow creators to add, delete, and copy widgets within a dashboard.

#### Scenario: Creator adds widget
- **WHEN** the creator adds a widget
- **THEN** the system inserts the widget into an available grid location and selects it for configuration

#### Scenario: Creator copies widget
- **WHEN** the creator copies an existing widget
- **THEN** the system creates a duplicate widget with the same configuration in an available grid location

### Requirement: MVP Widget Types
The system SHALL support line chart, area chart, bar chart, single value, table, and TopN ranking widgets.

#### Scenario: Creator selects chart type
- **WHEN** the creator selects one of the MVP widget types
- **THEN** the system renders the widget using that type and exposes compatible configuration options

### Requirement: Widget Configuration
The system SHALL support widget title, description, chart type, query configuration, unit, decimal places, legend configuration, threshold configuration, empty state, and error state.

#### Scenario: Creator edits widget configuration
- **WHEN** the creator changes widget configuration in the side panel
- **THEN** the system updates the selected widget preview and marks the dashboard as unsaved

### Requirement: Save Behavior
The system SHALL save dashboard edits immediately when the creator clicks save and SHALL retain editor state when saving fails.

#### Scenario: Save succeeds
- **WHEN** the creator saves a valid edited dashboard
- **THEN** the system persists the latest metadata, variables, widgets, layout, queries, and display configuration immediately

#### Scenario: Save fails
- **WHEN** dashboard save fails
- **THEN** the system keeps the current editor state and displays the failure reason with a retry path

### Requirement: Unsaved Change Prompt
The system SHALL warn creators before leaving the editor with unsaved changes.

#### Scenario: Creator navigates away with unsaved changes
- **WHEN** the creator attempts to leave the editor after unsaved modifications
- **THEN** the system prompts the creator to confirm leaving or continue editing
