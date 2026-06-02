// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package customdashboard

import (
	"encoding/json"
	"errors"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	cdmodel "github.com/CloudDetail/apo/backend/pkg/model/customdashboard"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
	"github.com/CloudDetail/apo/backend/pkg/model/profile"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/common"
	prommodel "github.com/prometheus/common/model"
)

func int64Ptr(value int64) *int64 {
	return &value
}

func TestValidateQueryRejectsSQLAndUnbalancedExpression(t *testing.T) {
	s := &service{}

	if result := s.ValidateQuery(nil, `SELECT * FROM metrics`); result.Valid {
		t.Fatalf("expected SQL-like query to be rejected")
	}

	if result := s.ValidateQuery(nil, `sum(rate(http_requests_total{svc_name="api"[5m]))`); result.Valid {
		t.Fatalf("expected unbalanced query to be rejected")
	}

	if result := s.ValidateQuery(nil, `sum(rate(http_requests_total{svc_name="api"}[5m]))`); !result.Valid {
		t.Fatalf("expected valid PromQL-like query, got %s", result.Message)
	}
}

func TestSubstituteVariables(t *testing.T) {
	expr := substituteVariables(
		`sum(rate(span_trace_duration_count{svc_name="$service",namespace="{{namespace}}"}[1m]))`,
		map[string][]string{"service": []string{"checkout"}, "namespace": []string{"prod"}},
	)
	want := `sum(rate(span_trace_duration_count{svc_name="checkout",namespace="prod"}[1m]))`
	if expr != want {
		t.Fatalf("unexpected substituted query\nwant: %s\n got: %s", want, expr)
	}
}

func TestSubstituteVariablesSupportsGrafanaFormats(t *testing.T) {
	expr := substituteVariables(
		`sum(rate(http_requests_total{cluster_id=~"${cluster_id:regex}",job_name="${job_name}",namespace="{{ namespace }}"}[$resolution]))`,
		map[string][]string{
			"cluster_id": {"dev", "prod.eu"},
			"job_name":   {"integrations/unix"},
			"namespace":  {"default"},
			"resolution": {"1m"},
		},
	)
	want := `sum(rate(http_requests_total{cluster_id=~"dev|prod\\.eu",job_name="integrations/unix",namespace="default"}[1m]))`
	if expr != want {
		t.Fatalf("unexpected substituted query\nwant: %s\n got: %s", want, expr)
	}
}

func TestSubstituteVariablesUsesRegexAlternationInRegexMatchers(t *testing.T) {
	vars := map[string][]string{"cluster_id": {"dev", "prod.eu"}}
	cases := []struct {
		name string
		expr string
		want string
	}{
		{
			name: "single quoted dollar token",
			expr: `up{cluster_id=~'$cluster_id'}`,
			want: `up{cluster_id=~'dev|prod\\.eu'}`,
		},
		{
			name: "backtick dollar token",
			expr: "up{cluster_id=~`$cluster_id`}",
			want: `up{cluster_id=~` + "`dev|prod\\\\.eu`" + `}`,
		},
		{
			name: "mustache token",
			expr: `up{cluster_id=~"{{ cluster_id }}"}`,
			want: `up{cluster_id=~"dev|prod\\.eu"}`,
		},
	}
	for _, tt := range cases {
		t.Run(tt.name, func(t *testing.T) {
			if got := substituteVariables(tt.expr, vars); got != tt.want {
				t.Fatalf("unexpected substituted query\nwant: %s\n got: %s", tt.want, got)
			}
		})
	}
}

func TestSubstituteVariablesKeepsAllValueRaw(t *testing.T) {
	expr := substituteVariables(
		`sum(rate(http_requests_total{cluster_id=~"${cluster_id:regex}"}[1m]))`,
		map[string][]string{"cluster_id": {".*"}},
	)
	want := `sum(rate(http_requests_total{cluster_id=~".*"}[1m]))`
	if expr != want {
		t.Fatalf("unexpected all substitution\nwant: %s\n got: %s", want, expr)
	}
}

func TestSubstituteVariablesTreatsAllInExactMatcherAsRegexWildcard(t *testing.T) {
	expr := substituteVariables(
		`sum(rate(http_requests_total{namespace="$namespace",cluster_id=~"${cluster_id:regex}",svc_name="$service",pod="${pod:raw}",dst_namespace="${dst_namespace:regex}"}[1m]))`,
		map[string][]string{
			"namespace":     {".*"},
			"cluster_id":    {".*"},
			"service":       {"checkout"},
			"pod":           {".*"},
			"dst_namespace": {".*"},
		},
	)
	want := `sum(rate(http_requests_total{namespace=~".*",cluster_id=~".*",svc_name="checkout",pod=~".*",dst_namespace=~".*"}[1m]))`
	if expr != want {
		t.Fatalf("unexpected all exact matcher substitution\nwant: %s\n got: %s", want, expr)
	}
}

func TestSubstituteVariablesUsesResolutionForGrafanaRateInterval(t *testing.T) {
	expr := substituteVariables(
		`sum(rate(container_cpu_usage_seconds_total{pod=~"$pod"}[$__rate_interval]))`,
		map[string][]string{"pod": {".*"}, "resolution": {"1m"}},
	)
	want := `sum(rate(container_cpu_usage_seconds_total{pod=~".*"}[1m]))`
	if expr != want {
		t.Fatalf("unexpected Grafana rate interval substitution\nwant: %s\n got: %s", want, expr)
	}
}

func TestInjectFilter(t *testing.T) {
	filter := `cluster_id=~"c1",namespace=~"prod",svc_name=~"checkout"`

	got := injectFilter(`up`, filter)
	want := `up{cluster_id=~"c1",namespace=~"prod",svc_name=~"checkout"}`
	if got != want {
		t.Fatalf("unexpected selector injection\nwant: %s\n got: %s", want, got)
	}

	got = injectFilter(`sum(rate(http_requests_total{code!="500"}[5m]))`, filter)
	want = `sum(rate(http_requests_total{cluster_id=~"c1",namespace=~"prod",svc_name=~"checkout",code!="500"}[5m]))`
	if got != want {
		t.Fatalf("unexpected existing selector injection\nwant: %s\n got: %s", want, got)
	}

	got = injectFilter(`sum(rate(http_requests_total[5m])) / sum(up{job="api"})`, filter)
	want = `sum(rate(http_requests_total{cluster_id=~"c1",namespace=~"prod",svc_name=~"checkout"}[5m])) / sum(up{cluster_id=~"c1",namespace=~"prod",svc_name=~"checkout",job="api"})`
	if got != want {
		t.Fatalf("unexpected nested selector injection\nwant: %s\n got: %s", want, got)
	}

	got = injectFilter(`sum by (svc_name) (rate(http_requests_total[5m])) or vector(0)`, filter)
	want = `sum by (svc_name) (rate(http_requests_total{cluster_id=~"c1",namespace=~"prod",svc_name=~"checkout"}[5m])) or vector(0)`
	if got != want {
		t.Fatalf("unexpected grouping injection\nwant: %s\n got: %s", want, got)
	}

	got = injectFilter("label_replace(up, \"metric\", `http_requests_total{svc_name=\"api\"}`, \"job\", \".*\")", filter)
	want = "label_replace(up{cluster_id=~\"c1\",namespace=~\"prod\",svc_name=~\"checkout\"}, \"metric\", `http_requests_total{svc_name=\"api\"}`, \"job\", \".*\")"
	if got != want {
		t.Fatalf("unexpected backtick string handling\nwant: %s\n got: %s", want, got)
	}
}

func TestScopeFilterString(t *testing.T) {
	got := scopeFilterString([]datagroup.DataScope{
		{ScopeLabels: datagroup.ScopeLabels{ClusterID: "c2", Namespace: "prod", Service: "checkout"}},
		{ScopeLabels: datagroup.ScopeLabels{ClusterID: "c1", Namespace: "staging", Service: "cart"}},
	})

	want := `cluster_id=~"c1|c2",namespace=~"prod|staging",svc_name=~"cart|checkout"`
	if got != want {
		t.Fatalf("unexpected filter\nwant: %s\n got: %s", want, got)
	}
}

func TestCandidateUpstreamSelectionsExcludeCurrentVariableAliases(t *testing.T) {
	selected := map[string][]string{"namespace": []string{"prod"}, "service": []string{"checkout"}}
	upstream := candidateUpstreamSelections(selected, "", "label", "svc_name")
	if _, exists := upstream["service"]; exists {
		t.Fatalf("current service variable must not filter its own svc_name candidates: %#v", upstream)
	}
	upstream = candidateUpstreamSelections(selected, "", "label", "svc")
	if _, exists := upstream["service"]; exists {
		t.Fatalf("current service variable must not filter its own svc candidates: %#v", upstream)
	}
	if upstream["namespace"][0] != "prod" {
		t.Fatalf("expected upstream namespace selection to remain, got %#v", upstream)
	}

	filter := selectedVariableFilter(upstream, "svc_name")
	if filter != `namespace=~"prod"` {
		t.Fatalf("unexpected upstream filter: %s", filter)
	}
}

func TestCandidateUpstreamSelectionsExcludeCurrentVariableByName(t *testing.T) {
	selected := map[string][]string{"namespace": []string{"prod"}, "var1": []string{"api.v1", "cart+checkout", "foo|bar"}}
	upstream := candidateUpstreamSelections(selected, "var1", "label", "svc_name")
	if _, exists := upstream["var1"]; exists {
		t.Fatalf("current variable name must be excluded even when it is not semantic: %#v", upstream)
	}
	if upstream["namespace"][0] != "prod" {
		t.Fatalf("expected upstream namespace selection to remain, got %#v", upstream)
	}
	if got := selectedVariableFilter(selected, "namespace"); got != `var1=~"api\\.v1|cart\\+checkout|foo\\|bar"` {
		t.Fatalf("expected escaped regex matcher, got %s", got)
	}
}

func TestEstimatePointsAndMetricSelector(t *testing.T) {
	start := time.Unix(0, 0).UnixMilli()
	end := time.Unix(3600, 0).UnixMilli()
	if got := estimatePoints(start, end, 60); got != 60 {
		t.Fatalf("expected 60 points, got %d", got)
	}

	if got := metricSelector("container_cpu_usage_seconds_total", `pod=~"api.*"`); got != `container_cpu_usage_seconds_total{pod=~"api.*"}` {
		t.Fatalf("unexpected metric selector: %s", got)
	}
}

