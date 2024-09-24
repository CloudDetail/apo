package kubernetes

import (
	"fmt"
	"strings"
	"sync"

	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/prometheus/common/model"
	promfmt "github.com/prometheus/prometheus/model/rulefmt"
	"gopkg.in/yaml.v3"
)

type Metadata struct {
	alertRulesLock sync.RWMutex
	AlertRulesMap  map[string]*AlertRules

	amConfigLock           sync.RWMutex
	AlertManagerConfigsMap map[string]*AlertManagerConfig
}

func (m *Metadata) SetAlertRules(configFile string, rules *AlertRules) {
	m.alertRulesLock.Lock()
	defer m.alertRulesLock.Unlock()
	m.AlertRulesMap[configFile] = rules
}

func (m *Metadata) SetAMConfig(configFile string, configs *AlertManagerConfig) {
	m.amConfigLock.Lock()
	defer m.amConfigLock.Unlock()
	m.AlertManagerConfigsMap[configFile] = configs
}

func (m *Metadata) GetAlertRules(configFile string, filter *request.AlertRuleFilter, pageParam *request.PageParam) ([]*request.AlertRule, int) {
	m.alertRulesLock.RLock()
	defer m.alertRulesLock.RUnlock()

	alertRules, find := m.AlertRulesMap[configFile]
	if !find {
		return []*request.AlertRule{}, 0
	}

	var res []*request.AlertRule = make([]*request.AlertRule, 0)

	if filter == nil {
		res = alertRules.Rules
	} else {
		for _, rule := range alertRules.Rules {
			if matchAlertRuleFilter(filter, rule) {
				res = append(res, rule)
			}
		}
	}

	if pageParam == nil {
		return res, len(res)
	}

	return pageByParam(res, pageParam)
}

func (m *Metadata) GetAMConfigReceiver(configFile string, filter *request.AMConfigReceiverFilter, pageParam *request.PageParam) ([]*request.AMConfigReceiver, int) {
	m.amConfigLock.RLock()
	defer m.amConfigLock.RUnlock()

	amConfig, find := m.AlertManagerConfigsMap[configFile]
	if !find {
		return []*request.AMConfigReceiver{}, 0
	}

	var res []*request.AMConfigReceiver = make([]*request.AMConfigReceiver, 0)
	if filter == nil {
		res = amConfig.ReceiverList
	} else {
		for i := 0; i < len(amConfig.ReceiverList); i++ {
			receiver := amConfig.ReceiverList[i]
			if matchAMConfigReceiverFilter(filter, receiver) {
				res = append(res, receiver)
			}
		}
	}

	if pageParam == nil {
		return res, len(res)
	}

	return pageByParam(res, pageParam)
}

func (m *Metadata) AddorUpdateAMConfigReceiver(configFile string, receiver request.AMConfigReceiver) error {
	m.amConfigLock.Lock()
	defer m.alertRulesLock.Unlock()

	amConfig, find := m.AlertManagerConfigsMap[configFile]
	if !find {
		return fmt.Errorf("configfile %s is not found", configFile)
	}

	// Update Exist receiver
	for i := range amConfig.ReceiverList {
		if amConfig.ReceiverList[i].Name == receiver.Name {
			amConfig.ReceiverList[i] = &receiver
			return nil
		}
	}

	amConfig.ReceiverList = append(amConfig.ReceiverList, &receiver)
	return nil
}

func (m *Metadata) AddorUpdateAlertRule(configFile string, alertRule request.AlertRule) error {
	m.alertRulesLock.Lock()
	defer m.alertRulesLock.Unlock()
	alertRules, find := m.AlertRulesMap[configFile]
	if !find {
		return fmt.Errorf("configfile %s is not found", configFile)
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
			alertRules.Rules[i] = &alertRule
			return nil
		}
	}

	alertRules.Rules = append(alertRules.Rules, &alertRule)
	return nil
}

func (m *Metadata) DeleteAMConfigReceiver(configFile string, name string) bool {
	m.amConfigLock.Lock()
	defer m.amConfigLock.Unlock()

	amConfig, find := m.AlertManagerConfigsMap[configFile]
	if !find {
		return false
	}

	for i := 0; i < len(amConfig.ReceiverList); i++ {
		if amConfig.ReceiverList[i].Name == name {
			amConfig.ReceiverList = removeElement(amConfig.ReceiverList, i)
			return true
		}
	}

	return false
}

