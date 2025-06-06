// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package router

import (
	"github.com/CloudDetail/apo/backend/config"
	"github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/util"
	"github.com/CloudDetail/metadata/source"
)

type ExtraRouter func(mux *core.Mux, r *resource) error

var extraRouters = map[string]ExtraRouter{
	"metaserver": SetMetaServerRouter,
}

func SetMetaServerRouter(mux *core.Mux, _ *resource) error {
	if !config.Get().MetaServer.Enable {
		return nil
	}

	meta := source.CreateMetaSourceFromConfig(&config.Get().MetaServer.MetaSourceConfig)
	err := meta.Run()
	if err != nil {
		return err
	}

	api := mux.Group("/metadata")
	for path, handler := range meta.Handlers() {
		// This set of APIs supports both GET and POST
		api.POST_Gin(path, util.WrapHandlerFunctions(handler))
		api.GET_Gin(path, util.WrapHandlerFunctions(handler))
	}
	return nil
}