func TestLabelValueCandidatesUseMetricSelectorFromPromQL(t *testing.T) {
	db := &dataScopeRepoStub{userGroups: map[int64][]int64{101: {0}}}
	prom := &queryRecorderProm{}
	s := &service{dbRepo: db, prom: prom}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	_, err := s.VariableCandidates(ctx, &request.DashboardVariableRequest{
		Type:         "service",
		Label:        "svc_name",
		VariableName: "service",
		Query:        `sum by (content_key, svc_name) (increase(kindling_span_trace_duration_nanoseconds_count{content_key=~".*", cluster_id=~".*"}[1m]))`,
		GroupID:      int64Ptr(0),
	})
	if err != nil {
		t.Fatalf("variable candidates should accept full PromQL by extracting selector: %v", err)
	}
	if len(prom.labelQueries) != 1 {
		t.Fatalf("expected one label query, got %d", len(prom.labelQueries))
	}
	want := `kindling_span_trace_duration_nanoseconds_count{content_key=~".*", cluster_id=~".*"}`
	if prom.labelQueries[0] != want {
		t.Fatalf("LabelValues should receive metric selector\nwant: %s\n got: %s", want, prom.labelQueries[0])
	}
}

func TestLabelValueCandidatesResolveVariableQuerySelections(t *testing.T) {
	db := &dataScopeRepoStub{userGroups: map[int64][]int64{101: {0}}}
	prom := &queryRecorderProm{}
	s := &service{dbRepo: db, prom: prom}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	_, err := s.VariableCandidates(ctx, &request.DashboardVariableRequest{
		Type:          "label",
		Label:         "job_name",
		VariableName:  "job_name",
		Query:         `label_values(node_cpu_seconds_total{cluster_id=~"$cluster"},job_name)`,
		VariablesJSON: `{"cluster":["dev"],"cadvisor_job_name":["integrations/kubernetes/cadvisor"]}`,
		GroupID:       int64Ptr(0),
	})
	if err != nil {
		t.Fatalf("variable candidates should resolve upstream variables in label query: %v", err)
	}
	if len(prom.labelQueries) != 1 {
		t.Fatalf("expected one label query, got %d", len(prom.labelQueries))
	}
	want := `node_cpu_seconds_total{cluster_id=~"dev"}`
	if prom.labelQueries[0] != want {
		t.Fatalf("LabelValues should receive resolved selector\nwant: %s\n got: %s", want, prom.labelQueries[0])
	}
}

func TestLabelValueCandidatesResolveMultiValueRegexSelections(t *testing.T) {
	db := &dataScopeRepoStub{userGroups: map[int64][]int64{101: {0}}}
	prom := &queryRecorderProm{}
	s := &service{dbRepo: db, prom: prom}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	_, err := s.VariableCandidates(ctx, &request.DashboardVariableRequest{
		Type:          "label",
		Label:         "job_name",
		VariableName:  "job_name",
		Query:         `label_values(node_cpu_seconds_total{cluster_id=~"$cluster_id"},job_name)`,
		VariablesJSON: `{"cluster_id":["dev","dllg-x86"],"cadvisor_job_name":["integrations/kubernetes/cadvisor"]}`,
		GroupID:       int64Ptr(0),
	})
	if err != nil {
		t.Fatalf("variable candidates should resolve upstream variables in label query: %v", err)
	}
	if len(prom.labelQueries) != 1 {
		t.Fatalf("expected one label query, got %d", len(prom.labelQueries))
	}
	want := `node_cpu_seconds_total{cluster_id=~"dev|dllg-x86"}`
	if prom.labelQueries[0] != want {
		t.Fatalf("LabelValues should receive resolved selector\nwant: %s\n got: %s", want, prom.labelQueries[0])
	}
}

func TestLabelValueCandidatesDoNotInjectDashboardVariablesIntoGrafanaLabelValues(t *testing.T) {
	db := &dataScopeRepoStub{userGroups: map[int64][]int64{101: {0}}}
	prom := &queryRecorderProm{}
	s := &service{dbRepo: db, prom: prom}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	_, err := s.VariableCandidates(ctx, &request.DashboardVariableRequest{
		Type:         "label",
		Label:        "pod",
		VariableName: "pod",
		Query:        `label_values(container_memory_rss{namespace="$namespace",job_name="$cadvisor_job_name",cluster_id=~"${cluster_id:regex}"}, pod)`,
		VariablesJSON: `{
			"dataGroup":["0"],
			"cluster_id":[".*"],
			"namespace":["apisix"],
			"pod":[".*"],
			"related_nodes":[".*"],
			"resolution":["1m"],
			"diskdevices":["[a-z]+|nvme[0-9]+n[0-9]+|mmcblk[0-9]+"],
				"job_name":["integrations/unix"],
				"cadvisor_job_name":["integrations/kubernetes/cadvisor"]
			}`,
		GroupID: int64Ptr(0),
	})
	if err != nil {
		t.Fatalf("variable candidates should resolve Grafana label_values query: %v", err)
	}
	if len(prom.labelQueries) != 1 {
		t.Fatalf("expected one label query, got %d", len(prom.labelQueries))
	}
	want := `container_memory_rss{namespace="apisix",job_name="integrations/kubernetes/cadvisor",cluster_id=~".*"}`
	if prom.labelQueries[0] != want {
		t.Fatalf("LabelValues should not receive non-label dashboard variables\nwant: %s\n got: %s", want, prom.labelQueries[0])
	}
}

func TestLabelValueCandidatesExtractMetricFromGrafanaLabelValuesWithoutMatchers(t *testing.T) {
	db := &dataScopeRepoStub{userGroups: map[int64][]int64{101: {0}}}
	prom := &queryRecorderProm{}
	s := &service{dbRepo: db, prom: prom}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	_, err := s.VariableCandidates(ctx, &request.DashboardVariableRequest{
		Type:          "label",
		Label:         "job_name",
		VariableName:  "cadvisor_job_name",
		Query:         `label_values(container_cpu_usage_seconds_total, job_name)`,
		VariablesJSON: `{"dataGroup":["0"],"namespace":["apisix"],"resolution":["1m"],"job_name":["integrations/unix"]}`,
		GroupID:       int64Ptr(0),
	})
	if err != nil {
		t.Fatalf("variable candidates should handle label_values with a bare metric: %v", err)
	}
	if len(prom.labelQueries) != 1 {
		t.Fatalf("expected one label query, got %d", len(prom.labelQueries))
	}
	want := `container_cpu_usage_seconds_total`
	if prom.labelQueries[0] != want {
		t.Fatalf("LabelValues should receive the metric selector from label_values\nwant: %s\n got: %s", want, prom.labelQueries[0])
	}
}

func TestPodCandidatesApplyUpstreamVariableSelections(t *testing.T) {
	db := &dataScopeRepoStub{userGroups: map[int64][]int64{101: {0}}}
	prom := &queryRecorderProm{}
	s := &service{dbRepo: db, prom: prom}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	_, err := s.VariableCandidates(ctx, &request.DashboardVariableRequest{
		Type:          "pod",
		VariableName:  "pod",
		VariablesJSON: `{"namespace":["prod"],"service":["checkout"],"pod":["pod-a"]}`,
	})
	if err != nil {
		t.Fatalf("pod candidates should apply upstream filters: %v", err)
	}
	if len(prom.labelQueries) != 1 {
		t.Fatalf("expected one label query, got %d", len(prom.labelQueries))
	}
	want := `container_cpu_usage_seconds_total{namespace=~"prod",svc_name=~"checkout"}`
	if prom.labelQueries[0] != want {
		t.Fatalf("pod LabelValues should receive upstream-filtered selector\nwant: %s\n got: %s", want, prom.labelQueries[0])
	}
}

func TestLabelValueCandidatesUseJSONVariablesWindowAndCandidateLimit(t *testing.T) {
	values := make(prommodel.LabelValues, 0, 1005)
	for i := 0; i < 1005; i++ {
		values = append(values, prommodel.LabelValue("value-"+fmt.Sprint(i)))
	}
	db := &dataScopeRepoStub{userGroups: map[int64][]int64{101: {0}}}
	prom := &queryRecorderProm{labelValues: values}
	s := &service{dbRepo: db, prom: prom}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	resp, err := s.VariableCandidates(ctx, &request.DashboardVariableRequest{
		Type:         "label",
		Label:        "job_name",
		VariableName: "job_name",
		Query:        `label_values(node_cpu_seconds_total{cluster_id=~"$cluster",namespace="$namespace"}, job_name)`,
		Variables: map[string][]string{
			"cluster":    {"dev"},
			"namespace":  {"prod"},
			"resolution": {"1m"},
			"job_name":   {"old"},
		},
		StartTime: time.Unix(100, 0).UnixMilli(),
		EndTime:   time.Unix(200, 0).UnixMilli(),
		GroupID:   int64Ptr(0),
	})
	if err != nil {
		t.Fatalf("variable candidates should accept JSON array variables: %v", err)
	}
	if len(resp) != 1000 {
		t.Fatalf("expected candidate limit to cap response at 1000, got %d", len(resp))
	}
	if got, want := prom.labelQueries[0], `node_cpu_seconds_total{cluster_id=~"dev",namespace="prod"}`; got != want {
		t.Fatalf("LabelValues should receive resolved selector\nwant: %s\n got: %s", want, got)
	}
	if got, want := prom.labelStartTimes[0], time.Unix(100, 0).UnixMicro(); got != want {
		t.Fatalf("LabelValues should receive request start time\nwant: %d\n got: %d", want, got)
	}
	if got, want := prom.labelEndTimes[0], time.Unix(200, 0).UnixMicro(); got != want {
		t.Fatalf("LabelValues should receive request end time\nwant: %d\n got: %d", want, got)
	}
}

func TestResolveQueryRejectsUnsupportedMode(t *testing.T) {
	s := &service{}
	_, err := s.resolveQuery(&request.DashboardQueryRequest{Mode: "sql", Expr: "SELECT * FROM metrics"})
	if err == nil {
		t.Fatalf("expected unsupported query mode to be rejected")
	}
}

