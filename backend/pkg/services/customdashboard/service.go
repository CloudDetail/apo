// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package customdashboard

import (
	go_context "context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/customdashboard"
	"github.com/CloudDetail/apo/backend/pkg/model/datagroup"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
	"github.com/CloudDetail/apo/backend/pkg/services/common"
	"github.com/google/uuid"
	prommodel "github.com/prometheus/common/model"
)

const (
	maxDashboardWidgets      = 60
	fixedGridColumns         = 4
	defaultWidgetGridWidth   = 2
	legacyDashboardColumns   = 12
	legacyTwoColumnCount     = 2
	maxQueryPoints           = 12000
	defaultStepSeconds       = 60
	queryTimeout             = 30 * time.Second
	variableCandidateLimit   = 1000
	variableCandidateTimeout = 5 * time.Second
	maxConcurrentQuery       = 8
	defaultSeriesLimit       = 50
	maxSeriesLimit           = 500
)

var querySemaphore = make(chan struct{}, maxConcurrentQuery)

var ErrNoDataGroupPermission = errors.New("current user has no data group permission")

var errNoDataGroupPermission = ErrNoDataGroupPermission

type Service interface {
	List(ctx core.Context, req *request.DashboardListRequest) (*response.DashboardListResponse, error)
	Get(ctx core.Context, id string) (*customdashboard.Dashboard, error)
	Create(ctx core.Context, req *request.CreateDashboardRequest) (*customdashboard.Dashboard, error)
	Update(ctx core.Context, req *request.UpdateDashboardRequest) (*customdashboard.Dashboard, error)
	Copy(ctx core.Context, id string) (*customdashboard.Dashboard, error)
	Delete(ctx core.Context, id string) error
	Templates(ctx core.Context) []customdashboard.Template
	PresetQueries(ctx core.Context) []customdashboard.PresetQuery
	VariableCandidates(ctx core.Context, req *request.DashboardVariableRequest) ([]customdashboard.VariableCandidate, error)
	ValidateQuery(ctx core.Context, expr string) response.DashboardValidateQueryResponse
	PreviewQuery(ctx core.Context, req *request.DashboardQueryRequest) (*response.DashboardQueryResponse, error)
}

type service struct {
	dbRepo database.Repo
	prom   prometheus.Repo
}

func New(dbRepo database.Repo, prom prometheus.Repo) Service {
	return &service{dbRepo: dbRepo, prom: prom}
}

func (s *service) List(ctx core.Context, req *request.DashboardListRequest) (*response.DashboardListResponse, error) {
	userID := ctx.UserID()
	teams, err := s.dbRepo.GetUserTeams(ctx, userID)
	if err != nil {
		return nil, err
	}
	rows, err := s.dbRepo.ListCustomDashboards(ctx, userID, teams, req.Keyword)
	if err != nil {
		return nil, err
	}
	dashboards := make([]customdashboard.Dashboard, 0, len(rows))
	for _, row := range rows {
		dashboard, err := s.rowToDashboard(ctx, &row, userID, true)
		if err != nil {
			return nil, err
		}
		dashboards = append(dashboards, *dashboard)
	}
	return &response.DashboardListResponse{Dashboards: dashboards, CanCreate: true}, nil
}

func (s *service) Get(ctx core.Context, id string) (*customdashboard.Dashboard, error) {
	row, err := s.dbRepo.GetCustomDashboard(ctx, id)
	if err != nil {
		return nil, err
	}
	if ok, err := s.canView(ctx, row); err != nil || !ok {
		if err != nil {
			return nil, err
		}
		return nil, core.Error(code.UserNoPermissionError, "")
	}
	dashboard, err := s.rowToDashboard(ctx, row, ctx.UserID(), true)
	if err != nil {
		return nil, err
	}
	_ = s.dbRepo.TouchCustomDashboardVisit(ctx, id, time.Now())
	return dashboard, nil
}

func (s *service) Create(ctx core.Context, req *request.CreateDashboardRequest) (*customdashboard.Dashboard, error) {
	dashboard := blankDashboard()
	if !req.Blank {
		templateID := req.TemplateID
		if templateID == "" {
			templateID = "blank"
		}
		template, err := findTemplate(templateID)
		if err != nil {
			return nil, err
		}
		dashboard = template.Dashboard
	}
	if req.Name != "" {
		dashboard.Name = req.Name
	}
	if req.Description != "" {
		dashboard.Description = req.Description
	}
	dashboard.ID = uuid.NewString()
	dashboard.OwnerID = ctx.UserID()
	if dashboard.Visibility == "" {
		dashboard.Visibility = customdashboard.VisibilityEveryone
	}
	if len(dashboard.Widgets) > maxDashboardWidgets {
		return nil, errors.New("dashboard has too many widgets")
	}
	row, err := s.dashboardToRow(&dashboard)
	if err != nil {
		return nil, err
	}
	if err := s.dbRepo.CreateCustomDashboardWithMenu(ctx, row); err != nil {
		return nil, err
	}
	return s.Get(ctx, dashboard.ID)
}

func (s *service) Update(ctx core.Context, req *request.UpdateDashboardRequest) (*customdashboard.Dashboard, error) {
	row, err := s.dbRepo.GetCustomDashboard(ctx, req.ID)
	if err != nil {
		return nil, err
	}
	if row.OwnerID != ctx.UserID() {
		return nil, core.Error(code.UserNoPermissionError, "")
	}
	if len(req.Widgets) > maxDashboardWidgets {
		return nil, errors.New("dashboard has too many widgets")
	}
	row.Name = req.Name
	row.Description = req.Description
	row.Visibility = normalizeVisibility(req.Visibility)
	teamJSON, _ := json.Marshal(req.TeamIDs)
	row.TeamIDs = string(teamJSON)
	config, err := marshalDashboardConfig(&req.Dashboard)
	if err != nil {
		return nil, err
	}
	row.Config = config
	if err := s.dbRepo.UpdateCustomDashboardWithMenu(ctx, row); err != nil {
		return nil, err
	}
	return s.Get(ctx, req.ID)
}

func (s *service) Copy(ctx core.Context, id string) (*customdashboard.Dashboard, error) {
	dashboard, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	dashboard.ID = uuid.NewString()
	dashboard.OwnerID = ctx.UserID()
	dashboard.Name = copyDashboardName(dashboard.Name, ctx.LANG())
	dashboard.Visibility = customdashboard.VisibilityPersonal
	dashboard.TeamIDs = nil
	row, err := s.dashboardToRow(dashboard)
	if err != nil {
		return nil, err
	}
	if err := s.dbRepo.CreateCustomDashboardWithMenu(ctx, row); err != nil {
		return nil, err
	}
	return s.Get(ctx, dashboard.ID)
}

func (s *service) Delete(ctx core.Context, id string) error {
	row, err := s.dbRepo.GetCustomDashboard(ctx, id)
	if err != nil {
		return err
	}
	if row.OwnerID != ctx.UserID() {
		return errors.New("only dashboard creator can delete")
	}
	return s.dbRepo.DeleteCustomDashboardWithMenu(ctx, id)
}

func (s *service) Templates(ctx core.Context) []customdashboard.Template {
	return catalogTemplates()
}

func (s *service) PresetQueries(ctx core.Context) []customdashboard.PresetQuery {
	return catalogPresetQueries()
}

func (s *service) VariableCandidates(ctx core.Context, req *request.DashboardVariableRequest) ([]customdashboard.VariableCandidate, error) {
	variables := variableValuesFromRequest(req)
	if req.Type == "timeRange" {
		return []customdashboard.VariableCandidate{}, nil
	}
	if req.Type == "labels" {
		return labelKeyCandidates(req.Query), nil
	}
	if req.Type == "label" || req.Label != "" {
		return s.labelValueCandidates(ctx, req)
	}
	userID := ctx.UserID()
	groupIDs, err := s.dbRepo.GetDataGroupIDsByUserId(ctx, userID)
	if err != nil {
		return nil, err
	}
	if req.Type == "dataGroup" {
		groups, err := s.dbRepo.GetDataGroupByUserID(ctx, userID)
		if err != nil {
			return nil, err
		}
		result := make([]customdashboard.VariableCandidate, 0, len(groups))
		for _, group := range groups {
			result = append(result, customdashboard.VariableCandidate{Label: group.GroupName, Value: fmt.Sprint(group.GroupID), Type: "dataGroup"})
		}
		if slices.Contains(groupIDs, 0) {
			result = append([]customdashboard.VariableCandidate{{Label: "All Data", Value: "0", Type: "dataGroup"}}, result...)
		}
		return result, nil
	}
	if req.Type == "pod" || req.Type == "instance" {
		filter := ""
		if !slices.Contains(groupIDs, 0) {
			filter, err = s.userFilter(ctx, groupIDs)
			if err != nil {
				return nil, err
			}
			if filter == "" {
				return nil, errNoDataGroupPermission
			}
		}
		if selected := selectedVariableFilter(candidateUpstreamSelections(variables, req.VariableName, req.Type, req.Label), "pod"); selected != "" {
			if filter == "" {
				filter = selected
			} else {
				filter = filter + "," + selected
			}
		}
		return s.podCandidates(ctx, req.Type, filter)
	}
	if slices.Contains(groupIDs, 0) {
		tree, err := s.dbRepo.LoadScopes(ctx)
		if err != nil {
			return nil, err
		}
		return candidatesFromTree(req.Type, tree.DataScopeTreeNode, candidateUpstreamSelections(variables, req.VariableName, req.Type, req.Label)), nil
	}
	scopeIDs, err := s.dbRepo.GetScopeIDsSelectedByPermGroupIDs(ctx, groupIDs)
	if err != nil {
		return nil, err
	}
	scopes, err := s.dbRepo.GetScopesByScopeIDs(ctx, scopeIDs)
	if err != nil {
		return nil, err
	}
	return candidatesFromScopes(req.Type, scopes, candidateUpstreamSelections(variables, req.VariableName, req.Type, req.Label)), nil
}

