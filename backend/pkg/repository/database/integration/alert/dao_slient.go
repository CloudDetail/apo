package alert

import sc "github.com/CloudDetail/apo/backend/pkg/model/amconfig/slienceconfig"

func (repo *subRepo) GetAlertSlience() ([]sc.AlertSlienceConfig, error) {
	var result []sc.AlertSlienceConfig
	err := repo.db.Find(&result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (repo *subRepo) AddAlertSlience(SlienceConfig *sc.AlertSlienceConfig) error {
	return repo.db.Create(SlienceConfig).Error
}

func (repo *subRepo) UpdateAlertSlience(SlienceConfig *sc.AlertSlienceConfig) error {
	return repo.db.Where("alert_id = ?", SlienceConfig.AlertID).Updates(SlienceConfig).Error
}

func (repo *subRepo) DeleteAlertSlience(alertID string) error {
	return repo.db.Delete(&sc.AlertSlienceConfig{}, "alert_id = ? ", alertID).Error
}