func TestCopyDashboardNameUsesProductSuffix(t *testing.T) {
	if got := copyDashboardName("Service Health", code.LANG_ZH); got != "Service Health - 副本" {
		t.Fatalf("unexpected copied dashboard name: %s", got)
	}
	if got := copyDashboardName("Service Health", code.LANG_EN); got != "Service Health - Copy" {
		t.Fatalf("unexpected copied dashboard name: %s", got)
	}
}

type dashboardCreatePermissionRepoStub struct {
	database.Repo
	featureLookups int
}

func (r *dashboardCreatePermissionRepoStub) GetUserTeams(ctx core.Context, userID int64) ([]int64, error) {
	return nil, nil
}

func (r *dashboardCreatePermissionRepoStub) ListCustomDashboards(ctx core.Context, userID int64, teamIDs []int64, keyword string) ([]database.CustomDashboard, error) {
	return []database.CustomDashboard{
		{
			ID:         "dashboard-1",
			Name:       "业务大盘",
			OwnerID:    userID,
			Visibility: cdmodel.VisibilityEveryone,
			TeamIDs:    "[]",
			Config:     "{}",
			CreatedAt:  time.Unix(100, 0),
			UpdatedAt:  time.Unix(200, 0),
		},
	}, nil
}

func (r *dashboardCreatePermissionRepoStub) GetFeatureByName(ctx core.Context, name string) (int, error) {
	r.featureLookups++
	return 0, errors.New("create-dashboard feature should not be queried")
}

func (r *dashboardCreatePermissionRepoStub) GetUserInfo(ctx core.Context, userID int64) (profile.User, error) {
	return profile.User{UserID: userID, Username: "owner"}, nil
}

func TestDashboardListDoesNotRequireCreateFeaturePermission(t *testing.T) {
	repo := &dashboardCreatePermissionRepoStub{}
	s := &service{dbRepo: repo}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	resp, err := s.List(ctx, &request.DashboardListRequest{})
	if err != nil {
		t.Fatalf("list dashboards: %v", err)
	}
	if repo.featureLookups != 0 {
		t.Fatalf("list should not query standalone create feature permission, got %d lookups", repo.featureLookups)
	}
	if !resp.CanCreate {
		t.Fatalf("dashboard list should allow creating from the menu page without a standalone create feature")
	}
	if len(resp.Dashboards) != 1 || !resp.Dashboards[0].CanCopy {
		t.Fatalf("visible dashboards should be copyable without a standalone create feature, got %#v", resp.Dashboards)
	}
}

type dashboardRoutePermissionRepoStub struct {
	database.Repo
	userPerms    []int
	routerRoutes []database.Router
	menuRoutes   []database.Router
	dashboard    *database.CustomDashboard
}

func (r *dashboardRoutePermissionRepoStub) GetUserRole(ctx core.Context, userID int64) ([]profile.UserRole, error) {
	return nil, nil
}

func (r *dashboardRoutePermissionRepoStub) GetSubjectsPermission(ctx core.Context, subIDs []int64, subType string, typ string) ([]database.AuthPermission, error) {
	return nil, nil
}

func (r *dashboardRoutePermissionRepoStub) GetSubjectPermission(ctx core.Context, subID int64, subType string, typ string) ([]int, error) {
	return r.userPerms, nil
}

func (r *dashboardRoutePermissionRepoStub) GetUserTeams(ctx core.Context, userID int64) ([]int64, error) {
	return nil, nil
}

func (r *dashboardRoutePermissionRepoStub) GetCustomDashboard(ctx core.Context, id string) (*database.CustomDashboard, error) {
	return r.dashboard, nil
}

func (r *dashboardRoutePermissionRepoStub) GetFeatureMappingByFeature(ctx core.Context, featureIDs []int, mappedType string) ([]database.FeatureMapping, error) {
	if mappedType == model.MAPPED_TYP_ROUTER {
		return []database.FeatureMapping{{FeatureID: featureIDs[0], MappedID: 10, MappedType: mappedType}}, nil
	}
	return []database.FeatureMapping{{FeatureID: featureIDs[0], MappedID: 20, MappedType: mappedType}}, nil
}

func (r *dashboardRoutePermissionRepoStub) GetRouterByIDs(ctx core.Context, routerIDs []int) ([]database.Router, error) {
	return r.routerRoutes, nil
}

func (r *dashboardRoutePermissionRepoStub) GetItemsRouter(ctx core.Context, itemIDs []int) ([]database.Router, error) {
	return r.menuRoutes, nil
}

func TestCanAccessDashboardRouteRequiresConcreteMenuAndRouterPermission(t *testing.T) {
	ctx := core.EmptyCtx()
	s := &service{dbRepo: &dashboardRoutePermissionRepoStub{
		userPerms:    []int{7},
		routerRoutes: []database.Router{{RouterTo: "/custom-dashboard/:id"}},
		menuRoutes:   []database.Router{{RouterTo: "/custom-dashboard/dashboard-1"}},
	}}
	if ok, err := s.canAccessDashboardRoute(ctx, "dashboard-1", 101); err != nil || ok {
		t.Fatalf("generic router permission must not authorize concrete dashboard route, ok=%v err=%v", ok, err)
	}

	s.dbRepo = &dashboardRoutePermissionRepoStub{
		userPerms:    []int{7},
		routerRoutes: []database.Router{{RouterTo: "/custom-dashboard/dashboard-1"}},
		menuRoutes:   []database.Router{{RouterTo: "/custom-dashboard/dashboard-1"}},
	}
	if ok, err := s.canAccessDashboardRoute(ctx, "dashboard-1", 101); err != nil || !ok {
		t.Fatalf("expected exact menu and router permission to authorize dashboard route, ok=%v err=%v", ok, err)
	}
}

func TestCanAccessDashboardRouteAllowsRootDashboardPermission(t *testing.T) {
	ctx := core.EmptyCtx()
	s := &service{dbRepo: &dashboardRoutePermissionRepoStub{
		userPerms:    []int{7},
		routerRoutes: []database.Router{{RouterTo: "/custom-dashboard/:id"}},
		menuRoutes:   []database.Router{{RouterTo: "/custom-dashboard"}},
	}}

	if ok, err := s.canAccessDashboardRoute(ctx, "dashboard-1", 101); err != nil || !ok {
		t.Fatalf("root dashboard permission should authorize visible dashboard route before visibility filtering, ok=%v err=%v", ok, err)
	}
}

func TestGetDashboardWithoutVisibilityReturnsNoPermissionCode(t *testing.T) {
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))
	s := &service{dbRepo: &dashboardRoutePermissionRepoStub{
		userPerms:    []int{7},
		routerRoutes: []database.Router{{RouterTo: "/custom-dashboard/:id"}},
		menuRoutes:   []database.Router{{RouterTo: "/custom-dashboard"}},
		dashboard: &database.CustomDashboard{
			ID:         "dashboard-1",
			OwnerID:    202,
			Visibility: cdmodel.VisibilityTeam,
			TeamIDs:    "[900]",
			Config:     "{}",
		},
	}}

	_, err := s.Get(ctx, "dashboard-1")
	if err == nil {
		t.Fatalf("expected invisible dashboard to be rejected")
	}
	var businessErr core.BusinessError
	if !errors.As(err, &businessErr) {
		t.Fatalf("expected business error, got %T", err)
	}
	if businessErr.BusinessCode() != code.UserNoPermissionError {
		t.Fatalf("expected no-permission code, got %s", businessErr.BusinessCode())
	}
	if businessErr.Message() != "" {
		t.Fatalf("expected no-permission message to use localized default, got %q", businessErr.Message())
	}
}

func TestUpdateDashboardByNonCreatorReturnsNoPermissionCode(t *testing.T) {
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))
	s := &service{dbRepo: &dashboardRoutePermissionRepoStub{
		dashboard: &database.CustomDashboard{
			ID:         "dashboard-1",
			OwnerID:    202,
			Visibility: cdmodel.VisibilityEveryone,
			TeamIDs:    "[]",
			Config:     "{}",
		},
	}}

	_, err := s.Update(ctx, &request.UpdateDashboardRequest{Dashboard: cdmodel.Dashboard{ID: "dashboard-1"}})
	if err == nil {
		t.Fatalf("expected non-creator update to be rejected")
	}
	var businessErr core.BusinessError
	if !errors.As(err, &businessErr) {
		t.Fatalf("expected business error, got %T", err)
	}
	if businessErr.BusinessCode() != code.UserNoPermissionError {
		t.Fatalf("expected no-permission code, got %s", businessErr.BusinessCode())
	}
	if businessErr.Message() != "" {
		t.Fatalf("expected no-permission message to use localized default, got %q", businessErr.Message())
	}
}

type dataScopeRepoStub struct {
	database.Repo
	userGroups   map[int64][]int64
	groupScopes  map[int64][]string
	scopes       map[string]datagroup.DataScope
	groupAllowed map[int64]bool
	audits       []database.CustomDashboardQueryAudit
}

func (r *dataScopeRepoStub) GetDataGroupIDsByUserId(ctx core.Context, userID int64) ([]int64, error) {
	return r.userGroups[userID], nil
}

func (r *dataScopeRepoStub) GetScopeIDsSelectedByPermGroupIDs(ctx core.Context, groupIDs []int64) ([]string, error) {
	var scopeIDs []string
	for _, groupID := range groupIDs {
		scopeIDs = append(scopeIDs, r.groupScopes[groupID]...)
	}
	return scopeIDs, nil
}

func (r *dataScopeRepoStub) GetScopesByScopeIDs(ctx core.Context, scopeIDs []string) ([]datagroup.DataScope, error) {
	scopes := make([]datagroup.DataScope, 0, len(scopeIDs))
	for _, scopeID := range scopeIDs {
		scopes = append(scopes, normalizeScopeForTest(r.scopes[scopeID]))
	}
	return scopes, nil
}

