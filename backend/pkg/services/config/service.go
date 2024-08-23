package config

import (
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
	"github.com/CloudDetail/apo/backend/pkg/repository/clickhouse"
)

var _ Service = (*service)(nil)

type Service interface {
	SetTTL(req *request.SetTTLRequest) error
	SetSingleTableTTL(req *request.SetSingleTTLRequest) error
	GetTTL() (*response.GetTTLResponse, error)
}

type service struct {
	chRepo clickhouse.Repo
}

func New(chRepo clickhouse.Repo) Service {
	return &service{
		chRepo: chRepo,
	}
}