func (m *Metadata) DeleteAlertRule(configFile string, group, alert string) bool {
	m.alertRulesLock.Lock()
	defer m.alertRulesLock.Unlock()
	alertRules, find := m.AlertRulesMap[configFile]
	if !find {
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

func removeElement[T any](slice []T, index int) []T {
	return append(slice[:index], slice[index+1:]...)
}

func (m *Metadata) AlertRuleMarshalToYaml(configFile string) ([]byte, error) {
	m.alertRulesLock.Lock()
	defer m.alertRulesLock.Unlock()
	alertRules, find := m.AlertRulesMap[configFile]
	if !find {
		return nil, fmt.Errorf("configfile %s is not found", configFile)
	}

	var content = promfmt.RuleGroups{}
	for _, group := range alertRules.Groups {

		var rules []promfmt.RuleNode
		for _, rule := range alertRules.Rules {
			if rule.Group == group.Name {
				forDuration, err := model.ParseDuration(rule.For)
				if err != nil {
					return nil, err
				}
				keepFiringFor, err := model.ParseDuration(rule.KeepFiringFor)
				if err != nil {
					return nil, err
				}
				ruleNode := promfmt.RuleNode{
					For:           forDuration,
					KeepFiringFor: keepFiringFor,
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

func (m *Metadata) AlertManagerConfigMarshalToYaml(configFile string) ([]byte, error) {
	m.amConfigLock.RLock()
	defer m.amConfigLock.RUnlock()

	amConfig, find := m.AlertManagerConfigsMap[configFile]
	if !find {
		return nil, fmt.Errorf("configfile %s is not found", configFile)
	}

	amConfigDef := amconfig.Config{}
	for _, receiver := range amConfig.ReceiverList {
		if rDef := receiver.ToReceiverDef(); rDef != nil {
			amConfigDef.Receivers = append(amConfigDef.Receivers, *rDef)
		}
	}

	for _, receiver := range amConfig.UnsupportReceiver {
		amConfigDef.Receivers = append(amConfigDef.Receivers, *receiver)
	}

	return yaml.Marshal(amConfigDef)
}

func matchAMConfigReceiverFilter(filter *request.AMConfigReceiverFilter, receiver *request.AMConfigReceiver) bool {
	if len(filter.Name) > 0 {
		if !strings.Contains(receiver.Name, filter.Name) {
			return false
		}
	}

	if len(filter.RType) > 0 {
		if !strings.Contains(receiver.RType, filter.RType) {
			return false
		}
	}

	return true
}

func matchAlertRuleFilter(filter *request.AlertRuleFilter, rule *request.AlertRule) bool {
	if len(filter.Alert) > 0 {
		if !strings.Contains(rule.Alert, filter.Alert) {
			return false
		}
	}

	if len(filter.Group) > 0 {
		if !strings.Contains(rule.Group, filter.Group) {
			return false
		}
	}

	if len(filter.Keyword) > 0 {
		var isFind bool

		if strings.Contains(rule.Alert, filter.Keyword) ||
			strings.Contains(rule.Group, filter.Keyword) {
			isFind = true
		}

		if !isFind {
			for k, v := range rule.Labels {
				if strings.Contains(k, filter.Keyword) ||
					strings.Contains(v, filter.Keyword) {
					isFind = true
					break
				}
			}
		}

		if !isFind {
			return false
		}
	}

	if len(filter.Severity) > 0 {
		severity := rule.Labels["severity"]
		if !ContainsIn(filter.Severity, severity) {
			return false
		}
	}
	return true
}

func ContainsIn(slices []string, expected string) bool {
	for _, item := range slices {
		if item == expected {
			return true
		}
	}
	return false
}

func pageByParam[T any](list []T, param *request.PageParam) ([]T, int) {
	totalCount := len(list)
	if param == nil {
		return list, totalCount
	}

	if totalCount < param.PageSize {
		return list, totalCount
	}

	startIdx := (param.CurrentPage - 1) * param.PageSize
	endIdx := startIdx + param.PageSize
	if endIdx > totalCount {
		endIdx = totalCount
	}
	return list[startIdx:endIdx], totalCount
}
