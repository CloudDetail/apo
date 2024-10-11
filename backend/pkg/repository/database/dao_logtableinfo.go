package database

type LogTableInfo struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	DataBase  string `gorm:"type:varchar(100);column:database"`
	Table     string `gorm:"type:varchar(100);column:tablename"`
	Cluster   string `gorm:"type:varchar(100)"`
	Fields    string `gorm:"type:varchar(100)"`
	ParseName string `gorm:"type:varchar(100);column:parsename"`
	RouteRule string `gorm:"type:varchar(100);column:routerule"`
	ParseRule string `gorm:"type:varchar(100);column:parserule"`
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

func (repo *daoRepo) OperateLogTableInfo(model *LogTableInfo, op Operator) error {
	var err error
	switch op {
	case INSERT:
		err = repo.db.Create(model).Error
	case QUERY:
		err = repo.db.Where("database=? AND tablename=?", model.DataBase, model.Table).First(model).Error
	case UPDATE:
		err = repo.db.Model(&LogTableInfo{}).Where("database=? AND tablename=?", model.DataBase, model.Table).Update("fields", model.Fields).Error
	case DELETE:
		return repo.db.Where("database=? AND tablename=?", model.DataBase, model.Table).Delete(&LogTableInfo{}).Error
	}
	return err
}

func (repo *daoRepo) GetAllLogTable() ([]LogTableInfo, error) {
	var logTableInfo []LogTableInfo
	err := repo.db.Find(&logTableInfo).Error
	return logTableInfo, err
}