func labelKeyCandidates(query string) []customdashboard.VariableCandidate {
	keys := extractMetricLabelKeys(query)
	result := make([]customdashboard.VariableCandidate, 0, len(keys))
	for _, key := range keys {
		result = append(result, customdashboard.VariableCandidate{Label: key, Value: key, Type: "label"})
	}
	return result
}

func extractMetricLabelKeys(query string) []string {
	seen := map[string]struct{}{}
	for i := 0; i < len(query); i++ {
		if query[i] != '{' {
			continue
		}
		end := findSelectorEnd(query, i)
		if end <= i {
			continue
		}
		selector := query[i+1 : end]
		for _, part := range strings.Split(selector, ",") {
			part = strings.TrimSpace(part)
			if part == "" {
				continue
			}
			keyEnd := strings.IndexAny(part, "=!~")
			if keyEnd <= 0 {
				continue
			}
			key := strings.TrimSpace(part[:keyEnd])
			if key != "" {
				seen[key] = struct{}{}
			}
		}
		i = end
	}
	keys := make([]string, 0, len(seen))
	for key := range seen {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func (s *service) labelValueCandidates(ctx core.Context, req *request.DashboardVariableRequest) ([]customdashboard.VariableCandidate, error) {
	if s.prom == nil {
		return []customdashboard.VariableCandidate{}, nil
	}
	label := strings.TrimSpace(req.Label)
	if label == "" {
		label = strings.TrimSpace(req.Type)
	}
	if label == "" || label == "label" {
		return []customdashboard.VariableCandidate{}, nil
	}
	variables := variableValuesFromRequest(req)
	filter, _, err := s.dataScopeFilter(ctx, req.GroupID, variables)
	if err != nil {
		return nil, err
	}
	upstream := candidateUpstreamSelections(variables, req.VariableName, req.Type, label)
	isGrafanaLabelValues := isGrafanaLabelValuesQuery(req.Query)
	selector := selectorForLabelValues(substituteVariables(req.Query, upstream), label)
	if selector == "" {
		selector = "up"
	}
	if !isGrafanaLabelValues {
		selected := selectedVariableFilter(excludeSelectionsForSelectorLabels(upstream, selector), label)
		if selected != "" {
			if filter == "" {
				filter = selected
			} else {
				filter = filter + "," + selected
			}
		}
	}
	selector = injectFilter(selector, filter)
	start, end := variableCandidateTimeWindow(req)
	values, err := s.labelValuesWithTimeout(ctx, selector, label, start, end)
	if err != nil {
		return nil, err
	}
	result := make([]customdashboard.VariableCandidate, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		result = append(result, customdashboard.VariableCandidate{Label: string(value), Value: string(value), Type: "label", Meta: map[string]string{"label": label}})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Label < result[j].Label })
	if len(result) > variableCandidateLimit {
		result = result[:variableCandidateLimit]
	}
	return result, nil
}

func (s *service) labelValuesWithTimeout(ctx core.Context, selector string, label string, startTime, endTime int64) (prommodel.LabelValues, error) {
	return s.labelValuesWithTimeoutFor(ctx, selector, label, startTime, endTime, variableCandidateTimeout)
}

func (s *service) labelValuesWithTimeoutFor(ctx core.Context, selector string, label string, startTime, endTime int64, timeout time.Duration) (prommodel.LabelValues, error) {
	timeoutCtx, cancel := withVariableCandidateTimeout(ctx, timeout)
	defer cancel()
	values, err := s.prom.LabelValues(timeoutCtx, selector, label, startTime, endTime)
	if errors.Is(err, go_context.DeadlineExceeded) || errors.Is(timeoutCtx.Err(), go_context.DeadlineExceeded) {
		return nil, errors.New("dashboard variable candidate query timeout")
	}
	return values, err
}

type variableCandidateTimeoutContext struct {
	core.Context
	ctx go_context.Context
}

func withVariableCandidateTimeout(ctx core.Context, timeout time.Duration) (core.Context, go_context.CancelFunc) {
	if timeout <= 0 {
		timeout = variableCandidateTimeout
	}
	baseCtx := go_context.Background()
	if ctx != nil {
		baseCtx = ctx.GetContext()
	}
	boundedCtx, cancel := go_context.WithTimeout(baseCtx, timeout)
	return &variableCandidateTimeoutContext{
		Context: ctx,
		ctx:     boundedCtx,
	}, cancel
}

func (c *variableCandidateTimeoutContext) GetContext() go_context.Context {
	return c.ctx
}

func (c *variableCandidateTimeoutContext) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *variableCandidateTimeoutContext) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

func (c *variableCandidateTimeoutContext) Err() error {
	return c.ctx.Err()
}

func variableCandidateTimeWindow(req *request.DashboardVariableRequest) (int64, int64) {
	if req != nil && req.StartTime > 0 && req.EndTime > req.StartTime {
		return normalizeCandidateTimestamp(req.StartTime), normalizeCandidateTimestamp(req.EndTime)
	}
	end := time.Now()
	return end.Add(-time.Hour).UnixMicro(), end.UnixMicro()
}

func normalizeCandidateTimestamp(value int64) int64 {
	if value > 1_000_000_000_000_000 {
		return value
	}
	return value * int64(time.Millisecond/time.Microsecond)
}

func isGrafanaLabelValuesQuery(query string) bool {
	return grafanaLabelValuesSelector(query) != ""
}

func grafanaLabelValuesSelector(query string) string {
	trimmed := strings.TrimSpace(query)
	const prefix = "label_values("
	if !strings.HasPrefix(trimmed, prefix) || !strings.HasSuffix(trimmed, ")") {
		return ""
	}
	args := splitTopLevelArgs(trimmed[len(prefix) : len(trimmed)-1])
	if len(args) != 2 {
		return ""
	}
	return strings.TrimSpace(args[0])
}

func splitTopLevelArgs(raw string) []string {
	args := []string{}
	start := 0
	braceDepth := 0
	bracketDepth := 0
	parenDepth := 0
	var quote rune
	for index, char := range raw {
		if quote != 0 {
			if char == quote {
				quote = 0
			}
			continue
		}
		switch char {
		case '"', '\'', '`':
			quote = char
		case '{':
			braceDepth++
		case '}':
			if braceDepth > 0 {
				braceDepth--
			}
		case '[':
			bracketDepth++
		case ']':
			if bracketDepth > 0 {
				bracketDepth--
			}
		case '(':
			parenDepth++
		case ')':
			if parenDepth > 0 {
				parenDepth--
			}
		case ',':
			if braceDepth == 0 && bracketDepth == 0 && parenDepth == 0 {
				args = append(args, strings.TrimSpace(raw[start:index]))
				start = index + len(string(char))
			}
		}
	}
	args = append(args, strings.TrimSpace(raw[start:]))
	return args
}

func excludeSelectionsForSelectorLabels(selected map[string][]string, selector string) map[string][]string {
	if len(selected) == 0 || selector == "" {
		return selected
	}
	labels := extractMetricLabelKeys(selector)
	if len(labels) == 0 {
		return selected
	}
	selectorLabels := make(map[string]struct{}, len(labels))
	for _, label := range labels {
		selectorLabels[normalizedVariableLabel(label)] = struct{}{}
	}
	filtered := make(map[string][]string, len(selected))
	for key, values := range selected {
		if _, exists := selectorLabels[normalizedVariableLabel(key)]; exists {
			continue
		}
		filtered[key] = values
	}
	return filtered
}

func selectorForLabelValues(query string, label string) string {
	if selector := grafanaLabelValuesSelector(query); selector != "" {
		return selector
	}
	selectors := metricSelectorsFromPromQL(query)
	if len(selectors) == 0 {
		return strings.TrimSpace(query)
	}
	for _, selector := range selectors {
		if selectorHasLabel(selector, label) {
			return selector
		}
	}
	return selectors[0]
}

