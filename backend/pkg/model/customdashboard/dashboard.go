// Copyright 2026 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package customdashboard

const (
	FeatureMyDashboard = "我的大盘"

	VisibilityEveryone = "everyone"
	VisibilityTeam     = "team"
	VisibilityPersonal = "personal"

	QueryModePreset = "preset"
	QueryModeCustom = "custom"

	WidgetLine     = "line"
	WidgetArea     = "area"
	WidgetBar      = "bar"
	WidgetStat     = "stat"
	WidgetBarGauge = "barGauge"
	WidgetSingle   = "single"
	WidgetTable    = "table"
	WidgetTopN     = "topn"
)

type Dashboard struct {
	ID             string     `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	OwnerID        int64      `json:"ownerId"`
	Creator        string     `json:"creator"`
	Visibility     string     `json:"visibility"`
	TeamIDs        []int64    `json:"teamIds,omitempty"`
	Variables      []Variable `json:"variables"`
	Widgets        []Widget   `json:"widgets"`
	TimeRange      TimeRange  `json:"timeRange,omitempty"`
	RefreshSeconds int        `json:"refreshSeconds,omitempty"`
	CanEdit        bool       `json:"canEdit"`
	CanDelete      bool       `json:"canDelete"`
	CanCopy        bool       `json:"canCopy"`
	CreatedAt      int64      `json:"createdAt"`
	UpdatedAt      int64      `json:"updatedAt"`
	LastVisitedAt  int64      `json:"lastVisitedAt"`
}

type TimeRange struct {
	From string `json:"from"`
	To   string `json:"to"`
}

type Variable struct {
	Name         string           `json:"name"`
	Label        string           `json:"label"`
	Type         string           `json:"type"`
	Value        string           `json:"value"`
	Multi        bool             `json:"multi"`
	Required     bool             `json:"required"`
	DefaultValue string           `json:"defaultValue"`
	Labels       []string         `json:"labels,omitempty"`
	Query        string           `json:"query,omitempty"`
	Options      []VariableOption `json:"options,omitempty"`
	IncludeAll   bool             `json:"includeAll,omitempty"`
	AllValue     string           `json:"allValue,omitempty"`
}

type VariableOption struct {
	Label string `json:"label"`
	Value string `json:"value"`
}

type Widget struct {
	ID          string        `json:"id"`
	Title       string        `json:"title"`
	Description string        `json:"description"`
	Type        string        `json:"type"`
	Layout      WidgetLayout  `json:"layout"`
	Query       WidgetQuery   `json:"query"`
	Display     WidgetDisplay `json:"display"`
}

type WidgetLayout struct {
	X int `json:"x"`
	Y int `json:"y"`
	W int `json:"w,omitempty"`
	H int `json:"h,omitempty"`
}

type WidgetQuery struct {
	Mode       string            `json:"mode"`
	PresetID   string            `json:"presetId"`
	Expression string            `json:"expression"`
	Params     map[string]string `json:"params"`
	Legend     string            `json:"legend"`
	Queries    []WidgetQueryItem `json:"queries"`
}

type WidgetQueryItem struct {
	ID         string `json:"id"`
	Expression string `json:"expression"`
	Legend     string `json:"legend"`
	Hidden     bool   `json:"hidden,omitempty"`
}

type WidgetDisplay struct {
	Unit               string              `json:"unit"`
	UnitPreset         string              `json:"unitPreset"`
	AutoScale          bool                `json:"autoScale"`
	Decimals           int                 `json:"decimals"`
	EmptyText          string              `json:"emptyText,omitempty"`
	ErrorText          string              `json:"errorText,omitempty"`
	TableColumns       []WidgetTableColumn `json:"tableColumns,omitempty"`
	TableHiddenColumns []string            `json:"tableHiddenColumns,omitempty"`
	ShowValueColumn    *bool               `json:"showValueColumn,omitempty"`
	WarningThreshold   *float64            `json:"warningThreshold,omitempty"`
	CriticalThreshold  *float64            `json:"criticalThreshold,omitempty"`
}

type WidgetTableColumn struct {
	Name  string `json:"name"`
	Label string `json:"label"`
}

type Template struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Category    string    `json:"category"`
	Dashboard   Dashboard `json:"dashboard"`
}

type PresetQuery struct {
	ID               string            `json:"id"`
	Name             string            `json:"name"`
	Description      string            `json:"description"`
	Category         string            `json:"category"`
	Expression       string            `json:"expression"`
	CompatibleCharts []string          `json:"compatibleCharts"`
	Parameters       []PresetParameter `json:"parameters"`
}

type PresetParameter struct {
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	Required    bool   `json:"required"`
	Default     string `json:"default"`
}

type VariableCandidate struct {
	Label string            `json:"label"`
	Value string            `json:"value"`
	Type  string            `json:"type"`
	Meta  map[string]string `json:"meta,omitempty"`
}
