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

var DefaultConfigMapKey = client.ObjectKey{
	Namespace: DefaultAPONS,
	Name:      DefaultCMNAME,
}

// GetAlertManagerRule implements Repo.
func (k *k8sApi) GetAlertManagerRule() (string, error) {
	obj := &v1.ConfigMap{}
	key := DefaultConfigMapKey
	err := k.cli.Get(context.Background(), key, obj)
	if err != nil {
		return "", err
	}

	return obj.Data["alert-rules.yaml"], nil
}

func (k *k8sApi) UpdateAlertManagerRule(alertRules string) error {
	obj := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      DefaultCMNAME,
			Namespace: DefaultAPONS,
		},
	}

	// TODO 处理未更新的情况
	_, err := controllerutil.CreateOrUpdate(context.Background(), k.cli, obj, func() error {
		obj.Data = map[string]string{
			"alert-rules.yaml": alertRules,
		}
		return nil
	})

	return err
}