func metricSelectorsFromPromQL(query string) []string {
	result := []string{}
	seen := map[string]struct{}{}
	for i := 0; i < len(query); i++ {
		if query[i] != '{' {
			continue
		}
		end := findSelectorEnd(query, i)
		if end <= i {
			continue
		}
		start := i - 1
		for start >= 0 && isMetricNameChar(query[start]) {
			start--
		}
		selector := strings.TrimSpace(query[start+1 : end+1])
		if selector == "" || strings.HasPrefix(selector, "{") {
			i = end
			continue
		}
		if _, ok := seen[selector]; !ok {
			result = append(result, selector)
			seen[selector] = struct{}{}
		}
		i = end
	}
	return result
}

func isMetricNameChar(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		(ch >= '0' && ch <= '9') ||
		ch == '_' ||
		ch == ':'
}

func selectorHasLabel(selector string, label string) bool {
	if label == "" {
		return false
	}
	start := strings.Index(selector, "{")
	end := strings.LastIndex(selector, "}")
	if start < 0 || end <= start {
		return false
	}
	for _, part := range strings.Split(selector[start+1:end], ",") {
		part = strings.TrimSpace(part)
		keyEnd := strings.IndexAny(part, "=!~")
		if keyEnd <= 0 {
			continue
		}
		if strings.TrimSpace(part[:keyEnd]) == label {
			return true
		}
	}
	return false
}

func (s *service) ValidateQuery(ctx core.Context, expr string) response.DashboardValidateQueryResponse {
	if strings.TrimSpace(expr) == "" {
		return response.DashboardValidateQueryResponse{Valid: false, Message: "query expression is empty"}
	}
	upper := strings.ToUpper(expr)
	for _, keyword := range []string{"SELECT ", "INSERT ", "UPDATE ", "DELETE ", "DROP ", " JOIN "} {
		if strings.Contains(upper, keyword) {
			return response.DashboardValidateQueryResponse{Valid: false, Message: "only PromQL/VictoriaMetrics expressions are supported", Position: strings.Index(upper, keyword)}
		}
	}
	if pos := firstUnbalanced(expr); pos >= 0 {
		return response.DashboardValidateQueryResponse{Valid: false, Message: "unbalanced brackets or braces", Position: pos}
	}
	return response.DashboardValidateQueryResponse{Valid: true}
}

func (s *service) PreviewQuery(ctx core.Context, req *request.DashboardQueryRequest) (*response.DashboardQueryResponse, error) {
	items, err := s.resolveQueryItems(req)
	if err != nil {
		return nil, err
	}
	exprText := queryItemsExpressionText(items)
	for index, item := range items {
		validation := s.ValidateQuery(ctx, item.Expression)
		if !validation.Valid {
			return &response.DashboardQueryResponse{
				Status:   "syntax_error",
				Message:  indexedQueryMessage(len(items), index, validation.Message),
				Position: validation.Position,
				Query:    item.Expression,
			}, nil
		}
	}

	resolvedItems := make([]dashboardResolvedQueryItem, 0, len(items))
	for index, item := range items {
		resolved := substituteVariables(item.Expression, req.Variables)
		query, scopeDesc, err := s.enforceDataScope(ctx, resolved, req.GroupID, req.Variables)
		if err != nil {
			if errors.Is(err, errNoDataGroupPermission) {
				message := indexedQueryMessage(len(items), index, err.Error())
				audit := s.audit(ctx, req, exprText, resolved, req.Variables, "", "no_permission", message, 0, 0)
				return &response.DashboardQueryResponse{Status: "no_permission", Message: message, Query: resolved, AuditID: audit}, nil
			}
			return nil, err
		}
		resolvedItems = append(resolvedItems, dashboardResolvedQueryItem{
			ID:       item.ID,
			Legend:   item.Legend,
			Query:    query,
			Scope:    scopeDesc,
			Resolved: resolved,
		})
	}

	points := estimatePoints(req.StartTime, req.EndTime, req.Step) * int64(len(resolvedItems))
	queryText := resolvedQueryText(resolvedItems)
	scopeDesc := resolvedScopeText(resolvedItems)
	if points > maxQueryPoints {
		message := "query result would exceed maximum returned points"
		audit := s.audit(ctx, req, exprText, queryText, req.Variables, scopeDesc, "too_large", message, 0, 0)
		return &response.DashboardQueryResponse{Status: "too_large", Message: message, Query: queryText, AuditID: audit, DataScope: scopeDesc}, nil
	}

	start, end, step := normalizeRange(req.StartTime, req.EndTime, req.Step)
	series := make([]prometheus.MetricResult, 0)
	for index, item := range resolvedItems {
		result, err := s.executeMetricQuery(ctx, req.Instant, item.Query, start, end, step)
		if err != nil {
			status := classifyQueryError(err)
			message := indexedQueryMessage(len(items), index, err.Error())
			audit := s.audit(ctx, req, exprText, item.Query, req.Variables, item.Scope, status, message, 0, 0)
			return &response.DashboardQueryResponse{Status: status, Message: message, Query: item.Query, AuditID: audit, DataScope: item.Scope}, nil
		}
		attachQueryMetadata(result, item)
		series = append(series, result...)
	}

	status := "success"
	message := ""
	if len(series) == 0 {
		status = "no_data"
		message = "no data returned for current time range or variables"
	}
	series, seriesLimit := limitDashboardQuerySeries(req, series)
	pointCount := 0
	for _, item := range series {
		pointCount += len(item.Values)
	}
	auditID := s.audit(ctx, req, exprText, queryText, req.Variables, scopeDesc, status, message, len(series), pointCount)
	return &response.DashboardQueryResponse{
		Status:      status,
		Message:     message,
		Query:       queryText,
		Series:      series,
		AuditID:     auditID,
		Truncated:   seriesLimit != nil && seriesLimit.Limited,
		SeriesLimit: seriesLimit,
		DataScope:   scopeDesc,
	}, nil
}

func (s *service) resolveQuery(req *request.DashboardQueryRequest) (string, error) {
	if req.Mode == customdashboard.QueryModePreset {
		for _, preset := range catalogPresetQueries() {
			if preset.ID == req.PresetID {
				expr := preset.Expression
				for key, val := range req.Params {
					expr = strings.ReplaceAll(expr, "${"+key+":regex}", val)
					expr = strings.ReplaceAll(expr, "${"+key+":raw}", val)
					expr = strings.ReplaceAll(expr, "$"+key, val)
					expr = strings.ReplaceAll(expr, "{{"+key+"}}", val)
				}
				return expr, nil
			}
		}
		return "", errors.New("preset query not found")
	}
	if req.Mode != "" && req.Mode != customdashboard.QueryModeCustom {
		return "", errors.New("unsupported dashboard query mode")
	}
	return req.Expr, nil
}

type dashboardResolvedQueryItem struct {
	ID       string
	Legend   string
	Query    string
	Scope    string
	Resolved string
}

func (s *service) resolveQueryItems(req *request.DashboardQueryRequest) ([]request.DashboardQueryItem, error) {
	if req.Mode == customdashboard.QueryModePreset {
		expr, err := s.resolveQuery(req)
		if err != nil {
			return nil, err
		}
		return []request.DashboardQueryItem{{ID: "q-1", Expression: expr}}, nil
	}
	if req.Mode != "" && req.Mode != customdashboard.QueryModeCustom {
		return nil, errors.New("unsupported dashboard query mode")
	}
	if req.Queries == nil {
		return []request.DashboardQueryItem{{ID: "q-1", Expression: req.Expr}}, nil
	}
	items := make([]request.DashboardQueryItem, 0, len(req.Queries))
	for index, item := range req.Queries {
		if item.Hidden {
			continue
		}
		if item.ID == "" {
			item.ID = fmt.Sprintf("q-%d", index+1)
		}
		items = append(items, item)
	}
	return items, nil
}

func queryItemsExpressionText(items []request.DashboardQueryItem) string {
	expressions := make([]string, 0, len(items))
	for _, item := range items {
		expressions = append(expressions, item.Expression)
	}
	return strings.Join(expressions, "\n")
}

func resolvedQueryText(items []dashboardResolvedQueryItem) string {
	queries := make([]string, 0, len(items))
	for _, item := range items {
		queries = append(queries, item.Query)
	}
	return strings.Join(queries, "\n")
}

func resolvedScopeText(items []dashboardResolvedQueryItem) string {
	seen := map[string]struct{}{}
	scopes := make([]string, 0, len(items))
	for _, item := range items {
		if item.Scope == "" {
			continue
		}
		if _, exists := seen[item.Scope]; exists {
			continue
		}
		seen[item.Scope] = struct{}{}
		scopes = append(scopes, item.Scope)
	}
	return strings.Join(scopes, "\n")
}

func indexedQueryMessage(total, index int, message string) string {
	if total <= 1 {
		return message
	}
	return fmt.Sprintf("query %d: %s", index+1, message)
}

func attachQueryMetadata(series []prometheus.MetricResult, item dashboardResolvedQueryItem) {
	for index := range series {
		series[index].Metric.QueryID = item.ID
		series[index].Metric.QueryLegend = item.Legend
	}
}

