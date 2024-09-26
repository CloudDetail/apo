package kubernetes

import (
	"fmt"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"strings"
	"sync"

	errmodel "github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/prometheus/common/model"
	promfmt "github.com/prometheus/prometheus/model/rulefmt"
	"gopkg.in/yaml.v3"
)

type Metadata struct {
	alertRulesLock sync.RWMutex
	AlertRulesMap  map[string]*AlertRules

	amConfigLock sync.RWMutex
	AMConfigMap  map[string]*amconfig.Config
}

func (m *Metadata) SetAlertRules(configFile string, rules *AlertRules) {
	m.alertRulesLock.Lock()
	defer m.alertRulesLock.Unlock()
	m.AlertRulesMap[configFile] = rules
}

func (m *Metadata) SetAMConfig(configFile string, configs *amconfig.Config) {
	m.amConfigLock.Lock()
	defer m.amConfigLock.Unlock()
	m.AMConfigMap[configFile] = configs
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

func (m *Metadata) GetAMConfigReceiver(configFile string, filter *request.AMConfigReceiverFilter, pageParam *request.PageParam) ([]amconfig.Receiver, int) {
	m.amConfigLock.RLock()
	defer m.amConfigLock.RUnlock()

	amConfig, find := m.AMConfigMap[configFile]
	if !find {
		return []amconfig.Receiver{}, 0
	}

	var res []amconfig.Receiver = make([]amconfig.Receiver, 0)
	for i := 0; i < len(amConfig.Receivers); i++ {
		receiver := amConfig.Receivers[i]
		if !amconfig.HasEmailOrWebhookConfig(receiver) {
			continue
		}
		if matchAMConfigReceiverFilter(filter, receiver) {
			res = append(res, receiver)
		}
	}

	if pageParam == nil {
		return res, len(res)
	}

	return pageByParam(res, pageParam)
}

func (m *Metadata) AddorUpdateAMConfigReceiver(configFile string, receiver amconfig.Receiver) error {
	m.amConfigLock.Lock()
	defer m.alertRulesLock.Unlock()

	amConfig, find := m.AMConfigMap[configFile]
	if !find {
		return fmt.Errorf("configfile %s is not found", configFile)
	}

	// Update Exist receiver
	for i := range amConfig.Receivers {
		if amConfig.Receivers[i].Name == receiver.Name {
			amConfig.Receivers[i].WebhookConfigs = receiver.WebhookConfigs
			amConfig.Receivers[i].EmailConfigs = receiver.EmailConfigs
			return nil
		}
	}

	amConfig.Receivers = append(amConfig.Receivers, receiver)
	return nil
}

func (m *Metadata) UpdateAlertRule(configFile string, alertRule request.AlertRule, oldGroup, oldAlert string) error {
	var remove = true
	if oldGroup == alertRule.Group && oldAlert == alertRule.Alert {
		remove = false
	}
	var err errmodel.ErrorWithMessage

	m.alertRulesLock.Lock()
	defer m.alertRulesLock.Unlock()

	alertRules, find := m.AlertRulesMap[configFile]
	if !find {
		err.Err = fmt.Errorf("can not find specific config: %s", configFile)
		err.Code = code.AlertConfigFileNotExistError
		return err
	}

	// 先检查旧告警的存在性
	if !checkGroupExists(oldGroup, *alertRules) {
		err.Err = fmt.Errorf("old group not exists: %s", oldGroup)
		err.Code = code.AlertOldGroupNotExistError
		return err
	}

	if !checkAlertExists(oldGroup, oldAlert, *alertRules) {
		err.Err = fmt.Errorf("old alert not exists: %s", oldAlert)
		err.Code = code.AlertAlertNotExistError
		return err
	}

	// 如果是移动操作，需要检查新告警的存在性
	if remove {
		if !checkGroupExists(alertRule.Group, *alertRules) {
			err.Err = fmt.Errorf("can not find specific group: %s", alertRule.Group)
			err.Code = code.AlertTargetGroupNotExistError
			return err
		}

		if checkAlertExists(alertRule.Group, alertRule.Alert, *alertRules) {
			err.Err = fmt.Errorf("target alert already exists: %s", alertRule.Alert)
			err.Code = code.AlertAlertAlreadyExistError
			return err
		}

	}

	// 前面已经检查了旧告警的存在性
	checkAndRemoveAlertRule(oldGroup, oldAlert, alertRules)
	alertRules.Rules = append(alertRules.Rules, &alertRule)

	return nil
}

func checkGroupExists(group string, alertRules AlertRules) bool {
	for _, g := range alertRules.Groups {
		if g.Name == group {
			return true
		}
	}

	return false
}

func checkAlertExists(group, alert string, alertRules AlertRules) bool {
	for i := 0; i < len(alertRules.Rules); i++ {
		if alertRules.Rules[i].Alert == alert &&
			alertRules.Rules[i].Group == group {
			return true
		}
	}

	return false
}

func (m *Metadata) CheckAlertRule(configFile, group, alert string) (bool, error) {
	m.alertRulesLock.RLock()
	defer m.alertRulesLock.RUnlock()

	var err errmodel.ErrorWithMessage
	alertRules, find := m.AlertRulesMap[configFile]
	if !find {
		err.Err = fmt.Errorf("configfile %s is not found", configFile)
		err.Code = code.AlertConfigFileNotExistError
		return false, err
	}

	find = checkGroupExists(group, *alertRules)
	if !find {
		err.Err = fmt.Errorf("can not find specific group: %s", group)
		err.Code = code.AlertTargetGroupNotExistError
		return false, err
	}

	find = checkAlertExists(group, alert, *alertRules)
	if find {
		return false, nil
	}

	return true, nil
}

func (m *Metadata) AddAlertRule(configFile string, alertRule request.AlertRule) error {
	m.alertRulesLock.Lock()
	defer m.alertRulesLock.Unlock()

	var err errmodel.ErrorWithMessage

	alertRules, find := m.AlertRulesMap[configFile]
	if !find {
		err.Err = fmt.Errorf("can not find specific config: %s", configFile)
		err.Code = code.AlertConfigFileNotExistError
		return err
	}

	// 检查group是否存在
	var isGroupExists bool
	for _, group := range alertRules.Groups {
		if group.Name == alertRule.Group {
			isGroupExists = true
			break
		}
	}
	if !isGroupExists {
		err.Err = fmt.Errorf("can not find specific group: %s", alertRule.Group)
		err.Code = code.AlertTargetGroupNotExistError
		return err
	}

	// 检查alert是否可用
	var isAlertExist bool
	for i := 0; i < len(alertRules.Rules); i++ {
		if alertRules.Rules[i].Alert == alertRule.Alert &&
			alertRules.Rules[i].Group == alertRule.Group {
			isAlertExist = true
		}
	}

	if isAlertExist {
		err.Err = fmt.Errorf("alert already exists: %s", alertRule.Alert)
		err.Code = code.AlertAlertNotExistError
	}

	alertRules.Rules = append(alertRules.Rules, &alertRule)
	return nil
}

func (m *Metadata) DeleteAMConfigReceiver(configFile string, name string) bool {
	m.amConfigLock.Lock()
	defer m.amConfigLock.Unlock()

	amConfig, find := m.AMConfigMap[configFile]
	if !find {
		return false
	}

	for i := 0; i < len(amConfig.Receivers); i++ {
		if amConfig.Receivers[i].Name == name {
			amConfig.Receivers = removeElement(amConfig.Receivers, i)
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

func checkAndRemoveAlertRule(group, alert string, alertRules *AlertRules) bool {
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

				var keepFiringFor model.Duration
				if len(rule.KeepFiringFor) > 0 {
					keepFiringFor, err = model.ParseDuration(rule.KeepFiringFor)
					if err != nil {
						return nil, err
					}
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

	amConfig, find := m.AMConfigMap[configFile]
	if !find {
		return nil, fmt.Errorf("configfile %s is not found", configFile)
	}

	return yaml.Marshal(amConfig)
}

func matchAMConfigReceiverFilter(filter *request.AMConfigReceiverFilter, receiver amconfig.Receiver) bool {
	if filter == nil {
		return true
	}

	if len(filter.Name) > 0 {
		return receiver.Name == filter.Name
	}

	if len(filter.RType) > 0 {
		switch filter.RType {
		case "webhook":
			return len(receiver.WebhookConfigs) > 0
		case "email":
			return len(receiver.EmailConfigs) > 0
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
