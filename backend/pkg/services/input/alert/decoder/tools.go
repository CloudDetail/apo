// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package decoder

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"sort"
)

// calculate AlertID based on alertName and raw_tag
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
