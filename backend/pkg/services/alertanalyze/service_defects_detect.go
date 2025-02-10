package alertanalyze

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/repository/kubernetes"
	"github.com/CloudDetail/apo/backend/pkg/repository/prometheus"
)

// DetectDefects 缺陷检测 查询当前输入时间段内的缺陷,
func (s *service) DetectDefects(req *request.DetectMutationRequest) error {
	startTime := time.UnixMicro(req.StartTime)
	endTime := time.UnixMicro(req.EndTime)
	step := time.Duration(req.Step * int64(time.Microsecond))

	if len(req.MutationCheck.Group) == 0 {
		req.MutationCheck.Group = kubernetes.MutationCustomLabelVal
	}

	// 准备告警事件所需的Detail模版
	pql := req.MutationCheck.String()
	upperLimit := req.UpperLimit.String()
	lowerLimit := req.LowerLimit.String()
	var summary string
	if req.UpperLimit != nil || req.LowerLimit != nil {
		summary = fmt.Sprintf("(%s) 异常 (metric: %s,lower: %s,upper: %s)", req.DetectName, pql, lowerLimit, upperLimit)
	} else {
		summary = fmt.Sprintf("(%s) 异常 (metric: %s)", req.DetectName, pql)
	}

	mutationPQLCheck := &prometheus.MutationPQLCheck{
		PQL:        pql,
		UpperLimit: upperLimit,
		LowerLimit: lowerLimit,
	}

	if req.SynchronizeToAlertRules {
		group, find := kubernetes.GetLabel(req.MutationCheck.Group)
		if !find {
			group = kubernetes.MutationCustomLabelKey
		}
		err := s.k8sRepo.AddAlertRule("", request.AlertRule{
			Group: req.MutationCheck.Group,
			Alert: req.DetectName,
			Expr:  mutationPQLCheck.GetExecutedPQL(),
			For:   req.For,
			Labels: map[string]string{
				"severity": "warning",
				"group":    group,
			},
			Annotations: map[string]string{
				"description": req.DetectName + " 检测到异常\n  VALUE = {{ $value }}\n  LABELS = {{ $labels }}",
				"summary":     summary,
			},
		})

		if err != nil {
			// 创建告警规则失败
			return model.ErrWithMessage{
				Err:  err,
				Code: code.DetectDefectsCreatAlertError,
			}
		}
	}

	metricResults, err := s.promRepo.ExecutedMutationCheck(mutationPQLCheck, startTime, endTime, step)
	if err != nil {
		return err
	}

	var forMicroSecond int64 = 0
	if len(req.For) > 0 {
		duration, err := time.ParseDuration(req.For)
		if err != nil {
			return err
		}
		forMicroSecond = duration.Microseconds()
	}

	var totalAlertEvents []*model.AlertEvent

	for _, metricResult := range metricResults {
		points := metricResult.Values
		mutationRanges := splitContinuePoint(points, forMicroSecond, req.Step)
		for i := 0; i < len(mutationRanges); i++ {
			alertEvents := transferMetricResult2AlertEvent(
				req.DetectName, req.MutationCheck.Group,
				req.EndTime, req.Step,
				forMicroSecond, summary,
				metricResult.Metric,
				mutationRanges[i],
			)

			totalAlertEvents = append(totalAlertEvents, alertEvents...)
		}
	}

	// 将时间周期内生成的告警事件存入Clickhouse
	err = s.chRepo.InsertBatchAlertEvents(context.Background(), totalAlertEvents)
	if err != nil {
		return err
	}

	detectMutation := model.DetectMutation{
		DetectName:    req.DetectName,
		Step:          req.Step,
		StartTime:     req.StartTime,
		EndTime:       req.EndTime,
		For:           req.For,
		MutationCheck: mutationPQLCheck.GetExecutedPQL(),
		Timestamp:     time.Now().UnixMicro(),
		Group:         req.MutationCheck.Group,
	}
	if err = s.chRepo.AddDetectMutation(detectMutation); err != nil {
		return model.NewErrWithMessage(errors.New("failed to add record"), code.AddExecRecordError)
	}
	return nil
}

// splitContinuePoint 将时间范围内的所有点,按是否连续分隔成多段
func splitContinuePoint(points []prometheus.Points, forMicroSecond int64, stepMicroSecond int64) [][]prometheus.Points {
	if len(points) == 0 {
		return [][]prometheus.Points{}
	}

	var res [][]prometheus.Points
	var rangeStart int64 = points[0].TimeStamp
	var lastSeenTime int64 = points[0].TimeStamp
	var tmpGroup []prometheus.Points = []prometheus.Points{points[0]}
	for i := 1; i < len(points); i++ {
		// 距离上次告警时间超过了For时间,中断这次告警
		if points[i].TimeStamp-lastSeenTime > stepMicroSecond {
			// 之前是否已经发生告警
			if forMicroSecond <= lastSeenTime-rangeStart && // 之前是否已经发生告警
				points[i].TimeStamp-lastSeenTime <= forMicroSecond { // 告警间隔是否超过等待时间
				// 连续
				tmpGroup = append(tmpGroup, points[i])
			} else {
				// 非连续
				res = append(res, tmpGroup)
				tmpGroup = []prometheus.Points{points[i]}
				rangeStart = points[i].TimeStamp
			}
		} else {
			// 连续
			tmpGroup = append(tmpGroup, points[i])
		}
		lastSeenTime = points[i].TimeStamp
	}
	if len(tmpGroup) > 0 {
		res = append(res, tmpGroup)
	}

	return res
}

