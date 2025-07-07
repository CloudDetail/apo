// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package kubernetes

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/core"

	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	prommodel "github.com/prometheus/common/model"
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
		filteredWebhookConfigs := []*amconfig.WebhookConfig{}
		// DingTalk configuration exists in the webhook, which needs to be ignored and obtained from db.
		for j := range receiver.WebhookConfigs {
			if !strings.Contains(receiver.WebhookConfigs[j].URL.String(), "/outputs/dingtalk/") {
				filteredWebhookConfigs = append(filteredWebhookConfigs, receiver.WebhookConfigs[j])
			}
		}

		receiver.WebhookConfigs = filteredWebhookConfigs
		if !amconfig.HasEmailOrWebhookConfig(receiver) {
			continue
		}
		if matchAMConfigReceiverFilter(filter, &receiver) {
			res = append(res, receiver)
		}
	}

	if pageParam == nil {
		return res, len(res)
	}

	return pageByParam(res, pageParam)
}

func (m *Metadata) AddAMConfigReceiver(configFile string, receiver amconfig.Receiver) error {
	m.amConfigLock.Lock()
	defer m.amConfigLock.Unlock()

	amConfig, find := m.AMConfigMap[configFile]
	if !find {
		return core.Error(code.AlertConfigFileNotExistError, "configfile not found")
	}

	for i := range amConfig.Receivers {
		if amConfig.Receivers[i].Name == receiver.Name {
			return core.Error(code.AlertManagerReceiverAlreadyExistsError, fmt.Sprintf("add receiver failed,receiver '%s' already exists", receiver.Name))
		}
	}

	routeIsExist := false
	for _, route := range amConfig.Route.Routes {
		if route.Receiver == receiver.Name {
			routeIsExist = true
			break
		}
	}

	if !routeIsExist {
		amConfig.Route.Routes = append(amConfig.Route.Routes, &amconfig.Route{
			Receiver: receiver.Name,
			Continue: true,
		})
	}

	amConfig.Receivers = append(amConfig.Receivers, receiver)
	return nil
}

func (m *Metadata) UpdateAMConfigReceiver(configFile string, receiver amconfig.Receiver, oldName string) error {
	m.amConfigLock.Lock()
	defer m.amConfigLock.Unlock()

	amConfig, find := m.AMConfigMap[configFile]
	if !find {
		return core.Error(code.AlertConfigFileNotExistError, "configfile not found")
	}

	// Update Exist receiver
	var receiverIsExist bool
	if len(oldName) == 0 || oldName == receiver.Name {
		for i := range amConfig.Receivers {
			if amConfig.Receivers[i].Name == oldName {
				receiverIsExist = true
				if len(receiver.WebhookConfigs) > 0 {
					amConfig.Receivers[i].WebhookConfigs = receiver.WebhookConfigs
				} else if len(receiver.EmailConfigs) > 0 {
					amConfig.Receivers[i].EmailConfigs = receiver.EmailConfigs
				} else if len(receiver.WechatConfigs) > 0 {
					amConfig.Receivers[i].WechatConfigs = receiver.WechatConfigs
				}
			}
		}
	} else if len(oldName) > 0 && oldName != receiver.Name {
		for i := range amConfig.Receivers {
			if amConfig.Receivers[i].Name == oldName {
				receiverIsExist = true
				amConfig.Receivers[i].Name = receiver.Name
				if len(receiver.WebhookConfigs) > 0 {
					amConfig.Receivers[i].WebhookConfigs = receiver.WebhookConfigs
				} else if len(receiver.EmailConfigs) > 0 {
					amConfig.Receivers[i].EmailConfigs = receiver.EmailConfigs
				} else if len(receiver.WechatConfigs) > 0 {
					amConfig.Receivers[i].WechatConfigs = receiver.WechatConfigs
				}

				for _, route := range amConfig.Route.Routes {
					if route.Receiver == oldName {
						route.Receiver = receiver.Name
					}
				}
				return nil
			}
		}
	}
	if !receiverIsExist {
		return core.Error(code.AlertManagerReceiverNotExistsError, fmt.Sprintf("update receiver failed, '%s' not found", oldName))
	}

	return nil
}

func (m *Metadata) DeleteAMConfigReceiver(configFile string, name string) (bool, error) {
	m.amConfigLock.Lock()
	defer m.amConfigLock.Unlock()

	amConfig, find := m.AMConfigMap[configFile]
	if !find {
		return false, core.Error(code.AlertConfigFileNotExistError, "configfile not found")
	}

	if name == amConfig.Route.Receiver {
		return false, core.Error(code.AlertManagerDefaultReceiverCannotDelete, fmt.Sprintf("delete receiver failed, '%s' is the default receiver", name))
	}

	for i := 0; i < len(amConfig.Receivers); i++ {
		if amConfig.Receivers[i].Name == name {
			amConfig.Receivers = removeElement(amConfig.Receivers, i)

			// Remove non-existent routes
			for i, route := range amConfig.Route.Routes {
				if route.Receiver == name {
					amConfig.Route.Routes = removeElement(amConfig.Route.Routes, i)
					break
				}
			}
			return true, nil
		}
	}

	return false, core.Error(code.AlertManagerReceiverNotExistsError, fmt.Sprintf("delete receiver failed, '%s' not found", name))
}

