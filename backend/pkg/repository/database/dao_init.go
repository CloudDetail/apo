package database

import "os"

// createMenuItems Auto migrate table and execute init sql script.
// Make sure sql script exists.
func (repo *daoRepo) initSql(model interface{}, sqlScript string) error {
	if err := repo.db.AutoMigrate(&model); err != nil {
		return err
	}

	var count int64
	if err := repo.db.Model(&model).Count(&count).Error; err != nil {
		return err
	}
	if count > 0 {
		return nil
	}

	if _, err := os.Stat(sqlScript); err == nil {
		sql, err := os.ReadFile(sqlScript)
		if err != nil {
			return err
		}
		if err := repo.db.Exec(string(sql)).Error; err != nil {
			return err
		}
	}
	return nil
}
