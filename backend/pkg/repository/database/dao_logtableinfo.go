// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import core "github.com/CloudDetail/apo/backend/pkg/core"

type LogTableInfo struct {
	ID           uint   `gorm:"primaryKey;autoIncrement"`
	DataBase     string `gorm:"type:varchar(255);column:database"`
	Table        string `gorm:"type:varchar(255);column:tablename"`
	Cluster      string `gorm:"type:varchar(255)"`
	Fields       string `gorm:"type:varchar(5000)"` // log field type
	ParseName    string `gorm:"type:varchar(255);column:parsename"`
	RouteRule    string `gorm:"type:varchar(1000);column:routerule"` // routing rule
	ParseRule    string `gorm:"type:varchar(5000);column:parserule"` // parsing rules
	ParseInfo    string `gorm:"type:varchar(255);column:parseinfo"`
	Service      string `gorm:"type:varchar(255)"`
	IsStructured bool   `gorm:"type:bool;column:structured"`
}

func (LogTableInfo) TableName() string {
	return "logtableinfo"
}

type Operator uint

const (
	INSERT Operator = iota
	QUERY
	UPDATE
	DELETE
)

func (repo *daoRepo) OperateLogTableInfo(ctx core.Context, model *LogTableInfo, op Operator) error {
	var err error
	switch op {
	case INSERT:
		err = repo.GetContextDB(ctx).Create(model).Error
	case QUERY:
		err = repo.GetContextDB(ctx).Where(`"database" = ? AND "tablename" = ?`, model.DataBase, model.Table).First(model).Error
	case UPDATE:
		err = repo.GetContextDB(ctx).Model(&LogTableInfo{}).Where(`"database" = ? AND "tablename" = ?`, model.DataBase, model.Table).Update("fields", model.Fields).Error
	case DELETE:
		return repo.GetContextDB(ctx).Where(`"database" = ? AND "tablename" = ?`, model.DataBase, model.Table).Delete(&LogTableInfo{}).Error
	}
	return err
}

func (repo *daoRepo) GetAllLogTable(ctx core.Context) ([]LogTableInfo, error) {
	var logTableInfo []LogTableInfo
	err := repo.GetContextDB(ctx).Find(&logTableInfo).Error
	return logTableInfo, err
}

func (repo *daoRepo) UpdateLogParseRule(ctx core.Context, model *LogTableInfo) error {
	return repo.GetContextDB(ctx).Model(&LogTableInfo{}).Where(`"database" = ? AND "tablename" = ?`, model.DataBase, model.Table).Updates(map[string]interface{}{
		"parseinfo":  model.ParseInfo,
		"parserule":  model.ParseRule,
		"routerule":  model.RouteRule,
		"service":    model.Service,
		"fields":     model.Fields,
		"structured": model.IsStructured,
	}).Error
}
