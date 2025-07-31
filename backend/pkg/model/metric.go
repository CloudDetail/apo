package model

type Pod struct {
	NodeName  string `json:"nodeName"`
	Namespace string `json:"namespace"`
	Pod       string `json:"pod"`
}
