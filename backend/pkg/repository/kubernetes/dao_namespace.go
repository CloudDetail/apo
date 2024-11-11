package kubernetes

import (
	"context"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (k *k8sApi) GetNamespaceList() (*v1.NamespaceList, error) {
	list := &v1.NamespaceList{}
	err := k.cli.List(context.Background(), list)
	if err != nil {
		k.logger.Error("Get namespace error: ", zap.Error(err))
		return nil, err
	}
	return list, nil
}

func (k *k8sApi) GetNamespaceInfo(namespace string) (*v1.Namespace, error) {
	namespaceInfo := &v1.Namespace{}
	err := k.cli.Get(context.Background(), client.ObjectKey{Name: namespace}, namespaceInfo)
	if err != nil {
		k.logger.Error("Get namespace error: ", zap.Error(err))
		return nil, err
	}
	return namespaceInfo, nil
}
