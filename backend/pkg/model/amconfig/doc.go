// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package amconfig

// Configuration file for parsing and deleting alertmanager
// Body from github.com/prometheus/alertmanager@v0.27.0/config

/**
修改内容:
- 更改SecretURL为 URL, 避免在Marshal/Unmarshal中丢失信息
- 修改所有配置的JSON tag为形式小驼峰
- 移除在Validate中将不同验证信息转移到Header的逻辑,只做检查
- 修改EmailConfigs/WebhookConfigs外的其他告警方式的JSON tag为隐藏,后续逐步开放
**/
