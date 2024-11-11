package database

type LogTableInfo struct {
	ID        uint   `gorm:"primaryKey;autoIncrement"`
	DataBase  string `gorm:"type:varchar(100);column:database"`
	Table     string `gorm:"type:varchar(100);column:tablename"`
	Cluster   string `gorm:"type:varchar(100)"`
	Fields    string `gorm:"type:varchar(5000)"` // 日志字段类型
	ParseName string `gorm:"type:varchar(100);column:parsename"`
	RouteRule string `gorm:"type:varchar(1000);column:routerule"` // 路由规则
	ParseRule string `gorm:"type:varchar(5000);column:parserule"` // 解析规则
	ParseInfo string `gorm:"type:varchar(100);column:parseinfo"`
	Service   string `gorm:"type:varchar(100)"`
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
		err = repo.db.Where("`database` = ? AND `tablename` = ?", model.DataBase, model.Table).First(model).Error
	case UPDATE:
		err = repo.db.Model(&LogTableInfo{}).Where("`database` = ? AND `tablename` = ?", model.DataBase, model.Table).Update("fields", model.Fields).Error
	case DELETE:
		return repo.db.Where("`database` = ? AND `tablename` = ?", model.DataBase, model.Table).Delete(&LogTableInfo{}).Error
	}
	return err
}

func (repo *daoRepo) GetAllLogTable() ([]LogTableInfo, error) {
	var logTableInfo []LogTableInfo
	err := repo.db.Find(&logTableInfo).Error
	return logTableInfo, err
}

func (repo *daoRepo) UpdateLogParseRule(model *LogTableInfo) error {
	return repo.db.Model(&LogTableInfo{}).Where("`database` = ? AND `tablename` = ?", model.DataBase, model.Table).Updates(LogTableInfo{
		ParseInfo: model.ParseInfo,
		ParseRule: model.ParseRule,
		RouteRule: model.RouteRule,
		Service:   model.Service,
		Fields:    model.Fields,
	}).Error
}
