// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package util

func RoundDownToMinute(timestamp int64) int64 {
	const minuteMicros = int64(60 * 1e6)
	return timestamp - timestamp%minuteMicros
}
