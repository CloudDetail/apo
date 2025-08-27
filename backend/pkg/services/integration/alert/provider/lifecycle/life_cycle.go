// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package lifecycle

import (
	"sync"
	"time"

	"github.com/CloudDetail/apo/backend/pkg/model/integration/alert"
)

var AlertLifeCycle *AlertCache

func init() {
	// 初始化缓存
	AlertLifeCycle = NewAlertCache()
}

const maxCacheSize = 1e6
const cleanupInterval = 1 * time.Hour // Regular cleanup interval

// AlertCache records the first trigger time of alerts with the same ID
type AlertCache struct {
	sync.RWMutex
	records     map[string]time.Time // Key is alert ID, value is first trigger time
	order       []string             // Maintain insertion order of records
	lastCleanup time.Time            // Last cleanup time
	deletedKeys map[string]struct{}  // Mark deleted keys to avoid frequent memory copying
}

// NewAlertCache creates a new alert cache instance
func NewAlertCache() *AlertCache {
	return &AlertCache{
		records:     make(map[string]time.Time),
		order:       make([]string, 0),
		lastCleanup: time.Now(),
		deletedKeys: make(map[string]struct{}),
	}
}

func (ac *AlertCache) RecordEvent(alertID string, status string) (time.Time, bool) {
	ac.Lock()
	defer ac.Unlock()

	switch status {
	case alert.StatusFiring:
		if createTime, exists := ac.records[alertID]; exists {
			return createTime, true
		}
		// 检查是否需要清理
		ac.maybeCleanup()
		// 确保缓存大小不超过限制
		ac.ensureCacheSize()
		ac.records[alertID] = time.Now()
		ac.order = append(ac.order, alertID)
	case alert.StatusResolved:
		if createTime, exists := ac.records[alertID]; exists {
			// 标记为已删除
			ac.deletedKeys[alertID] = struct{}{}
			return createTime, true
		}
	}
	return time.Time{}, false
}

// maybeCleanup periodically cleans up deleted records and expired records
func (ac *AlertCache) maybeCleanup() {
	now := time.Now()
	// Check if cleanup interval has been reached
	if now.Sub(ac.lastCleanup) < cleanupInterval {
		return
	}

	// Clean up records marked as deleted
	for key := range ac.deletedKeys {
		delete(ac.records, key)
	}

	// Rebuild order slice, removing deleted elements
	if len(ac.deletedKeys) > 0 {
		newOrder := make([]string, 0, len(ac.order))
		for _, id := range ac.order {
			if _, deleted := ac.deletedKeys[id]; !deleted {
				newOrder = append(newOrder, id)
			}
		}
		ac.order = newOrder
	}

	// Clear deleted key markers
	clear(ac.deletedKeys)

	// Ensure cache size does not exceed limit
	ac.ensureCacheSize()

	ac.lastCleanup = now
}

// ensureCacheSize ensures the cache size does not exceed the limit
func (ac *AlertCache) ensureCacheSize() {
	// If the cache size exceeds the limit, remove the oldest records in batches
	// To reduce the number of array copies, remove more records at once
	const batchRemoveRatio = 0.1 // Batch removal ratio
	batchRemoveCount := int(float64(maxCacheSize) * batchRemoveRatio)
	if batchRemoveCount < 1 {
		batchRemoveCount = 1
	}

	for len(ac.records) >= maxCacheSize && len(ac.order) > 0 {
		// Calculate the number of records to remove this time
		removeCount := batchRemoveCount
		if removeCount > len(ac.order) {
			removeCount = len(ac.order)
		}

		// Remove the oldest records in batches
		oldestIDs := ac.order[:removeCount]
		for _, id := range oldestIDs {
			delete(ac.records, id)
		}

		// Update the order slice
		newOrder := make([]string, len(ac.order)-removeCount)
		copy(newOrder, ac.order[removeCount:])
		ac.order = newOrder
	}
}
