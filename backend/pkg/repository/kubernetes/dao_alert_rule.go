package kubernetes

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/hashicorp/go-multierror"
	"github.com/prometheus/common/model"
	promfmt "github.com/prometheus/prometheus/model/rulefmt"
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

func (k *k8sApi) GetAlertRules(configFile string, filter *request.AlertRuleFilter, pageParam *request.PageParam) ([]*request.AlertRule, int) {
	if len(configFile) == 0 {
		configFile = k.MetadataSettings.AlertRuleFileName
	}
	return k.Metadata.GetAlertRules(configFile, filter, pageParam)
}

type ErrAlertRuleValidate struct {
	err error
}

func (e ErrAlertRuleValidate) Error() string {
	return e.err.Error()
}

func (k *k8sApi) AddOrUpdateAlertRule(configFile string, alertRule request.AlertRule) error {
	if len(configFile) == 0 {
		configFile = k.MetadataSettings.AlertRuleFileName
	}

	err := ValidateAlertRule(alertRule)
	if err != nil {
		return ErrAlertRuleValidate{err: err}
	}

	err = k.Metadata.AddorUpdateAlertRule(configFile, alertRule)
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
	var err error

	var keepFiringFor model.Duration
	if len(rule.KeepFiringFor) > 0 {
		keepFiringFor, err = model.ParseDuration(rule.KeepFiringFor)
		if err != nil {
			return fmt.Errorf("'keepFiringFor' in alertRule is illegal: %s", rule.KeepFiringFor)
		}
	}

	forDuration, err := model.ParseDuration(rule.For)
	if err != nil {
		return fmt.Errorf("'for' in alertRule is illegal: %s", rule.KeepFiringFor)
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
