// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package response

type LogTableInfoResponse struct {
	Parses    []Parse    `json:"parses"`
	Instances []Instance `json:"instances"`
}

type Parse struct {
	DataBase  string `json:"dataBase"`
	TableName string `json:"tableName"`
	ParseName string `json:"parseName"`
	ParseInfo string `json:"parseInfo"`
}

type Instance struct {
	InstanceName string   `json:"instanceName"`
	DataBases    []DBInfo `json:"dataBases"`
}

type DBInfo struct {
	DataBase string         `json:"dataBase"`
	Tables   []LogTableInfo `json:"tables"`
}

type LogTableInfo struct {
	Cluster   string `json:"cluster"`
	TableName string `json:"tableName"`
	TimeField string `json:"timeField"`
	LogField  string `json:"logField"`
}
