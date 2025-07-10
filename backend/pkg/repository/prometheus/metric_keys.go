// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

type EndpointKey struct {
	ContentKey string // URL
	SvcName    string // Name of the service to which the url belongs
}

type InstanceKey struct {
	ServiceName string `json:"service_name"`
	PID         string `json:"pid"`
	ContainerId string `json:"container_id"`
	Pod         string `json:"pod"`
	Namespace   string `json:"namespace"`
	NodeName    string `json:"node_name"`
	NodeIP      string `json:"node_ip"`
	ClusterID   string `json:"cluster_id"`
}

func (i InstanceKey) ConvertFromLabels(labels Labels) ConvertFromLabels {
	return InstanceKey{
		ServiceName: labels.SvcName,
		PID:         labels.PID,
		ContainerId: labels.ContainerID,
		Pod:         labels.POD,
		Namespace:   labels.Namespace,
		NodeName:    labels.NodeName,
		NodeIP:      labels.NodeIP,
		ClusterID:   labels.ClusterID,
	}
}

func (i InstanceKey) GenInstanceName() string {
	name := ""
	if len(i.Pod) > 0 {
		name = i.Pod
	} else if len(i.ContainerId) > 0 {
		name += i.ServiceName + "@" + i.NodeName + "@" + i.ContainerId
	} else if len(i.PID) > 0 {
		name += i.ServiceName + "@" + i.NodeName + "@" + i.PID
	}

	return name
}

func (e EndpointKey) ConvertFromLabels(labels Labels) ConvertFromLabels {
	return EndpointKey{
		SvcName:    labels.SvcName,
		ContentKey: labels.ContentKey,
	}
}

type SQLKey struct {
	Service string `json:"service"`
	// DBSystem -> ${SQL Type}, e.g: Mysql
	DBSystem string `json:"dbSystem"`
	// DBName -> ${database}
	DBName string `json:"dbName"`
	// DBOperation -> ${operation} ${table}, e.g: SELECT trip
	DBOperation string `json:"dbOperation"`
	DBUrl       string `json:"dbUrl"`
}

func (k SQLKey) ConvertFromLabels(labels Labels) ConvertFromLabels {
	return SQLKey{
		Service:     labels.SvcName,
		DBSystem:    labels.DBSystem,
		DBName:      labels.DBName,
		DBOperation: labels.Name,
		DBUrl:       labels.DBUrl,
	}
}

type ServiceKey struct {
	SvcName string // Name of the service to which the url belongs
}

func (S ServiceKey) ConvertFromLabels(labels Labels) ConvertFromLabels {
	return ServiceKey{
		SvcName: labels.SvcName,
	}
}
