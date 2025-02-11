// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"sync"
	"time"
)

type Repo interface {
	AddToken(token string) error
	IsInBlacklist(token string) (bool, error)
}

type cache struct {
	blackList sync.Map
}

func (c *cache) refreshLoop() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for range ticker.C {
		c.refreshTokenCache()
	}
}

func New() (Repo, error) {
	c := cache{
		blackList: sync.Map{},
	}
	go c.refreshLoop()
	return &c, nil
}