func normalizeScopeForTest(scope datagroup.DataScope) datagroup.DataScope {
	if scope.ScopeID == "" {
		scope.ScopeID = scope.ScopeLabels.ToScopeID()
	}
	if scope.Type == "" {
		switch {
		case scope.Service != "":
			scope.Type = datagroup.DATASOURCE_TYP_SERVICE
		case scope.Namespace != "":
			scope.Type = datagroup.DATASOURCE_TYP_NAMESPACE
		case scope.ClusterID != "":
			scope.Type = datagroup.DATASOURCE_TYP_CLUSTER
		}
	}
	if scope.Name == "" {
		switch scope.Type {
		case datagroup.DATASOURCE_TYP_SERVICE:
			scope.Name = scope.Service
		case datagroup.DATASOURCE_TYP_NAMESPACE:
			scope.Name = scope.Namespace
		case datagroup.DATASOURCE_TYP_CLUSTER:
			scope.Name = scope.ClusterID
		}
	}
	return scope
}

func (r *dataScopeRepoStub) GetScopesByGroupIDAndCat(ctx core.Context, groupID int64, category string) ([]datagroup.DataScope, error) {
	return r.GetScopesByScopeIDs(ctx, r.groupScopes[groupID])
}

func (r *dataScopeRepoStub) CreateCustomDashboardQueryAudit(ctx core.Context, audit *database.CustomDashboardQueryAudit) error {
	r.audits = append(r.audits, *audit)
	return nil
}

func (r *dataScopeRepoStub) CheckGroupPermission(ctx core.Context, userID, groupID int64, typ string) (bool, error) {
	allowed, ok := r.groupAllowed[groupID]
	if !ok {
		return false, nil
	}
	return allowed, nil
}

type queryRecorderProm struct {
	prometheus.Repo
	queries         []string
	labelQueries    []string
	labelStartTimes []int64
	labelEndTimes   []int64
	labelValues     prommodel.LabelValues
	err             error
	results         map[string][]prometheus.MetricResult
}

func (p *queryRecorderProm) QueryRangeData(ctx core.Context, startTime time.Time, endTime time.Time, query string, step time.Duration) ([]prometheus.MetricResult, error) {
	p.queries = append(p.queries, query)
	if p.err != nil {
		return nil, p.err
	}
	if p.results != nil {
		return p.results[query], nil
	}
	return nil, nil
}

func (p *queryRecorderProm) LabelValues(ctx core.Context, expr string, label string, startTime, endTime int64) (prommodel.LabelValues, error) {
	p.labelQueries = append(p.labelQueries, expr)
	p.labelStartTimes = append(p.labelStartTimes, startTime)
	p.labelEndTimes = append(p.labelEndTimes, endTime)
	if p.labelValues != nil {
		return p.labelValues, nil
	}
	return prommodel.LabelValues{"pod-a"}, nil
}

type cancellableLabelProm struct {
	prometheus.Repo
	done chan struct{}
}

func (p *cancellableLabelProm) LabelValues(ctx core.Context, expr string, label string, startTime, endTime int64) (prommodel.LabelValues, error) {
	<-ctx.GetContext().Done()
	close(p.done)
	return nil, ctx.GetContext().Err()
}

func TestLabelValuesWithTimeoutCancelsPrometheusContext(t *testing.T) {
	prom := &cancellableLabelProm{done: make(chan struct{})}
	s := &service{prom: prom}

	_, err := s.labelValuesWithTimeoutFor(core.EmptyCtx(), "up", "pod", 1, 2, 10*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "dashboard variable candidate query timeout") {
		t.Fatalf("expected candidate timeout error, got %v", err)
	}
	select {
	case <-prom.done:
	case <-time.After(time.Second):
		t.Fatalf("expected prometheus LabelValues context to be cancelled by timeout")
	}
}

type cancellableMetricProm struct {
	prometheus.Repo
	done chan struct{}
}

func (p *cancellableMetricProm) QueryRangeData(ctx core.Context, startTime time.Time, endTime time.Time, query string, step time.Duration) ([]prometheus.MetricResult, error) {
	<-ctx.GetContext().Done()
	close(p.done)
	return nil, ctx.GetContext().Err()
}

func TestMetricQueryWithTimeoutCancelsPrometheusContext(t *testing.T) {
	prom := &cancellableMetricProm{done: make(chan struct{})}
	s := &service{prom: prom}

	_, err := s.executeMetricQueryWithTimeout(core.EmptyCtx(), false, "up", 1, 2, 60, 10*time.Millisecond)
	if err == nil || !strings.Contains(err.Error(), "dashboard query timeout") {
		t.Fatalf("expected metric timeout error, got %v", err)
	}
	select {
	case <-prom.done:
	case <-time.After(time.Second):
		t.Fatalf("expected prometheus metric query context to be cancelled by timeout")
	}
}

func TestPreviewQueryRecordsDashboardIDInAudit(t *testing.T) {
	db := &dataScopeRepoStub{
		userGroups:  map[int64][]int64{101: {0}},
		groupScopes: map[int64][]string{},
		scopes:      map[string]datagroup.DataScope{},
	}
	prom := &queryRecorderProm{}
	s := &service{dbRepo: db, prom: prom}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	resp, err := s.PreviewQuery(ctx, &request.DashboardQueryRequest{
		DashboardID: "dashboard-1",
		WidgetID:    "widget-1",
		Mode:        "custom",
		Expr:        "up",
		GroupID:     int64Ptr(0),
		StartTime:   time.Unix(0, 0).UnixMilli(),
		EndTime:     time.Unix(60, 0).UnixMilli(),
		Step:        60,
	})
	if err != nil {
		t.Fatalf("preview query: %v", err)
	}
	if resp.AuditID == 0 && len(db.audits) != 1 {
		t.Fatalf("expected exactly one audit row, got response audit %d and rows %#v", resp.AuditID, db.audits)
	}
	if len(db.audits) != 1 || db.audits[0].DashboardID != "dashboard-1" {
		t.Fatalf("expected audit dashboard id dashboard-1, got %#v", db.audits)
	}
}

func TestSharedDashboardQueryUsesCurrentViewerDataScope(t *testing.T) {
	db := &dataScopeRepoStub{
		userGroups: map[int64][]int64{
			101: {201},
			102: {202},
		},
		groupScopes: map[int64][]string{
			201: {"301"},
			202: {"302"},
		},
		scopes: map[string]datagroup.DataScope{
			"301": {ScopeLabels: datagroup.ScopeLabels{ClusterID: "cluster-a", Namespace: "prod", Service: "checkout"}},
			"302": {ScopeLabels: datagroup.ScopeLabels{ClusterID: "cluster-b", Namespace: "staging", Service: "cart"}},
		},
	}
	prom := &queryRecorderProm{}
	s := &service{dbRepo: db, prom: prom}
	req := &request.DashboardQueryRequest{
		WidgetID:  "shared-widget",
		Mode:      "custom",
		Expr:      "up",
		StartTime: time.Unix(0, 0).UnixMilli(),
		EndTime:   time.Unix(60, 0).UnixMilli(),
		Step:      60,
	}

	viewerA := core.EmptyCtx()
	viewerA.Set(core.UserIDKey, int64(101))
	req.GroupID = int64Ptr(201)
	if _, err := s.PreviewQuery(viewerA, req); err != nil {
		t.Fatalf("preview for viewer A: %v", err)
	}

	viewerB := core.EmptyCtx()
	viewerB.Set(core.UserIDKey, int64(102))
	req.GroupID = int64Ptr(202)
	if _, err := s.PreviewQuery(viewerB, req); err != nil {
		t.Fatalf("preview for viewer B: %v", err)
	}

	if len(prom.queries) != 2 {
		t.Fatalf("expected two backend metric queries, got %d", len(prom.queries))
	}
	wantA := `up{namespace="prod",cluster_id="cluster-a",svc_name=~"checkout"}`
	wantB := `up{namespace="staging",cluster_id="cluster-b",svc_name=~"cart"}`
	if prom.queries[0] != wantA {
		t.Fatalf("viewer A query should be constrained to viewer A data scope\nwant: %s\n got: %s", wantA, prom.queries[0])
	}
	if prom.queries[1] != wantB {
		t.Fatalf("viewer B query should be constrained to viewer B data scope\nwant: %s\n got: %s", wantB, prom.queries[1])
	}
}

func TestPreviewQueryReturnsNoPermissionState(t *testing.T) {
	db := &dataScopeRepoStub{groupAllowed: map[int64]bool{201: false}}
	s := &service{dbRepo: db, prom: &queryRecorderProm{}}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	resp, err := s.PreviewQuery(ctx, &request.DashboardQueryRequest{
		WidgetID:  "w-1",
		Mode:      "custom",
		Expr:      "up",
		GroupID:   int64Ptr(201),
		StartTime: time.Unix(0, 0).UnixMilli(),
		EndTime:   time.Unix(60, 0).UnixMilli(),
		Step:      60,
	})
	if err != nil {
		t.Fatalf("preview should return widget state, not handler error: %v", err)
	}
	if resp.Status != "no_permission" {
		t.Fatalf("expected no_permission state, got %q", resp.Status)
	}
}

func TestPreviewQueryUsesPageGroupIDAndIgnoresDataGroupVariable(t *testing.T) {
	db := &dataScopeRepoStub{
		userGroups:   map[int64][]int64{101: {202}},
		groupAllowed: map[int64]bool{201: true, 202: true},
		groupScopes: map[int64][]string{
			201: {"301"},
			202: {"302"},
		},
		scopes: map[string]datagroup.DataScope{
			"301": {ScopeLabels: datagroup.ScopeLabels{ClusterID: "ignored", Namespace: "old", Service: "orders"}},
			"302": {ScopeLabels: datagroup.ScopeLabels{ClusterID: "selected", Namespace: "prod", Service: "checkout"}},
		},
	}
	prom := &queryRecorderProm{}
	s := &service{dbRepo: db, prom: prom}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	_, err := s.PreviewQuery(ctx, &request.DashboardQueryRequest{
		WidgetID:  "w-1",
		Mode:      "custom",
		Expr:      "up",
		GroupID:   int64Ptr(202),
		Variables: map[string][]string{"dataGroup": {"201"}},
		StartTime: time.Unix(0, 0).UnixMilli(),
		EndTime:   time.Unix(60, 0).UnixMilli(),
		Step:      60,
	})
	if err != nil {
		t.Fatalf("preview query: %v", err)
	}

	if got, want := prom.queries[0], `up{namespace="prod",cluster_id="selected",svc_name=~"checkout"}`; got != want {
		t.Fatalf("query should use page groupId scope\nwant: %s\n got: %s", want, got)
	}
}

