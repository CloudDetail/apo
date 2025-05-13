// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package amconfig

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
)

type EmailConfigs []*EmailConfig
type WebhookConfigs []*WebhookConfig
type DingTalkConfigs []*DingTalkConfig
type WechatConfigs []*WechatConfig

func (e EmailConfigs) Value() (driver.Value, error) {
	if e == nil {
		return "", nil
	}
	val, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return string(val), nil
}

func (e *EmailConfigs) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	var val []byte
	switch s := value.(type) {
	case string:
		val = []byte(s)
	case []byte:
		val = s
	default:
		return fmt.Errorf("failed to scan JSONField, expected string, got %T", value)
	}
	return json.Unmarshal(val, e)
}

func (e WebhookConfigs) Value() (driver.Value, error) {
	if e == nil {
		return "", nil
	}
	val, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return string(val), nil
}

func (e *WebhookConfigs) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	var val []byte
	switch s := value.(type) {
	case string:
		val = []byte(s)
	case []byte:
		val = s
	default:
		return fmt.Errorf("failed to scan JSONField, expected string, got %T", value)
	}
	return json.Unmarshal(val, e)
}

func (e DingTalkConfigs) Value() (driver.Value, error) {
	if e == nil {
		return "", nil
	}
	val, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return string(val), nil
}

func (e *DingTalkConfigs) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	var val []byte
	switch s := value.(type) {
	case string:
		val = []byte(s)
	case []byte:
		val = s
	default:
		return fmt.Errorf("failed to scan JSONField, expected string, got %T", value)
	}
	if len(val) == 0 {
		return nil
	}
	return json.Unmarshal(val, e)
}

func (e WechatConfigs) Value() (driver.Value, error) {
	if e == nil {
		return "", nil
	}
	val, err := json.Marshal(e)
	if err != nil {
		return nil, err
	}
	return string(val), nil
}

func (e *WechatConfigs) Scan(value interface{}) error {
	if value == nil {
		return nil
	}
	var val []byte
	switch s := value.(type) {
	case string:
		val = []byte(s)
	case []byte:
		val = s
	default:
		return fmt.Errorf("failed to scan JSONField, expected string, got %T", value)
	}
	if len(val) == 0 {
		return nil
	}
	return json.Unmarshal(val, e)
}
