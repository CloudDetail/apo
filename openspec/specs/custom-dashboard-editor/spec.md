## Purpose
This specification defines the custom dashboard capability behavior.
## Requirements
### Requirement: Native Grid Editor
The system SHALL provide an APO-native fixed two-column dashboard editor with a top toolbar, central grid canvas, and right-side panel that switches between a widget palette and selected-widget configuration.

#### Scenario: Creator opens editor
- **WHEN** the dashboard creator opens edit mode
- **THEN** the system displays the native editor without embedding a Grafana iframe

#### Scenario: No widget is selected
- **WHEN** the creator has no widget selected in the editor
- **THEN** the right-side panel displays all supported widget types with widget names and mocked visual previews

#### Scenario: Widget is selected
- **WHEN** the creator selects an existing widget
- **THEN** the right-side panel displays configuration for that selected widget only

### Requirement: Grid Layout Editing
The system SHALL allow creators to drag widgets between fixed cells in a two-column grid, snap widgets to the nearest valid cell, and avoid overlapping widgets. The system MUST NOT allow creators to resize widgets in the grid canvas or edit widget size in the configuration panel.

#### Scenario: Creator moves a widget
- **WHEN** the creator drags a widget to a new valid grid position
- **THEN** the system snaps the widget into the target fixed cell and updates the widget layout coordinates without overlapping other widgets

#### Scenario: Creator drags over an occupied cell
- **WHEN** the creator drags a widget over a cell that cannot accept the widget
- **THEN** the system prevents invalid placement and keeps the previous valid layout until a valid cell is chosen

#### Scenario: Creator attempts to resize a widget
- **WHEN** the creator views or interacts with a dashboard widget in edit mode
- **THEN** the system provides no resize handles, width controls, or height controls for that widget

### Requirement: Widget Lifecycle
The system SHALL allow creators to add, delete, and copy widgets within a dashboard.

#### Scenario: Creator adds widget from palette
- **WHEN** the creator selects a widget type from the right-side palette
- **THEN** the system creates a real widget instance in placement mode that can be dragged into the dashboard grid

#### Scenario: Creator places new widget
- **WHEN** the creator drops the new widget into a valid fixed cell
- **THEN** the system inserts the widget at that cell and selects it for configuration

#### Scenario: Creator deletes widget
- **WHEN** the creator deletes an existing widget
- **THEN** the system removes the widget from the dashboard and frees its fixed cell for later placement

#### Scenario: Creator copies widget
- **WHEN** the creator copies an existing widget
- **THEN** the system creates a duplicate widget with the same supported configuration in an available fixed cell

### Requirement: MVP Widget Types
The system SHALL support line chart, area chart, bar chart, single value, table, and TopN ranking widgets.

#### Scenario: Creator selects chart type
- **WHEN** the creator selects one of the MVP widget types
- **THEN** the system renders the widget using that type and exposes compatible configuration options

### Requirement: Widget Configuration
The system SHALL support widget title, description, chart type, query configuration, unit, and decimal places. The system MUST NOT expose widget size editing, literal dashboard variable fields such as `$service`, or legend, threshold, empty-state text, or error-state text fields unless those fields produce visible runtime behavior.

#### Scenario: Creator edits widget configuration
- **WHEN** the creator changes supported widget configuration in the side panel
- **THEN** the system updates the selected widget preview and marks the dashboard as unsaved

#### Scenario: Creator opens widget configuration
- **WHEN** the creator opens the selected-widget configuration panel
- **THEN** unsupported or non-effective fields are not displayed

### Requirement: Save Behavior
The system SHALL save dashboard edits immediately when the creator clicks save and SHALL retain editor state when saving fails.

#### Scenario: Save succeeds
- **WHEN** the creator saves a valid edited dashboard
- **THEN** the system persists the latest metadata, variables, widgets, layout, queries, and display configuration immediately

#### Scenario: Save fails
- **WHEN** dashboard save fails
- **THEN** the system keeps the current editor state and displays the failure reason with a retry path

### Requirement: Unsaved Change Prompt
The system SHALL warn creators before leaving the editor with unsaved changes and SHALL keep preview mode reversible.

#### Scenario: Creator navigates away with unsaved changes
- **WHEN** the creator attempts to leave the editor after unsaved modifications
- **THEN** the system prompts the creator to confirm leaving or continue editing

#### Scenario: Creator previews dashboard
- **WHEN** the creator clicks preview from the editor
- **THEN** the system displays a dashboard preview without discarding editor state and provides a path back to editing

### Requirement: Drag Feedback
The system SHALL provide drag feedback that visually moves only the widget being dragged and its placement indicator.

#### Scenario: Creator drags a widget
- **WHEN** the creator drags a widget in the dashboard editor
- **THEN** toolbar, panel, and non-dragged widget header UI remain visually stable while only the dragged widget or placeholder follows the pointer