func limitDashboardQuerySeries(req *request.DashboardQueryRequest, series []prometheus.MetricResult) ([]prometheus.MetricResult, *response.DashboardSeriesLimitInfo) {
	if req == nil || !chartTypeSupportsSeriesLimit(req.ChartType) {
		return series, nil
	}
	limit := normalizeSeriesLimit(req.SeriesLimit)
	total := len(series)
	if total <= limit {
		return series, nil
	}
	scored := make([]dashboardSeriesScore, 0, total)
	for index, item := range series {
		score, hasScore := metricResultScore(item)
		scored = append(scored, dashboardSeriesScore{
			series:   item,
			index:    index,
			score:    score,
			hasScore: hasScore,
		})
	}
	sort.SliceStable(scored, func(i, j int) bool {
		left, right := scored[i], scored[j]
		if left.hasScore && right.hasScore && left.score != right.score {
			return left.score > right.score
		}
		if left.hasScore != right.hasScore {
			return left.hasScore
		}
		return left.index < right.index
	})
	limited := make([]prometheus.MetricResult, 0, limit)
	for _, item := range scored[:limit] {
		limited = append(limited, item.series)
	}
	return limited, &response.DashboardSeriesLimitInfo{
		Limited: true,
		Total:   total,
		Shown:   len(limited),
		Limit:   limit,
	}
}

type dashboardSeriesScore struct {
	series   prometheus.MetricResult
	index    int
	score    float64
	hasScore bool
}

func chartTypeSupportsSeriesLimit(chartType string) bool {
	switch strings.ToLower(chartType) {
	case "line", "area", "bar":
		return true
	default:
		return false
	}
}

func normalizeSeriesLimit(value int) int {
	if value <= 0 {
		return defaultSeriesLimit
	}
	if value > maxSeriesLimit {
		return maxSeriesLimit
	}
	return value
}

func metricResultScore(result prometheus.MetricResult) (float64, bool) {
	var score float64
	hasScore := false
	for _, point := range result.Values {
		if point.Value != point.Value {
			continue
		}
		if !hasScore || point.Value > score {
			score = point.Value
			hasScore = true
		}
	}
	return score, hasScore
}

func (s *service) executeMetricQuery(ctx core.Context, instant bool, query string, start, end, step int64) ([]prometheus.MetricResult, error) {
	return s.executeMetricQueryWithTimeout(ctx, instant, query, start, end, step, queryTimeout)
}

func (s *service) executeMetricQueryWithTimeout(ctx core.Context, instant bool, query string, start, end, step int64, timeout time.Duration) ([]prometheus.MetricResult, error) {
	if s.prom == nil {
		return nil, errors.New("prometheus repository is not configured")
	}
	select {
	case querySemaphore <- struct{}{}:
		defer func() { <-querySemaphore }()
	default:
		return nil, errors.New("dashboard query concurrency limit exceeded")
	}

	timeoutCtx, cancel := withMetricQueryTimeout(ctx, timeout)
	defer cancel()

	var series []prometheus.MetricResult
	var err error
	if instant {
		series, err = s.prom.QueryData(timeoutCtx, time.Unix(end/1000, 0), query)
	} else {
		series, err = s.prom.QueryRangeData(timeoutCtx, time.Unix(start/1000, 0), time.Unix(end/1000, 0), query, time.Duration(step)*time.Second)
	}
	if errors.Is(err, go_context.DeadlineExceeded) || errors.Is(timeoutCtx.GetContext().Err(), go_context.DeadlineExceeded) {
		return nil, errors.New("dashboard query timeout")
	}
	return series, err
}

type metricQueryTimeoutContext struct {
	core.Context
	ctx go_context.Context
}

func withMetricQueryTimeout(ctx core.Context, timeout time.Duration) (core.Context, go_context.CancelFunc) {
	if timeout <= 0 {
		timeout = queryTimeout
	}
	baseCtx := go_context.Background()
	if ctx != nil {
		baseCtx = ctx.GetContext()
	}
	boundedCtx, cancel := go_context.WithTimeout(baseCtx, timeout)
	return &metricQueryTimeoutContext{
		Context: ctx,
		ctx:     boundedCtx,
	}, cancel
}

func (c *metricQueryTimeoutContext) GetContext() go_context.Context {
	return c.ctx
}

func (c *metricQueryTimeoutContext) Done() <-chan struct{} {
	return c.ctx.Done()
}

func (c *metricQueryTimeoutContext) Deadline() (deadline time.Time, ok bool) {
	return c.ctx.Deadline()
}

func (c *metricQueryTimeoutContext) Err() error {
	return c.ctx.Err()
}

func (s *service) podCandidates(ctx core.Context, typ string, filter string) ([]customdashboard.VariableCandidate, error) {
	if s.prom == nil {
		return []customdashboard.VariableCandidate{}, nil
	}
	end := time.Now()
	start := end.Add(-time.Hour)
	values, err := s.prom.LabelValues(ctx, metricSelector("container_cpu_usage_seconds_total", filter), "pod", start.UnixMicro(), end.UnixMicro())
	if err != nil {
		return nil, err
	}
	result := make([]customdashboard.VariableCandidate, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		result = append(result, customdashboard.VariableCandidate{Label: string(value), Value: string(value), Type: typ})
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Label < result[j].Label })
	return result, nil
}

func (s *service) enforceDataScope(ctx core.Context, expr string, requestGroupID *int64, vars map[string][]string) (string, string, error) {
	filter, scopeDesc, err := s.dataScopeFilter(ctx, requestGroupID, vars)
	if err != nil {
		return "", "", err
	}
	if filter == "" {
		return expr, scopeDesc, nil
	}
	return injectFilter(expr, filter), filter, nil
}

func (s *service) dataScopeFilter(ctx core.Context, requestGroupID *int64, vars map[string][]string) (string, string, error) {
	groupID, explicitGroupID := requestDataGroupID(requestGroupID, vars)
	if explicitGroupID {
		ok, err := s.canViewPageGroup(ctx, groupID)
		if err != nil || !ok {
			if err != nil {
				return "", "", err
			}
			return "", "", errNoDataGroupPermission
		}
		if groupID == 0 {
			return "", "all", nil
		}
		filter, err := s.groupFilter(ctx, groupID)
		if err != nil {
			return "", "", err
		}
		if filter == "" {
			return "", "", errNoDataGroupPermission
		}
		return filter, filter, nil
	}
	return "", "", errNoDataGroupPermission
}

func requestDataGroupID(requestGroupID *int64, _ map[string][]string) (int64, bool) {
	if requestGroupID != nil {
		return *requestGroupID, true
	}
	return 0, false
}

func (s *service) canViewPageGroup(ctx core.Context, groupID int64) (bool, error) {
	groupIDs, err := s.dbRepo.GetDataGroupIDsByUserId(ctx, ctx.UserID())
	if err != nil {
		return false, err
	}
	if len(groupIDs) == 0 {
		return false, nil
	}
	if groupID == 0 {
		return slices.Contains(groupIDs, 0), nil
	}
	if slices.Contains(groupIDs, 0) || slices.Contains(groupIDs, groupID) {
		return true, nil
	}
	if common.DataGroupStorage == nil || common.DataGroupStorage.DataGroupTreeNode == nil {
		return false, nil
	}
	return common.DataGroupStorage.CheckGroupPermission(groupID, groupIDs), nil
}

func (s *service) groupFilter(ctx core.Context, groupID int64) (string, error) {
	filter, err := common.GetPQLFilterByGroupID(ctx, s.dbRepo, datagroup.DATASOURCE_CATEGORY_APM, groupID)
	if err != nil {
		return "", err
	}
	if filter == nil {
		return "", nil
	}
	return filter.String(), nil
}

func (s *service) userFilter(ctx core.Context, groupIDs []int64) (string, error) {
	if len(groupIDs) == 0 {
		return "", nil
	}
	scopeIDs, err := s.dbRepo.GetScopeIDsSelectedByPermGroupIDs(ctx, groupIDs)
	if err != nil {
		return "", err
	}
	if len(scopeIDs) == 0 {
		return "", nil
	}
	if common.DataGroupStorage != nil && common.DataGroupStorage.DataScopeTree != nil {
		scopeTree, _ := common.DataGroupStorage.CloneWithCategory(scopeIDs, datagroup.DATASOURCE_CATEGORY_APM)
		filter := common.ConvertScopeNodeToPQLFilter(scopeTree)
		if filter == nil {
			return "", nil
		}
		return filter.String(), nil
	}
	scopes, err := s.dbRepo.GetScopesByScopeIDs(ctx, scopeIDs)
	if err != nil {
		return "", err
	}
	return scopeFilterString(scopes), nil
}

