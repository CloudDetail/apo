package app

import "github.com/CloudDetail/apo/backend/pkg/core"

type Handler interface {
	// GetDeployMode Get deploy mode.
	// @Tags API.app
	// @Router /api/app/deployMode [get]
	GetDeployMode() core.HandlerFunc
}

type handler struct {
}

func New() Handler {
	return &handler{}
}
