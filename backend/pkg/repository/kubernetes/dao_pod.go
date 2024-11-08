package kubernetes

import (
	"context"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func (k *k8sApi) GetPodList(namespace string) (*v1.PodList, error) {
	list := &v1.PodList{}
	err := k.cli.List(context.Background(), list, &client.ListOptions{Namespace: namespace})
	if err != nil {
		return nil, err
	}
	return list, nil
}

func (k *k8sApi) GetPodInfo(namespace string, pod string) (*v1.Pod, error) {
	podInfo := &v1.Pod{}
	key := client.ObjectKey{
		Name:      pod,
		Namespace: namespace,
	}
	err := k.cli.Get(context.Background(), key, podInfo)
	if err != nil {
		k.logger.Error("Get pod error: ", zap.Error(err))
		return nil, err
	}

	return podInfo, nil
}
