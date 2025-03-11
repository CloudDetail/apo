// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

func (s *service) GetAlertEnrichRuleTags(ctx core.Context) ([]alert.TargetTag, error) {
	return s.dbRepo.ListAlertTargetTags(ctx.LANG())
}
