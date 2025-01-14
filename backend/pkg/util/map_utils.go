// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package util

// MapKeysToArray takes a map and returns its keys as a slice.
func MapKeysToArray[K comparable, V any](inputMap map[K]V) []K {
	keys := make([]K, 0, len(inputMap))
	for key := range inputMap {
		keys = append(keys, key)
	}
	return keys
}