func (s *service) userFeatureIDs(ctx core.Context, userID int64) ([]int, error) {
	roleRefs, err := s.dbRepo.GetUserRole(ctx, userID)
	if err != nil {
		return nil, err
	}
	roleIDs := make([]int64, len(roleRefs))
	for i := range roleRefs {
		roleIDs[i] = int64(roleRefs[i].RoleID)
	}
	rolePerms, err := s.dbRepo.GetSubjectsPermission(ctx, roleIDs, model.PERMISSION_SUB_TYP_ROLE, model.PERMISSION_TYP_FEATURE)
	if err != nil {
		return nil, err
	}
	userPerms, err := s.dbRepo.GetSubjectPermission(ctx, userID, model.PERMISSION_SUB_TYP_USER, model.PERMISSION_TYP_FEATURE)
	if err != nil {
		return nil, err
	}
	teamIDs, err := s.dbRepo.GetUserTeams(ctx, userID)
	if err != nil {
		return nil, err
	}
	teamPerms, err := s.dbRepo.GetSubjectsPermission(ctx, teamIDs, model.PERMISSION_SUB_TYP_TEAM, model.PERMISSION_TYP_FEATURE)
	if err != nil {
		return nil, err
	}
	featureIDs := make([]int, 0, len(rolePerms)+len(userPerms)+len(teamPerms))
	for _, perm := range rolePerms {
		featureIDs = append(featureIDs, perm.PermissionID)
	}
	featureIDs = append(featureIDs, userPerms...)
	for _, perm := range teamPerms {
		featureIDs = append(featureIDs, perm.PermissionID)
	}
	return featureIDs, nil
}

func (s *service) canView(ctx core.Context, row *database.CustomDashboard) (bool, error) {
	if ok, err := s.canAccessDashboardRoute(ctx, row.ID, ctx.UserID()); err != nil || !ok {
		return ok, err
	}
	if row.OwnerID == ctx.UserID() || row.Visibility == customdashboard.VisibilityEveryone {
		return true, nil
	}
	if row.Visibility == customdashboard.VisibilityPersonal {
		return row.OwnerID == ctx.UserID(), nil
	}
	if row.Visibility != customdashboard.VisibilityTeam {
		return false, nil
	}
	teamIDs, err := decodeTeamIDs(row.TeamIDs)
	if err != nil {
		return false, err
	}
	userTeams, err := s.dbRepo.GetUserTeams(ctx, ctx.UserID())
	if err != nil {
		return false, err
	}
	for _, teamID := range teamIDs {
		if slices.Contains(userTeams, teamID) {
			return true, nil
		}
	}
	return false, nil
}

func (s *service) canAccessDashboardRoute(ctx core.Context, id string, userID int64) (bool, error) {
	featureIDs, err := s.userFeatureIDs(ctx, userID)
	if err != nil || len(featureIDs) == 0 {
		return false, err
	}
	route := customDashboardRoute(id)

	routerMappings, err := s.dbRepo.GetFeatureMappingByFeature(ctx, featureIDs, model.MAPPED_TYP_ROUTER)
	if err != nil {
		return false, err
	}
	routerIDs := make([]int, 0, len(routerMappings))
	for _, mapping := range routerMappings {
		routerIDs = append(routerIDs, mapping.MappedID)
	}
	routers, err := s.dbRepo.GetRouterByIDs(ctx, routerIDs)
	if err != nil || (!hasExactDashboardRoute(routers, route) && !hasDashboardRootRouter(routers)) {
		return false, err
	}

	menuMappings, err := s.dbRepo.GetFeatureMappingByFeature(ctx, featureIDs, model.MAPPED_TYP_MENU)
	if err != nil {
		return false, err
	}
	menuIDs := make([]int, 0, len(menuMappings))
	for _, mapping := range menuMappings {
		menuIDs = append(menuIDs, mapping.MappedID)
	}
	menuRouters, err := s.dbRepo.GetItemsRouter(ctx, menuIDs)
	if err != nil {
		return false, err
	}
	if hasExactDashboardRoute(routers, route) && hasExactDashboardRoute(menuRouters, route) {
		return true, nil
	}
	return hasDashboardRootRouter(routers) && hasDashboardRootMenu(menuRouters), nil
}

func hasExactDashboardRoute(routers []database.Router, route string) bool {
	for _, router := range routers {
		if router.RouterTo == route {
			return true
		}
	}
	return false
}

func hasDashboardRootRouter(routers []database.Router) bool {
	for _, router := range routers {
		if router.RouterTo == "/custom-dashboard/:id" {
			return true
		}
	}
	return false
}

func hasDashboardRootMenu(routers []database.Router) bool {
	for _, router := range routers {
		if router.RouterTo == "/custom-dashboard" {
			return true
		}
	}
	return false
}

func customDashboardRoute(id string) string {
	return "/custom-dashboard/" + id
}

func copyDashboardName(name, lang string) string {
	if lang == code.LANG_EN {
		return name + " - Copy"
	}
	return name + " - 副本"
}

func (s *service) rowToDashboard(ctx core.Context, row *database.CustomDashboard, userID int64, canCreate bool) (*customdashboard.Dashboard, error) {
	dashboard, err := unmarshalDashboardConfig(row.Config)
	if err != nil {
		return nil, err
	}
	teamIDs, _ := decodeTeamIDs(row.TeamIDs)
	dashboard.ID = row.ID
	dashboard.Name = row.Name
	dashboard.Description = row.Description
	dashboard.OwnerID = row.OwnerID
	dashboard.Visibility = row.Visibility
	dashboard.TeamIDs = teamIDs
	dashboard.CanEdit = row.OwnerID == userID
	dashboard.CanDelete = row.OwnerID == userID
	dashboard.CanCopy = canCreate
	dashboard.CreatedAt = row.CreatedAt.UnixMilli()
	dashboard.UpdatedAt = row.UpdatedAt.UnixMilli()
	if row.LastVisitedAt != nil {
		dashboard.LastVisitedAt = row.LastVisitedAt.UnixMilli()
	}
	if user, err := s.dbRepo.GetUserInfo(ctx, row.OwnerID); err == nil {
		dashboard.Creator = user.Username
	}
	return &dashboard, nil
}

func (s *service) dashboardToRow(dashboard *customdashboard.Dashboard) (*database.CustomDashboard, error) {
	config, err := marshalDashboardConfig(dashboard)
	if err != nil {
		return nil, err
	}
	teamJSON, _ := json.Marshal(dashboard.TeamIDs)
	return &database.CustomDashboard{
		ID:          dashboard.ID,
		Name:        dashboard.Name,
		Description: dashboard.Description,
		OwnerID:     dashboard.OwnerID,
		Visibility:  normalizeVisibility(dashboard.Visibility),
		TeamIDs:     string(teamJSON),
		Config:      config,
	}, nil
}

func (s *service) audit(ctx core.Context, req *request.DashboardQueryRequest, expr, query string, vars map[string][]string, scope, status, message string, series, points int) uint {
	varsJSON, _ := json.Marshal(vars)
	audit := database.CustomDashboardQueryAudit{
		UserID: ctx.UserID(), DashboardID: req.DashboardID, WidgetID: req.WidgetID, QueryMode: req.Mode,
		Expression: expr, ResolvedQuery: query, ResolvedVars: string(varsJSON),
		ConstrainedScope: scope, Status: status, ErrorMessage: message,
		ResultSeries: series, ResultPoints: points,
	}
	_ = s.dbRepo.CreateCustomDashboardQueryAudit(ctx, &audit)
	return audit.ID
}

func findTemplate(id string) (*customdashboard.Template, error) {
	for _, template := range catalogTemplates() {
		if template.ID == id {
			return &template, nil
		}
	}
	return nil, errors.New("template not found")
}

func marshalDashboardConfig(dashboard *customdashboard.Dashboard) (string, error) {
	copy := *dashboard
	normalizeDashboardDefinition(&copy, nil)
	copy.ID = ""
	copy.OwnerID = 0
	copy.Creator = ""
	copy.CanEdit = false
	copy.CanDelete = false
	copy.CanCopy = false
	copy.CreatedAt = 0
	copy.UpdatedAt = 0
	copy.LastVisitedAt = 0
	bytes, err := json.Marshal(copy)
	return string(bytes), err
}

func unmarshalDashboardConfig(raw string) (customdashboard.Dashboard, error) {
	if raw == "" {
		return blankDashboard(), nil
	}
	var dashboard customdashboard.Dashboard
	if err := json.Unmarshal([]byte(raw), &dashboard); err != nil {
		return customdashboard.Dashboard{}, err
	}
	normalizeDashboardDefinition(&dashboard, dashboardVariablesWithExplicitMulti(raw))
	return dashboard, nil
}

