package response

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
)

type GetTTLResponse struct {
	Logs     []model.ModifyTableTTLMap `json:"logs"`
	Trace    []model.ModifyTableTTLMap `json:"trace"`
	K8s      []model.ModifyTableTTLMap `json:"k8s"`
	Topology []model.ModifyTableTTLMap `json:"topology"`
	Other    []model.ModifyTableTTLMap `json:"other"`
}