func TestPreviewQueryTreatsTopLevelAllDataAsAuthoritative(t *testing.T) {
	db := &dataScopeRepoStub{
		userGroups: map[int64][]int64{101: {0}},
		groupAllowed: map[int64]bool{
			201: true,
		},
		groupScopes: map[int64][]string{
			201: {"301"},
		},
		scopes: map[string]datagroup.DataScope{
			"301": {ScopeLabels: datagroup.ScopeLabels{ClusterID: "ignored", Namespace: "old", Service: "orders"}},
		},
	}
	prom := &queryRecorderProm{}
	s := &service{dbRepo: db, prom: prom}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	resp, err := s.PreviewQuery(ctx, &request.DashboardQueryRequest{
		WidgetID:  "w-1",
		Mode:      "custom",
		Expr:      "up",
		GroupID:   int64Ptr(0),
		Variables: map[string][]string{"dataGroup": {"201"}},
		StartTime: time.Unix(0, 0).UnixMilli(),
		EndTime:   time.Unix(60, 0).UnixMilli(),
		Step:      60,
	})
	if err != nil {
		t.Fatalf("preview query: %v", err)
	}
	if resp.Status == "no_permission" {
		t.Fatalf("all-data top-level groupId should not fall back to legacy dataGroup variable")
	}
	if got, want := prom.queries[0], "up"; got != want {
		t.Fatalf("query should use all-data scope\nwant: %s\n got: %s", want, got)
	}
}

func TestPreviewQueryDoesNotUseDataGroupVariableAsScope(t *testing.T) {
	db := &dataScopeRepoStub{
		userGroups:   map[int64][]int64{101: {}},
		groupAllowed: map[int64]bool{201: true},
		groupScopes: map[int64][]string{
			201: {"301"},
		},
		scopes: map[string]datagroup.DataScope{
			"301": {ScopeLabels: datagroup.ScopeLabels{ClusterID: "legacy", Namespace: "prod", Service: "checkout"}},
		},
	}
	prom := &queryRecorderProm{}
	s := &service{dbRepo: db, prom: prom}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	resp, err := s.PreviewQuery(ctx, &request.DashboardQueryRequest{
		WidgetID:  "w-1",
		Mode:      "custom",
		Expr:      "up",
		Variables: map[string][]string{"dataGroup": {"201"}},
		StartTime: time.Unix(0, 0).UnixMilli(),
		EndTime:   time.Unix(60, 0).UnixMilli(),
		Step:      60,
	})
	if err != nil {
		t.Fatalf("preview query: %v", err)
	}
	if resp.Status != "no_permission" {
		t.Fatalf("missing page groupId must not use dataGroup variable, got status %q query %q", resp.Status, resp.Query)
	}
	if len(prom.queries) != 0 {
		t.Fatalf("query should not execute without an allowed page groupId, got %#v", prom.queries)
	}
}

func TestPreviewQueryRequiresPageGroupIDEvenForRootUser(t *testing.T) {
	db := &dataScopeRepoStub{
		userGroups: map[int64][]int64{101: {0}},
		groupScopes: map[int64][]string{
			201: {"301"},
		},
		scopes: map[string]datagroup.DataScope{
			"301": {ScopeLabels: datagroup.ScopeLabels{ClusterID: "legacy", Namespace: "prod", Service: "checkout"}},
		},
	}
	prom := &queryRecorderProm{}
	s := &service{dbRepo: db, prom: prom}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	resp, err := s.PreviewQuery(ctx, &request.DashboardQueryRequest{
		WidgetID:  "w-1",
		Mode:      "custom",
		Expr:      "up",
		Variables: map[string][]string{"dataGroup": {"201"}},
		StartTime: time.Unix(0, 0).UnixMilli(),
		EndTime:   time.Unix(60, 0).UnixMilli(),
		Step:      60,
	})
	if err != nil {
		t.Fatalf("preview query: %v", err)
	}
	if resp.Status != "no_permission" {
		t.Fatalf("missing page groupId must not use root user fallback, got status %q query %q", resp.Status, resp.Query)
	}
	if len(prom.queries) != 0 {
		t.Fatalf("query should not execute without page groupId, got %#v", prom.queries)
	}
}

func TestPreviewQueryUsesTreePermissionForPageGroup(t *testing.T) {
	previousStorage := common.DataGroupStorage
	defer func() { common.DataGroupStorage = previousStorage }()
	common.DataGroupStorage = &common.DataGroupStore{
		DataGroupTreeNode: &datagroup.DataGroupTreeNode{
			DataGroup: datagroup.DataGroup{GroupID: 0, GroupName: "ALL", ParentGroupID: -1},
			SubGroups: []*datagroup.DataGroupTreeNode{
				{DataGroup: datagroup.DataGroup{GroupID: 770368178368, GroupName: "train-ticket", ParentGroupID: 0}},
			},
		},
	}

	db := &dataScopeRepoStub{
		userGroups: map[int64][]int64{101: {0}},
		groupAllowed: map[int64]bool{
			770368178368: false,
		},
		groupScopes: map[int64][]string{
			770368178368: {"dev#train-ticket"},
		},
		scopes: map[string]datagroup.DataScope{
			"dev#train-ticket": {
				Name:        "train-ticket",
				Type:        datagroup.DATASOURCE_TYP_NAMESPACE,
				ScopeLabels: datagroup.ScopeLabels{ClusterID: "dev", Namespace: "train-ticket"},
			},
		},
	}
	prom := &queryRecorderProm{}
	s := &service{dbRepo: db, prom: prom}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	_, err := s.PreviewQuery(ctx, &request.DashboardQueryRequest{
		WidgetID:  "w-1",
		Mode:      "custom",
		Expr:      "up",
		GroupID:   int64Ptr(770368178368),
		Variables: map[string][]string{},
		StartTime: time.Unix(0, 0).UnixMilli(),
		EndTime:   time.Unix(60, 0).UnixMilli(),
		Step:      60,
	})
	if err != nil {
		t.Fatalf("preview query: %v", err)
	}
	if len(prom.queries) != 1 {
		t.Fatalf("expected one query, got %#v", prom.queries)
	}
	if strings.Contains(prom.queries[0], `svc_name=~`) {
		t.Fatalf("namespace page scope must not be narrowed to selected services: %s", prom.queries[0])
	}
}

func TestLabelValueCandidatesUseTopLevelGroupIDScope(t *testing.T) {
	db := &dataScopeRepoStub{
		userGroups:   map[int64][]int64{101: {0}},
		groupAllowed: map[int64]bool{201: true},
		groupScopes: map[int64][]string{
			201: {"301"},
		},
		scopes: map[string]datagroup.DataScope{
			"301": {ScopeLabels: datagroup.ScopeLabels{ClusterID: "cluster-a", Namespace: "prod", Service: "checkout"}},
		},
	}
	prom := &queryRecorderProm{}
	s := &service{dbRepo: db, prom: prom}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	_, err := s.VariableCandidates(ctx, &request.DashboardVariableRequest{
		Type:         "label",
		Label:        "job_name",
		VariableName: "job_name",
		Query:        `label_values(node_cpu_seconds_total,job_name)`,
		GroupID:      int64Ptr(201),
	})
	if err != nil {
		t.Fatalf("variable candidates: %v", err)
	}
	if len(prom.labelQueries) != 1 {
		t.Fatalf("expected one label query, got %d", len(prom.labelQueries))
	}
	if got, want := prom.labelQueries[0], `node_cpu_seconds_total{namespace="prod",cluster_id="cluster-a",svc_name=~"checkout"}`; got != want {
		t.Fatalf("label query should include selected group scope\nwant: %s\n got: %s", want, got)
	}
}

func TestPreviewQueryExecutesMultipleCustomQueries(t *testing.T) {
	prom := &queryRecorderProm{
		results: map[string][]prometheus.MetricResult{
			`up{job="api"}`: {
				{Metric: prometheus.Labels{Job: "api"}, Values: []prometheus.Points{{TimeStamp: 1, Value: 1}}},
			},
			`up{job="worker"}`: {
				{Metric: prometheus.Labels{Job: "worker"}, Values: []prometheus.Points{{TimeStamp: 1, Value: 2}}},
			},
		},
	}
	s := &service{dbRepo: &dataScopeRepoStub{userGroups: map[int64][]int64{101: {0}}}, prom: prom}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	resp, err := s.PreviewQuery(ctx, &request.DashboardQueryRequest{
		WidgetID: "w-1",
		Mode:     "custom",
		GroupID:  int64Ptr(0),
		Queries: []request.DashboardQueryItem{
			{ID: "q-a", Expression: `up{job="api"}`, Legend: "API"},
			{ID: "q-b", Expression: `up{job="worker"}`, Legend: "Worker"},
		},
		StartTime: 1,
		EndTime:   2,
		Step:      60,
	})
	if err != nil {
		t.Fatalf("preview should not return error: %v", err)
	}
	if resp.Status != "success" {
		t.Fatalf("expected success, got %q: %s", resp.Status, resp.Message)
	}
	if !slices.Equal(prom.queries, []string{`up{job="api"}`, `up{job="worker"}`}) {
		t.Fatalf("unexpected executed queries: %v", prom.queries)
	}
	if len(resp.Series) != 2 {
		t.Fatalf("expected merged series from both queries, got %d", len(resp.Series))
	}
	if resp.Series[0].Metric.QueryLegend != "API" || resp.Series[1].Metric.QueryLegend != "Worker" {
		t.Fatalf("query legends should be attached to response series: %#v", resp.Series)
	}
}

