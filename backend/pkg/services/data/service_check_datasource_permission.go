// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"fmt"

	core "github.com/CloudDetail/apo/backend/pkg/core"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/model/response"
)

func (s *service) CheckDatasourcePermission(ctx core.Context, userID, groupID int64, namespaces, services interface{}, fillCategory string) (err error) {
	// TODO
	return nil
}

func toStringSlice(input interface{}) ([]string, error) {
	if input == nil {
		return nil, nil
	}
	switch v := input.(type) {
	case *string:
		if len(*input.(*string)) == 0 {
			return nil, nil
		}
		return []string{*v}, nil
	case *[]string:
		if len(*input.(*[]string)) == 0 {
			return nil, nil
		}
		return *v, nil
	default:
		return nil, fmt.Errorf("unsupported type: %T, please use *string or *[]string", input)
	}
}

func setInterface(dest interface{}, value []string) {
	if dest == nil {
		return
	}
	switch v := dest.(type) {
	case *string:
		if len(value) > 0 {
			*v = value[0]
		}
	case *[]string:
		*v = value
	}
}

func inAllDatasource(all response.GetDatasourceResponse, datasource string, typ string) bool {
	switch typ {
	case model.DATASOURCE_TYP_NAMESPACE:
		for _, namespace := range all.NamespaceList {
			if datasource == namespace.Datasource {
				return true
			}
		}
	case model.DATASOURCE_TYP_SERVICE:
		for _, service := range all.ServiceList {
			if datasource == service.Datasource {
				return true
			}
		}
	}

	return false
}
