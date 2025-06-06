// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"gorm.io/gorm"
)

func (repo *daoRepo) CheckAMReceiverCount(ctx core.Context) int64 {
	var count int64
	err := repo.GetContextDB(ctx).Model(&amconfig.Receiver{}).Count(&count).Error
	if err != nil {
		return -1
	}
	return count
}

func (repo *daoRepo) MigrateAMReceiver(ctx core.Context, receivers []amconfig.Receiver) ([]amconfig.Receiver, error) {
	extraReceiver := skipAPOReceiver(receivers)
	err := repo.GetContextDB(ctx).Transaction(func(tx *gorm.DB) error {
		// TODO Read Dingtalk config from Database, transform into xxx
		if err := tx.AutoMigrate(&amconfig.Receiver{}); err != nil {
			return err
		}

		if len(extraReceiver) == 0 {
			return nil
		}
		return tx.CreateInBatches(extraReceiver, 100).Error
	})

	return extraReceiver, err
}

func skipAPOReceiver(receivers []amconfig.Receiver) []amconfig.Receiver {
	var res []amconfig.Receiver
	for i := 0; i < len(receivers); i++ {
		if receivers[i].Name == "APO Mutation Check" || receivers[i].Name == "APO Alert Collector" {
			continue
		}
		res = append(res, receivers[i])
	}
	return res
}
