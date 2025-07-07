// Copyright 2024 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package request

import (
	"sort"

	"github.com/CloudDetail/apo/backend/pkg/model/amconfig"
	"github.com/prometheus/common/model"
)

type InputAlertManagerRequest struct {
	Receiver          string `json:"receiver"`
	Status            string `json:"status"`
	Alerts            Alerts `json:"alerts"`
	GroupLabels       KV     `json:"groupLabels"`
	CommonLabels      KV     `json:"commonLabels"`
	CommonAnnotations KV     `json:"commonAnnotations"`
	TruncatedAlerts   int    `json:"truncatedAlerts"`
	ExternalURL       string `json:"ExternalURL"`
}

type Alerts []Alert

func (as Alerts) Firing() []Alert {
	res := []Alert{}
	for _, a := range as {
		if a.Status == string(model.AlertFiring) {
			res = append(res, a)
		}
	}
	return res
}

// Resolved returns the subset of alerts that are resolved.
func (as Alerts) Resolved() []Alert {
	res := []Alert{}
	for _, a := range as {
		if a.Status == string(model.AlertResolved) {
			res = append(res, a)
		}
	}
	return res
}

type KV map[string]string

// SortedPairs returns a sorted list of key/value pairs.
func (kv KV) SortedPairs() Pairs {
	var (
		pairs     = make([]Pair, 0, len(kv))
		keys      = make([]string, 0, len(kv))
		sortStart = 0
	)
	for k := range kv {
		if k == string(model.AlertNameLabel) {
			keys = append([]string{k}, keys...)
			sortStart = 1
		} else {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys[sortStart:])

	for _, k := range keys {
		pairs = append(pairs, Pair{k, kv[k]})
	}
	return pairs
}

// Remove returns a copy of the key/value set without the given keys.
func (kv KV) Remove(keys []string) KV {
	keySet := make(map[string]struct{}, len(keys))
	for _, k := range keys {
		keySet[k] = struct{}{}
	}

	res := KV{}
	for k, v := range kv {
		if _, ok := keySet[k]; !ok {
			res[k] = v
		}
	}
	return res
}

// Names returns the names of the label names in the LabelSet.
func (kv KV) Names() []string {
	return kv.SortedPairs().Names()
}

// Values returns a list of the values in the LabelSet.
func (kv KV) Values() []string {
	return kv.SortedPairs().Values()
}

// Pair is a key/value string pair.
type Pair struct {
	Name, Value string
}

// Pairs is a list of key/value string pairs.
type Pairs []Pair

// Names returns a list of names of the pairs.
func (ps Pairs) Names() []string {
	ns := make([]string, 0, len(ps))
	for _, p := range ps {
		ns = append(ns, p.Name)
	}
	return ns
}

// Values returns a list of values of the pairs.
func (ps Pairs) Values() []string {
	vs := make([]string, 0, len(ps))
	for _, p := range ps {
		vs = append(vs, p.Value)
	}
	return vs
}

type Alert struct {
	Status       string `json:"status"`
	Labels       KV     `json:"labels"`
	Annotations  KV     `json:"annotations"`
	StartsAt     string `json:"startsAt"`
	EndsAt       string `json:"endsAt"`
	GeneratorURL string `json:"generatorURL"`
	Fingerprint  string `json:"fingerprint"`
}

type GetAlertRuleConfigRequest struct {
	AlertRuleFile string `form:"alertRuleFile" json:"alertRuleFile"`
}

type GetAlertRuleRequest struct {
	AlertRuleFile string `form:"alertRuleFile" json:"alertRuleFile"`
	RefreshCache  bool   `form:"refreshCache" json:"refreshCache"`

	*AlertRuleFilter `json:",inline"`
	*PageParam       `json:",inline"`
}

type GetAlertManagerConfigReceverRequest struct {
	AMConfigFile string `form:"amConfigFile" json:"amConfigFile"`
	RefreshCache bool   `form:"refreshCache" json:"refreshCache"`

	*AMConfigReceiverFilter
	*PageParam
}

type AlertRuleFilter struct {
	Group    string   `form:"group" json:"group"`
	Groups   []string `form:"groups" json:"groups"`
	Alert    string   `form:"alert" json:"alert"`
	Severity []string `form:"severity" json:"severity"` // alarm level info warning...
	Keyword  string   `form:"keyword" json:"keyword"`
}

type AMConfigReceiverFilter struct {
	Name  string `form:"name" json:"name"`
	RType string `form:"rType" json:"rType"`
}

type UpdateAlertRuleConfigRequest struct {
	AlertRuleFile string `json:"alertRuleFile"`
	Content       string `json:"content"`
}

type UpdateAlertRuleRequest struct {
	AlertRuleFile string `json:"alertRuleFile"`

	OldGroup  string    `json:"oldGroup" binding:"required"`
	OldAlert  string    `json:"oldAlert" binding:"required"`
	AlertRule AlertRule `json:"alertRule"`

	GroupID int64 `json:"groupId"`
}

type AddAlertManagerConfigReceiver UpdateAlertManagerConfigReceiver

type UpdateAlertManagerConfigReceiver struct {
	AMConfigFile string `form:"amConfigFile" json:"amConfigFile"`

	Type             string            `form:"type" json:"type"` // receiver type
	OldName          string            `form:"oldName" json:"oldName"`
	AMConfigReceiver amconfig.Receiver `form:"amConfigReceiver" json:"amConfigReceiver"`
}

type DeleteAlertRuleRequest struct {
	AlertRuleFile string `form:"alertRuleFile" json:"alertRuleFile"`

	Group string `form:"group" json:"group" binding:"required"`
	Alert string `form:"alert" json:"alert" binding:"required"`
}

type DeleteAlertManagerConfigReceiverRequest struct {
	AMConfigFile string `form:"amConfigFile" json:"amConfigFile"`
	Type         string `form:"type" json:"type"`
	Name         string `form:"name" json:"name" binding:"required"`
}

type AlertRule struct {
	Group string `json:"group" binding:"required"`

	Record        string            `json:"record"`
	Alert         string            `json:"alert" binding:"required"`
	Expr          string            `json:"expr"`
	For           string            `json:"for,omitempty"`
	KeepFiringFor string            `json:"keepFiringFor,omitempty"`
	Labels        map[string]string `json:"labels,omitempty"`
	Annotations   map[string]string `json:"annotations,omitempty"`
}

type AddAlertRuleRequest struct {
	AlertRuleFile string `json:"alertRuleFile"`

	AlertRule AlertRule `json:"alertRule"`
	GroupID   int64     `json:"groupId"`
}

type CheckAlertRuleRequest struct {
	AlertRuleFile string `form:"alertRuleFile,omitempty"`
	Group         string `form:"group" binding:"required"`
	Alert         string `form:"alert" binding:"required"`
}

type ForwardToDingTalkRequest InputAlertManagerRequest
