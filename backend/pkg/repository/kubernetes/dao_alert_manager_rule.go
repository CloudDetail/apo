package kubernetes

import (
	"context"

	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const DefaultAPONS = "apo"
const DefaultCMNAME = "apo-victoria-metrics-alert-server-alert-rules-config"

var DefaultConfigMapKey = client.ObjectKey{
	Namespace: DefaultAPONS,
	Name:      DefaultCMNAME,
}

// GetAlertManagerRule implements Repo.
func (k *k8sApi) GetAlertManagerRule() error {

	obj := &v1.ConfigMap{}
	key := DefaultConfigMapKey
	k.cli.Get(context.Background(), key, obj)

	panic("unimplemented")
}

// UpdateAlertManagerRule implements Repo.
func (k *k8sApi) UpdateAlertManagerRule() error {
	panic("unimplemented")
}
