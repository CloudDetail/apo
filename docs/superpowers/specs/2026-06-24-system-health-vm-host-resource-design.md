# System Health VM Host Resource Compatibility Design

## Context

The system health detail page renders `resource` and `resourceShift` from
`system.metricDetails`. Today those details are populated by container-only
Prometheus queries in `backend/pkg/api/systemhealth/func_getoverview.go`.

Current resource data depends on Kubernetes labels and cAdvisor metrics:

- `container_cpu_usage_seconds_total`
- `container_spec_cpu_quota`
- `container_spec_cpu_period`
- `container_memory_working_set_bytes`
- `container_spec_memory_limit_bytes`
- `pod`, `namespace`, and `container` labels

VM applications do not have pod or container resource metrics. The repository
already models VM service instances as `serviceName + nodeName + pid`, but only
host-level CPU and memory metrics are available for this requirement.

## Goal

Support VM applications in system health overview/detail resource indicators
without changing the frontend API contract.

The existing response fields remain:

- `metricDetails.resource`
- `metricDetails.resourceShift`
- `metrics.resource`
- `metrics.resourceShift`

For VM-only services, these fields represent host-level resource pressure for
the node where the service instance runs. The UI must make that source explicit
through `note` and `source`.

## Non-Goals

- Do not infer process-level CPU or memory usage from host-level metrics.
- Do not change frontend component behavior or response shape.
- Do not change existing container resource semantics.
- Do not introduce new alert thresholds for VM resources in this pass.

## Data Semantics

Container services continue to use the existing pod/container resource path.

VM services use host-level fallback when no pod/container resource data is
available. A VM service instance is identified by:

```text
PodName == "" && ContainerId == "" && NodeName != "" && Pid > 1
```

Because only host-level metrics are available, all services on the same
`node_name` share the same host resource waterline and host resource shift.
This is intentionally presented as "service host resource pressure", not
"service process resource usage".

## Resource Waterline

For VM services:

```text
service.resource = max(host CPU waterline, host memory waterline)
```

For a service with multiple VM instances on multiple hosts:

```text
service.resource = max(resource waterline of all related hosts)
```

For system-level details:

```text
system.resource = max(resource waterline of all hosts related to services in the system)
```

Thresholds stay aligned with current container behavior:

```text
>= 90% red
>= 70% yellow
< 70% green
```

The detail should include:

```json
{
  "source": "host/<node_name>",
  "note": "Current host CPU resource waterline"
}
```

or:

```json
{
  "source": "host/<node_name>",
  "note": "Current host memory resource waterline"
}
```

## Resource Shift

For VM services, compute the host-level 24h baseline shift:

```text
cpuRatio = current window host CPU / same window 24h ago host CPU
memRatio = current window host memory / same window 24h ago host memory
resourceShift = max(cpuRatio, 1 / cpuRatio, memRatio, 1 / memRatio)
shiftPercent = (resourceShift - 1) * 100
```

Thresholds stay aligned with current container behavior:

```text
>= 60% red
>= 30% yellow
< 30% green
```

For a service with multiple VM hosts, choose the largest shift. For system-level
trend data, choose the largest shift per timestamp across all related VM hosts
and existing container shift series.

## Backend Architecture

Add VM host resource collection beside the existing container resource
collection in `querySystemHealthExtraMetrics`.

Recommended new helpers:

```go
type systemHealthVMHostResourceScope struct {
    NodesByService map[string]map[string]struct{}
    SystemNodes    map[string]struct{}
}

func buildSystemHealthVMHostResourceScope(
    ctx core.Context,
    startTime time.Time,
    endTime time.Time,
    groupFilter prometheus.PQLFilter,
    services []response.SystemHealthService,
) (systemHealthVMHostResourceScope, error)

func buildSystemHealthVMHostCPUWaterlinePQL(nodes []string, window string) string
func buildSystemHealthVMHostMemoryWaterlinePQL(nodes []string, window string) string
func buildSystemHealthVMHostResourceShiftPQL(nodes []string, window string) string
```

The scope helper should use existing service instance lookup where possible:

- Query service instances for the services in the selected data group.
- Keep instances with VM shape only.
- Build service-to-node and system-node sets.

The query helpers should return Prometheus results labeled by `node_name`. The
conversion helpers then map host results back to services through
`NodesByService`.

## Merge Behavior

Existing container results keep priority for container applications.

For each service:

1. Build the existing container `resource` and `resourceShift` details.
2. Build VM host fallback details when VM host nodes are available.
3. Merge by worse health status.
4. If statuses tie, keep the higher numeric representative value.
5. Preserve source clarity in `note` and `source`.

For system-level details, aggregate across both container and VM host sources by
worst status and highest representative value.

## Error Handling

- If host metrics are missing, keep the current `gray` fallback.
- If service instance lookup fails, log a warning and keep existing container
  behavior.
- If a 24h baseline denominator is missing or zero, omit that shift sample
  rather than emitting infinite noise.
- VM fallback must not cause the whole overview/detail API to fail.

## Tests

Add focused tests in `backend/pkg/api/systemhealth/func_getoverview_test.go`:

1. VM-only service maps `node_name` host CPU/memory waterline into
   `metricDetails.resource`.
2. Multiple services on the same host receive the same host-level detail with a
   host-level note/source.
3. Mixed container and VM services aggregate the worse `resource` status.
4. VM host resource shift converts fold-change to percent and applies existing
   30/60 thresholds.
5. System resource shift trend picks the max host/container shift per timestamp.
6. Missing host metrics leaves resource details gray and does not break the API.

## Open Contract

The API contract remains unchanged. The only user-visible semantic difference
is that VM service resource notes and sources explicitly say the data is
host-level.
