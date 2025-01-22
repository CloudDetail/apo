// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package model

func (a *AlertEvent) GetDatabaseURL() string {
	if dbURL, find := a.Tags["db_url"]; find && len(dbURL) > 0 {
		return dbURL
	}

	if a.Group == "database" {
		instance := a.Tags["instance"]
		dbURL := dbURLRegex.FindSubmatch([]byte(instance))
		if len(dbURL) > 0 {
			return string(dbURL[1])
		}
	}
	return ""
}

func (a *AlertEvent) GetDatabaseIP() string {
	if dbIP := a.Tags["db_ip"]; len(dbIP) > 0 {
		return dbIP
	}

	if a.Group == "database" {
		instance := a.Tags["instance"]
		dbIP := dbIPRegex.FindSubmatch([]byte(instance))
		if len(dbIP) > 0 {
			return string(dbIP[1])
		}
	}
	return ""
}

func (a *AlertEvent) GetDatabasePort() string {
	if dbPort := a.Tags["db_port"]; len(dbPort) > 0 {
		return dbPort
	}

	if a.Group == "database" {
		instance := a.Tags["instance"]
		dbPort := dbPortRegex.FindSubmatch([]byte(instance))
		if len(dbPort) > 0 {
			return string(dbPort[1])
		}
	}
	return ""
}
