// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"errors"
	"fmt"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/util"
	"time"
)

func (s *service) CheckDatasourcePermission(userID int64, namespaces, services interface{}) (err error) {
	var (
		namespaceMap = map[string]struct{}{}
		serviceMap   = map[string]struct{}{}
		endTime      = time.Now()
		startTime    = endTime.Add(-24 * time.Hour)
		serviceList  []string
	)

	// Get user's data group
	groups, err := s.getUserDataGroup(userID, "")
	if err != nil {
		return err
	}

	for _, group := range groups {
		for _, gs := range group.DatasourceList {
			ds := gs.Datasource
			if gs.Type == model.DATASOURCE_TYP_NAMESPACE {
				namespaceMap[ds] = struct{}{}
			} else if gs.Type == model.DATASOURCE_TYP_SERVICE {
				serviceMap[ds] = struct{}{}
			}
		}
	}

	namespaceDs := util.MapKeysToArray[string, struct{}](namespaceMap)

	namespacesSlice, err := toStringSlice(namespaces)
	if err != nil {
		return err
	}
	servicesSlice, err := toStringSlice(services)
	if err != nil {
		return err
	}

	if len(namespaceDs) > 0 {
		serviceList, err = s.promRepo.GetServiceList(startTime.UnixMicro(), endTime.UnixMicro(), namespaceDs)
		if err != nil {
			return err
		}
	}

	for _, srv := range serviceList {
		serviceMap[srv] = struct{}{}
	}

	serviceDs := util.MapKeysToArray[string, struct{}](serviceMap)

	// The request didn't offer parameters.
	// Fill with datasource which user is authorized to view.
	if len(namespacesSlice) == 0 && len(servicesSlice) == 0 {
		if len(namespaceDs) > 0 {
			setInterface(namespaces, namespacesSlice)
		} else if len(serviceDs) > 0 {
			setInterface(services, serviceDs)
		}
		return nil
	}

	filteredNamespaces := filterByMap(namespacesSlice, namespaceMap)
	filteredServices := filterByMap(servicesSlice, serviceMap)

	// This means all the namespaces and services are filtered.
	if len(filteredNamespaces) == 0 && len(filteredServices) == 0 {
		return model.NewErrWithMessage(errors.New("no permission"), code.UserNoPermissionError)
	}

	setInterface(namespaces, filteredNamespaces)
	setInterface(services, filteredServices)

	return nil
}

func toStringSlice(input interface{}) ([]string, error) {
	if input == nil {
		return nil, nil
	}
	switch v := input.(type) {
	case *string:
		return []string{*v}, nil
	case *[]string:
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

func filterByMap(items []string, allowedMap map[string]struct{}) []string {
	filtered := make([]string, 0, len(items))
	for _, item := range items {
		if _, exists := allowedMap[item]; exists {
			filtered = append(filtered, item)
		}
	}
	return filtered
}
