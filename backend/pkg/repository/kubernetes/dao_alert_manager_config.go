package kubernetes

import (
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (k *k8sApi) syncAMConfig() error {
	res, err := k.GetAlertManagerConfigFile("")
	if err != nil {
		return err
	}
	for key, config := range res {
		amConfig, err := amconfig.Load(config)
		if err != nil {
			continue
		}

		k.Metadata.SetAMConfig(key, amConfig)
	}
	return nil
}

func (k *k8sApi) GetAMConfigReceiver(configFile string, filter *request.AMConfigReceiverFilter, pageParam *request.PageParam) ([]amconfig.Receiver, int) {
	if len(configFile) == 0 {
		configFile = k.MetadataSettings.AlertManagerFileName
	}
	return k.Metadata.GetAMConfigReceiver(configFile, filter, pageParam)
}

func (k *k8sApi) AddOrUpdateAMConfigReceiver(configFile string, receiver amconfig.Receiver) error {
	if len(configFile) == 0 {
		configFile = k.MetadataSettings.AlertManagerFileName
	}

	err := k.Metadata.AddorUpdateAMConfigReceiver(configFile, receiver)
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
	isDeleted := k.Metadata.DeleteAMConfigReceiver(configFile, name)
	if !isDeleted {
		return nil
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
