## ADDED Requirements

### Requirement: Viewport-Aware Widget Query Scheduling
The system SHALL schedule custom dashboard widget preview queries according to widget viewport eligibility in custom dashboard view and edit pages. A widget is eligible when it is visible, near the viewport within the configured preload margin, explicitly prioritized by user action, or required for fullscreen display. The system MUST start a widget preview only when the current preview is missing, stale, failed and explicitly retried, or uses a changed request signature. Except when viewport observation is unavailable, the system MUST NOT request previews for offscreen widgets solely because the dashboard opened or a shared filter changed.

#### Scenario: Dashboard opens with many widgets
- **WHEN** a user opens a custom dashboard view or edit page containing widgets above and below the initial viewport
- **THEN** the system schedules preview queries for visible and near-visible widgets before offscreen widgets

#### Scenario: User scrolls toward deferred widgets
- **WHEN** a deferred widget enters the preload margin or viewport
- **THEN** the system schedules that widget's preview query using the current global time range and dashboard variable selections

#### Scenario: Loaded widget re-enters viewport
- **WHEN** a widget re-enters the preload margin or viewport and already has a successful or no-data preview for the current request signature
- **THEN** the system does not schedule another preview query for that widget unless the widget is stale, the request signature changed, or the viewer explicitly retries a failed request

#### Scenario: Viewport observation is unavailable
- **WHEN** the browser cannot provide widget viewport observation
- **THEN** the system treats all dashboard widgets as eligible and schedules widget previews through the same dashboard-level concurrency limit instead of bypassing request scheduling

### Requirement: Dashboard-Level Preview Concurrency
The system SHALL limit the number of concurrent widget preview requests started by one custom dashboard page instance to a default maximum of four. Queued preview work MUST keep only the newest request signature per widget, and stale queued or in-flight work MUST NOT overwrite newer widget state.

#### Scenario: Many widgets become eligible together
- **WHEN** more eligible widgets exist than the dashboard preview concurrency limit allows
- **THEN** the system starts only up to the limit and queues the remaining eligible widget preview requests

#### Scenario: Default concurrency is enforced
- **WHEN** at least five widget preview requests are eligible for one dashboard viewer
- **THEN** the system has no more than four widget preview requests in flight at the same time for that page instance

#### Scenario: Request context changes while work is queued
- **WHEN** a queued widget preview request is superseded by a newer request signature
- **THEN** the system skips the superseded queued request and keeps only the newest work for that widget

#### Scenario: Stale response returns
- **WHEN** an in-flight preview response returns after the widget's newest request signature has changed
- **THEN** the system ignores that stale response and preserves the newer widget state

### Requirement: Refresh Invalidation For Deferred Widgets
The system SHALL handle refresh triggers by invalidating affected widgets and scheduling only eligible affected widgets immediately. Affected widgets that are not eligible MUST remain stale until they become eligible.

#### Scenario: Global time range changes
- **WHEN** the user changes the APO global time range
- **THEN** the system marks dashboard widgets stale, schedules eligible widgets with the new time range, and defers offscreen widgets until they become eligible

#### Scenario: Dashboard variable changes
- **WHEN** the user changes a dashboard variable
- **THEN** the system marks widgets affected by that variable stale, schedules eligible affected widgets, and does not refresh unaffected widgets solely because the variable changed

#### Scenario: Manual refresh is requested
- **WHEN** the viewer manually refreshes the dashboard
- **THEN** the system marks all widgets stale and schedules eligible widgets through the dashboard-level preview scheduler

### Requirement: Deferred Widget States
The system SHALL distinguish widgets that have not been queried yet from widgets that queried successfully with no data. Deferred, queued, loading, stale, successful, no-data, and failed widget states MUST remain scoped to the individual widget.

#### Scenario: Widget has not reached the viewport
- **WHEN** a widget has not yet been eligible for preview query dispatch
- **THEN** the widget displays a deferred or pending state rather than a no-data state

#### Scenario: Stale widget keeps prior data
- **WHEN** a widget has previous preview data and becomes stale due to a refresh trigger
- **THEN** the widget continues displaying the previous preview while indicating that a refreshed preview is pending or queued

#### Scenario: Widget query fails
- **WHEN** a scheduled widget preview query fails
- **THEN** only that widget displays the query failure state and other widget query states remain independent

#### Scenario: Viewer retries failed widget
- **WHEN** a viewer retries a failed widget preview
- **THEN** the system schedules only that widget using the current dashboard request context and does not refresh unrelated widgets

### Requirement: Fullscreen Widget Priority
The system SHALL prioritize a widget preview request when a viewer opens that widget in fullscreen and the current preview is missing or stale.

#### Scenario: Viewer opens unloaded widget fullscreen
- **WHEN** a viewer opens fullscreen for a widget whose preview has not loaded
- **THEN** the system immediately schedules that widget's preview ahead of ordinary grid preload work

#### Scenario: Viewer opens stale widget fullscreen
- **WHEN** a viewer opens fullscreen for a widget whose current preview is stale
- **THEN** the system schedules a preview using the current dashboard request context and displays fullscreen loading or error state until the request completes
