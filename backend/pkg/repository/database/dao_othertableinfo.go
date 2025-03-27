// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

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

func (repo *daoRepo) GetAllOtherLogTable() ([]OtherLogTable, error) {
	var logTableInfo []OtherLogTable
	err := repo.db.Find(&logTableInfo).Error
	return logTableInfo, err
}

func (repo *daoRepo) OperatorOtherLogTable(model *OtherLogTable, op Operator) error {
	var err error
	switch op {
	case INSERT:
		err = repo.db.Create(model).Error
	case QUERY:
		err = repo.db.Where("database=? AND tablename=? And instance=?", model.DataBase, model.Table, model.Instance).First(model).Error
	case DELETE:
		err = repo.db.Where("database=? AND tablename=? And instance=?", model.DataBase, model.Table, model.Instance).Delete(&OtherLogTable{}).Error
	}
	return err
}
