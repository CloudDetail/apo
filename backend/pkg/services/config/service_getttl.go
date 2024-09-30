package config

import (
	"log"

	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetTTL() (*response.GetTTLResponse, error) {
	tables, err := s.chRepo.GetTables(model.GetAllTables())
	if err != nil {
		log.Println("[GetTTL] Error getting tables: ", err)
		return nil, err
	}
	tableInfo := prepareTTLInfo(tables)
	result := map[string][]model.ModifyTableTTLMap{
		"logs":     {},
		"trace":    {},
		"k8s":      {},
		"topology": {},
		"other":    {},
	}
	TableToType := model.TableToType()

	for _, item := range tableInfo {
		if typ, found := TableToType[item.Name]; found {
			result[typ] = append(result[typ], item)
		}
	}

	return &response.GetTTLResponse{
		Logs:     result["logs"],
		Trace:    result["trace"],
		K8s:      result["k8s"],
		Other:    result["other"],
		Topology: result["topology"],
	}, nil
}
