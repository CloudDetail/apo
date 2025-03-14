package alert

import (
	"fmt"
	"strings"
)

type ErrAlertImpactNotFit struct {
	TagGroup
	Group string
}

func (e ErrAlertImpactNotFit) Error() string {
	return fmt.Sprintf("cannon fit search condition: group %s cannot search instance by %s", e.Group, e.TagGroup)
}

type ErrAlertImpactMissingTag struct {
	TagGroups []TagGroup
	Event     *Alert
}

type TagGroup []string

func (e ErrAlertImpactMissingTag) Error() string {
	return fmt.Sprintf("Unable to find any of the following label group %s", e.TagGroups)
}

func (e ErrAlertImpactMissingTag) CheckedTagGroups() string {
	return fmt.Sprintf("%s", e.TagGroups)
}

func (e *ErrAlertImpactMissingTag) AddCheckedGroup(err ErrAlertImpactMissingTag) {
	e.TagGroups = append(e.TagGroups, err.TagGroups...)
}

type ErrMutationCheckFailed struct {
	PQL        string
	UpperLimit string
	LowerLimit string
	UserMsg    string
	Err        error
}

func (e ErrMutationCheckFailed) Error() string {
	return fmt.Sprintf("failed to check mutation for (%s) outside [%s,%s]: %v,", e.PQL, e.LowerLimit, e.UpperLimit, e.Err)
}

func (e ErrMutationCheckFailed) Msg() string {
	return fmt.Sprintf("指标突变查询失败 (%s), %s", e.PQL, e.UserMsg)
}

type ErrAlertImpactNoMatchedService struct {
	TagGroup  TagGroup
	TagValues []string
}

func (e ErrAlertImpactNoMatchedService) Error() string {
	var checkedGroup []string
	for idx, group := range e.TagGroup {
		if idx == len(e.TagValues) {
			break
		}
		checkedGroup = append(checkedGroup, fmt.Sprintf("%s:%s", group, e.TagValues[idx]))
	}

	return fmt.Sprintf("no service found for [ %s ]", strings.Join(checkedGroup, ", "))
}

func (e ErrAlertImpactNoMatchedService) CheckedTagGroup() string {
	var checkedGroup []string
	for idx, group := range e.TagGroup {
		if idx == len(e.TagValues) {
			break
		}
		checkedGroup = append(checkedGroup, fmt.Sprintf("%s:%s", group, e.TagValues[idx]))
	}

	return fmt.Sprintf("[%s]", strings.Join(checkedGroup, ", "))
}

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