func TestPreviewQueryCapsChartSeriesAndReportsLimit(t *testing.T) {
	const query = `up`
	results := make([]prometheus.MetricResult, 60)
	for index := range results {
		results[index] = prometheus.MetricResult{
			Metric: prometheus.Labels{POD: fmt.Sprintf("pod-%02d", index+1)},
			Values: []prometheus.Points{
				{TimeStamp: 1, Value: float64(index + 1)},
				{TimeStamp: 2, Value: float64(index)},
			},
		}
	}
	prom := &queryRecorderProm{
		results: map[string][]prometheus.MetricResult{query: results},
	}
	s := &service{dbRepo: &dataScopeRepoStub{userGroups: map[int64][]int64{101: {0}}}, prom: prom}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	resp, err := s.PreviewQuery(ctx, &request.DashboardQueryRequest{
		WidgetID:    "w-1",
		Mode:        "custom",
		Expr:        query,
		GroupID:     int64Ptr(0),
		ChartType:   "line",
		SeriesLimit: 50,
		StartTime:   1,
		EndTime:     2,
		Step:        60,
	})
	if err != nil {
		t.Fatalf("preview should not return error: %v", err)
	}
	if len(resp.Series) != 50 {
		t.Fatalf("expected response series to be capped to 50, got %d", len(resp.Series))
	}
	if resp.Series[0].Metric.POD != "pod-60" || resp.Series[49].Metric.POD != "pod-11" {
		t.Fatalf("expected top series by max value, got first %q last %q", resp.Series[0].Metric.POD, resp.Series[49].Metric.POD)
	}
	if !resp.Truncated {
		t.Fatalf("expected response to report truncation")
	}
	if resp.SeriesLimit == nil || !resp.SeriesLimit.Limited || resp.SeriesLimit.Total != 60 || resp.SeriesLimit.Shown != 50 || resp.SeriesLimit.Limit != 50 {
		t.Fatalf("unexpected series limit metadata: %#v", resp.SeriesLimit)
	}
}

func TestNormalizeSeriesLimitBounds(t *testing.T) {
	tests := []struct {
		name  string
		value int
		want  int
	}{
		{name: "default", value: 0, want: 50},
		{name: "negative", value: -1, want: 50},
		{name: "custom", value: 200, want: 200},
		{name: "max", value: 1000, want: 500},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := normalizeSeriesLimit(tt.value); got != tt.want {
				t.Fatalf("normalizeSeriesLimit(%d) = %d, want %d", tt.value, got, tt.want)
			}
		})
	}
}

func TestPreviewQueryPresetUsesRegexForMultiSelectResourceVariables(t *testing.T) {
	prom := &queryRecorderProm{
		results: map[string][]prometheus.MetricResult{
			`sum(rate(span_trace_duration_count{svc_name=~"checkout|cart"}[1m]))`: {
				{Metric: prometheus.Labels{}, Values: []prometheus.Points{{TimeStamp: 1, Value: 2}}},
			},
		},
	}
	s := &service{dbRepo: &dataScopeRepoStub{userGroups: map[int64][]int64{101: {0}}}, prom: prom}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	resp, err := s.PreviewQuery(ctx, &request.DashboardQueryRequest{
		Mode:     "preset",
		PresetID: "service-qps",
		Params:   map[string]string{"service": "$service"},
		Variables: map[string][]string{
			"service": {"checkout", "cart"},
		},
		GroupID:   int64Ptr(0),
		StartTime: 1,
		EndTime:   2,
		Step:      60,
	})
	if err != nil {
		t.Fatalf("preview should not return error: %v", err)
	}
	if resp.Status != "success" {
		t.Fatalf("expected success, got %q: %s", resp.Status, resp.Message)
	}
	want := []string{`sum(rate(span_trace_duration_count{svc_name=~"checkout|cart"}[1m]))`}
	if !slices.Equal(prom.queries, want) {
		t.Fatalf("preset should use regex matcher for multi-select resource variable\nwant: %v\n got: %v", want, prom.queries)
	}
}

func TestPreviewQueryCountsAllVisibleQueriesTowardPointLimit(t *testing.T) {
	prom := &queryRecorderProm{}
	s := &service{dbRepo: &dataScopeRepoStub{userGroups: map[int64][]int64{101: {0}}}, prom: prom}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	pointsPerQuery := int64(maxQueryPoints/2 + 1)
	start := time.Unix(1, 0).UnixMilli()
	end := start + pointsPerQuery*60*1000
	resp, err := s.PreviewQuery(ctx, &request.DashboardQueryRequest{
		Mode:    "custom",
		GroupID: int64Ptr(0),
		Queries: []request.DashboardQueryItem{
			{ID: "q-a", Expression: `up{job="api"}`},
			{ID: "q-b", Expression: `up{job="worker"}`},
		},
		StartTime: start,
		EndTime:   end,
		Step:      60,
	})
	if err != nil {
		t.Fatalf("preview should not return error: %v", err)
	}
	if resp.Status != "too_large" {
		t.Fatalf("expected too_large for total multi-query points, got %q: %s", resp.Status, resp.Message)
	}
	if len(prom.queries) != 0 {
		t.Fatalf("too_large queries should not execute Prometheus queries, got %v", prom.queries)
	}
}

func TestPreviewQuerySkipsHiddenCustomQueries(t *testing.T) {
	prom := &queryRecorderProm{
		results: map[string][]prometheus.MetricResult{
			`up{job="worker"}`: {
				{Metric: prometheus.Labels{Job: "worker"}, Values: []prometheus.Points{{TimeStamp: 1, Value: 2}}},
			},
		},
	}
	s := &service{dbRepo: &dataScopeRepoStub{userGroups: map[int64][]int64{101: {0}}}, prom: prom}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	resp, err := s.PreviewQuery(ctx, &request.DashboardQueryRequest{
		WidgetID: "w-1",
		Mode:     "custom",
		GroupID:  int64Ptr(0),
		Queries: []request.DashboardQueryItem{
			{ID: "q-a", Expression: `up{job="api"}`, Legend: "API", Hidden: true},
			{ID: "q-b", Expression: `up{job="worker"}`, Legend: "Worker"},
		},
		StartTime: 1,
		EndTime:   2,
		Step:      60,
	})
	if err != nil {
		t.Fatalf("preview should not return error: %v", err)
	}
	if resp.Status != "success" {
		t.Fatalf("expected success, got %q: %s", resp.Status, resp.Message)
	}
	if !slices.Equal(prom.queries, []string{`up{job="worker"}`}) {
		t.Fatalf("hidden query should not execute, got %v", prom.queries)
	}
	if len(resp.Series) != 1 || resp.Series[0].Metric.QueryLegend != "Worker" {
		t.Fatalf("expected visible query series only: %#v", resp.Series)
	}
}

func TestPreviewQueryTreatsExplicitEmptyCustomQueriesAsNoData(t *testing.T) {
	prom := &queryRecorderProm{}
	s := &service{dbRepo: &dataScopeRepoStub{userGroups: map[int64][]int64{101: {0}}}, prom: prom}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	resp, err := s.PreviewQuery(ctx, &request.DashboardQueryRequest{
		Mode:      "custom",
		Expr:      "",
		Queries:   []request.DashboardQueryItem{},
		StartTime: 1,
		EndTime:   2,
		Step:      60,
	})
	if err != nil {
		t.Fatalf("preview should not return error: %v", err)
	}
	if resp.Status != "no_data" {
		t.Fatalf("expected no_data for all hidden queries, got %q: %s", resp.Status, resp.Message)
	}
	if len(prom.queries) != 0 {
		t.Fatalf("explicit empty queries should not execute Prometheus queries, got %v", prom.queries)
	}
}

func TestPreviewQueryReportsMultiQuerySyntaxError(t *testing.T) {
	s := &service{dbRepo: &dataScopeRepoStub{userGroups: map[int64][]int64{101: {0}}}, prom: &queryRecorderProm{}}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	resp, err := s.PreviewQuery(ctx, &request.DashboardQueryRequest{
		Mode: "custom",
		Queries: []request.DashboardQueryItem{
			{ID: "q-a", Expression: "up", Legend: "OK"},
			{ID: "q-b", Expression: "sum(up", Legend: "Broken"},
		},
		StartTime: 1,
		EndTime:   2,
		Step:      60,
	})
	if err != nil {
		t.Fatalf("preview should return widget state, not handler error: %v", err)
	}
	if resp.Status != "syntax_error" {
		t.Fatalf("expected syntax_error, got %q", resp.Status)
	}
	if !strings.Contains(resp.Message, "query 2") {
		t.Fatalf("expected message to identify query 2, got %q", resp.Message)
	}
}

func TestCatalogPresetQueriesExposeOnlySupportedWidgetTypes(t *testing.T) {
	for _, preset := range catalogPresetQueries() {
		for _, chart := range preset.CompatibleCharts {
			if chart == cdmodel.WidgetSingle || chart == cdmodel.WidgetTopN {
				t.Fatalf("preset %s exposes removed chart type %q", preset.ID, chart)
			}
		}
	}

	for _, template := range catalogTemplates() {
		for _, widget := range template.Dashboard.Widgets {
			if widget.Type == cdmodel.WidgetSingle || widget.Type == cdmodel.WidgetTopN {
				t.Fatalf("template %s exposes removed widget type %q", template.ID, widget.Type)
			}
		}
	}
}

func TestServiceHealthCatalogIncludesLatencyPresets(t *testing.T) {
	requiredPresets := map[string]bool{
		"service-latency-avg": false,
		"service-latency-p99": false,
	}
	for _, preset := range catalogPresetQueries() {
		if _, required := requiredPresets[preset.ID]; required {
			requiredPresets[preset.ID] = true
			if !slices.Contains(preset.CompatibleCharts, cdmodel.WidgetLine) {
				t.Fatalf("%s preset should support line charts", preset.ID)
			}
		}
	}
	for presetID, exists := range requiredPresets {
		if !exists {
			t.Fatalf("service health presets should include %s", presetID)
		}
	}

	for _, template := range catalogTemplates() {
		if template.ID != "service-health" {
			continue
		}
		for _, widget := range template.Dashboard.Widgets {
			if widget.Query.PresetID == "service-latency-p99" {
				return
			}
		}
		t.Fatalf("service health template should include a P99 latency widget")
	}
	t.Fatalf("service health template is missing")
}

func TestInstanceCatalogIncludesPodNetworkTrafficPreset(t *testing.T) {
	for _, preset := range catalogPresetQueries() {
		if preset.ID != "pod-network-traffic" {
			continue
		}
		if preset.Category != "instance" {
			t.Fatalf("pod network traffic should be an instance preset, got %q", preset.Category)
		}
		if !slices.Contains(preset.CompatibleCharts, cdmodel.WidgetLine) || !slices.Contains(preset.CompatibleCharts, cdmodel.WidgetBar) {
			t.Fatalf("pod network traffic should support line and bar charts")
		}
		if len(preset.Parameters) == 0 || preset.Parameters[0].Name != "pod" {
			t.Fatalf("pod network traffic should expose a pod parameter")
		}
		return
	}
	t.Fatalf("preset pod-network-traffic is missing")
}

