// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package database

import (
	core "github.com/CloudDetail/apo/backend/pkg/core"
	"gorm.io/gorm"
)

func (repo *daoRepo) GetContextDB(ctx core.Context) *gorm.DB {
	ctxDB, exist := ctx.Get(_transactionCtxKey)
	if exist && ctxDB != nil {
		tx, ok := ctxDB.(*gorm.DB)
		if !ok {
			return nil
		}
		return tx
	}
	return repo.db.WithContext(ctx.GetContext())
}

func (repo *daoRepo) WithTransaction(ctx core.Context, tx *gorm.DB) core.Context {
	ctx.Set(_transactionCtxKey, tx)
	return ctx
}

func (repo *daoRepo) Transaction(ctx core.Context, funcs ...func(txCtx core.Context) error) (err error) {
	tx := repo.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	txCtx := repo.WithTransaction(ctx, tx)
	for _, f := range funcs {
		if err = f(txCtx); err != nil {
			tx.Rollback()
			return
		}
	}
	if err = tx.Commit().Error; err != nil {
		tx.Rollback()
	}
	return
}
