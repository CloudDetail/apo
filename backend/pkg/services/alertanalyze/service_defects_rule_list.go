package alertanalyze

import (
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) GetDefectDetectRuleList(req *request.GetDefectDetectRuleListRequest) (*response.GetDefectDetectRuleListResponse, error) {
	list, totalCount, err := s.chRepo.GetDistinctDetectExecList(req)
	if err != nil {
		return nil, err
	}
	filter := &request.AlertRuleFilter{
		Group: "异常检测",
	}
	rules, _ := s.k8sRepo.GetAlertRules("", filter, nil, false)
	for i := range list {
		for j := range rules {
			if list[i].Group == rules[j].Group && list[i].DetectName == rules[j].Alert && list[i].MutationCheck == rules[j].Expr {
				list[i].SynchronizeToAlertRules = true
				break
			}
		}
	}
	resp := &response.GetDefectDetectRuleListResponse{
		Pagination: model.Pagination{
			CurrentPage: req.CurrentPage,
			Total:       totalCount,
			PageSize:    req.PageSize,
		},
		AlertRule: list,
	}

	return resp, nil
}
