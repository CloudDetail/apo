# System Health VM Host Resource Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add host-level VM resource support for `resource` and `resourceShift` while preserving the existing system health overview/detail API contract.

**Architecture:** Keep the current container resource path unchanged and add a VM host fallback path in `backend/pkg/api/systemhealth/func_getoverview.go`. VM services are mapped from trace service instances to `node_name`; host CPU and memory are queried from Node Exporter metrics labeled by `instance_name`, normalized to `node_name` with PromQL `label_replace`, and then merged into existing `metricDetails.resource` and `metricDetails.resourceShift` semantics.

**Tech Stack:** Go, Prometheus PromQL, existing `prometheus.Repo`, existing system health response models, Go unit tests.

---

## File Structure

- Modify: `backend/pkg/api/systemhealth/func_getoverview.go`
  - Add VM host resource scope helpers.
  - Add host CPU/memory waterline PromQL builders.
  - Add host 24h resource shift PromQL builder.
  - Add host result conversion helpers.
  - Merge VM host details with existing container details.
- Modify: `backend/pkg/api/systemhealth/func_getoverview_test.go`
  - Add unit tests for VM host scope selection.
  - Add unit tests for host PromQL builders.
  - Add unit tests for host waterline and shift conversion.
  - Add unit tests for mixed container/VM merging and system trend aggregation.

No frontend files are required because the response contract remains unchanged.

## Important Existing Behavior

- Container `resource` currently uses cAdvisor pod/container CPU and memory metrics.
- Container `resourceShift` currently compares current window CPU/memory with an `offset 24h` baseline.
- `applySystemHealthExtraMetrics` recalculates service scores, then recalculates the system aggregate, then applies system-level current values and trends.
- `systemHealthMaxResourceShiftSeries` already takes fold-change values and emits max shift percent per timestamp, so host shift trend results can be appended to `SystemResourceShiftTrend` after they are normalized to the same fold-change shape.

## PromQL Metric Contract

Use the same Node Exporter metrics already referenced in `backend/static/predefined-metrics/default/Node-Exporter-Full.json`:

- CPU: `node_cpu_seconds_total`
- Memory available: `node_memory_MemAvailable_bytes`
- Memory total: `node_memory_MemTotal_bytes`

Node Exporter identifies hosts with `instance_name`. Service instances identify VM hosts with `NodeName`, which corresponds to the same host name string. PromQL builders must use `label_replace(..., "node_name", "$1", "instance_name", "(.*)")` so existing Go conversion code can read `result.Metric.NodeName`.

---

### Task 1: Add VM Host Scope And PromQL Builder Tests

**Files:**
- Modify: `backend/pkg/api/systemhealth/func_getoverview_test.go`
- Modify: `backend/pkg/api/systemhealth/func_getoverview.go`

- [ ] **Step 1: Write failing tests for VM host scope and PromQL builders**

Add these tests near the existing resource waterline tests in `backend/pkg/api/systemhealth/func_getoverview_test.go`:

