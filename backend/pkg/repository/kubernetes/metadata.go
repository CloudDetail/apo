package kubernetes

import (
	"fmt"
	"strings"
	"sync"

	"github.com/CloudDetail/apo/backend/pkg/code"

	"github.com/CloudDetail/apo/backend/pkg/model"
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

func (m *Metadata) AddAMConfigReceiver(configFile string, receiver amconfig.Receiver) error {
	m.amConfigLock.Lock()
	defer m.amConfigLock.Unlock()

	amConfig, find := m.AMConfigMap[configFile]
	if !find {
		return model.NewErrWithMessage(fmt.Errorf("configfile %s is not found", configFile), code.AlertConfigFileNotExistError)
	}

	for i := range amConfig.Receivers {
		if amConfig.Receivers[i].Name == receiver.Name {
			return model.NewErrWithMessage(fmt.Errorf("add receiver failed,receiver '%s' already exists", receiver.Name), code.AlertManagerReceiverAlreadyExistsError)
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
		return model.NewErrWithMessage(fmt.Errorf("configfile %s is not found", configFile), code.AlertConfigFileNotExistError)
	}

	if len(oldName) > 0 && oldName != receiver.Name {
		// Update Exist receiver
		var receiverIsExist bool
		for i := range amConfig.Receivers {
			if amConfig.Receivers[i].Name == oldName {
				receiverIsExist = true
				amConfig.Receivers[i].Name = receiver.Name
				amConfig.Receivers[i].WebhookConfigs = receiver.WebhookConfigs
				amConfig.Receivers[i].EmailConfigs = receiver.EmailConfigs

				for _, route := range amConfig.Route.Routes {
					if route.Receiver == oldName {
						route.Receiver = receiver.Name
					}
				}
				return nil
			}
		}
		if !receiverIsExist {
			return model.NewErrWithMessage(fmt.Errorf("update receiver failed, '%s' not found", oldName), code.AlertManagerReceiverNotExistsError)
		}

	}

	var receiverIsExist bool
	for i := range amConfig.Receivers {
		if amConfig.Receivers[i].Name == receiver.Name {
			receiverIsExist = true
			amConfig.Receivers[i].WebhookConfigs = receiver.WebhookConfigs
			amConfig.Receivers[i].EmailConfigs = receiver.EmailConfigs
			return nil
		}
	}
	if !receiverIsExist {
		return model.NewErrWithMessage(fmt.Errorf("update receiver failed, '%s' not found", oldName), code.AlertManagerReceiverNotExistsError)
	}

	return nil
}

func (m *Metadata) DeleteAMConfigReceiver(configFile string, name string) (bool, error) {
	m.amConfigLock.Lock()
	defer m.amConfigLock.Unlock()

	amConfig, find := m.AMConfigMap[configFile]
	if !find {
		return false, model.NewErrWithMessage(fmt.Errorf("configfile %s is not found", configFile), code.AlertConfigFileNotExistError)
	}

	if name == amConfig.Route.Receiver {
		return false, model.NewErrWithMessage(fmt.Errorf("delete receiver failed, '%s' is the default receiver", name), code.AlertManagerDefaultReceiverCannotDelete)
	}

	for i := 0; i < len(amConfig.Receivers); i++ {
		if amConfig.Receivers[i].Name == name {
			amConfig.Receivers = removeElement(amConfig.Receivers, i)

			// 移除不存在的路由
			for i, route := range amConfig.Route.Routes {
				if route.Receiver == name {
					amConfig.Route.Routes = removeElement(amConfig.Route.Routes, i)
					break
				}
			}
			return true, nil
		}
	}

	return false, model.NewErrWithMessage(fmt.Errorf("delete receiver failed, '%s' not found", name), code.AlertManagerReceiverNotExistsError)
}

func (m *Metadata) AddAlertRule(configFile string, alertRule request.AlertRule) error {
	m.alertRulesLock.Lock()
	defer m.alertRulesLock.Unlock()

	alertRules, find := m.AlertRulesMap[configFile]
	if !find {
		return model.NewErrWithMessage(
			fmt.Errorf("can not find specific config: %s", configFile),
			code.AlertConfigFileNotExistError)
	}

	// 检查group是否存在
	if checkGroupExists(alertRule.Group, *alertRules) {
		// 组存在, 检查alert是否可用
		if checkAlertExists(alertRule.Group, alertRule.Alert, *alertRules) {
			return model.NewErrWithMessage(
				fmt.Errorf("alert already exists: %s", alertRule.Alert),
				code.AlertAlertAlreadyExistError)
		}
	} else {
		name, _ := GetLabel(alertRule.Group)
		alertRules.Groups = append(alertRules.Groups, AlertGroup{Name: name})
	}

	alertRules.Rules = append(alertRules.Rules, &alertRule)
	return nil
}

func (m *Metadata) UpdateAlertRule(configFile string, alertRule request.AlertRule, oldGroup, oldAlert string) error {

	m.alertRulesLock.Lock()
	defer m.alertRulesLock.Unlock()

	alertRules, find := m.AlertRulesMap[configFile]
	if !find {
		return model.NewErrWithMessage(
			fmt.Errorf("can not find specific config: %s", configFile),
			code.AlertConfigFileNotExistError)
	}

	// 先检查旧告警的存在性
	if !checkGroupExists(oldGroup, *alertRules) {
		return model.NewErrWithMessage(
			fmt.Errorf("old group not exists: %s", oldGroup),
			code.AlertOldGroupNotExistError)
	}

	if !checkAlertExists(oldGroup, oldAlert, *alertRules) {
		return model.NewErrWithMessage(
			fmt.Errorf("old alert not exists: %s", oldAlert),
			code.AlertAlertNotExistError)
	}

	// 如果是移动操作，需要检查新告警的存在性
	if oldGroup != alertRule.Group || oldAlert != alertRule.Alert {
		// 组中存在这个告警
		if checkAlertExists(alertRule.Group, alertRule.Alert, *alertRules) {
			return model.NewErrWithMessage(
				fmt.Errorf("alert already exists: %s", alertRule.Alert),
				code.AlertAlertAlreadyExistError)
		} else if !checkGroupExists(alertRule.Group, *alertRules) {
			// 新增一个组
			name, _ := GetLabel(alertRule.Group)
			alertRules.Groups = append(alertRules.Groups, AlertGroup{Name: name})
		}
	}

	// 前面已经检查了旧告警的存在性
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

	var err model.ErrWithMessage
	alertRules, find := m.AlertRulesMap[configFile]
	if !find {
		err.Err = fmt.Errorf("configfile %s is not found", configFile)
		err.Code = code.AlertConfigFileNotExistError
		return false, err
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

// checkGroupExists 检查在alertRules中group是否存在，调用时默认已经取锁
func checkGroupExists(group string, alertRules AlertRules) bool {
	for _, g := range alertRules.Groups {
		if g.Name == group {
			return true
		}
	}

	return false
}

// checkAlertExists 检查alertRules中group下alert是否存在，调用时默认已经取锁
// group不存在或者group中的alert不存在返回false
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
