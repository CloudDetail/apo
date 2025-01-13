// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package decoder

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"sort"
)

// fastAlertID 通过告警名和原始Tag计算AlertID
// 注意会跳过所有非String类型的Tag
func fastAlertID(alertName string, tags map[string]any) string {
	buf := new(bytes.Buffer)
	buf.WriteString(alertName)

	keys := make([]string, 0)
	for k, v := range tags {
		if _, ok := v.(string); ok {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	for _, key := range keys {
		buf.WriteString(tags[key].(string))
	}

	hash := md5.Sum(buf.Bytes())
	return fmt.Sprintf("%x", hash)
}