func TestInfrastructureCatalogIncludesNetworkTraffic(t *testing.T) {
	hasPreset := false
	for _, preset := range catalogPresetQueries() {
		if preset.ID == "node-network-traffic" {
			hasPreset = true
			if preset.Category != "infrastructure" {
				t.Fatalf("node network traffic should be an infrastructure preset, got %q", preset.Category)
			}
			if !slices.Contains(preset.CompatibleCharts, cdmodel.WidgetLine) {
				t.Fatalf("node network traffic should support line charts")
			}
		}
	}
	if !hasPreset {
		t.Fatalf("preset node-network-traffic is missing")
	}

	for _, template := range catalogTemplates() {
		if template.ID != "infrastructure-resources" {
			continue
		}
		for _, widget := range template.Dashboard.Widgets {
			if widget.Query.PresetID == "node-network-traffic" {
				return
			}
		}
		t.Fatalf("infrastructure resources template should include network traffic")
	}
	t.Fatalf("infrastructure resources template is missing")
}

func TestCatalogTemplatesIncludeMVPRecommendedOperationalWidgets(t *testing.T) {
	requiredPresets := map[string]string{
		"service-instance-status": "service",
		"service-slow-endpoints":  "service",
		"pod-status-distribution": "infrastructure",
		"resource-usage-top":      "infrastructure",
	}
	seenPresets := map[string]cdmodel.PresetQuery{}
	for _, preset := range catalogPresetQueries() {
		if _, required := requiredPresets[preset.ID]; required {
			seenPresets[preset.ID] = preset
		}
	}
	for presetID, category := range requiredPresets {
		preset, exists := seenPresets[presetID]
		if !exists {
			t.Fatalf("catalog preset %s is missing", presetID)
		}
		if preset.Category != category {
			t.Fatalf("preset %s should be category %q, got %q", presetID, category, preset.Category)
		}
		if slices.Contains(preset.CompatibleCharts, cdmodel.WidgetSingle) || slices.Contains(preset.CompatibleCharts, cdmodel.WidgetTopN) {
			t.Fatalf("preset %s must not reintroduce removed widget types: %v", presetID, preset.CompatibleCharts)
		}
		if !slices.Contains(preset.CompatibleCharts, cdmodel.WidgetTable) {
			t.Fatalf("preset %s should support table charts", presetID)
		}
	}

	requiredByTemplate := map[string][]string{
		"service-health": {
			"service-instance-status",
			"service-slow-endpoints",
		},
		"infrastructure-resources": {
			"pod-status-distribution",
			"resource-usage-top",
		},
	}
	templates := map[string]cdmodel.Template{}
	for _, template := range catalogTemplates() {
		templates[template.ID] = template
	}
	for templateID, presetIDs := range requiredByTemplate {
		template, exists := templates[templateID]
		if !exists {
			t.Fatalf("template %s is missing", templateID)
		}
		for _, presetID := range presetIDs {
			found := false
			for _, widget := range template.Dashboard.Widgets {
				if widget.Query.PresetID == presetID {
					found = true
					break
				}
			}
			if !found {
				t.Fatalf("template %s should include preset %s", templateID, presetID)
			}
		}
	}
}

func TestNormalizeDashboardDefinitionConvertsHistoricalTopNOnly(t *testing.T) {
	dashboard, err := unmarshalDashboardConfig(`{
		"widgets": [
			{"id": "top", "type": "topn", "layout": {"x": 0, "y": 0}},
			{"id": "single", "type": "single", "layout": {"x": 1, "y": 0}}
		]
	}`)
	if err != nil {
		t.Fatalf("unmarshal dashboard: %v", err)
	}
	if dashboard.Widgets[0].Type != cdmodel.WidgetTable {
		t.Fatalf("historical topn should normalize to table, got %q", dashboard.Widgets[0].Type)
	}
	if dashboard.Widgets[1].Type != cdmodel.WidgetSingle {
		t.Fatalf("historical single should be preserved for unsupported rendering, got %q", dashboard.Widgets[1].Type)
	}
}

func TestNormalizeDashboardDefinitionPreservesWidgetLayoutSize(t *testing.T) {
	dashboard, err := unmarshalDashboardConfig(`{
		"widgets": [
			{"id": "wide", "type": "line", "layout": {"x": 0, "y": 0, "w": 4, "h": 3}},
			{"id": "below", "type": "line", "layout": {"x": 0, "y": 3, "w": 1, "h": 2}}
		]
	}`)
	if err != nil {
		t.Fatalf("unmarshal dashboard: %v", err)
	}
	if got := dashboard.Widgets[0].Layout; got.X != 0 || got.Y != 0 || got.W != 4 || got.H != 3 {
		t.Fatalf("wide layout size was not preserved: %#v", got)
	}
	if got := dashboard.Widgets[1].Layout; got.X != 0 || got.Y != 3 || got.W != 1 || got.H != 2 {
		t.Fatalf("below layout size was not preserved: %#v", got)
	}

	dashboard, err = unmarshalDashboardConfig(`{
		"widgets": [
			{"id": "left", "type": "line", "layout": {"x": 0, "y": 0, "w": 1, "h": 1}},
			{"id": "right", "type": "line", "layout": {"x": 1, "y": 0, "w": 1, "h": 1}},
			{"id": "wide", "type": "line", "layout": {"x": 0, "y": 0, "w": 2, "h": 2}}
		]
	}`)
	if err != nil {
		t.Fatalf("unmarshal colliding dashboard: %v", err)
	}
	if got := dashboard.Widgets[2].Layout; got.X != 2 || got.Y != 0 || got.W != 2 || got.H != 2 {
		t.Fatalf("wide layout should move without shrinking: %#v", got)
	}
}

func TestNormalizeDashboardDefinitionExpandsLegacyTwoColumnLayouts(t *testing.T) {
	dashboard, err := unmarshalDashboardConfig(`{
		"widgets": [
			{"id": "left", "type": "line", "layout": {"x": 0, "y": 0}},
			{"id": "right", "type": "line", "layout": {"x": 1, "y": 0}}
		]
	}`)
	if err != nil {
		t.Fatalf("unmarshal dashboard: %v", err)
	}
	if got := dashboard.Widgets[0].Layout; got.X != 0 || got.Y != 0 || got.W != 2 || got.H != 1 {
		t.Fatalf("left legacy layout should expand to half-width: %#v", got)
	}
	if got := dashboard.Widgets[1].Layout; got.X != 2 || got.Y != 0 || got.W != 2 || got.H != 1 {
		t.Fatalf("right legacy layout should expand to half-width: %#v", got)
	}
}

func TestNormalizeDashboardDefinitionPreservesWidgetStateText(t *testing.T) {
	dashboard, err := unmarshalDashboardConfig(`{
		"widgets": [
			{
				"id": "state-text",
				"type": "line",
				"display": {
					"emptyText": "Nothing matched this service",
					"errorText": "Metric query failed"
				}
			}
		]
	}`)
	if err != nil {
		t.Fatalf("unmarshal dashboard: %v", err)
	}
	display := dashboard.Widgets[0].Display
	if display.EmptyText != "Nothing matched this service" {
		t.Fatalf("empty text was not preserved: %q", display.EmptyText)
	}
	if display.ErrorText != "Metric query failed" {
		t.Fatalf("error text was not preserved: %q", display.ErrorText)
	}
}

func TestNormalizeDashboardDefinitionPreservesWidgetThresholds(t *testing.T) {
	dashboard, err := unmarshalDashboardConfig(`{
		"widgets": [
			{
				"id": "thresholds",
				"type": "line",
				"display": {
					"warningThreshold": 80,
					"criticalThreshold": 95
				}
			}
		]
	}`)
	if err != nil {
		t.Fatalf("unmarshal dashboard: %v", err)
	}
	display := dashboard.Widgets[0].Display
	if display.WarningThreshold == nil || *display.WarningThreshold != 80 {
		t.Fatalf("warning threshold was not preserved: %#v", display.WarningThreshold)
	}
	if display.CriticalThreshold == nil || *display.CriticalThreshold != 95 {
		t.Fatalf("critical threshold was not preserved: %#v", display.CriticalThreshold)
	}
}

func TestNormalizeDashboardDefinitionPreservesWidgetUnitPreset(t *testing.T) {
	dashboard, err := unmarshalDashboardConfig(`{
		"widgets": [
			{
				"id": "unit-preset",
				"type": "stat",
				"display": {
					"unit": "B",
					"unitPreset": "bytes",
					"autoScale": false,
					"decimals": 2
				}
			}
		]
	}`)
	if err != nil {
		t.Fatalf("unmarshal dashboard: %v", err)
	}
	display := dashboard.Widgets[0].Display
	if display.UnitPreset != "bytes" {
		t.Fatalf("unit preset was not preserved: %q", display.UnitPreset)
	}
	if display.AutoScale {
		t.Fatalf("auto scale false should be preserved")
	}
}

func TestNormalizeDashboardDefinitionPreservesWidgetTableColumns(t *testing.T) {
	dashboard, err := unmarshalDashboardConfig(`{
		"widgets": [
			{
				"id": "table-transform",
				"type": "table",
				"display": {
					"tableColumns": [
						{"name": "service_name", "label": "服务名"},
						{"name": "service_version", "label": "版本"}
					],
					"tableHiddenColumns": ["__name__", "instance"],
					"showValueColumn": false
				}
			}
		]
	}`)
	if err != nil {
		t.Fatalf("unmarshal dashboard: %v", err)
	}
	bytes, err := json.Marshal(dashboard.Widgets[0].Display)
	if err != nil {
		t.Fatalf("marshal display: %v", err)
	}
	var display map[string]any
	if err := json.Unmarshal(bytes, &display); err != nil {
		t.Fatalf("unmarshal display map: %v", err)
	}
	if got, ok := display["showValueColumn"].(bool); !ok || got {
		t.Fatalf("showValueColumn=false was not preserved: %#v", display["showValueColumn"])
	}
	columns, ok := display["tableColumns"].([]any)
	if !ok || len(columns) != 2 {
		t.Fatalf("table columns were not preserved: %#v", display["tableColumns"])
	}
	first, ok := columns[0].(map[string]any)
	if !ok || first["name"] != "service_name" || first["label"] != "服务名" {
		t.Fatalf("first table column was not preserved: %#v", columns[0])
	}
	hidden, ok := display["tableHiddenColumns"].([]any)
	if !ok || len(hidden) != 2 || hidden[1] != "instance" {
		t.Fatalf("table hidden columns were not preserved: %#v", display["tableHiddenColumns"])
	}
}