func normalizeDashboardDefinition(dashboard *customdashboard.Dashboard, explicitMulti map[string]bool) {
	dashboard.TimeRange = customdashboard.TimeRange{}
	dashboard.RefreshSeconds = 0
	dashboard.Variables = normalizeVariables(dashboard.Variables, explicitMulti)
	occupied := map[[2]int]struct{}{}
	for i := range dashboard.Widgets {
		widget := &dashboard.Widgets[i]
		if widget.Type == customdashboard.WidgetTopN {
			widget.Type = customdashboard.WidgetTable
		}
		widget.Layout = normalizeWidgetLayout(widget.Layout, occupied)
		widget.Display = customdashboard.WidgetDisplay{
			Unit:               widget.Display.Unit,
			UnitPreset:         widget.Display.UnitPreset,
			AutoScale:          widget.Display.AutoScale,
			Decimals:           widget.Display.Decimals,
			EmptyText:          widget.Display.EmptyText,
			ErrorText:          widget.Display.ErrorText,
			TableColumns:       widget.Display.TableColumns,
			TableHiddenColumns: widget.Display.TableHiddenColumns,
			ShowValueColumn:    widget.Display.ShowValueColumn,
			WarningThreshold:   widget.Display.WarningThreshold,
			CriticalThreshold:  widget.Display.CriticalThreshold,
		}
		if widget.Display.Decimals < 0 {
			widget.Display.Decimals = 0
		}
		if widget.Query.Params == nil {
			widget.Query.Params = map[string]string{}
		}
		normalizeWidgetQuery(&widget.Query)
	}
}

func normalizeWidgetQuery(query *customdashboard.WidgetQuery) {
	if query.Params == nil {
		query.Params = map[string]string{}
	}
	if query.Mode == "" && (query.Expression != "" || len(query.Queries) > 0) {
		query.Mode = customdashboard.QueryModeCustom
	}
	if query.Mode != customdashboard.QueryModeCustom {
		return
	}
	items := query.Queries
	if len(items) == 0 {
		items = []customdashboard.WidgetQueryItem{{
			ID:         "q-1",
			Expression: query.Expression,
			Legend:     query.Legend,
		}}
	}
	for index := range items {
		if items[index].ID == "" {
			items[index].ID = fmt.Sprintf("q-%d", index+1)
		}
	}
	query.Queries = items
	query.Expression = items[0].Expression
	query.Legend = items[0].Legend
}

func dashboardVariablesWithExplicitMulti(raw string) map[string]bool {
	var config struct {
		Variables []map[string]json.RawMessage `json:"variables"`
	}
	if err := json.Unmarshal([]byte(raw), &config); err != nil {
		return nil
	}
	result := map[string]bool{}
	for _, variable := range config.Variables {
		nameRaw, ok := variable["name"]
		if !ok {
			continue
		}
		var name string
		if err := json.Unmarshal(nameRaw, &name); err != nil || name == "" {
			continue
		}
		if _, ok := variable["multi"]; ok {
			result[name] = true
		}
	}
	return result
}

func normalizeVariables(variables []customdashboard.Variable, explicitMulti map[string]bool) []customdashboard.Variable {
	result := make([]customdashboard.Variable, 0, len(variables))
	for _, variable := range variables {
		if variable.Type == "timeRange" || variable.Name == "timeRange" || variable.Type == "dataGroup" || variable.Name == "dataGroup" {
			continue
		}
		if explicitMulti != nil && !explicitMulti[variable.Name] {
			variable.Multi = true
		}
		if resourceType := builtinResourceVariableType(variable); resourceType != "" {
			variable.Type = resourceType
			variable.IncludeAll = true
			if variable.AllValue == "" {
				variable.AllValue = ".*"
			}
			variable.Value = variable.AllValue
			variable.DefaultValue = variable.AllValue
		}
		result = append(result, variable)
	}
	return result
}

var builtinResourceVariableAliases = map[string]string{
	"cluster":        "cluster",
	"cluster_id":     "cluster",
	"clusterId":      "cluster",
	"clusterID":      "cluster",
	"cluster_name":   "cluster",
	"namespace":      "namespace",
	"namespace_name": "namespace",
	"namespaceName":  "namespace",
	"src_namespace":  "namespace",
	"dst_namespace":  "namespace",
	"svc":            "service",
	"svc_name":       "service",
	"service":        "service",
	"service_name":   "service",
	"endpoint":       "endpoint",
	"endpoint_name":  "endpoint",
	"content_key":    "endpoint",
	"contentKey":     "endpoint",
	"pod":            "pod",
	"pod_name":       "pod",
	"src_pod":        "pod",
	"dst_pod":        "pod",
	"instance":       "instance",
	"instance_name":  "instance",
	"node_name":      "instance",
	"nodeName":       "instance",
	"host_name":      "instance",
	"hostName":       "instance",
}

var customLabelVariableNames = map[string]struct{}{
	"related_nodes":     {},
	"related_pid":       {},
	"jvm_instance":      {},
	"job_name":          {},
	"cadvisor_job_name": {},
}

func builtinResourceVariableType(variable customdashboard.Variable) string {
	if _, ok := customLabelVariableNames[variable.Name]; ok {
		return ""
	}
	values := append([]string{variable.Type, variable.Name}, variable.Labels...)
	for _, value := range values {
		if resourceType, ok := builtinResourceVariableAliases[value]; ok {
			return resourceType
		}
	}
	return ""
}

func normalizeWidgetLayout(layout customdashboard.WidgetLayout, occupied map[[2]int]struct{}) customdashboard.WidgetLayout {
	x := layout.X
	y := layout.Y
	w := layout.W
	h := layout.H
	if x >= fixedGridColumns || w > fixedGridColumns {
		w = scaleLegacyDashboardWidth(layout.W)
		x = scaleLegacyDashboardX(layout.X, w)
		rowHeight := layout.H
		if rowHeight <= 0 {
			rowHeight = 4
		}
		if rowHeight > 1 && y > 0 {
			y = y / rowHeight
		}
		h = 1
	} else if w <= 0 {
		x = legacyTwoColumnX(x)
		w = defaultWidgetGridWidth
	}
	if x < 0 {
		x = 0
	}
	if x >= fixedGridColumns {
		x = fixedGridColumns - 1
	}
	if y < 0 {
		y = 0
	}
	if w <= 0 {
		w = 1
	}
	maxWidth := fixedGridColumns - x
	if w > maxWidth {
		w = maxWidth
	}
	if h <= 0 {
		h = 1
	}
	desiredWidth := w
	for {
		if desiredWidth <= fixedGridColumns-x && widgetLayoutOpen(x, y, desiredWidth, h, occupied) {
			occupyWidgetLayout(x, y, desiredWidth, h, occupied)
			return customdashboard.WidgetLayout{X: x, Y: y, W: desiredWidth, H: h}
		}
		x++
		if x >= fixedGridColumns {
			x = 0
			y++
		}
	}
}

func legacyTwoColumnX(x int) int {
	if x < 0 {
		x = 0
	}
	if x >= legacyTwoColumnCount {
		x = legacyTwoColumnCount - 1
	}
	return x * defaultWidgetGridWidth
}

func scaleLegacyDashboardWidth(w int) int {
	if w <= 0 {
		return defaultWidgetGridWidth
	}
	width := (w*fixedGridColumns + legacyDashboardColumns/2) / legacyDashboardColumns
	if width < 1 {
		return 1
	}
	if width > fixedGridColumns {
		return fixedGridColumns
	}
	return width
}

func scaleLegacyDashboardX(x int, w int) int {
	if x < 0 {
		x = 0
	}
	scaled := x * fixedGridColumns / legacyDashboardColumns
	maxX := fixedGridColumns - w
	if scaled > maxX {
		return maxX
	}
	return scaled
}

func widgetLayoutOpen(x, y, w, h int, occupied map[[2]int]struct{}) bool {
	for row := y; row < y+h; row++ {
		for column := x; column < x+w; column++ {
			if _, exists := occupied[[2]int{column, row}]; exists {
				return false
			}
		}
	}
	return true
}

func occupyWidgetLayout(x, y, w, h int, occupied map[[2]int]struct{}) {
	for row := y; row < y+h; row++ {
		for column := x; column < x+w; column++ {
			occupied[[2]int{column, row}] = struct{}{}
		}
	}
}

func decodeTeamIDs(raw string) ([]int64, error) {
	if raw == "" {
		return nil, nil
	}
	var ids []int64
	err := json.Unmarshal([]byte(raw), &ids)
	return ids, err
}

func normalizeVisibility(visibility string) string {
	switch visibility {
	case customdashboard.VisibilityTeam, customdashboard.VisibilityPersonal:
		return visibility
	default:
		return customdashboard.VisibilityEveryone
	}
}

func substituteVariables(expr string, vars map[string][]string) string {
	if rateInterval := grafanaRateInterval(vars); rateInterval != "" {
		expr = strings.ReplaceAll(expr, "$__rate_interval", rateInterval)
		expr = strings.ReplaceAll(expr, "${__rate_interval}", rateInterval)
		expr = strings.ReplaceAll(expr, "$__interval", rateInterval)
		expr = strings.ReplaceAll(expr, "${__interval}", rateInterval)
	}
	keys := make([]string, 0, len(vars))
	for key := range vars {
		keys = append(keys, key)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })
	for _, key := range keys {
		values := vars[key]
		raw := variableRawValue(values)
		regex := variableRegexValue(values)
		if isAllVariableValues(values) {
			expr = replaceExactMatcherVariable(expr, key, regex)
		}
		expr = strings.ReplaceAll(expr, "${"+key+":regex}", regex)
		expr = replaceRegexMatcherVariable(expr, key, regex)
		expr = strings.ReplaceAll(expr, "${"+key+":raw}", raw)
		expr = strings.ReplaceAll(expr, "${"+key+"}", raw)
		expr = regexp.MustCompile(`\{\{\s*`+regexp.QuoteMeta(key)+`\s*\}\}`).ReplaceAllString(expr, raw)
		expr = regexp.MustCompile(`\$`+regexp.QuoteMeta(key)+`([^A-Za-z0-9_]|$)`).ReplaceAllString(expr, raw+"$1")
	}
	return expr
}

