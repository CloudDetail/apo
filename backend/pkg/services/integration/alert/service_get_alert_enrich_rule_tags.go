// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
	"github.com/CloudDetail/apo/backend/pkg/model/request"
)

func (s *service) GetAlertEnrichRuleTags(req *request.ListTargetTagsRequest) ([]alert.TargetTag, error) {
	return s.dbRepo.ListAlertTargetTags(req.Language)
}
