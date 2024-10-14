package database

type OtherLogTable struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	DataBase  string `gorm:"type:varchar(100);column:database"`
	Table     string `gorm:"type:varchar(100);column:tablename"`
	Cluster   string `gorm:"type:varchar(100)"`
	TimeField string `gorm:"type:varchar(100);column:timefield"`
	LogField  string `gorm:"type:varchar(100);column:logfield"`
}

func (OtherLogTable) TableName() string {
	return "otherlogtable"
}

func (repo *daoRepo) GetAllOtherLogTable() ([]OtherLogTable, error) {
	var logTableInfo []OtherLogTable
	err := repo.db.Find(&logTableInfo).Error
	return logTableInfo, err
}