func transferMetricResult2AlertEvent(
	detectName string, group string,
	endTimeMicroSecond int64,
	stepMicroSecond int64,
	forMicroSecond int64,
	summary string,
	labels prometheus.Labels,
	mutationRange []prometheus.Points,
) []*model.AlertEvent {
	if len(mutationRange) == 0 {
		return nil
	}
	rangeStart := mutationRange[0].TimeStamp
	rangeEnd := mutationRange[len(mutationRange)-1].TimeStamp

	groupLabel, find := kubernetes.GetLabel(group)
	if !find {
		groupLabel = kubernetes.MutationCustomLabelKey
	}

	if rangeEnd-rangeStart+stepMicroSecond < forMicroSecond {
		// 不生成告警事件
		return nil
	}

	tags := extractTagsFromLabels(detectName, group, groupLabel, labels)
	detailTemplate := extractDetailTemplate(detectName, tags, summary)

	// 为故障事件范围内所有超过ForDuration的点生成事件和更新事件
	var alertEvents []*model.AlertEvent
	var noEnd time.Time = time.Unix(0, 0)
	alertStart := rangeStart + forMicroSecond - stepMicroSecond
	for _, point := range mutationRange {
		if point.TimeStamp >= alertStart {
			alertEvents = append(alertEvents, &model.AlertEvent{
				ID:           model.GenUUID(),
				Source:       "apo-backend",
				CreateTime:   time.UnixMicro(alertStart),
				UpdateTime:   time.UnixMicro(alertStart),
				EndTime:      noEnd,
				ReceivedTime: time.UnixMicro(point.TimeStamp),
				Severity:     model.SeverityLevelWarning,
				Group:        groupLabel,
				Name:         detectName,
				Detail:       fmt.Sprintf(detailTemplate, point.Value),
				Tags:         tags,
				Status:       model.StatusFiring,
			})
		}
	}

	if len(alertEvents) > 0 {
		lastPoint := mutationRange[len(mutationRange)-1]
		shouldResolvedTime := lastPoint.TimeStamp + forMicroSecond
		// 如果查询时间端包含了告警区间的后一个点
		if shouldResolvedTime <= endTimeMicroSecond {
			alertEvents = append(alertEvents, &model.AlertEvent{
				ID:           model.GenUUID(),
				Source:       "apo-backend",
				CreateTime:   time.UnixMicro(alertStart),
				UpdateTime:   time.UnixMicro(shouldResolvedTime),
				EndTime:      noEnd,
				ReceivedTime: time.UnixMicro(shouldResolvedTime),
				Severity:     model.SeverityLevelInfo,
				Group:        groupLabel,
				Name:         detectName,
				Detail:       extractResolvedDetail(detectName, tags, summary),
				Tags:         tags,
				Status:       model.StatusResolved,
			})
		}
	}

	return alertEvents
}

func extractTagsFromLabels(name, alertGroup, group string, labels prometheus.Labels) map[string]string {
	var res = map[string]string{}
	res["alertname"] = name
	res["alertgroup"] = alertGroup
	res["group"] = group
	res["severity"] = "warning"

	if len(labels.ContainerID) > 0 {
		res["container_id"] = labels.ContainerID
	}

	if len(labels.ContentKey) > 0 {
		res["content_key"] = labels.ContentKey
	}

	if len(labels.Instance) > 0 {
		res["instance"] = labels.Instance
	}

	if len(labels.IsError) > 0 {
		res["is_error"] = labels.IsError
	}

	if len(labels.Job) > 0 {
		res["job"] = labels.Job
	}

	if len(labels.NodeName) > 0 {
		res["node_name"] = labels.NodeName
	}

	if len(labels.POD) > 0 {
		res["pod"] = labels.POD
	}

	if len(labels.SvcName) > 0 {
		res["svc_name"] = labels.SvcName
	}

	if len(labels.TopSpan) > 0 {
		res["top_span"] = labels.TopSpan
	}

	if len(labels.PID) > 0 {
		res["pid"] = labels.PID
	}

	if len(labels.PodName) > 0 {
		res["pod_name"] = labels.PodName
	}

	if len(labels.Namespace) > 0 {
		res["namespace"] = labels.Namespace
	}

	if len(labels.NodeIP) > 0 {
		res["node_ip"] = labels.NodeIP
	}

	if len(labels.DBSystem) > 0 {
		res["db_system"] = labels.DBSystem
	}

	if len(labels.DBName) > 0 {
		res["db_name"] = labels.DBName
	}

	if len(labels.DBUrl) > 0 {
		res["db_url"] = labels.DBUrl
	}

	if len(labels.MonitorName) > 0 {
		res["monitor_name"] = labels.MonitorName
	}

	return res
}

func extractDetailTemplate(name string, tags map[string]string, summary string) string {
	tagsStr, err := json.Marshal(tags)
	if err != nil {
		tagsStr = []byte("tags marshal error")
	}
	var summaryStr bytes.Buffer
	json.HTMLEscape(&summaryStr, []byte(summary))
	return `{"description":"` + name + ` 检测到异常\n  VALUE = %f\n  LABELS = ` + string(tagsStr) + `","summary":"` + summaryStr.String() + `"}`
}

func extractResolvedDetail(name string, tags map[string]string, summary string) string {
	tagsStr, err := json.Marshal(tags)
	if err != nil {
		tagsStr = []byte("tags marshal error")
	}
	return `{"description":"` + name + ` 检测到异常\n  LABELS = ` + string(tagsStr) + `","summary":"` + summary + `"}`
}
