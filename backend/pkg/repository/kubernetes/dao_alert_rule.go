package kubernetes

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/prometheus/common/model"
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

func (k *k8sApi) AddOrUpdateAlertRule(configFile string, alertRule request.AlertRule) error {
	if len(configFile) == 0 {
		configFile = k.MetadataSettings.AlertRuleFileName
	}

	var err error
	if len(alertRule.KeepFiringFor) > 0 {
		_, err = model.ParseDuration(alertRule.KeepFiringFor)
		if err != nil {
			return fmt.Errorf("'keepFiringFor' in alertRule is illegal: %s", alertRule.KeepFiringFor)
		}
	}
	_, err = model.ParseDuration(alertRule.For)
	if err != nil {
		return fmt.Errorf("'For' in alertRule is illegal: %s", alertRule.KeepFiringFor)
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
