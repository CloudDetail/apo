// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package utils

type Filter interface {
	ExtractFilterStr() []string
}
