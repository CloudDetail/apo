# System Health Permission and Report Retention

## Scope

This design defines the authorization boundary for system health reports and global score thresholds, and bounds persisted report history growth.

## Authorization

- System health overview, trend, pending issue, report generation, and report history are scoped to the caller's visible data groups.
- Data group `0` represents the aggregate root and is available only when the caller has root data-group permission.
- A report record is authorized through its persisted `group_id`; knowing a `record_id` does not grant access.
- The score threshold configuration is a global singleton. Only users with the `admin` role may update it. The backend is authoritative; the frontend hides the control for other roles.
- Authorization failures return HTTP 403 and do not fall back to broader data.

## Report Persistence

- Report generation keeps its existing behavior: every successfully generated report is saved as an independent history record.
- Report history has a 30-day retention period. A save deletes expired records for the same data group in the same database transaction.
- History list and detail queries exclude expired records immediately, including when a data group has not generated another report to trigger physical cleanup.
- Each data group retains at most 500 non-expired records as a secondary guard against abnormal request volume. After a successful save, the oldest excess records for that group are deleted in the same database transaction.
- Listing report history requires an explicit non-empty set of authorized data-group IDs. An empty set returns an empty result.

## Failure Behavior

- Permission lookup errors fail closed and are returned as authorization failures.
- Report persistence remains non-critical to report generation: save or retention cleanup failures are logged and the generated report is still returned.
