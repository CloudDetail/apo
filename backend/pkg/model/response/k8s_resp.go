package response

type GetNamespaceListResponse struct {
	NamespaceList string `json:"namespaceList"`
}

type GetPodListResponse struct {
	PodList string `json:"podList"`
}

type GetNamespaceInfoResponse struct {
	NamespaceInfo string `json:"namespaceInfo"`
}

type GetPodInfoResponse struct {
	PodInfo string `json:"podInfo"`
}
