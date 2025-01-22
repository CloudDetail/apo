package database

import (
	"context"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// DataGroup is a collection of Datasource.
type DataGroup struct {
	GroupID     int64  `gorm:"column:group_id;primary_key;auto_increment" json:"groupId"`
	GroupName   string `gorm:"column:group_name;type:varchar(20)" json:"groupName"`
	Description string `gorm:"column:description;type:varchar(50)" json:"description"` // The description of data group.

	DatasourceList []DatasourceGroup `gorm:"foreignKey:GroupID;references:GroupID" json:"datasourceList"`
	AuthType       string            `json:"authType,omitempty"`
	Source         string            `gorm:"-" json:"source,omitempty"`
}

// DatasourceGroup is a mapping table of Datasource and DataGroup.
type DatasourceGroup struct {
	GroupID    int64  `gorm:"column:group_id;primary_key" json:"-"`
	Datasource string `gorm:"column:datasource;primary_key" json:"datasource"`
	Type       string `gorm:"column:type" json:"type"`         // service or namespace
	Category   string `gorm:"column:category" json:"category"` // apm or normal
}

func (dg *DataGroup) TableName() string {
	return "data_group"
}

func (dsg *DatasourceGroup) TableName() string {
	return "datasource_group"
}

func (repo *daoRepo) CreateDataGroup(ctx context.Context, group *DataGroup) error {
	return repo.GetContextDB(ctx).Create(group).Error
}

func (repo *daoRepo) DeleteDataGroup(ctx context.Context, groupID int64) error {
	group := DataGroup{
		GroupID: groupID,
	}
	if err := repo.GetContextDB(ctx).Delete(&group).Error; err != nil {
		return err
	}

	return repo.GetContextDB(ctx).Model(&AuthDataGroup{}).Where("data_group_id = ?", groupID).Delete(nil).Error
}

func (repo *daoRepo) CreateDatasourceGroup(ctx context.Context, datasource []model.Datasource, dataGroupID int64) error {
	if len(datasource) == 0 {
		return nil
	}

	datasourceGroups := make([]DatasourceGroup, len(datasource))
	for _, ds := range datasource {
		dsGroup := DatasourceGroup{
			GroupID:    dataGroupID,
			Datasource: ds.Datasource,
			Type:       ds.Type,
			Category:   ds.Category,
		}
		datasourceGroups = append(datasourceGroups, dsGroup)
	}

	return repo.GetContextDB(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "group_id"}, {Name: "datasource"}},
		DoNothing: true,
	}).Create(&datasourceGroups).Error
}

func (repo *daoRepo) DeleteDSGroup(ctx context.Context, groupID int64) error {
	return repo.GetContextDB(ctx).Model(&DatasourceGroup{}).Where("group_id = ?", groupID).Delete(&DatasourceGroup{}).Error
}

func (repo *daoRepo) UpdateDataGroup(ctx context.Context, groupID int64, groupName string, description string) error {
	return repo.GetContextDB(ctx).
		Model(&DataGroup{}).
		Where("group_id = ?", groupID).
		Update("group_name", groupName).
		Update("description", description).
		Error
}

// DataGroupExist check whether the group exists for the given condition.
func (repo *daoRepo) DataGroupExist(filter model.DataGroupFilter) (bool, error) {
	var count int64
	query := repo.db
	if len(filter.Name) > 0 {
		query = query.Where("group_name = ?", filter.Name)
	}
	if filter.ID > 0 {
		query = query.Where("group_id = ?", filter.ID)
	}
	if len(filter.IDs) > 0 {
		query = query.Where("group_id IN ?", filter.IDs)
	}
	if err := query.Model(&DataGroup{}).Count(&count).Error; err != nil {
		return false, err
	}

	return count > 0, nil
}

func (repo *daoRepo) GetDataGroup(filter model.DataGroupFilter) ([]DataGroup, int64, error) {
	var (
		dataGroups []DataGroup
		count      int64
	)
	query := repo.db
	if len(filter.Name) > 0 {
		query = query.Where("group_name like ?", "%"+filter.Name+"%")
	}
	if len(filter.Names) > 0 {
		query = query.Where("group_name in ?", filter.Names)
	}
	if filter.ID != 0 {
		query = query.Where("group_id = ?", filter.ID)
	}
	if len(filter.IDs) > 0 {
		query = query.Where("group_id in ?", filter.IDs)
	}
	if len(filter.DatasourceList) > 0 {
		conditions := make([][]interface{}, 0, len(filter.DatasourceList))
		for _, item := range filter.DatasourceList {
			conditions = append(conditions, []interface{}{item.Datasource, item.Type})
		}

		subQuery := repo.db.Model(&DatasourceGroup{}).
			Select("group_id").
			Where("(datasource, type) IN ?", conditions)

		query.Where("group_id IN (?)", subQuery)
	}

	if err := query.Model(&DataGroup{}).Count(&count).Error; err != nil {
		return nil, 0, err
	}

	if filter.CurrentPage != nil && filter.PageSize != nil {
		query = query.Offset((*filter.CurrentPage - 1) * (*filter.PageSize)).Limit(*filter.PageSize)
	}

	err := query.Preload("DatasourceList").Find(&dataGroups).Error
	return dataGroups, count, err
}

func (repo *daoRepo) RetrieveDataFromGroup(ctx context.Context, groupID int64, datasource []string) error {
	return repo.GetContextDB(ctx).
		Model(&DatasourceGroup{}).Where("group_id = ? AND datasource in ?", groupID, datasource).Delete(nil).Error
}

func (repo *daoRepo) GetGroupDatasource(groupID ...int64) ([]DatasourceGroup, error) {
	var dsGroup []DatasourceGroup
	err := repo.db.Where("group_id in ?", groupID).Find(&dsGroup).Error
	return dsGroup, err
}

func (repo *daoRepo) GetSubjectDataGroupList(subjectID int64, subjectType string, category string) ([]DataGroup, error) {
	var dataGroups []DataGroup

	preloadQuery := func(db *gorm.DB) *gorm.DB {
		if len(category) > 0 {
			return db.Where("category = ?", category)
		}
		return db
	}

	err := repo.db.Table("data_group").
		Preload("DatasourceList", preloadQuery).
		Select("data_group.group_id, data_group.group_name, data_group.description, auth_data_group.type as auth_type").
		Joins("JOIN auth_data_group ON auth_data_group.data_group_id = data_group.group_id").
		Where("auth_data_group.subject_id = ? AND auth_data_group.subject_type = ?", subjectID, subjectType).
		Find(&dataGroups).Error

	if err != nil {
		return nil, err
	}

	return dataGroups, nil
}
