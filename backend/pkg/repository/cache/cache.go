// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"sync"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/util/jwt"
	core "github.com/CloudDetail/apo/backend/pkg/core"
)

type Repo interface {
	AddToken(ctx_core core.Context, token string) error
	IsInBlacklist(ctx_core core.Context, token string) (bool, error)
}

type cache struct {
	blackList sync.Map
}

func (c *cache) IsInBlacklist(ctx_core core.Context, token string) (bool, error) {
	_, ok := c.blackList.Load(token)
	return ok, nil
}

func (c *cache) AddToken(ctx_core core.Context, token string) error {
	c.blackList.Store(token, struct{}{})
	return nil
}

func (c *cache) refreshLoop() {
	ticker := time.NewTicker(time.Second * 10)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.refreshCache()
		}
	}
}

func (c *cache) refreshCache() {
	c.blackList.Range(func(token, _ any) bool {
		if jwt.IsExpire(token.(string)) {
			c.blackList.Delete(token)
		}
		return true
	})
}

func New() (Repo, error) {
	c := cache{
		blackList: sync.Map{},
	}
	go c.refreshLoop()
	return &c, nil
}
