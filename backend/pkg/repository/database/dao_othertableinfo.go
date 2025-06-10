// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import core "github.com/CloudDetail/apo/backend/pkg/core"

type OtherLogTable struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	DataBase  string `gorm:"type:varchar(255);column:database"`
	Table     string `gorm:"type:varchar(255);column:tablename"`
	Cluster   string `gorm:"type:varchar(255)"`
	TimeField string `gorm:"type:varchar(255);column:timefield"`
	LogField  string `gorm:"type:varchar(255);column:logfield"`
	Instance  string `gorm:"type:varchar(255)"`
}

func (OtherLogTable) TableName() string {
	return "otherlogtable"
}

func (repo *daoRepo) GetAllOtherLogTable(ctx core.Context) ([]OtherLogTable, error) {
	var logTableInfo []OtherLogTable
	err := repo.GetContextDB(ctx).Find(&logTableInfo).Error
	return logTableInfo, err
}

func (repo *daoRepo) OperatorOtherLogTable(ctx core.Context, model *OtherLogTable, op Operator) error {
	var err error
	switch op {
	case INSERT:
		err = repo.GetContextDB(ctx).Create(model).Error
	case QUERY:
		err = repo.GetContextDB(ctx).Where("database=? AND tablename=? And instance=?", model.DataBase, model.Table, model.Instance).First(model).Error
	case DELETE:
		err = repo.GetContextDB(ctx).Where("database=? AND tablename=? And instance=?", model.DataBase, model.Table, model.Instance).Delete(&OtherLogTable{}).Error
	}
	return err
}
