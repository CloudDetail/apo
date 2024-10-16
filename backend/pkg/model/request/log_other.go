package request

import "github.com/CloudDetail/apo/backend/config"

type OtherTableRequest struct {
}

type OtherTableInfoRequest struct {
	DataBase  string `form:"dataBase" json:"dataBase"`
	TableName string `form:"tableName" json:"tableName"`
}

type AddOtherTableRequest struct {
	DataBase  string `json:"dataBase"`
	Table     string `json:"tableName"`
	Cluster   string `json:"cluster"`
	Instance  string `json:"instance"`
	TimeField string `json:"timeField"`
	LogField  string `json:"logField"`
}

func (req *AddOtherTableRequest) FillerValue() {
	if req.Cluster == "" {
		req.Cluster = config.Get().ClickHouse.Cluster
	}
	if req.DataBase == "" {
		req.DataBase = "apo"
	}
	if req.Instance == "" {
		req.Instance = "default"
	}
}

type DeleteOtherTableRequest struct {
	DataBase  string `json:"dataBase"`
	TableName string `json:"tableName"`
	Instance  string `json:"instance"`
}
