// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package data

import (
	"errors"
	"fmt"
	"github.com/CloudDetail/apo/backend/pkg/code"
	"github.com/CloudDetail/apo/backend/pkg/model"
	"github.com/CloudDetail/apo/backend/pkg/repository/database"
	"time"
)

func (s *service) CheckDatasourcePermission(userID, groupID int64, namespaces, services interface{}, fillCategory string) (err error) {
	var (
		namespaceMap    = map[string]bool{}
		serviceMap      = map[string]struct{}{}
		namespaceSrvMap = map[string][]string{}
		endTime         = time.Now()
		startTime       = endTime.Add(-24 * time.Hour)
		serviceList     []string
		namespaceDs     []string
		serviceDs       []string
		filteredNs      []string
		filteredSrv     []string
		filteredSrvMap  = map[string]struct{}{}
		groups          = make([]database.DataGroup, 0)
	)

	// Get user's data group
	if groupID != 0 {
		has, err := s.dbRepo.CheckGroupPermission(userID, groupID, "view")
		if err != nil {
			return err
		}

		if !has {
			return model.NewErrWithMessage(errors.New("does not have group permission"), code.UserNoPermissionError)
		}
		filter := model.DataGroupFilter{
			ID: groupID,
		}
		groups, _, err = s.dbRepo.GetDataGroup(filter)
	} else {
		groups, err = s.getUserDataGroup(userID, "")
	}
	if err != nil {
		return err
	}

	if len(groups) == 0 {
		defaultGroup, err := s.getDefaultDataGroup("")
		if err != nil {
			return err
		}

		groups = append(groups, defaultGroup)
	}

	for _, group := range groups {
		for _, gs := range group.DatasourceList {
			if len(fillCategory) > 0 && gs.Category != fillCategory {
				continue
			}
			ds := gs.Datasource
			if gs.Type == model.DATASOURCE_TYP_NAMESPACE {
				namespaceMap[ds] = true
			} else if gs.Type == model.DATASOURCE_TYP_SERVICE {
				namespaceList, err := s.promRepo.GetServiceNamespace(startTime.UnixMicro(), endTime.UnixMicro(), ds)
				if err != nil {
					return err
				}

				for _, namespace := range namespaceList {
					// Doesn't really have the auth
					if _, ok := namespaceMap[namespace]; !ok {
						namespaceMap[namespace] = false
					}
					namespaceSrvMap[namespace] = append(namespaceSrvMap[namespace], ds)
				}
				serviceMap[ds] = struct{}{}
			}
		}
	}

	for namespace, has := range namespaceMap {
		if !has {
			continue
		}

		namespaceDs = append(namespaceDs, namespace)
	}

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

	for service := range serviceMap {
		serviceDs = append(serviceDs, service)
	}

	// The request didn't offer parameters.
	// Fill with datasource which user is authorized to view.
	if len(namespacesSlice) == 0 && len(servicesSlice) == 0 {
		if namespaces != nil && len(namespaceDs) > 0 {
			setInterface(namespaces, namespaceDs)
		} else if services != nil && len(serviceDs) > 0 {
			setInterface(services, serviceDs)
		} else {
			return model.NewErrWithMessage(errors.New("does not have data permission"), code.UserNoPermissionError)
		}
		return nil
	}

	filteredNs = make([]string, 0, len(namespacesSlice))
	filteredSrv = make([]string, 0, len(servicesSlice))
	for _, srv := range servicesSlice {
		_, exists := serviceMap[srv]
		if !exists {
			continue
		}

		if _, ok := filteredSrvMap[srv]; !ok {
			filteredSrv = append(filteredSrv, srv)
		}
		filteredSrvMap[srv] = struct{}{}
	}

	for _, ns := range namespacesSlice {
		has, exists := namespaceMap[ns]
		if !exists {
			continue
		}

		if has {
			filteredNs = append(filteredNs, ns)
			continue
		}
		// user has permissions to some of the services under this namespace
		namespaceServices := namespaceSrvMap[ns]
		var toAppend []string
		needed := true
		for _, service := range namespaceServices {
			if _, ok := filteredSrvMap[service]; !ok {
				toAppend = append(toAppend, service)
			} else {
				needed = false
			}
		}
		if needed {
			filteredSrv = append(filteredSrv, toAppend...)
		}
	}

	// This means all the namespaces and services are filtered.
	if len(filteredNs) == 0 && len(filteredSrv) == 0 {
		return model.NewErrWithMessage(errors.New("no permission"), code.UserNoPermissionError)
	}

	setInterface(namespaces, filteredNs)
	setInterface(services, filteredSrv)

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