```go
func TestSystemHealthVMHostResourceScopeKeepsOnlyVMInstances(t *testing.T) {
	scope := systemHealthVMHostResourceScope{
		NodesByService: map[string]map[string]struct{}{},
		SystemNodes:    map[string]struct{}{},
	}

	scope.addInstance("order-api", &model.ServiceInstance{
		ServiceName: "order-api",
		NodeName:    "vm-a",
		Pid:         2201,
	})
	scope.addInstance("order-api", &model.ServiceInstance{
		ServiceName: "order-api",
		NodeName:    "vm-b",
		Pid:         2202,
	})
	scope.addInstance("order-api", &model.ServiceInstance{
		ServiceName: "order-api",
		NodeName:    "node-with-pod",
		PodName:     "order-api-0",
		Pid:         1,
	})
	scope.addInstance("order-worker", &model.ServiceInstance{
		ServiceName: "order-worker",
		NodeName:    "node-with-container",
		ContainerId: "container-1",
		Pid:         3310,
	})
	scope.addInstance("ignored", &model.ServiceInstance{
		ServiceName: "ignored",
		NodeName:    "pid-one",
		Pid:         1,
	})

	if !scope.hasNodes() {
		t.Fatalf("expected VM host scope to contain nodes")
	}
	if got := scope.sortedNodes(); !reflect.DeepEqual(got, []string{"vm-a", "vm-b"}) {
		t.Fatalf("expected only VM hosts, got %+v", got)
	}
	if got := scope.sortedServiceNodes("order-api"); !reflect.DeepEqual(got, []string{"vm-a", "vm-b"}) {
		t.Fatalf("expected order-api VM hosts, got %+v", got)
	}
	if got := scope.sortedServiceNodes("order-worker"); len(got) != 0 {
		t.Fatalf("expected container service to have no VM host nodes, got %+v", got)
	}
}

func TestBuildSystemHealthVMHostResourcePQLUsesNodeExporterInstanceName(t *testing.T) {
	nodes := []string{"vm-a", "vm-b"}

	cpuQuery := buildSystemHealthVMHostCPUWaterlinePQL(nodes, "15m", "")
	memoryQuery := buildSystemHealthVMHostMemoryWaterlinePQL(nodes, "15m", "")
	shiftQuery := buildSystemHealthVMHostResourceShiftPQL(nodes, "15m")

	for _, part := range []string{
		`node_cpu_seconds_total{instance_name=~"vm-a|vm-b",mode!~"idle|iowait|steal"}[15m]`,
		`label_replace(rate(`,
		`"node_name", "$1", "instance_name", "(.*)"`,
		`avg by (node_name)`,
	} {
		if !strings.Contains(cpuQuery, part) {
			t.Fatalf("expected VM host CPU query to contain %q, got %s", part, cpuQuery)
		}
	}

	for _, part := range []string{
		`node_memory_MemAvailable_bytes{instance_name=~"vm-a|vm-b"}[15m]`,
		`node_memory_MemTotal_bytes{instance_name=~"vm-a|vm-b"}[15m]`,
		`label_replace(avg_over_time(`,
		`"node_name", "$1", "instance_name", "(.*)"`,
		`max by (node_name)`,
	} {
		if !strings.Contains(memoryQuery, part) {
			t.Fatalf("expected VM host memory query to contain %q, got %s", part, memoryQuery)
		}
	}

	for _, part := range []string{
		`offset 24h`,
		`max by (node_name)`,
		`1 / (`,
		`node_cpu_seconds_total{instance_name=~"vm-a|vm-b",mode!~"idle|iowait|steal"}[15m]`,
		`node_memory_MemAvailable_bytes{instance_name=~"vm-a|vm-b"}[15m]`,
	} {
		if !strings.Contains(shiftQuery, part) {
			t.Fatalf("expected VM host shift query to contain %q, got %s", part, shiftQuery)
		}
	}
}
```

Also add `reflect` and `github.com/CloudDetail/apo/backend/pkg/model` to the test imports if they are not already present:

```go
import (
	"reflect"
	"strings"
	"testing"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```powershell
go test ./backend/pkg/api/systemhealth -run "TestSystemHealthVMHostResourceScopeKeepsOnlyVMInstances|TestBuildSystemHealthVMHostResourcePQLUsesNodeExporterInstanceName" -count=1
```

Expected: FAIL because `systemHealthVMHostResourceScope`, `buildSystemHealthVMHostCPUWaterlinePQL`, `buildSystemHealthVMHostMemoryWaterlinePQL`, and `buildSystemHealthVMHostResourceShiftPQL` are not defined.

- [ ] **Step 3: Add VM host scope and PQL builders**

Add this code in `backend/pkg/api/systemhealth/func_getoverview.go` near the existing resource PQL helpers:

```go
type systemHealthVMHostResourceScope struct {
	NodesByService map[string]map[string]struct{}
	SystemNodes    map[string]struct{}
}

func newSystemHealthVMHostResourceScope() systemHealthVMHostResourceScope {
	return systemHealthVMHostResourceScope{
		NodesByService: map[string]map[string]struct{}{},
		SystemNodes:    map[string]struct{}{},
	}
}

func (scope *systemHealthVMHostResourceScope) addInstance(serviceName string, instance *model.ServiceInstance) {
	if scope == nil || instance == nil {
		return
	}
	if serviceName == "" {
		serviceName = instance.ServiceName
	}
	if serviceName == "" || instance.PodName != "" || instance.ContainerId != "" || instance.NodeName == "" || instance.Pid <= 1 {
		return
	}
	if scope.NodesByService == nil {
		scope.NodesByService = map[string]map[string]struct{}{}
	}
	if scope.SystemNodes == nil {
		scope.SystemNodes = map[string]struct{}{}
	}
	if _, ok := scope.NodesByService[serviceName]; !ok {
		scope.NodesByService[serviceName] = map[string]struct{}{}
	}
	scope.NodesByService[serviceName][instance.NodeName] = struct{}{}
	scope.SystemNodes[instance.NodeName] = struct{}{}
}

func (scope systemHealthVMHostResourceScope) hasNodes() bool {
	return len(scope.SystemNodes) > 0
}

func (scope systemHealthVMHostResourceScope) sortedNodes() []string {
	nodes := make([]string, 0, len(scope.SystemNodes))
	for node := range scope.SystemNodes {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)
	return nodes
}

func (scope systemHealthVMHostResourceScope) sortedServiceNodes(serviceName string) []string {
	nodeSet := scope.NodesByService[serviceName]
	nodes := make([]string, 0, len(nodeSet))
	for node := range nodeSet {
		nodes = append(nodes, node)
	}
	sort.Strings(nodes)
	return nodes
}

func (h *handler) buildSystemHealthVMHostResourceScope(ctx core.Context, startTime time.Time, endTime time.Time, groupFilter prometheus.PQLFilter, services []response.SystemHealthService) (systemHealthVMHostResourceScope, error) {
	scope := newSystemHealthVMHostResourceScope()
	serviceNames := systemHealthServiceNames(services)
	if len(serviceNames) == 0 {
		return scope, nil
	}

	spanFilter := prometheus.Clone(groupFilter).RegexMatch(prometheus.ServiceNameKey, prometheus.RegexMultipleValue(serviceNames...))
	instancesByService, err := h.promRepo.GetMultiSVCInstanceListByPQLFilter(ctx, startTime.UnixMicro(), endTime.UnixMicro(), spanFilter)
	if err != nil {
		return scope, err
	}

	for _, serviceName := range serviceNames {
		serviceInstances := instancesByService[serviceName]
		if serviceInstances == nil {
			continue
		}
		for _, instance := range serviceInstances.GetInstances() {
			scope.addInstance(serviceName, instance)
		}
	}
	return scope, nil
}

func buildSystemHealthVMHostCPUWaterlinePQL(nodes []string, window string, offset string) string {
	series := fmt.Sprintf(
		`label_replace(rate(%s), "node_name", "$1", "instance_name", "(.*)")`,
		systemHealthVMHostRangeSelector("node_cpu_seconds_total", nodes, window, offset, `mode!~"idle|iowait|steal"`),
	)
	return fmt.Sprintf(`avg by (node_name) (sum by (node_name,cpu) (%s))`, series)
}

func buildSystemHealthVMHostMemoryWaterlinePQL(nodes []string, window string, offset string) string {
	available := fmt.Sprintf(
		`max by (node_name) (label_replace(avg_over_time(%s), "node_name", "$1", "instance_name", "(.*)"))`,
		systemHealthVMHostRangeSelector("node_memory_MemAvailable_bytes", nodes, window, offset),
	)
	total := fmt.Sprintf(
		`max by (node_name) (label_replace(avg_over_time(%s), "node_name", "$1", "instance_name", "(.*)"))`,
		systemHealthVMHostRangeSelector("node_memory_MemTotal_bytes", nodes, window, offset),
	)
	return fmt.Sprintf(`1 - ((%s) / (%s))`, available, total)
}

func buildSystemHealthVMHostResourceShiftPQL(nodes []string, window string) string {
	cpuRatio := systemHealthVMHostResourceRatioPQL(buildSystemHealthVMHostCPUWaterlinePQL, nodes, window)
	memoryRatio := systemHealthVMHostResourceRatioPQL(buildSystemHealthVMHostMemoryWaterlinePQL, nodes, window)
	return fmt.Sprintf(
		`max by (node_name) (((%s) >= 1) or (1 / (%s)) or ((%s) >= 1) or (1 / (%s)))`,
		cpuRatio,
		cpuRatio,
		memoryRatio,
		memoryRatio,
	)
}

func systemHealthVMHostResourceRatioPQL(build func([]string, string, string) string, nodes []string, window string) string {
	now := build(nodes, window, "")
	lastDay := build(nodes, window, "offset 24h")
	return fmt.Sprintf(`(%s) / (%s)`, now, lastDay)
}

func systemHealthVMHostRangeSelector(metric string, nodes []string, window string, offset string, extraMatchers ...string) string {
	matchers := []string{fmt.Sprintf(`instance_name=~"%s"`, prometheus.RegexMultipleValue(nodes...))}
	matchers = append(matchers, extraMatchers...)
	offsetText := strings.TrimSpace(offset)
	if offsetText != "" {
		offsetText = " " + offsetText
	}
	return fmt.Sprintf(`%s{%s}[%s]%s`, metric, strings.Join(matchers, ","), window, offsetText)
}
```

- [ ] **Step 4: Run tests to verify pass**

Run:

```powershell
go test ./backend/pkg/api/systemhealth -run "TestSystemHealthVMHostResourceScopeKeepsOnlyVMInstances|TestBuildSystemHealthVMHostResourcePQLUsesNodeExporterInstanceName" -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```powershell
git add backend/pkg/api/systemhealth/func_getoverview.go backend/pkg/api/systemhealth/func_getoverview_test.go
git commit -m "test: cover VM host resource query helpers"
```

---

### Task 2: Add VM Host Detail Conversion

**Files:**
- Modify: `backend/pkg/api/systemhealth/func_getoverview_test.go`
- Modify: `backend/pkg/api/systemhealth/func_getoverview.go`

- [ ] **Step 1: Write failing tests for VM host waterline and shift conversion**

Add these tests in `backend/pkg/api/systemhealth/func_getoverview_test.go` near the existing `applySystemHealthExtraMetrics` tests:

```go
func TestSystemHealthVMHostResourceWaterlineDetailsMapHostsToServices(t *testing.T) {
	scope := newSystemHealthVMHostResourceScope()
	scope.addInstance("order-api", &model.ServiceInstance{ServiceName: "order-api", NodeName: "vm-a", Pid: 2201})
	scope.addInstance("order-api", &model.ServiceInstance{ServiceName: "order-api", NodeName: "vm-b", Pid: 2202})
	scope.addInstance("order-worker", &model.ServiceInstance{ServiceName: "order-worker", NodeName: "vm-b", Pid: 3301})

	details := systemHealthVMHostResourceWaterlineDetails(
		scope,
		[]prometheus.MetricResult{
			{Metric: prometheus.Labels{NodeName: "vm-a"}, Values: []prometheus.Points{{Value: 0.72}}},
			{Metric: prometheus.Labels{NodeName: "vm-b"}, Values: []prometheus.Points{{Value: 0.42}}},
		},
		[]prometheus.MetricResult{
			{Metric: prometheus.Labels{NodeName: "vm-a"}, Values: []prometheus.Points{{Value: 0.65}}},
			{Metric: prometheus.Labels{NodeName: "vm-b"}, Values: []prometheus.Points{{Value: 0.93}}},
		},
	)

	api := details["order-api"]
	if api.Status != healthStatusRed || api.Value != "93.00%" || api.Source != "host/vm-b" || api.Note != "Current host memory resource waterline" {
		t.Fatalf("expected order-api to use highest related host waterline, got %+v", api)
	}

	worker := details["order-worker"]
	if worker.Status != healthStatusRed || worker.Value != "93.00%" || worker.Source != "host/vm-b" {
		t.Fatalf("expected order-worker to inherit vm-b host waterline, got %+v", worker)
	}
}

func TestSystemHealthVMHostResourceShiftDetailsMapHostsToServices(t *testing.T) {
	scope := newSystemHealthVMHostResourceScope()
	scope.addInstance("order-api", &model.ServiceInstance{ServiceName: "order-api", NodeName: "vm-a", Pid: 2201})
	scope.addInstance("order-api", &model.ServiceInstance{ServiceName: "order-api", NodeName: "vm-b", Pid: 2202})
	scope.addInstance("order-worker", &model.ServiceInstance{ServiceName: "order-worker", NodeName: "vm-b", Pid: 3301})

	details := systemHealthVMHostResourceShiftDetails(scope, []prometheus.MetricResult{
		{Metric: prometheus.Labels{NodeName: "vm-a"}, Values: []prometheus.Points{{Value: 1.35}}},
		{Metric: prometheus.Labels{NodeName: "vm-b"}, Values: []prometheus.Points{{Value: 1.75}}},
	})

	api := details["order-api"]
	if api.Status != healthStatusRed || api.Value != "75.00%" || api.Source != "host/vm-b" || api.Note != "Host CPU/memory resource shift against 24h baseline" {
		t.Fatalf("expected order-api to use highest related host shift, got %+v", api)
	}

	worker := details["order-worker"]
	if worker.Status != healthStatusRed || worker.Value != "75.00%" || worker.Source != "host/vm-b" {
		t.Fatalf("expected order-worker to inherit vm-b host shift, got %+v", worker)
	}
}
```

- [ ] **Step 2: Run tests to verify failure**

Run:

```powershell
go test ./backend/pkg/api/systemhealth -run "TestSystemHealthVMHostResourceWaterlineDetailsMapHostsToServices|TestSystemHealthVMHostResourceShiftDetailsMapHostsToServices" -count=1
```

Expected: FAIL because the VM host detail conversion helpers are not defined.

- [ ] **Step 3: Add VM host detail conversion helpers**

Add this code in `backend/pkg/api/systemhealth/func_getoverview.go` near the existing resource detail conversion helpers:

```go
type systemHealthVMHostResourceCandidate struct {
	value  float64
	note   string
	source string
}

func systemHealthVMHostResourceWaterlineDetails(scope systemHealthVMHostResourceScope, cpuResults []prometheus.MetricResult, memoryResults []prometheus.MetricResult) map[string]response.SystemHealthMetricDetail {
	hostCandidates := systemHealthVMHostResourceWaterlineCandidates(cpuResults, memoryResults)
	details := make(map[string]response.SystemHealthMetricDetail, len(scope.NodesByService))
	for serviceName := range scope.NodesByService {
		if detail, ok := systemHealthVMHostDetailForServiceNodes(scope.sortedServiceNodes(serviceName), hostCandidates, 70, 90); ok {
			details[serviceName] = detail
		}
	}
	return details
}

func systemHealthVMHostSystemResourceWaterlineDetail(scope systemHealthVMHostResourceScope, cpuResults []prometheus.MetricResult, memoryResults []prometheus.MetricResult) (response.SystemHealthMetricDetail, bool) {
	return systemHealthVMHostDetailForServiceNodes(scope.sortedNodes(), systemHealthVMHostResourceWaterlineCandidates(cpuResults, memoryResults), 70, 90)
}

func systemHealthVMHostResourceWaterlineCandidates(cpuResults []prometheus.MetricResult, memoryResults []prometheus.MetricResult) map[string]systemHealthVMHostResourceCandidate {
	candidates := map[string]systemHealthVMHostResourceCandidate{}
	collect := func(results []prometheus.MetricResult, note string) {
		for _, result := range results {
			nodeName := systemHealthVMHostNodeName(result.Metric)
			if nodeName == "" || len(result.Values) == 0 {
				continue
			}
			value := result.Values[0].Value
			if !isValidSystemHealthMetricValue(value) {
				continue
			}
			percentValue := value * 100
			current, ok := candidates[nodeName]
			if !ok || percentValue > current.value {
				candidates[nodeName] = systemHealthVMHostResourceCandidate{
					value:  percentValue,
					note:   note,
					source: systemHealthVMHostResourceSource(nodeName),
				}
			}
		}
	}
	collect(cpuResults, "Current host CPU resource waterline")
	collect(memoryResults, "Current host memory resource waterline")
	return candidates
}

func systemHealthVMHostResourceShiftDetails(scope systemHealthVMHostResourceScope, results []prometheus.MetricResult) map[string]response.SystemHealthMetricDetail {
	hostCandidates := systemHealthVMHostResourceShiftCandidates(results)
	details := make(map[string]response.SystemHealthMetricDetail, len(scope.NodesByService))
	for serviceName := range scope.NodesByService {
		if detail, ok := systemHealthVMHostDetailForServiceNodes(scope.sortedServiceNodes(serviceName), hostCandidates, 30, 60); ok {
			details[serviceName] = detail
		}
	}
	return details
}

func systemHealthVMHostResourceShiftCandidates(results []prometheus.MetricResult) map[string]systemHealthVMHostResourceCandidate {
	candidates := map[string]systemHealthVMHostResourceCandidate{}
	for _, result := range results {
		nodeName := systemHealthVMHostNodeName(result.Metric)
		if nodeName == "" || len(result.Values) == 0 {
			continue
		}
		value := result.Values[0].Value
		if math.IsNaN(value) || value < 1 {
			continue
		}
		shiftPercent := (value - 1) * 100
		current, ok := candidates[nodeName]
		if !ok || shiftPercent > current.value {
			candidates[nodeName] = systemHealthVMHostResourceCandidate{
				value:  shiftPercent,
				note:   "Host CPU/memory resource shift against 24h baseline",
				source: systemHealthVMHostResourceSource(nodeName),
			}
		}
	}
	return candidates
}

func systemHealthVMHostDetailForServiceNodes(nodes []string, candidates map[string]systemHealthVMHostResourceCandidate, yellowThreshold float64, redThreshold float64) (response.SystemHealthMetricDetail, bool) {
	var selected *systemHealthVMHostResourceCandidate
	for _, node := range nodes {
		candidate, ok := candidates[node]
		if !ok {
			continue
		}
		if selected == nil || candidate.value > selected.value {
			next := candidate
			selected = &next
		}
	}
	if selected == nil {
		return response.SystemHealthMetricDetail{}, false
	}
	status := healthStatusGreen
	if math.IsInf(selected.value, 0) || selected.value >= redThreshold {
		status = healthStatusRed
	} else if selected.value >= yellowThreshold {
		status = healthStatusYellow
	}
	return response.SystemHealthMetricDetail{
		Status: status,
		Value:  formatPercentValue(selected.value),
		Unit:   "%",
		Source: selected.source,
		Note:   selected.note,
	}, true
}

func systemHealthVMHostNodeName(labels prometheus.Labels) string {
	return strings.TrimSpace(labels.NodeName)
}

func systemHealthVMHostResourceSource(nodeName string) string {
	if nodeName == "" {
		return ""
	}
	return "host/" + nodeName
}
```

- [ ] **Step 4: Run tests to verify pass**

Run:

```powershell
go test ./backend/pkg/api/systemhealth -run "TestSystemHealthVMHostResourceWaterlineDetailsMapHostsToServices|TestSystemHealthVMHostResourceShiftDetailsMapHostsToServices" -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```powershell
git add backend/pkg/api/systemhealth/func_getoverview.go backend/pkg/api/systemhealth/func_getoverview_test.go
git commit -m "feat: map VM host resource metrics to services"
```

---

### Task 3: Wire VM Host Queries Into Extra Metrics

**Files:**
- Modify: `backend/pkg/api/systemhealth/func_getoverview.go`
- Modify: `backend/pkg/api/systemhealth/func_getoverview_test.go`

- [ ] **Step 1: Write failing merge tests**

Add this test in `backend/pkg/api/systemhealth/func_getoverview_test.go`:

```go
func TestApplySystemHealthExtraMetricsMergesVMHostResourceDetails(t *testing.T) {
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{
				Name:          "order-api",
				Metrics:       defaultSystemMetricStatus(),
				MetricDetails: defaultSystemMetricDetails(),
			},
			{
				Name:          "order-worker",
				Metrics:       defaultSystemMetricStatus(),
				MetricDetails: defaultSystemMetricDetails(),
			},
		},
	}

	scope := newSystemHealthVMHostResourceScope()
	scope.addInstance("order-api", &model.ServiceInstance{ServiceName: "order-api", NodeName: "vm-a", Pid: 2201})
	scope.addInstance("order-worker", &model.ServiceInstance{ServiceName: "order-worker", NodeName: "vm-b", Pid: 3301})

	extra := systemHealthExtraMetricResults{
		VMHostResource: systemHealthVMHostResourceResults{
			Scope: scope,
			CPU: []prometheus.MetricResult{
				{Metric: prometheus.Labels{NodeName: "vm-a"}, Values: []prometheus.Points{{Value: 0.72}}},
				{Metric: prometheus.Labels{NodeName: "vm-b"}, Values: []prometheus.Points{{Value: 0.32}}},
			},
			Memory: []prometheus.MetricResult{
				{Metric: prometheus.Labels{NodeName: "vm-a"}, Values: []prometheus.Points{{Value: 0.64}}},
				{Metric: prometheus.Labels{NodeName: "vm-b"}, Values: []prometheus.Points{{Value: 0.93}}},
			},
			Shift: []prometheus.MetricResult{
				{Metric: prometheus.Labels{NodeName: "vm-a"}, Values: []prometheus.Points{{Value: 1.35}}},
				{Metric: prometheus.Labels{NodeName: "vm-b"}, Values: []prometheus.Points{{Value: 1.05}}},
			},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	apiResource := system.Services[0].MetricDetails[healthMetricResource]
	if system.Services[0].Metrics[healthMetricResource] != healthStatusYellow || apiResource.Value != "72.00%" || apiResource.Source != "host/vm-a" {
		t.Fatalf("expected order-api host CPU waterline yellow, metrics=%+v detail=%+v", system.Services[0].Metrics, apiResource)
	}
	apiShift := system.Services[0].MetricDetails[healthMetricResourceShift]
	if system.Services[0].Metrics[healthMetricResourceShift] != healthStatusYellow || apiShift.Value != "35.00%" || apiShift.Source != "host/vm-a" {
		t.Fatalf("expected order-api host resource shift yellow, metrics=%+v detail=%+v", system.Services[0].Metrics, apiShift)
	}

	workerResource := system.Services[1].MetricDetails[healthMetricResource]
	if system.Services[1].Metrics[healthMetricResource] != healthStatusRed || workerResource.Value != "93.00%" || workerResource.Source != "host/vm-b" {
		t.Fatalf("expected order-worker host memory waterline red, metrics=%+v detail=%+v", system.Services[1].Metrics, workerResource)
	}

	systemResource := system.MetricDetails[healthMetricResource]
	if system.Metrics[healthMetricResource] != healthStatusRed || systemResource.Value != "93.00%" || systemResource.Source != "host/vm-b" {
		t.Fatalf("expected system resource to use highest VM host waterline, metrics=%+v detail=%+v", system.Metrics, systemResource)
	}
}
```

- [ ] **Step 2: Run test to verify failure**

Run:

```powershell
go test ./backend/pkg/api/systemhealth -run TestApplySystemHealthExtraMetricsMergesVMHostResourceDetails -count=1
```

Expected: FAIL because `systemHealthVMHostResourceResults` and `VMHostResource` are not defined or not applied.

- [ ] **Step 3: Add VM host result fields and merge logic**

Modify `systemHealthExtraMetricResults` in `backend/pkg/api/systemhealth/func_getoverview.go`:

```go
type systemHealthExtraMetricResults struct {
	ResourceCPU              []prometheus.MetricResult
	ResourceMemory           []prometheus.MetricResult
	SystemResourceCPU        []prometheus.MetricResult
	SystemResourceMemory     []prometheus.MetricResult
	VMHostResource           systemHealthVMHostResourceResults
	Latency                  []prometheus.MetricResult
	ErrorRate                []prometheus.MetricResult
	ErrorLog                 []prometheus.MetricResult
	BadRequests              []prometheus.MetricResult
	ResourceShift            []prometheus.MetricResult
	LockCost                 []prometheus.MetricResult
	SystemLatency            []prometheus.MetricResult
	SystemErrorRate          []prometheus.MetricResult
	SystemLatencyTrend       []prometheus.MetricResult
	SystemErrorRateTrend     []prometheus.MetricResult
	SystemErrorLogTrend      []prometheus.MetricResult
	SystemResourceShiftTrend []prometheus.MetricResult
}

type systemHealthVMHostResourceResults struct {
	Scope      systemHealthVMHostResourceScope
	CPU        []prometheus.MetricResult
	Memory     []prometheus.MetricResult
	Shift      []prometheus.MetricResult
	ShiftTrend []prometheus.MetricResult
}
```

Update `querySystemHealthExtraMetrics` after the existing container resource shift trend query:

```go
	if vmHostScope, err := h.buildSystemHealthVMHostResourceScope(ctx, startTime, endTime, groupFilter, services); err == nil && vmHostScope.hasNodes() {
		nodes := vmHostScope.sortedNodes()
		result.VMHostResource.Scope = vmHostScope
		if resourceCPU, err := h.promRepo.QueryData(ctx, queryTime, buildSystemHealthVMHostCPUWaterlinePQL(nodes, window, "")); err == nil {
			result.VMHostResource.CPU = resourceCPU
		}
		if resourceMemory, err := h.promRepo.QueryData(ctx, queryTime, buildSystemHealthVMHostMemoryWaterlinePQL(nodes, window, "")); err == nil {
			result.VMHostResource.Memory = resourceMemory
		}
		hostShiftQuery := buildSystemHealthVMHostResourceShiftPQL(nodes, window)
		if resourceShift, err := h.promRepo.QueryData(ctx, queryTime, hostShiftQuery); err == nil {
			result.VMHostResource.Shift = resourceShift
		}
		hostShiftTrendQuery := buildSystemHealthVMHostResourceShiftPQL(nodes, prometheus.VecFromDuration(trendStep))
		if resourceShiftTrend, err := h.promRepo.QueryRangeData(ctx, startTime, endTime, hostShiftTrendQuery, trendStep); err == nil {
			result.VMHostResource.ShiftTrend = resourceShiftTrend
			result.SystemResourceShiftTrend = append(result.SystemResourceShiftTrend, resourceShiftTrend...)
		}
	} else if err != nil {
		systemHealthWarnCtx(ctx, "system-health VM host resource scope skipped", zap.Error(err))
	}
```

Update `applySystemHealthExtraMetrics` after the existing `resourceByService` and `resourceShiftByService` assignments:

```go
	vmHostResourceByService := systemHealthVMHostResourceWaterlineDetails(extra.VMHostResource.Scope, extra.VMHostResource.CPU, extra.VMHostResource.Memory)
	vmHostResourceShiftByService := systemHealthVMHostResourceShiftDetails(extra.VMHostResource.Scope, extra.VMHostResource.Shift)
```

Update the service loop in `applySystemHealthExtraMetrics` after the existing container resource/shift applications:

```go
		if detail, ok := vmHostResourceByService[service.Name]; ok {
			applySystemHealthResourceWaterline(service, detail)
		}
		if detail, ok := vmHostResourceShiftByService[service.Name]; ok {
			applySystemHealthResourceShift(service, detail)
		}
```

Add this helper near `applySystemHealthResourceWaterline`:

```go
func shouldKeepCurrentSystemHealthDetail(currentStatus string, current response.SystemHealthMetricDetail, next response.SystemHealthMetricDetail) bool {
	mergedStatus := worseHealthStatus(currentStatus, next.Status)
	if mergedStatus != next.Status && currentStatus != "" && currentStatus != healthStatusGray {
		return true
	}
	if mergedStatus == next.Status && currentStatus == next.Status {
		currentRank, currentOK := parseSystemMetricDetailRank(current.Value)
		nextRank, nextOK := parseSystemMetricDetailRank(next.Value)
		return currentOK && nextOK && currentRank > nextRank
	}
	return false
}

func applySystemHealthResourceWaterline(service *response.SystemHealthService, detail response.SystemHealthMetricDetail) {
	if service == nil {
		return
	}
	currentStatus := service.Metrics[healthMetricResource]
	currentDetail := service.MetricDetails[healthMetricResource]
	if shouldKeepCurrentSystemHealthDetail(currentStatus, currentDetail, detail) {
		return
	}
	mergedStatus := worseHealthStatus(currentStatus, detail.Status)
	detail.Status = mergedStatus
	service.Metrics[healthMetricResource] = mergedStatus
	service.MetricDetails[healthMetricResource] = detail
}

func applySystemHealthResourceShift(service *response.SystemHealthService, detail response.SystemHealthMetricDetail) {
	if service == nil {
		return
	}
	currentStatus := service.Metrics[healthMetricResourceShift]
	currentDetail := service.MetricDetails[healthMetricResourceShift]
	if shouldKeepCurrentSystemHealthDetail(currentStatus, currentDetail, detail) {
		return
	}
	mergedStatus := worseHealthStatus(currentStatus, detail.Status)
	detail.Status = mergedStatus
	service.Metrics[healthMetricResourceShift] = mergedStatus
	service.MetricDetails[healthMetricResourceShift] = detail
}
```

Update `applySystemHealthSystemMetricDetails` after the existing container system resource waterline block:

```go
	if detail, ok := systemHealthVMHostSystemResourceWaterlineDetail(extra.VMHostResource.Scope, extra.VMHostResource.CPU, extra.VMHostResource.Memory); ok {
		applySystemHealthSystemResourceWaterline(system, detail)
	}
```

Refactor the existing system resource waterline logic into this helper and call it from both container and VM paths:

```go
func applySystemHealthSystemResourceWaterline(system *response.SystemHealthSystem, detail response.SystemHealthMetricDetail) {
	if system == nil {
		return
	}
	existingStatus := system.Metrics[healthMetricResource]
	existingDetail := system.MetricDetails[healthMetricResource]
	if shouldKeepCurrentSystemHealthDetail(existingStatus, existingDetail, detail) {
		return
	}
	mergedStatus := worseHealthStatus(existingStatus, detail.Status)
	detail.Status = mergedStatus
	system.Metrics[healthMetricResource] = mergedStatus
	system.MetricDetails[healthMetricResource] = detail
}
```

Replace the original container system resource block with:

```go
	if detail, ok := systemHealthSystemResourceWaterlineDetail(extra.SystemResourceCPU, extra.SystemResourceMemory); ok {
		applySystemHealthSystemResourceWaterline(system, detail)
	}
```

- [ ] **Step 4: Run test to verify pass**

Run:

```powershell
go test ./backend/pkg/api/systemhealth -run TestApplySystemHealthExtraMetricsMergesVMHostResourceDetails -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```powershell
git add backend/pkg/api/systemhealth/func_getoverview.go backend/pkg/api/systemhealth/func_getoverview_test.go
git commit -m "feat: merge VM host resource health details"
```

---

### Task 4: Add Mixed Source And Trend Tests

**Files:**
- Modify: `backend/pkg/api/systemhealth/func_getoverview_test.go`
- Modify: `backend/pkg/api/systemhealth/func_getoverview.go`

- [ ] **Step 1: Write failing tests for mixed container/VM aggregation and trend**

Add these tests in `backend/pkg/api/systemhealth/func_getoverview_test.go`:

```go
func TestApplySystemHealthExtraMetricsKeepsWorseContainerOrVMResource(t *testing.T) {
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{
				Name:          "order-api",
				Metrics:       defaultSystemMetricStatus(),
				MetricDetails: defaultSystemMetricDetails(),
			},
		},
	}

	scope := newSystemHealthVMHostResourceScope()
	scope.addInstance("order-api", &model.ServiceInstance{ServiceName: "order-api", NodeName: "vm-a", Pid: 2201})

	extra := systemHealthExtraMetricResults{
		ResourceCPU: []prometheus.MetricResult{
			{Metric: prometheus.Labels{SvcName: "order-api"}, Values: []prometheus.Points{{Value: 0.91}}},
		},
		VMHostResource: systemHealthVMHostResourceResults{
			Scope: scope,
			Memory: []prometheus.MetricResult{
				{Metric: prometheus.Labels{NodeName: "vm-a"}, Values: []prometheus.Points{{Value: 0.72}}},
			},
		},
	}

	applySystemHealthExtraMetrics(&system, extra)

	detail := system.Services[0].MetricDetails[healthMetricResource]
	if system.Services[0].Metrics[healthMetricResource] != healthStatusRed || detail.Value != "91.00%" || detail.Source != "" {
		t.Fatalf("expected worse container resource detail to win, metrics=%+v detail=%+v", system.Services[0].Metrics, detail)
	}
}

func TestApplySystemHealthExtraMetricsAddsVMHostShiftTrendToSystemTrend(t *testing.T) {
	system := response.SystemHealthSystem{
		Metrics:       defaultSystemMetricStatus(),
		MetricDetails: defaultSystemMetricDetails(),
		Services: []response.SystemHealthService{
			{
				Name:          "order-api",
				Metrics:       map[string]string{healthMetricResourceShift: healthStatusYellow},
				MetricDetails: defaultSystemMetricDetails(),
			},
		},
	}

	scope := newSystemHealthVMHostResourceScope()
	scope.addInstance("order-api", &model.ServiceInstance{ServiceName: "order-api", NodeName: "vm-a", Pid: 2201})

	extra := systemHealthExtraMetricResults{
		VMHostResource: systemHealthVMHostResourceResults{
			Scope: scope,
			ShiftTrend: []prometheus.MetricResult{
				{
					Metric: prometheus.Labels{NodeName: "vm-a"},
					Values: []prometheus.Points{
						{TimeStamp: 1000, Value: 1.20},
						{TimeStamp: 2000, Value: 1.75},
					},
				},
			},
		},
		SystemResourceShiftTrend: []prometheus.MetricResult{
			{
				Metric: prometheus.Labels{SvcName: "container-api"},
				Values: []prometheus.Points{
					{TimeStamp: 1000, Value: 1.45},
					{TimeStamp: 2000, Value: 1.10},
				},
			},
		},
	}
	extra.SystemResourceShiftTrend = append(extra.SystemResourceShiftTrend, extra.VMHostResource.ShiftTrend...)

	applySystemHealthExtraMetrics(&system, extra)

	detail := system.MetricDetails[healthMetricResourceShift]
	if detail.Unit != "%" {
		t.Fatalf("expected resource shift trend unit percent, got %+v", detail)
	}
	if len(detail.Thresholds) != 2 || detail.Thresholds[0].Value != 30 || detail.Thresholds[1].Value != 60 {
		t.Fatalf("expected resource shift thresholds, got %+v", detail.Thresholds)
	}
	if len(detail.Series) != 2 {
		t.Fatalf("expected two trend points, got %+v", detail.Series)
	}
	if detail.Series[0].Timestamp != 1000 || detail.Series[0].Value != 45 {
		t.Fatalf("expected timestamp 1000 to use container 45%% shift, got %+v", detail.Series[0])
	}
	if detail.Series[1].Timestamp != 2000 || detail.Series[1].Value != 75 {
		t.Fatalf("expected timestamp 2000 to use VM host 75%% shift, got %+v", detail.Series[1])
	}
}
```

- [ ] **Step 2: Run tests to verify behavior**

Run:

```powershell
go test ./backend/pkg/api/systemhealth -run "TestApplySystemHealthExtraMetricsKeepsWorseContainerOrVMResource|TestApplySystemHealthExtraMetricsAddsVMHostShiftTrendToSystemTrend" -count=1
```

Expected: PASS.

- [ ] **Step 3: Confirm host shift trends are appended in query path**

In `querySystemHealthExtraMetrics`, verify this code is present inside the VM host scope block:

```go
		if resourceShiftTrend, err := h.promRepo.QueryRangeData(ctx, startTime, endTime, hostShiftTrendQuery, trendStep); err == nil {
			result.VMHostResource.ShiftTrend = resourceShiftTrend
			result.SystemResourceShiftTrend = append(result.SystemResourceShiftTrend, resourceShiftTrend...)
		}
```

- [ ] **Step 4: Run tests to verify pass**

Run:

```powershell
go test ./backend/pkg/api/systemhealth -run "TestApplySystemHealthExtraMetricsKeepsWorseContainerOrVMResource|TestApplySystemHealthExtraMetricsAddsVMHostShiftTrendToSystemTrend" -count=1
```

Expected: PASS.

- [ ] **Step 5: Commit**

Run:

```powershell
git add backend/pkg/api/systemhealth/func_getoverview.go backend/pkg/api/systemhealth/func_getoverview_test.go
git commit -m "test: cover mixed VM and container resource health"
```

---

### Task 5: Run Full System Health Tests And Polish

**Files:**
- Modify: `backend/pkg/api/systemhealth/func_getoverview.go`
- Modify: `backend/pkg/api/systemhealth/func_getoverview_test.go`

- [ ] **Step 1: Run focused system health test package**

Run:

```powershell
go test ./backend/pkg/api/systemhealth -count=1
```

Expected: PASS.

- [ ] **Step 2: Run broader backend compile/test check**

Run:

```powershell
go test ./backend/pkg/... -run TestNonExistentForCompileOnly -count=1
```

Expected: packages compile; packages with no matching tests report `ok` or `[no test files]`.

- [ ] **Step 3: Inspect final diff**

Run:

```powershell
git diff -- backend/pkg/api/systemhealth/func_getoverview.go backend/pkg/api/systemhealth/func_getoverview_test.go
git status --short
```

Expected:

- Only `backend/pkg/api/systemhealth/func_getoverview.go` and `backend/pkg/api/systemhealth/func_getoverview_test.go` are modified for this feature.
- Existing unrelated worktree changes, such as `frontend/src/pro/views/systemHealthOverview/index.jsx`, remain unstaged and untouched.

- [ ] **Step 4: Confirm no uncommitted feature changes remain**

Run:

```powershell
git diff -- backend/pkg/api/systemhealth/func_getoverview.go backend/pkg/api/systemhealth/func_getoverview_test.go
```

Expected: no output.

---

## Final Verification Checklist

- [ ] `go test ./backend/pkg/api/systemhealth -count=1` passes.
- [ ] `go test ./backend/pkg/api/systemhealth ./backend/pkg/repository/prometheus ./backend/pkg/model -count=1` passes, or the broader `go test ./backend/pkg/... -run TestNonExistentForCompileOnly -count=1` compile check passes.
- [ ] Container resource tests still pass.
- [ ] VM-only service resource details include `Source: "host/<node>"`.
- [ ] VM-only service resource notes explicitly say host-level.
- [ ] Mixed container/VM systems aggregate by worse status.
- [ ] Resource shift trend includes the max shift per timestamp across container and VM host data.
- [ ] No frontend response shape changes were introduced.