func TestNormalizeDashboardDefinitionPreservesHiddenQueryItems(t *testing.T) {
	dashboard, err := unmarshalDashboardConfig(`{
		"widgets": [
			{
				"id": "multi-query",
				"type": "line",
				"query": {
					"mode": "custom",
					"queries": [
						{"id": "q-a", "expression": "up", "legend": "Availability", "hidden": true},
						{"id": "q-b", "expression": "rate(http_requests_total[5m])", "legend": "Requests"}
					]
				}
			}
		]
	}`)
	if err != nil {
		t.Fatalf("unmarshal dashboard: %v", err)
	}
	queries := dashboard.Widgets[0].Query.Queries
	if len(queries) != 2 || !queries[0].Hidden || queries[1].Hidden {
		t.Fatalf("hidden query state was not preserved: %#v", queries)
	}
	bytes, err := json.Marshal(dashboard.Widgets[0].Query)
	if err != nil {
		t.Fatalf("marshal query: %v", err)
	}
	if !strings.Contains(string(bytes), `"hidden":true`) {
		t.Fatalf("hidden query state should be serialized: %s", string(bytes))
	}
}

func TestNormalizeDashboardDefinitionPreservesGrafanaStyleVariableMetadata(t *testing.T) {
	dashboard, err := unmarshalDashboardConfig(`{
		"variables": [
			{
				"name": "resolution",
				"label": "Resolution",
				"type": "custom",
				"defaultValue": "1m",
				"includeAll": true,
				"allValue": ".*",
				"options": [
					{"label": "1m", "value": "1m"},
					{"label": "5m", "value": "5m"}
				]
			},
			{
				"name": "job_name",
				"label": "Node Job",
				"type": "label",
				"labels": ["job_name"],
				"query": "label_values(node_cpu_seconds_total{cluster_id=~\"$cluster\"},job_name)"
			}
		]
	}`)
	if err != nil {
		t.Fatalf("unmarshal dashboard: %v", err)
	}
	if len(dashboard.Variables) != 2 {
		t.Fatalf("expected two variables, got %d", len(dashboard.Variables))
	}
	if got := dashboard.Variables[0].Options; len(got) != 2 || got[0].Value != "1m" || got[1].Value != "5m" {
		t.Fatalf("static variable options were not preserved: %#v", got)
	}
	if !dashboard.Variables[0].IncludeAll || dashboard.Variables[0].AllValue != ".*" {
		t.Fatalf("includeAll metadata was not preserved: %#v", dashboard.Variables[0])
	}
	if got := dashboard.Variables[1].Labels; len(got) != 1 || got[0] != "job_name" {
		t.Fatalf("label variable labels were not preserved: %#v", got)
	}
	if dashboard.Variables[1].Query == "" {
		t.Fatalf("label variable query should be preserved")
	}
}

func TestNormalizeDashboardDefinitionDefaultsBuiltinResourceVariablesToAll(t *testing.T) {
	dashboard, err := unmarshalDashboardConfig(`{
		"variables": [
			{"name": "cluster", "label": "Cluster", "type": "cluster"},
			{"name": "namespace", "label": "Namespace", "type": "namespace"},
			{"name": "service_name", "label": "Service", "type": "label", "value": "sandbox-default-user-3", "defaultValue": "sandbox-default-user-3", "labels": ["service_name"]},
			{"name": "pod", "label": "Pod", "type": "label", "labels": ["pod"]},
			{"name": "endpoint", "label": "Endpoint", "type": "endpoint", "multi": false},
			{"name": "related_nodes", "label": "Related Nodes", "type": "label", "labels": ["node_name"]},
			{"name": "job_name", "label": "Job", "type": "label", "labels": ["job_name"]}
		]
	}`)
	if err != nil {
		t.Fatalf("unmarshal dashboard: %v", err)
	}
	for _, name := range []string{"cluster", "namespace", "service_name", "pod"} {
		var found cdmodel.Variable
		for _, variable := range dashboard.Variables {
			if variable.Name == name {
				found = variable
				break
			}
		}
		if !found.IncludeAll || found.AllValue != ".*" {
			t.Fatalf("%s variable should default to All: %#v", name, found)
		}
		if found.Value != ".*" || found.DefaultValue != ".*" {
			t.Fatalf("%s variable should select All by default: %#v", name, found)
		}
		if !found.Multi {
			t.Fatalf("%s builtin variable should default to multi-select: %#v", name, found)
		}
	}
	var endpoint cdmodel.Variable
	for _, variable := range dashboard.Variables {
		if variable.Name == "endpoint" {
			endpoint = variable
			break
		}
	}
	if endpoint.Multi {
		t.Fatalf("explicit single-select builtin variable should be preserved: %#v", endpoint)
	}
	if !endpoint.IncludeAll || endpoint.AllValue != ".*" {
		t.Fatalf("single-select builtin variable should still default to All: %#v", endpoint)
	}
	for _, name := range []string{"related_nodes", "job_name"} {
		var found cdmodel.Variable
		for _, variable := range dashboard.Variables {
			if variable.Name == name {
				found = variable
				break
			}
		}
		if found.IncludeAll {
			t.Fatalf("%s custom label variable should not default to All: %#v", name, found)
		}
		if !found.Multi {
			t.Fatalf("%s custom label variable should default to multi-select: %#v", name, found)
		}
	}

	blank := blankDashboard()
	blankNames := map[string]bool{}
	for _, variable := range blank.Variables {
		blankNames[variable.Name] = true
		if !variable.IncludeAll || variable.AllValue != ".*" {
			t.Fatalf("blank dashboard builtin variable should default to All: %#v", variable)
		}
		if !variable.Multi {
			t.Fatalf("blank dashboard builtin variable should default to multi-select: %#v", variable)
		}
	}
	for _, name := range []string{"cluster", "namespace", "service", "endpoint", "pod", "instance"} {
		if !blankNames[name] {
			t.Fatalf("blank dashboard should include builtin variable %s: %#v", name, blank.Variables)
		}
	}
}

func TestPreviewQueryRejectsEmptyUserDataScope(t *testing.T) {
	cases := []struct {
		name string
		db   *dataScopeRepoStub
	}{
		{
			name: "no data groups",
			db:   &dataScopeRepoStub{userGroups: map[int64][]int64{101: {}}},
		},
		{
			name: "group has no scopes",
			db: &dataScopeRepoStub{
				userGroups:  map[int64][]int64{101: {201}},
				groupScopes: map[int64][]string{201: {}},
				scopes:      map[string]datagroup.DataScope{},
			},
		},
		{
			name: "scope has no filter labels",
			db: &dataScopeRepoStub{
				userGroups:  map[int64][]int64{101: {201}},
				groupScopes: map[int64][]string{201: {"301"}},
				scopes: map[string]datagroup.DataScope{
					"301": {},
				},
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			prom := &queryRecorderProm{}
			s := &service{dbRepo: tc.db, prom: prom}
			ctx := core.EmptyCtx()
			ctx.Set(core.UserIDKey, int64(101))

			resp, err := s.PreviewQuery(ctx, &request.DashboardQueryRequest{
				Mode:      "custom",
				Expr:      "up",
				StartTime: time.Unix(0, 0).UnixMilli(),
				EndTime:   time.Unix(60, 0).UnixMilli(),
				Step:      60,
			})
			if err != nil {
				t.Fatalf("preview should return widget state, not handler error: %v", err)
			}
			if resp.Status != "no_permission" {
				t.Fatalf("expected no_permission state, got %q", resp.Status)
			}
			if len(prom.queries) != 0 {
				t.Fatalf("expected no Prometheus query for no-permission scope, got %v", prom.queries)
			}
		})
	}
}

func TestPodCandidatesRejectEmptyUserDataScope(t *testing.T) {
	prom := &queryRecorderProm{}
	s := &service{
		dbRepo: &dataScopeRepoStub{userGroups: map[int64][]int64{101: {}}},
		prom:   prom,
	}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	_, err := s.VariableCandidates(ctx, &request.DashboardVariableRequest{Type: "pod"})
	if !errors.Is(err, errNoDataGroupPermission) {
		t.Fatalf("expected no data group permission error, got %v", err)
	}
	if len(prom.labelQueries) != 0 {
		t.Fatalf("expected no Prometheus label query for no-permission scope, got %v", prom.labelQueries)
	}
}

func TestPreviewQueryClassifiesTimeoutAndValidationPosition(t *testing.T) {
	s := &service{dbRepo: &dataScopeRepoStub{userGroups: map[int64][]int64{101: {0}}}, prom: &queryRecorderProm{err: errors.New("dashboard query timeout")}}
	ctx := core.EmptyCtx()
	ctx.Set(core.UserIDKey, int64(101))

	resp, err := s.PreviewQuery(ctx, &request.DashboardQueryRequest{Mode: "custom", Expr: "sum(up", StartTime: 1, EndTime: 2, Step: 60})
	if err != nil {
		t.Fatalf("syntax preview should not return error: %v", err)
	}
	if resp.Status != "syntax_error" || resp.Position == 0 {
		t.Fatalf("expected syntax_error with position, got status=%q position=%d", resp.Status, resp.Position)
	}

	resp, err = s.PreviewQuery(ctx, &request.DashboardQueryRequest{Mode: "custom", Expr: "up", GroupID: int64Ptr(0), StartTime: 1, EndTime: 2, Step: 60})
	if err != nil {
		t.Fatalf("timeout preview should not return error: %v", err)
	}
	if resp.Status != "timeout" {
		t.Fatalf("expected timeout status, got %q", resp.Status)
	}
}
