package database

import (
	"context"
	"gorm.io/gorm"
)

func (repo *daoRepo) GetContextDB(ctx context.Context) *gorm.DB {
	ctxDB := ctx.Value(repo.transactionCtx)

	if ctxDB != nil {
		tx, ok := ctxDB.(*gorm.DB)
		if !ok {
			return nil
		}
		return tx
	}
	return repo.db.WithContext(ctx)
}

func (repo *daoRepo) WithTransaction(ctx context.Context, tx *gorm.DB) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}
	return context.WithValue(ctx, repo.transactionCtx, tx)
}

func (repo *daoRepo) Transaction(ctx context.Context, funcs ...func(txCtx context.Context) error) (err error) {
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
