package kubernetes

import (
	"fmt"

	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (k *k8sApi) syncAMConfig() error {
	res, err := k.GetAlertManagerConfigFile("")
	if err != nil {
		return err
	}
	for key, config := range res {
		amConfig, err := ParseAlertManagerConfig(config)
		if err != nil {
			continue
		}

		k.Metadata.SetAMConfig(key, amConfig)
	}
	return nil
}

func (k *k8sApi) GetAMConfigReceiver(configFile string, filter *request.AMConfigReceiverFilter, pageParam *request.PageParam) ([]*request.AMConfigReceiver, int) {
	if len(configFile) == 0 {
		configFile = k.MetadataSettings.AlertManagerFileName
	}
	return k.Metadata.GetAMConfigReceiver(configFile, filter, pageParam)
}

func (k *k8sApi) AddOrUpdateAMConfigReceiver(configFile string, receiver request.AMConfigReceiver) error {
	if len(configFile) == 0 {
		configFile = k.MetadataSettings.AlertManagerFileName
	}

	// check Before update
	if rDef := receiver.ToReceiverDef(); rDef == nil {
		return fmt.Errorf("dont't support to add receiver as %s now", receiver.RType)
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
