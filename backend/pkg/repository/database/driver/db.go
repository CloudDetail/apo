package driver

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"gorm.io/gorm"
)

const (
	_transactionCtxKey = "__transaction__"
)

type DB struct {
	*gorm.DB
}

func (d *DB) GetContextDB(ctx core.Context) *gorm.DB {
	if ctx == nil {
		return d.DB
	}

	ctxDB, exist := ctx.Get(_transactionCtxKey)
	if exist && ctxDB != nil {
		tx, ok := ctxDB.(*gorm.DB)
		if !ok {
			return nil
		}
		return tx
	}
	return d.WithContext(ctx.GetContext())
}

func (d *DB) WithTransaction(ctx core.Context, tx *gorm.DB) core.Context {
	ctx.Set(_transactionCtxKey, tx)
	return ctx
}
