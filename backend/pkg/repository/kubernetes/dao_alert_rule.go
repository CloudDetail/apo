package kubernetes

import (
	"github.com/CloudDetail/apo/backend/pkg/code"
	errmodel "github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/hashicorp/go-multierror"
	"github.com/prometheus/common/model"
	promfmt "github.com/prometheus/prometheus/model/rulefmt"
	"go.uber.org/zap"
)

func (k *k8sApi) syncAlertRule() error {
	res, err := k.GetAlertRuleConfigFile("")
	if err != nil {
		return err
	}
	for key, config := range res {
		alertRules, err := ParseAlertRules(config)
		if err != nil {
			continue
		}
		k.Metadata.SetAlertRules(key, alertRules)
	}
	return nil
}

func (k *k8sApi) GetAlertRules(configFile string, filter *request.AlertRuleFilter, pageParam *request.PageParam, syncNow bool) ([]*request.AlertRule, int) {
	if len(configFile) == 0 {
		configFile = k.MetadataSettings.AlertRuleFileName
	}

	if syncNow {
		err := k.syncAlertRule()
		if err != nil {
			k.logger.Error("failed to sync alertRule with k8sAPI", zap.Error(err))
		}
	}
	return k.Metadata.GetAlertRules(configFile, filter, pageParam)
}

func (k *k8sApi) AddAlertRule(configFile string, alertRules request.AlertRule) error {
	if len(configFile) == 0 {
		configFile = k.MetadataSettings.AlertRuleFileName
	}

	if err := ValidateAlertRule(alertRules); err != nil {
		return err
	}

	if err := k.Metadata.AddAlertRule(configFile, alertRules); err != nil {
		return err
	}

	content, err := k.Metadata.AlertRuleMarshalToYaml(configFile)
	if err != nil {
		return err
	}

	return k.UpdateAlertRuleConfigFile(configFile, content)
}

func (k *k8sApi) CheckAlertRule(configFile, group, alert string) (bool, error) {
	if len(configFile) == 0 {
		configFile = k.MetadataSettings.AlertRuleFileName
	}

	return k.Metadata.CheckAlertRuleExists(configFile, group, alert)
}

func (k *k8sApi) UpdateAlertRule(configFile string, alertRule request.AlertRule, oldGroup, oldAlert string) error {
	if len(configFile) == 0 {
		configFile = k.MetadataSettings.AlertRuleFileName
	}

	err := ValidateAlertRule(alertRule)
	if err != nil {
		return err
	}

	err = k.Metadata.UpdateAlertRule(configFile, alertRule, oldGroup, oldAlert)
	if err != nil {
		return err
	}

	content, err := k.Metadata.AlertRuleMarshalToYaml(configFile)
	if err != nil {
		return err
	}

	return k.UpdateAlertRuleConfigFile(configFile, content)
}

func (k *k8sApi) DeleteAlertRule(configFile string, group, alert string) error {
	if len(configFile) == 0 {
		configFile = k.MetadataSettings.AlertRuleFileName
	}
	isDeleted := k.Metadata.DeleteAlertRule(configFile, group, alert)
	if !isDeleted {
		return nil
	}

	content, err := k.Metadata.AlertRuleMarshalToYaml(configFile)
	if err != nil {
		return err
	}

	return k.UpdateAlertRuleConfigFile(configFile, content)
}

func (k *k8sApi) GetAlertRuleConfigFile(alertRuleFile string) (map[string]string, error) {
	return k.getConfigMap(k.AlertRuleCMName, alertRuleFile)
}

func (k *k8sApi) UpdateAlertRuleConfigFile(configFile string, content []byte) error {
	return k.updateConfigMap(k.AlertRuleCMName, configFile, content)
}

func ValidateAlertRule(rule request.AlertRule) error {
	var err errmodel.ErrWithMessage
	var e error
	var keepFiringFor model.Duration
	if len(rule.KeepFiringFor) > 0 {
		keepFiringFor, e = model.ParseDuration(rule.KeepFiringFor)
		if e != nil {
			err.Err = e
			err.Code = code.AlertKeepFiringForIllegalError
			return err
		}
	}

	forDuration, e := model.ParseDuration(rule.For)
	if e != nil {
		err.Err = e
		err.Code = code.AlertForIllegalError
		return err
	}
	ruleNode := promfmt.RuleNode{
		For:           forDuration,
		KeepFiringFor: keepFiringFor,
		Labels:        rule.Labels,
		Annotations:   rule.Annotations,
	}
	ruleNode.Alert.SetString(rule.Alert)
	ruleNode.Expr.SetString(rule.Expr)

	var validateErr *multierror.Error
	for _, node := range ruleNode.Validate() {
		var ruleName string
		if ruleNode.Alert.Value != "" {
			ruleName = ruleNode.Alert.Value
		} else {
			ruleName = ruleNode.Record.Value
		}
		validateErr = multierror.Append(validateErr, &promfmt.Error{
			RuleName: ruleName,
			Err:      node,
		})
	}
	return validateErr.ErrorOrNil()
}
