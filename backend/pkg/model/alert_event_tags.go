// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0
package model

func (a *AlertEvent) GetDatabaseURL() string {
	return a.Tags["db_url"]
}

func (a *AlertEvent) GetDatabaseIP() string {
	return a.Tags["db_ip"]
}

func (a *AlertEvent) GetDatabasePort() string {
	return a.Tags["db_port"]
}
