package custom

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

var _ Service = (*service)(nil)

type Service interface {
	CreateCustomQuery(ctx core.Context, req *request.CreateCustomRequest) error
	GetCustomQuery(ctx core.Context, req *request.GetCustomRequest) (*response.GetCustomResponse, error)
	UpdateCustomQuery(ctx core.Context, req *request.UpdateCustomRequest) error
	DeleteCustomQuery(ctx core.Context, req *request.DeleteCustomRequest) error
}

type service struct {
	dbRepo database.Repo
}

func New(dbRepo database.Repo) Service {
	return &service{
		dbRepo: dbRepo,
	}
}
