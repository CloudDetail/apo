package response

import v1 "k8s.io/api/core/v1"

type GetNamespaceListResponse struct {
	*v1.NamespaceList
}

type GetPodListResponse struct {
	*v1.PodList
}

type GetNamespaceInfoResponse struct {
	*v1.Namespace
}

type GetPodInfoResponse struct {
	*v1.Pod
}
