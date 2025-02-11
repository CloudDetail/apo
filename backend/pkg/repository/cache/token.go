// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package cache

import (
	"github.com/CloudDetail/apo/backend/pkg/util"
)

func (c *cache) IsInBlacklist(token string) (bool, error) {
	_, ok := c.blackList.Load(token)
	return ok, nil
}

func (c *cache) AddToken(token string) error {
	c.blackList.Store(token, struct{}{})
	return nil
}

func (c *cache) refreshTokenCache() {
	c.blackList.Range(func(token, _ any) bool {
		if util.IsExpire(token.(string)) {
			c.blackList.Delete(token)
		}
		return true
	})
}
