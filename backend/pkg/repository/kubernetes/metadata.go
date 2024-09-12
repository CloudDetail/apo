package kubernetes

import (
	"fmt"
	"sync"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	promfmt "github.com/prometheus/prometheus/model/rulefmt"
	"gopkg.in/yaml.v3"
)

type Metadata struct {
	alertRulesLock sync.RWMutex
	AlertRulesMap  map[string]*AlertRules
}

func (m *Metadata) SetAlertRules(configFile string, rules *AlertRules) {
	m.alertRulesLock.Lock()
	defer m.alertRulesLock.Unlock()
	m.AlertRulesMap[configFile] = rules
}

func (m *Metadata) GetAlertRules(configFile string) []request.AlertRule {
	m.alertRulesLock.RLock()
	defer m.alertRulesLock.RUnlock()

	alertRules, find := m.AlertRulesMap[configFile]
	if !find {
		return []request.AlertRule{}
	}
	return alertRules.Rules
}

func (m *Metadata) AddorUpdateAlertRule(configFile string, alertRule request.AlertRule) error {
	m.alertRulesLock.Lock()
	defer m.alertRulesLock.Unlock()
	alertRules, find := m.AlertRulesMap[configFile]
	if !find {
		// TODO return error
		return fmt.Errorf("")
	}
	// check if group exists
	var isGroupExist bool = false
	for _, group := range alertRules.Groups {
		if group.Name == alertRule.Group {
			isGroupExist = true
			break
		}
	}

	if !isGroupExist {
		alertRules.Groups = append(alertRules.Groups, AlertGroup{
			Name: alertRule.Group,
		})
	}

	for i := 0; i < len(alertRules.Rules); i++ {
		if alertRules.Rules[i].Alert == alertRule.Alert &&
			alertRules.Rules[i].Group == alertRule.Group {
			alertRules.Rules[i] = alertRule
			return nil
		}
	}

	alertRules.Rules = append(alertRules.Rules, alertRule)
	return nil
}

func (m *Metadata) DeleteAlertRule(configFile string, group, alert string) bool {
	m.alertRulesLock.Lock()
	defer m.alertRulesLock.Unlock()
	alertRules, find := m.AlertRulesMap[configFile]
	if !find {
		// TODO return error
		return false
	}

	for i := 0; i < len(alertRules.Rules); i++ {
		if alertRules.Rules[i].Alert == alert &&
			alertRules.Rules[i].Group == group {
			alertRules.Rules = removeElement(alertRules.Rules, i)
			return true
		}
	}
	return false
}

func removeElement(slice []request.AlertRule, index int) []request.AlertRule {
	return append(slice[:index], slice[index+1:]...)
}

func (m *Metadata) MarshalToYaml(configFile string) ([]byte, error) {
	m.alertRulesLock.Lock()
	defer m.alertRulesLock.Unlock()
	alertRules, find := m.AlertRulesMap[configFile]
	if !find {
		// TODO return error
		return nil, nil
	}

	var content = promfmt.RuleGroups{}
	for _, group := range alertRules.Groups {

		var rules []promfmt.RuleNode
		for _, rule := range alertRules.Rules {
			if rule.Group == group.Name {

				ruleNode := promfmt.RuleNode{
					For:           rule.For,
					KeepFiringFor: rule.KeepFiringFor,
					Labels:        rule.Labels,
					Annotations:   rule.Annotations,
				}
				ruleNode.Alert.SetString(rule.Alert)
				ruleNode.Expr.SetString(rule.Expr)
				rules = append(rules, ruleNode)
			}
		}

		content.Groups = append(content.Groups, promfmt.RuleGroup{
			Name:        group.Name,
			Interval:    group.Interval,
			QueryOffset: group.QueryOffset,
			Limit:       group.Limit,
			Rules:       rules,
		})
	}

	return yaml.Marshal(content)
}
