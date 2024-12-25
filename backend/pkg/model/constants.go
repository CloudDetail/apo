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
