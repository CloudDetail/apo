## MODIFIED Requirements

### Requirement: Native Grid Editor
The system SHALL provide an APO-native fixed two-column dashboard editor with a top toolbar, central grid canvas, and right-side panel that switches between a widget palette and selected-widget configuration.

#### Scenario: Creator opens editor
- **WHEN** the dashboard creator opens edit mode
- **THEN** the system displays the native editor without embedding a Grafana iframe

#### Scenario: No widget is selected
- **WHEN** the creator has no widget selected in the editor
- **THEN** the right-side panel displays all supported widget types with widget names and visual thumbnails that match each widget type's runtime rendering shape

#### Scenario: Widget is selected
- **WHEN** the creator selects an existing widget
- **THEN** the right-side panel displays configuration for that selected widget only

### Requirement: Grid Layout Editing
The system SHALL allow creators to drag widgets between fixed cells in a two-column, multi-row grid, snap widgets to valid cells, avoid overlapping widgets, and automatically compact layout gaps between widgets. The system MUST NOT allow creators to resize widgets in the grid canvas or edit widget size in the configuration panel.

#### Scenario: Creator moves a widget
- **WHEN** the creator drags a widget to a new valid grid position
- **THEN** the system snaps the widget into the target fixed cell, updates the widget layout coordinates, avoids overlap, and compacts the grid so no empty cell remains between placed widgets in row-major order

#### Scenario: Creator drags to a lower row
- **WHEN** the creator drags a widget below the first row in the editor grid
- **THEN** the system calculates the target row and column from the pointer position within the two-column multi-row grid

#### Scenario: Creator drags over an occupied cell
- **WHEN** the creator drags a widget over a cell that cannot accept the widget
- **THEN** the system prevents invalid placement and keeps the previous valid layout until a valid cell is chosen

#### Scenario: Creator attempts to resize a widget
- **WHEN** the creator views or interacts with a dashboard widget in edit mode
- **THEN** the system provides no resize handles, width controls, or height controls for that widget

#### Scenario: New widget uses compact cell ratio
- **WHEN** the creator adds a new widget
- **THEN** the widget is placed in a fixed cell with an approximate 6:4 visual ratio

### Requirement: Widget Lifecycle
The system SHALL allow creators to add, delete, and copy widgets within a dashboard.

#### Scenario: Creator adds widget from palette
- **WHEN** the creator selects a widget type from the right-side palette
- **THEN** the system creates a real widget instance in placement mode that can be dragged into the dashboard grid

#### Scenario: Creator places new widget
- **WHEN** the creator drops the new widget into a valid fixed cell
- **THEN** the system inserts the widget at that cell, compacts the layout, and selects it for configuration

#### Scenario: Creator deletes widget
- **WHEN** the creator deletes an existing widget
- **THEN** the system removes the widget from the dashboard, frees its fixed cell, and compacts remaining widgets so no gap remains between them

#### Scenario: Creator copies widget
- **WHEN** the creator copies an existing widget
- **THEN** the system creates a duplicate widget with the same supported configuration in an available fixed cell

### Requirement: MVP Widget Types
The system SHALL support line chart, area chart, bar chart, and table widgets. The system MUST NOT offer single-value metric widgets or TopN ranking widgets for new widget creation or chart-type selection.

#### Scenario: Creator selects chart type
- **WHEN** the creator selects one of the supported widget types
- **THEN** the system renders the widget using that type and exposes compatible configuration options

#### Scenario: Creator opens widget palette
- **WHEN** the creator opens the widget palette
- **THEN** the palette shows line chart, area chart, bar chart, and table options only

#### Scenario: Creator edits historical TopN widget
- **WHEN** the creator opens a dashboard containing a historical TopN widget
- **THEN** the system treats the widget as a table-compatible widget

#### Scenario: Creator edits historical single-value widget
- **WHEN** the creator opens a dashboard containing a historical single-value widget
- **THEN** the system marks the widget type as removed or unsupported and provides a path to delete it or change it to a supported widget type

### Requirement: Widget Configuration
The system SHALL support widget title, description, chart type, query configuration, unit, and decimal places. The system MUST NOT expose widget size editing, literal dashboard variable fields such as `$service`, or legend, threshold, empty-state text, or error-state text fields unless those fields produce visible runtime behavior.

#### Scenario: Creator edits widget configuration
- **WHEN** the creator changes supported widget configuration in the side panel
- **THEN** the system updates the selected widget preview and marks the dashboard as unsaved

#### Scenario: Creator opens widget configuration
- **WHEN** the creator opens the selected-widget configuration panel
- **THEN** unsupported or non-effective fields are not displayed

#### Scenario: Creator configures unit
- **WHEN** the creator sets a widget unit
- **THEN** the widget preview uses that unit in chart axes, chart tooltip values, or table value cells according to the widget type

### Requirement: Drag Feedback
The system SHALL provide drag feedback that visually moves only the widget being dragged and its placement indicator.

#### Scenario: Creator drags a widget
- **WHEN** the creator drags a widget in the dashboard editor
- **THEN** toolbar, panel, and non-dragged widget header UI remain visually stable while only the dragged widget or placeholder follows the pointer

## ADDED Requirements

### Requirement: Variable Deletion In Editor
The system SHALL allow creators to delete enabled dashboard variables from the dashboard editor.

#### Scenario: Creator deletes variable
- **WHEN** the creator deletes an enabled dashboard variable
- **THEN** the system removes the variable from editor state, marks the dashboard as unsaved, and excludes the variable from the next save payload
