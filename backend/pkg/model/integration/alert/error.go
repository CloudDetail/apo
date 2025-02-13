// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package alert

import (
	"fmt"
)

type ErrAlertSourceAlreadyExist struct {
	Name string
}

func (e ErrAlertSourceAlreadyExist) Error() string {
	return fmt.Sprintf("alertsource %s is already existed", e.Name)
}

type ErrAlertSourceNotExist struct{}

func (e ErrAlertSourceNotExist) Error() string {
	return "alertSource is not existed"
}

type ErrNotAllowSchema struct {
	Table  string
	Column string
}

func (e ErrNotAllowSchema) Error() string {
	if len(e.Table) > 0 {
		return fmt.Sprintf("not allowed table: %s", e.Table)
	}

	return fmt.Sprintf("not allowed column: %s", e.Column)
}

type ErrIllegalAlertRule struct {
	Err error
}

func (e ErrIllegalAlertRule) Error() string {
	return fmt.Sprintf("illegal alert rule: %v", e.Err)
}
