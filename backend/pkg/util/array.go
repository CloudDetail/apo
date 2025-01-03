// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package util

func ContainsStr(arr []string, target string) bool {
	for _, value := range arr {
		if value == target {
			return true
		}
	}
	return false
}
