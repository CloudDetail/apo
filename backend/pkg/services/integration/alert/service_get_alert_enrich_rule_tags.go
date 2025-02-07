// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import "github.com/CloudDetail/apo/backend/pkg/model/integration/alert"

func (s *service) GetAlertEnrichRuleTags() ([]alert.TargetTag, error) {
	return s.dbRepo.ListAlertTargetTags()
}
