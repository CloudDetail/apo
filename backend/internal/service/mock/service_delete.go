// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package mock

import (
	"github.com/CloudDetail/apo/backend/internal/model/request"
)

func (s *service) Delete(req *request.DeleteRequest) error {
	return s.dbRepo.DeleteMockById(req.Id)
}