func replaceRegexMatcherVariable(expr string, key string, value string) string {
	if key == "" {
		return expr
	}
	quotedKey := regexp.QuoteMeta(key)
	tokenPatterns := []string{
		`\$\{` + quotedKey + `(?::[A-Za-z_][A-Za-z0-9_]*)?\}`,
		`\$` + quotedKey,
		`\{\{\s*` + quotedKey + `\s*\}\}`,
	}
	for _, operator := range []string{"=~", "!~"} {
		for _, quote := range []string{"\"", "'", "`"} {
			quotePattern := regexp.QuoteMeta(quote)
			for _, tokenPattern := range tokenPatterns {
				pattern := regexp.MustCompile(`(` + operator + `\s*` + quotePattern + `)` + tokenPattern + `(` + quotePattern + `)`)
				expr = replaceRegexMatcherPattern(expr, pattern, value)
			}
		}
	}
	return expr
}

func replaceExactMatcherVariable(expr string, key string, value string) string {
	if key == "" {
		return expr
	}
	quotedKey := regexp.QuoteMeta(key)
	tokenPatterns := []string{
		`\$\{` + quotedKey + `(?::[A-Za-z_][A-Za-z0-9_]*)?\}`,
		`\$` + quotedKey,
		`\{\{\s*` + quotedKey + `\s*\}\}`,
	}
	for _, quote := range []string{"\"", "'", "`"} {
		quotePattern := regexp.QuoteMeta(quote)
		for _, tokenPattern := range tokenPatterns {
			pattern := regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_:]*\s*)(=|!=)(\s*` + quotePattern + `)` + tokenPattern + `(` + quotePattern + `)`)
			expr = replaceExactMatcherPattern(expr, pattern, value)
		}
	}
	return expr
}

func replaceExactMatcherPattern(expr string, pattern *regexp.Regexp, value string) string {
	matches := pattern.FindAllStringSubmatchIndex(expr, -1)
	if len(matches) == 0 {
		return expr
	}
	var out strings.Builder
	out.Grow(len(expr) + len(value)*len(matches))
	last := 0
	for _, match := range matches {
		out.WriteString(expr[last:match[0]])
		out.WriteString(expr[match[2]:match[3]])
		operator := expr[match[4]:match[5]]
		if operator == "!=" {
			out.WriteString("!~")
		} else {
			out.WriteString("=~")
		}
		out.WriteString(expr[match[6]:match[7]])
		out.WriteString(value)
		out.WriteString(expr[match[8]:match[9]])
		last = match[1]
	}
	out.WriteString(expr[last:])
	return out.String()
}

func replaceRegexMatcherPattern(expr string, pattern *regexp.Regexp, value string) string {
	matches := pattern.FindAllStringSubmatchIndex(expr, -1)
	if len(matches) == 0 {
		return expr
	}
	var out strings.Builder
	out.Grow(len(expr) + len(value)*len(matches))
	last := 0
	for _, match := range matches {
		out.WriteString(expr[last:match[0]])
		out.WriteString(expr[match[2]:match[3]])
		out.WriteString(value)
		out.WriteString(expr[match[4]:match[5]])
		last = match[1]
	}
	out.WriteString(expr[last:])
	return out.String()
}

func grafanaRateInterval(vars map[string][]string) string {
	if values := vars["resolution"]; len(values) > 0 && strings.TrimSpace(values[0]) != "" {
		return strings.TrimSpace(values[0])
	}
	return "1m"
}

func variableRawValue(values []string) string {
	return strings.Join(values, ",")
}

func variableRegexValue(values []string) string {
	if isAllVariableValues(values) {
		return ".*"
	}
	return promRegexAlternation(values)
}

func isAllVariableValues(values []string) bool {
	for _, value := range values {
		switch strings.TrimSpace(value) {
		case ".*", "$__all", "__all":
			return true
		}
	}
	return false
}

func injectFilter(expr string, filter string) string {
	filter = strings.TrimSpace(filter)
	if filter == "" || filter == "all" {
		return expr
	}
	var out strings.Builder
	out.Grow(len(expr) + len(filter)*2)
	for i := 0; i < len(expr); {
		ch := rune(expr[i])
		if ch == '"' || ch == '\'' || ch == '`' {
			end := copyQuoted(&out, expr, i, ch)
			i = end
			continue
		}
		if ch == '[' {
			end := copyBracket(&out, expr, i)
			i = end
			continue
		}
		if !isMetricIdentStart(ch) {
			out.WriteByte(expr[i])
			i++
			continue
		}
		if i > 0 && unicode.IsDigit(rune(expr[i-1])) {
			out.WriteByte(expr[i])
			i++
			continue
		}
		start := i
		for i < len(expr) && isMetricIdentPart(rune(expr[i])) {
			i++
		}
		ident := expr[start:i]
		next := skipSpace(expr, i)
		if isGroupingKeyword(ident) && next < len(expr) && expr[next] == '(' {
			groupEnd := findParenEnd(expr, next)
			out.WriteString(ident)
			out.WriteString(expr[i : groupEnd+1])
			i = groupEnd + 1
			continue
		}
		if isPromQLKeyword(ident) || (next < len(expr) && expr[next] == '(') {
			out.WriteString(ident)
			continue
		}
		if next < len(expr) && expr[next] == '{' {
			selectorEnd := findSelectorEnd(expr, next)
			out.WriteString(expr[start : next+1])
			out.WriteString(filter)
			if next+1 < len(expr) && expr[next+1] != '}' {
				out.WriteByte(',')
			}
			out.WriteString(expr[next+1 : selectorEnd+1])
			i = selectorEnd + 1
			continue
		}
		if next >= len(expr) || isMetricBoundary(rune(expr[next])) {
			out.WriteString(ident)
			out.WriteByte('{')
			out.WriteString(filter)
			out.WriteByte('}')
			continue
		}
		out.WriteString(ident)
	}
	return out.String()
}

func copyBracket(out *strings.Builder, expr string, start int) int {
	for i := start; i < len(expr); i++ {
		out.WriteByte(expr[i])
		if expr[i] == ']' {
			return i + 1
		}
	}
	return len(expr)
}

func findSelectorEnd(expr string, open int) int {
	for i := open + 1; i < len(expr); i++ {
		switch expr[i] {
		case '"', '\'', '`':
			i = skipQuoted(expr, i, rune(expr[i])) - 1
		case '}':
			return i
		}
	}
	return len(expr) - 1
}

