package custom

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
)

var _ Service = (*service)(nil)

type Service interface {
	CreateCustomQuery(req *request.CreateCustomRequest) error
	GetCustomQuery(req *request.GetCustomRequest) (*response.GetCustomResponse, error)
	UpdateCustomQuery(req *request.UpdateCustomRequest) error
	DeleteCustomQuery(req *request.DeleteCustomRequest) error
}

type service struct {
	dbRepo database.Repo
}

func New(dbRepo database.Repo) Service {
	return &service{
		dbRepo: dbRepo,
	}
}
