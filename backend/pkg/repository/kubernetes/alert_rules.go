package kubernetes

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/prometheus/common/model"
	promfmt "github.com/prometheus/prometheus/model/rulefmt"
)

type AlertRules struct {
	Rules []*request.AlertRule

	Groups []AlertGroup
}

type AlertGroup struct {
	Name        string          `json:"name"`
	Interval    model.Duration  `json:"interval,omitempty"`
	QueryOffset *model.Duration `json:"query_offset,omitempty"`
	Limit       int             `json:"limit,omitempty"`

	// TODO record alert group label
	// e.g 应用指标 -> group: app
}

func ParseAlertRules(strContent string) (*AlertRules, error) {
	data := []byte(strContent)
	ruleGroups, errs := promfmt.Parse(data)
	if errs != nil {
		return nil, errs[0]
	}

	groups := make([]AlertGroup, 0)
	rules := make([]*request.AlertRule, 0)

	for i := 0; i < len(ruleGroups.Groups); i++ {
		ruleGroup := ruleGroups.Groups[i]
		groups = append(groups, AlertGroup{
			Name:        ruleGroup.Name,
			Interval:    ruleGroup.Interval,
			QueryOffset: ruleGroup.QueryOffset,
			Limit:       ruleGroup.Limit,
		})
		for s := 0; s < len(ruleGroup.Rules); s++ {
			ruleNode := ruleGroup.Rules[s]
			alertRule := &request.AlertRule{
				Group: ruleGroup.Name,

				Alert:         ruleNode.Alert.Value,
				Expr:          ruleNode.Expr.Value,
				For:           ruleNode.For.String(),
				KeepFiringFor: ruleNode.KeepFiringFor.String(),
				Labels:        ruleNode.Labels,
				Annotations:   ruleNode.Annotations,
			}

			rules = append(rules, alertRule)
		}
	}

	return &AlertRules{
		Rules:  rules,
		Groups: groups,
	}, nil
}
