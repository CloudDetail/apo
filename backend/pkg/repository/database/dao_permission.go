package database

import "context"

// AuthPermission Records which feature are authorised to which subjects.
type AuthPermission struct {
	ID           int    `gorm:"primary_key;auto_increment" json:"id"`
	Type         string `gorm:"column:type;index:idx_sub_id_type" json:"type"`            // feature data
	SubjectID    int64  `gorm:"column:subject_id;index:idx_sub_id_type" json:"subjectId"` // Role id, user id or team id.
	SubjectType  string `gorm:"column:subject_type" json:"subjectType"`                   // role user team.
	PermissionID int    `gorm:"column:permission_id" json:"permissionId"`
}

func (t *AuthPermission) TableName() string {
	return "auth_permission"
}

func (repo *daoRepo) GetSubjectPermission(subID int64, subType string, typ string) ([]int, error) {
	var permissionIDs []int
	err := repo.db.Model(&AuthPermission{}).
		Select("permission_id").
		Where("subject_id = ? AND subject_type = ? AND type = ?", subID, subType, typ).
		Find(&permissionIDs).Error
	return permissionIDs, err
}

func (repo *daoRepo) GetSubjectsPermission(subIDs []int64, typ string) ([]AuthPermission, error) {
	var permissions []AuthPermission
	err := repo.db.Model(&AuthPermission{}).
		Where("subject_id in ? AND type = ?", subIDs, typ).
		Find(&permissions).Error
	return permissions, err
}

func (repo *daoRepo) GrantPermission(ctx context.Context, subID int64, subType string, typ string, permissionIDs []int) error {
	db := repo.GetContextDB(ctx)
	permission := make([]AuthPermission, len(permissionIDs))
	for i := range permissionIDs {
		permission[i] = AuthPermission{
			SubjectID:    subID,
			SubjectType:  subType,
			Type:         typ,
			PermissionID: permissionIDs[i],
		}
	}

	return db.Create(&permission).Error
}

func (repo *daoRepo) RevokePermission(ctx context.Context, subID int64, subType string, typ string, permissionIDs []int) error {
	return repo.GetContextDB(ctx).
		Model(&AuthPermission{}).
		Where("subject_id = ? AND subject_type = ? AND type = ? AND permission_id in ?", subID, subType, typ, permissionIDs).
		Delete(nil).
		Error
}