func (m *Metadata) AddAlertRule(configFile string, alertRule request.AlertRule) error {
	m.alertRulesLock.Lock()
	defer m.alertRulesLock.Unlock()

	alertRules, find := m.AlertRulesMap[configFile]
	if !find {
		return core.Error(code.AlertConfigFileNotExistError, "configfile not found")
	}

	// Check whether the group exists
	if checkGroupExists(alertRule.Group, *alertRules) {
		// group exists, check whether alert is available
		if checkAlertExists(alertRule.Group, alertRule.Alert, *alertRules) {
			return core.Error(code.AlertAlertAlreadyExistError, fmt.Sprintf("alert already exists: %s", alertRule.Alert))
		}
	} else {
		// name, _ := GetLabel(alertRule.Group)
		alertRules.Groups = append(alertRules.Groups, AlertGroup{Name: alertRule.Group})
	}

	alertRules.Rules = append(alertRules.Rules, &alertRule)
	return nil
}

func (m *Metadata) UpdateAlertRule(configFile string, alertRule request.AlertRule, oldGroup, oldAlert string) error {

	m.alertRulesLock.Lock()
	defer m.alertRulesLock.Unlock()

	alertRules, find := m.AlertRulesMap[configFile]
	if !find {
		return core.Error(code.AlertConfigFileNotExistError, "configfile not found")
	}

	// Check the existence of old alarms first
	if !checkGroupExists(oldGroup, *alertRules) {
		return core.Error(code.AlertOldGroupNotExistError, fmt.Sprintf("old group not exists: %s", oldGroup))
	}

	if !checkAlertExists(oldGroup, oldAlert, *alertRules) {
		return core.Error(code.AlertAlertNotExistError, fmt.Sprintf("old alert not exists: %s", oldAlert))
	}

	// If it is a move operation, you need to check the existence of new alarms.
	if oldGroup != alertRule.Group || oldAlert != alertRule.Alert {
		// This alarm exists in the group
		if checkAlertExists(alertRule.Group, alertRule.Alert, *alertRules) {
			return core.Error(code.AlertAlertAlreadyExistError, fmt.Sprintf("alert already exists: %s", alertRule.Alert))
		} else if !checkGroupExists(alertRule.Group, *alertRules) {
			// Add a new group
			name, _ := GetLabel(alertRule.Group)
			alertRules.Groups = append(alertRules.Groups, AlertGroup{Name: name})
		}
	}

	// The existence of the old alarm has been checked earlier
	checkAndRemoveAlertRule(oldGroup, oldAlert, alertRules)
	alertRules.Rules = append(alertRules.Rules, &alertRule)

	return nil
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

func (m *Metadata) CheckAlertRuleExists(configFile, group, alert string) (bool, error) {
	m.alertRulesLock.RLock()
	defer m.alertRulesLock.RUnlock()

	alertRules, find := m.AlertRulesMap[configFile]
	if !find {
		return false, core.Error(code.AlertConfigFileNotExistError, "configfile not found")
	}

	if checkGroupExists(group, *alertRules) && checkAlertExists(group, alert, *alertRules) {
		return false, nil
	}

	return true, nil
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
				forDuration, err := prommodel.ParseDuration(rule.For)
				if err != nil {
					return nil, err
				}

				var keepFiringFor prommodel.Duration
				if len(rule.KeepFiringFor) > 0 {
					keepFiringFor, err = prommodel.ParseDuration(rule.KeepFiringFor)
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

func matchAMConfigReceiverFilter(filter *request.AMConfigReceiverFilter, receiver *amconfig.Receiver) bool {
	if filter == nil {
		return true
	}

	if len(filter.Name) > 0 {
		return strings.Contains(receiver.Name, filter.Name)
	}

	if len(filter.RType) > 0 {
		switch filter.RType {
		case "webhook":
			return len(receiver.WebhookConfigs) > 0
		case "email":
			return len(receiver.EmailConfigs) > 0
		case "wechat":
			return len(receiver.WechatConfigs) > 0
		default:
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

	if len(filter.Groups) > 0 {
		if !ContainsLike(filter.Groups, rule.Group) {
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

	if filter.GroupID > 0 {
		if rule.Annotations["groupId"] != strconv.FormatInt(filter.GroupID, 10) {
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

func ContainsLike(slices []string, expected string) bool {
	for _, item := range slices {
		if strings.Contains(item, expected) {
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
	if startIdx >= totalCount {
		return nil, totalCount
	}
	endIdx := startIdx + param.PageSize
	if endIdx > totalCount {
		endIdx = totalCount
	}
	return list[startIdx:endIdx], totalCount
}

// checkGroupExists check whether the group exists in the alertRules, the lock has been fetched by default when the call is made.
func checkGroupExists(group string, alertRules AlertRules) bool {
	for _, g := range alertRules.Groups {
		if g.Name == group {
			return true
		}
	}

	return false
}

// checkAlertExists check whether the alert under the group in the alertRules exists. By default, the lock has been fetched when the call is made.
// false is returned if the group does not exist or the alert in the group does not exist.
func checkAlertExists(group, alert string, alertRules AlertRules) bool {
	for i := 0; i < len(alertRules.Rules); i++ {
		if alertRules.Rules[i].Alert == alert && alertRules.Rules[i].Group == group {
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
