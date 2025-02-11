// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"context"
	"fmt"
	"time"
)

const (
	TEMPLATE_GET_NODES = `count(originx_agent_ready{}[%s])`
)

func (repo *promRepo) GetNodesCount(startTime int64, endTime int64) (int, error) {
	query := fmt.Sprintf(TEMPLATE_GET_NODES, VecFromS2E(startTime, endTime))
	value, _, err := repo.GetApi().Query(context.Background(), query, time.UnixMicro(endTime))

	if err != nil {
		return 0, err
	}

	vector, err := value.(Vec)
	// TODO return count
	
}
