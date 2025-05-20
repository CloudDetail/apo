// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"fmt"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"gorm.io/gorm"
)

func (repo *daoRepo) GetAMConfigReceiver(ctx core.Context, filter *request.AMConfigReceiverFilter, pageParam *request.PageParam) ([]amconfig.Receiver, int, error) {
	var count int64
	var countQuery = repo.GetContextDB(ctx).Model(&amconfig.Receiver{})
	if filter != nil && len(filter.Name) > 0 {
		countQuery.Where("name = ?", filter.Name)
	}
	err := countQuery.Count(&count).Error
	if err != nil || count == 0 {
		return nil, 0, err
	}

	var result []amconfig.Receiver
	query := repo.GetContextDB(ctx).Model(&amconfig.Receiver{})

	if filter != nil && len(filter.Name) > 0 {
		query.Where("name = ?", filter.Name)
	}

	if pageParam != nil {
		err = query.Limit(pageParam.PageSize).Offset((pageParam.CurrentPage - 1) * pageParam.PageSize).Find(&result).Error
	} else {
		err = query.Find(&result).Error
	}

	if err != nil {
		return []amconfig.Receiver{}, 0, err
	}
	return result, int(count), nil
}

func (repo *daoRepo) AddAMConfigReceiver(ctx core.Context, receiver amconfig.Receiver) error {
	if repo.CheckAMConfigReceiverExist(ctx, receiver.Name) {
		return fmt.Errorf("receiver name has been used: %s", receiver.Name)
	}
	return repo.GetContextDB(ctx).Create(receiver).Error
}

func (repo *daoRepo) UpdateAMConfigReceiver(ctx core.Context, receiver amconfig.Receiver, oldName string) error {
	if receiver.Name == oldName {
		return repo.GetContextDB(ctx).Model(&amconfig.Receiver{Name: oldName}).Updates(receiver).Error
	}

	if !repo.CheckAMConfigReceiverExist(ctx, oldName) {
		return fmt.Errorf("receiver not existed: %s", oldName)
	}

	if repo.CheckAMConfigReceiverExist(ctx, receiver.Name) {
		return fmt.Errorf("receiver name has been used: %s", receiver.Name)
	}

	return repo.GetContextDB(ctx).Transaction(func(tx *gorm.DB) error {
		err := repo.GetContextDB(ctx).Delete(&amconfig.Receiver{}, "name = ?", oldName).Error
		if err != nil {
			return err
		}
		return repo.GetContextDB(ctx).Create(receiver).Error
	})
}

func (repo *daoRepo) DeleteAMConfigReceiver(ctx core.Context, name string) error {
	return repo.GetContextDB(ctx).Delete(&amconfig.Receiver{}, "name = ?", name).Error
}

func (repo *daoRepo) CheckAMConfigReceiverExist(ctx core.Context, name string) bool {
	var count int64
	err := repo.GetContextDB(ctx).Model(&amconfig.Receiver{}).Where("name = ?", name).Count(&count).Error
	if err != nil {
		return false
	}
	return count > 0
}
