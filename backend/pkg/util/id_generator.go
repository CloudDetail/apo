// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package util

import (
	"sync"
	"time"
)

const (
	epoch        = int64(1726588800000)
	sequenceBits = 5
	maxSequence  = -1 ^ (-1 << sequenceBits)
	timeShift    = sequenceBits
)

type IDGenerator struct {
	mu       sync.Mutex
	lastTime int64
	sequence int64
}

var Generator = &IDGenerator{}

func (g *IDGenerator) GenerateID() int64 {
	g.mu.Lock()
	defer g.mu.Unlock()

	now := time.Now().UnixMilli()

	if now == g.lastTime {
		g.sequence = (g.sequence + 1) & maxSequence
		if g.sequence == 0 {
			for now <= g.lastTime {
				now = time.Now().UnixMilli()
			}
		}
	} else {
		g.sequence = 0
	}

	g.lastTime = now

	id := ((now - epoch) << timeShift) | g.sequence
	return id
}
