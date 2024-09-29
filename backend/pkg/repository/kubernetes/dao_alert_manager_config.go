package kubernetes

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"

	"go.uber.org/zap"
)

func (k *k8sApi) syncAMConfig() error {
	res, err := k.GetAlertManagerConfigFile("")
	if err != nil {
		return err
	}
	for key, config := range res {
		amConfig, err := amconfig.Load(config)
		if err != nil {
			k.logger.Error("failed to load amConfig", zap.String("key", key), zap.Error(err))
			continue
		}

		k.Metadata.SetAMConfig(key, amConfig)
	}
	return nil
}

func (k *k8sApi) GetAMConfigReceiver(configFile string, filter *request.AMConfigReceiverFilter, pageParam *request.PageParam, syncNow bool) ([]amconfig.Receiver, int) {
	if len(configFile) == 0 {
		configFile = k.MetadataSettings.AlertManagerFileName
	}

	if syncNow {
		err := k.syncAMConfig()
		if err != nil {
			k.logger.Error("failed to sync amConfig with k8sAPI", zap.Error(err))
		}
	}
	return k.Metadata.GetAMConfigReceiver(configFile, filter, pageParam)
}

func (k *k8sApi) AddAMConfigReceiver(configFile string, receiver amconfig.Receiver) error {
	if len(configFile) == 0 {
		configFile = k.MetadataSettings.AlertManagerFileName
	}

	err := ValidateAMConfigReceiver(receiver)
	if err != nil {
		return err
	}

	err = k.Metadata.AddAMConfigReceiver(configFile, receiver)
	if err != nil {
		return err
	}

	content, err := k.Metadata.AlertManagerConfigMarshalToYaml(configFile)
	if err != nil {
		return err
	}
	return k.UpdateAlertManagerConfigFile(configFile, content)
}

func (k *k8sApi) UpdateAMConfigReceiver(configFile string, receiver amconfig.Receiver, oldName string) error {
	if len(configFile) == 0 {
		configFile = k.MetadataSettings.AlertManagerFileName
	}

	err := ValidateAMConfigReceiver(receiver)
	if err != nil {
		return err
	}

	err = k.Metadata.UpdateAMConfigReceiver(configFile, receiver, oldName)
	if err != nil {
		return err
	}

	content, err := k.Metadata.AlertManagerConfigMarshalToYaml(configFile)
	if err != nil {
		return err
	}

	return k.UpdateAlertManagerConfigFile(configFile, content)
}

func (k *k8sApi) DeleteAMConfigReceiver(configFile string, name string) error {
	if len(configFile) == 0 {
		configFile = k.MetadataSettings.AlertManagerFileName
	}

	isDeleted, err := k.Metadata.DeleteAMConfigReceiver(configFile, name)
	if !isDeleted {
		return err
	}

	content, err := k.Metadata.AlertManagerConfigMarshalToYaml(configFile)
	if err != nil {
		return err
	}

	return k.UpdateAlertManagerConfigFile(configFile, content)
}

func (k *k8sApi) GetAlertManagerConfigFile(alertManagerConfig string) (map[string]string, error) {
	return k.getConfigMap(k.AlertManagerCMName, alertManagerConfig)
}

func (k *k8sApi) UpdateAlertManagerConfigFile(alertManagerConfig string, content []byte) error {
	return k.updateConfigMap(k.AlertManagerCMName, alertManagerConfig, content)
}

func ValidateAMConfigReceiver(receiver amconfig.Receiver) error {
	if len(receiver.WebhookConfigs) == 0 && len(receiver.EmailConfigs) == 0 {
		return model.NewErrWithMessage(fmt.Errorf("receiver %s has no webhook or email config", receiver.Name), code.AlertManagerEmptyReceiver)
	}

	if receiver.EmailConfigs != nil {
		for _, cfg := range receiver.EmailConfigs {
			if len(cfg.From) == 0 {
				return model.NewErrWithMessage(fmt.Errorf("receiver %s email config has no from", receiver.Name), code.AlertManagerReceiverEmailFromMissing)
			}
			if len(cfg.Smarthost.String()) == 0 {
				return model.NewErrWithMessage(fmt.Errorf("receiver %s email config has no smarthost", receiver.Name), code.AlertManagerReceiverEmailHostMissing)
			}
		}
	}

	return nil
}