func findParenEnd(expr string, open int) int {
	depth := 0
	for i := open; i < len(expr); i++ {
		switch expr[i] {
		case '"', '\'', '`':
			i = skipQuoted(expr, i, rune(expr[i])) - 1
		case '(':
			depth++
		case ')':
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return len(expr) - 1
}

func skipQuoted(expr string, start int, quote rune) int {
	for i := start + 1; i < len(expr); i++ {
		if quote != '`' && expr[i] == '\\' && i+1 < len(expr) {
			i++
			continue
		}
		if rune(expr[i]) == quote {
			return i + 1
		}
	}
	return len(expr)
}

func copyQuoted(out *strings.Builder, expr string, start int, quote rune) int {
	for i := start; i < len(expr); i++ {
		out.WriteByte(expr[i])
		if quote != '`' && rune(expr[i]) == '\\' && i+1 < len(expr) {
			i++
			out.WriteByte(expr[i])
			continue
		}
		if i > start && rune(expr[i]) == quote {
			return i + 1
		}
	}
	return len(expr)
}

func skipSpace(expr string, index int) int {
	for index < len(expr) && unicode.IsSpace(rune(expr[index])) {
		index++
	}
	return index
}

func isMetricIdentStart(ch rune) bool {
	return unicode.IsLetter(ch) || ch == '_' || ch == ':'
}

func isMetricIdentPart(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == ':'
}

func isMetricBoundary(ch rune) bool {
	return unicode.IsSpace(ch) || strings.ContainsRune("+-*/%^,),[]", ch)
}

func isPromQLKeyword(ident string) bool {
	switch ident {
	case "and", "or", "unless", "by", "without", "on", "ignoring", "group_left", "group_right", "bool", "offset":
		return true
	default:
		return false
	}
}

func isGroupingKeyword(ident string) bool {
	switch ident {
	case "by", "without", "on", "ignoring", "group_left", "group_right":
		return true
	default:
		return false
	}
}

func classifyQueryError(err error) string {
	msg := strings.ToLower(err.Error())
	switch {
	case strings.Contains(msg, "timeout"):
		return "timeout"
	case strings.Contains(msg, "concurrency") || strings.Contains(msg, "limit"):
		return "too_large"
	case strings.Contains(msg, "permission"):
		return "no_permission"
	case strings.Contains(msg, "not found") || strings.Contains(msg, "unknown metric"):
		return "metric_missing"
	default:
		return "error"
	}
}

func metricSelector(metric string, filter string) string {
	filter = strings.TrimSpace(filter)
	if filter == "" || filter == "all" {
		return metric
	}
	return metric + "{" + filter + "}"
}

func scopeFilterString(scopes []datagroup.DataScope) string {
	clusters := map[string]struct{}{}
	namespaces := map[string]struct{}{}
	services := map[string]struct{}{}
	for _, scope := range scopes {
		if scope.ClusterID != "" {
			clusters[scope.ClusterID] = struct{}{}
		}
		if scope.Namespace != "" {
			namespaces[scope.Namespace] = struct{}{}
		}
		if scope.Service != "" {
			services[scope.Service] = struct{}{}
		}
	}
	parts := []string{}
	if len(clusters) > 0 {
		parts = append(parts, `cluster_id=~"`+joinKeys(clusters)+`"`)
	}
	if len(namespaces) > 0 {
		parts = append(parts, `namespace=~"`+joinKeys(namespaces)+`"`)
	}
	if len(services) > 0 {
		parts = append(parts, `svc_name=~"`+joinKeys(services)+`"`)
	}
	return strings.Join(parts, ",")
}

func joinKeys(values map[string]struct{}) string {
	keys := make([]string, 0, len(values))
	for key := range values {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return strings.Join(keys, "|")
}

func candidatesFromScopes(typ string, scopes []datagroup.DataScope, selected map[string][]string) []customdashboard.VariableCandidate {
	seen := map[string]customdashboard.VariableCandidate{}
	for _, scope := range scopes {
		if !scopeMatchesSelected(scope, selected) {
			continue
		}
		switch typ {
		case "cluster":
			if scope.ClusterID != "" {
				seen[scope.ClusterID] = customdashboard.VariableCandidate{Label: scope.ClusterID, Value: scope.ClusterID, Type: typ}
			}
		case "namespace":
			if scope.Namespace != "" {
				seen[scope.Namespace] = customdashboard.VariableCandidate{Label: scope.Namespace, Value: scope.Namespace, Type: typ, Meta: map[string]string{"clusterId": scope.ClusterID}}
			}
		case "service":
			if scope.Service != "" {
				seen[scope.Service] = customdashboard.VariableCandidate{Label: scope.Service, Value: scope.Service, Type: typ, Meta: map[string]string{"clusterId": scope.ClusterID, "namespace": scope.Namespace}}
			}
		}
	}
	result := make([]customdashboard.VariableCandidate, 0, len(seen))
	for _, item := range seen {
		result = append(result, item)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Label < result[j].Label })
	return result
}

func candidatesFromTree(typ string, root *datagroup.DataScopeTreeNode, selected map[string][]string) []customdashboard.VariableCandidate {
	scopes := []datagroup.DataScope{}
	var walk func(node *datagroup.DataScopeTreeNode)
	walk = func(node *datagroup.DataScopeTreeNode) {
		if node == nil {
			return
		}
		scopes = append(scopes, node.DataScope)
		for _, child := range node.Children {
			walk(child)
		}
	}
	walk(root)
	return candidatesFromScopes(typ, scopes, selected)
}

func scopeMatchesSelected(scope datagroup.DataScope, selected map[string][]string) bool {
	if len(selected) == 0 {
		return true
	}
	for key, values := range selected {
		if len(values) == 0 {
			continue
		}
		if isAllVariableValues(values) {
			continue
		}
		value := ""
		switch normalizedVariableLabel(key) {
		case "cluster_id":
			value = scope.ClusterID
		case "namespace":
			value = scope.Namespace
		case "svc_name":
			value = scope.Service
		default:
			continue
		}
		if value != "" && !slices.Contains(values, value) {
			return false
		}
	}
	return true
}

func selectedVariableFilter(selected map[string][]string, excludeLabel string) string {
	if len(selected) == 0 {
		return ""
	}
	parts := []string{}
	exclude := normalizedVariableLabel(excludeLabel)
	for key, values := range selected {
		if len(values) == 0 {
			continue
		}
		if isAllVariableValues(values) {
			continue
		}
		label := normalizedVariableLabel(key)
		if label == exclude {
			continue
		}
		parts = append(parts, label+`=~"`+promRegexAlternation(values)+`"`)
	}
	sort.Strings(parts)
	return strings.Join(parts, ",")
}

func candidateUpstreamSelections(selected map[string][]string, variableName string, typ string, label string) map[string][]string {
	if len(selected) == 0 {
		return nil
	}
	exclude := normalizedVariableLabel(label)
	if exclude == "" || exclude == "label" {
		exclude = normalizedVariableLabel(typ)
	}
	upstream := make(map[string][]string, len(selected))
	for key, value := range selected {
		if variableName != "" && key == variableName {
			continue
		}
		if normalizedVariableLabel(key) == exclude {
			continue
		}
		upstream[key] = value
	}
	return upstream
}

func normalizedVariableLabel(key string) string {
	if strings.HasSuffix(key, "job_name") {
		return "job_name"
	}
	switch builtinResourceVariableAliases[key] {
	case "cluster":
		return "cluster_id"
	case "namespace":
		return "namespace"
	case "service":
		return "svc_name"
	case "endpoint":
		return "endpoint"
	case "pod":
		return "pod"
	case "instance":
		return "instance"
	default:
		return key
	}
}

func promRegexAlternation(values []string) string {
	escaped := make([]string, 0, len(values))
	for _, value := range values {
		escaped = append(escaped, escapePromQLString(regexp.QuoteMeta(value)))
	}
	return strings.Join(escaped, "|")
}

func escapePromQLString(value string) string {
	return strings.NewReplacer(
		`\`, `\\`,
		`"`, `\"`,
		"\n", `\n`,
	).Replace(value)
}

func variableValuesFromRequest(req *request.DashboardVariableRequest) map[string][]string {
	if strings.TrimSpace(req.VariablesJSON) != "" {
		var arrayValues map[string][]string
		if err := json.Unmarshal([]byte(req.VariablesJSON), &arrayValues); err == nil {
			return cleanVariableValues(arrayValues)
		}
		var stringValues map[string]string
		if err := json.Unmarshal([]byte(req.VariablesJSON), &stringValues); err == nil {
			return variableValuesFromStrings(stringValues)
		}
	}
	return cleanVariableValues(req.Variables)
}

func variableValuesFromStrings(raw map[string]string) map[string][]string {
	if len(raw) == 0 {
		return nil
	}
	values := make(map[string][]string, len(raw))
	for key, value := range raw {
		values[key] = splitVariableSelection(value)
	}
	return cleanVariableValues(values)
}

func cleanVariableValues(raw map[string][]string) map[string][]string {
	if len(raw) == 0 {
		return nil
	}
	values := make(map[string][]string, len(raw))
	for key, items := range raw {
		for _, item := range items {
			item = strings.TrimSpace(item)
			if item != "" {
				values[key] = append(values[key], item)
			}
		}
	}
	return values
}

func splitVariableSelection(raw string) []string {
	parts := strings.FieldsFunc(raw, func(r rune) bool {
		return r == '|' || r == ','
	})
	values := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			values = append(values, part)
		}
	}
	return values
}

func firstUnbalanced(expr string) int {
	stack := []rune{}
	for i, r := range expr {
		switch r {
		case '(', '{', '[':
			stack = append(stack, r)
		case ')':
			if len(stack) == 0 || stack[len(stack)-1] != '(' {
				return i
			}
			stack = stack[:len(stack)-1]
		case '}':
			if len(stack) == 0 || stack[len(stack)-1] != '{' {
				return i
			}
			stack = stack[:len(stack)-1]
		case ']':
			if len(stack) == 0 || stack[len(stack)-1] != '[' {
				return i
			}
			stack = stack[:len(stack)-1]
		}
	}
	if len(stack) > 0 {
		return len(expr) - 1
	}
	return -1
}

func estimatePoints(start, end, step int64) int64 {
	_, _, normalizedStep := normalizeRange(start, end, step)
	if normalizedStep <= 0 {
		normalizedStep = defaultStepSeconds
	}
	if end <= start {
		return 1
	}
	return (end - start) / 1000 / normalizedStep
}

func normalizeRange(start, end, step int64) (int64, int64, int64) {
	now := time.Now()
	if end == 0 {
		end = now.UnixMilli()
	}
	if start == 0 {
		start = now.Add(-time.Hour).UnixMilli()
	}
	if step <= 0 {
		step = defaultStepSeconds
	}
	return start, end, step
}
