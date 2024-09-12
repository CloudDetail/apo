package kubernetes

import (
	"context"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const DefaultAPONS = "apo"
const DefaultCMNAME = "apo-victoria-metrics-alert-server-alert-rules-config"
const DefaultAlertRuleFile = "alert-rules.yaml"

var DefaultConfigMapKey = client.ObjectKey{
	Namespace: DefaultAPONS,
	Name:      DefaultCMNAME,
}

// GetAlertManagerRule implements Repo.
func (k *k8sApi) GetAlertManagerRule(alertRuleFile string) (map[string]string, error) {
	obj := &v1.ConfigMap{}
	key := DefaultConfigMapKey
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

func (k *k8sApi) UpdateAlertManagerRule(alertRules map[string]string) error {
	obj := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultCMNAME,
			Namespace: DefaultAPONS,
		},
	}

	// TODO 处理未更新的情况
	_, err := controllerutil.CreateOrUpdate(context.Background(), k.cli, obj, func() error {
		for k, v := range alertRules {
			if len(v) > 0 {
				obj.Data[k] = v
			} else {
				delete(obj.Data, k)
			}
		}
		return nil
	})
	return err
}
