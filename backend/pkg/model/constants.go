// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package model

const (
	STATUS_NORMAL   = "normal"
	STATUS_WARNING  = "warning"
	STATUS_CRITICAL = "critical"
)

const (
	ROLE_ADMIN    = "admin"
	ROLE_MANAGER  = "manager"
	ROLE_VIEWER   = "viewer"
	ROLE_ANONYMOS = "anonymous"
)

const (
	PERMISSION_SUB_TYP_ROLE = "role"
	PERMISSION_SUB_TYP_USER = "user"
	PERMISSION_SUB_TYP_TEAM = "team"

	PERMISSION_TYP_FEATURE = "feature"
	PERMISSION_TYP_DATA    = "data"
)

const (
	TRANSLATION_EN          = "en"
	TRANSLATION_ZH          = "zh"
	TRANSLATION_TYP_FEATURE = "feature"
	TRANSLATION_TYP_MENU    = "menu"
)

const (
	MAPPED_TYP_MENU   = "menu"
	MAPPED_TYP_ROUTER = "router"
	MAPPED_TYP_API    = "api"
)

const (
	DATA_GROUP_SUB_TYP_USER    = "user"
	DATA_GROUP_SUB_TYP_TEAM    = "team"
	DATA_GROUP_SOURCE_DEFAULT  = "default"
	DATASOURCE_TYP_NAMESPACE   = "namespace"
	DATASOURCE_TYP_SERVICE     = "service"
	DATASOURCE_CATEGORY_APM    = "apm"
	DATASOURCE_CATEGORY_NORMAL = "normal"
)
