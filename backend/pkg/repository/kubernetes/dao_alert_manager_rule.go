package kubernetes

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// GetAlertRuleConfigFile implements Repo.
func (k *k8sApi) GetAlertRuleConfigFile(alertRuleFile string) (map[string]string, error) {
	obj := &v1.ConfigMap{}
	key := client.ObjectKey{
		Namespace: k.Namespace,
		Name:      k.AlertRuleCMName,
	}
	err := k.cli.Get(context.Background(), key, obj)
	if err != nil {
		return nil, err
	}

	if len(alertRuleFile) > 0 {
		return map[string]string{
			alertRuleFile: obj.Data[alertRuleFile],
		}, nil
	}

	return obj.Data, nil
}

func (k *k8sApi) UpdateAlertRuleConfigFile(configFile string, content []byte) error {
	obj := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      k.AlertRuleCMName,
			Namespace: k.Namespace,
		},
	}

	_, err := controllerutil.CreateOrUpdate(context.Background(), k.cli, obj, func() error {
		if content == nil {
			delete(obj.Data, configFile)
		} else {
			obj.Data[configFile] = string(content)
		}
		return nil
	})
	return err
}
