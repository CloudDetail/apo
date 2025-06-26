// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"fmt"
	"strings"
)

type PQLFilter interface {
	// { $k='$v' }
	Equal(k, v string) PQLFilter
	// { $k!='$v' }
	NotEqual(k, v string) PQLFilter
	// { $k=~'$regexPattern' }
	RegexMatch(k, regexPattern string) PQLFilter
	// compatible with old predefined FilterPattern
	//
	// e.g. { $pattern$v }
	AddPatternFilter(pattern, v string) PQLFilter

	Clone() PQLFilter

	// return PQL using VM-extended syntax syntax
	//
	// e.g. {a=1 or b=2}
	String() string

	// return PQL using strict prometheus syntax
	//
	// e.g. $__name__{a=1} or $__name__{b=2}
	_strictPQL(__name__ string, modifier string, offset string, rangeV string) []string
}

type AndFilter struct {
	Filters []string
}

func NewFilter() *AndFilter {
	return &AndFilter{}
}

func (f *AndFilter) EqualIfNotEmpty(k, v string) PQLFilter {
	if len(v) == 0 || len(k) == 0 {
		return f
	}
	f.Filters = append(f.Filters, k+`="`+v+`"`)
	return f
}

func (f *AndFilter) Equal(k, v string) PQLFilter {
	f.Filters = append(f.Filters, k+`="`+v+`"`)
	return f
}

func (f *AndFilter) NotEqual(k, v string) PQLFilter {
	f.Filters = append(f.Filters, k+`!="`+v+`"`)
	return f
}

func (f *AndFilter) RegexMatch(k, regexPattern string) PQLFilter {
	f.Filters = append(f.Filters, k+`=~"`+regexPattern+`"`)
	return f
}

func (f *AndFilter) AddPatternFilter(pattern, v string) PQLFilter {
	f.Filters = append(f.Filters, pattern+`"`+v+`"`)
	return f
}

func (f *AndFilter) _strictPQL(__name__ string, modifier string, offset string, rangeV string) []string {
	return []string{__name__ + fmt.Sprintf("{%s}", strings.Join(f.Filters, ","))}
}

func (f *AndFilter) String() string {
	return strings.Join(f.Filters, ",")
}

func (f *AndFilter) Clone() PQLFilter {
	newFilters := make([]string, len(f.Filters))
	copy(newFilters, f.Filters)
	return &AndFilter{
		Filters: newFilters,
	}
}

type OrFilter struct {
	Filters []AndFilter
}

// TODO Support Or(filter ...PQLFilter)
func Or(filters ...*AndFilter) *OrFilter {
	var options []AndFilter
	for i := 0; i < len(filters); i++ {
		if filters[i] == nil {
			continue
		}
		options = append(options, *filters[i].Clone().(*AndFilter))
	}
	return &OrFilter{Filters: options}
}

func And(filters ...PQLFilter) *OrFilter {
	var options []AndFilter
	for _, filter := range filters {
		if filter == nil {
			continue
		}
		switch f := filter.(type) {
		case *OrFilter:
			if len(options) == 0 {
				options = append(options, f.Filters...)
				continue
			}
			newOptions := make([]AndFilter, 0, len(options)*len(f.Filters))
			for _, option := range options {
				for _, filter := range f.Filters {
					opt := option.Clone().(*AndFilter)
					opt.Filters = append(opt.Filters, filter.Filters...)
					newOptions = append(newOptions, *opt)
				}
			}
			options = newOptions
		case *AndFilter:
			if len(options) == 0 {
				options = append(options, *f.Clone().(*AndFilter))
				continue
			}
			for i := 0; i < len(options); i++ {
				options[i].Filters = append(options[i].Filters, f.Filters...)
			}
		}
	}
	return &OrFilter{Filters: options}
}

func (f *OrFilter) Equal(k, v string) PQLFilter {
	for _, filter := range f.Filters {
		filter.Equal(k, v)
	}
	return f
}

func (f *OrFilter) NotEqual(k, v string) PQLFilter {
	for _, filter := range f.Filters {
		filter.NotEqual(k, v)
	}
	return f
}

func (f *OrFilter) RegexMatch(k, regexPattern string) PQLFilter {
	for _, filter := range f.Filters {
		filter.RegexMatch(k, regexPattern)
	}
	return f
}

func (f *OrFilter) AddPatternFilter(pattern, v string) PQLFilter {
	for _, filter := range f.Filters {
		filter.AddPatternFilter(pattern, v)
	}
	return f
}

func (f *OrFilter) _strictPQL(__name__ string, modifier string, offset string, rangeV string) []string {
	if len(f.Filters) == 0 {
		return []string{__name__}
	}

	// Standard Prometheus PQL
	var vectors []string
	for _, filter := range f.Filters {
		vectors = append(vectors, filter._strictPQL(__name__, modifier, offset, rangeV)...)
	}

	return vectors
}

func (f *OrFilter) String() string {
	var filters []string
	for _, filter := range f.Filters {
		filters = append(filters, filter.String())
	}
	return strings.Join(filters, " or ")
}

func (o *OrFilter) Clone() PQLFilter {
	newFilters := make([]AndFilter, len(o.Filters))
	for i, f := range o.Filters {
		newFilters[i] = *f.Clone().(*AndFilter)
	}
	return &OrFilter{Filters: newFilters}
}

// ############## Fast Filter ##############

// Fast Filter
func EqualFilter(k, v string) *AndFilter {
	return &AndFilter{Filters: []string{k + `="` + v + `"`}}
}

func EqualIfNotEmptyFilter(k, v string) *AndFilter {
	if len(v) == 0 {
		return nil
	}
	return &AndFilter{Filters: []string{k + `="` + v + `"`}}
}

func NotEqualFilter(k, v string) *AndFilter {
	return &AndFilter{Filters: []string{k + `!="` + v + `"`}}
}

func RegexMatchFilter(k, regexPattern string) *AndFilter {
	return &AndFilter{Filters: []string{k + `=~"` + regexPattern + `"`}}
}

func RegexMatchIfNotEmptyFilter(k, regexPattern string) *AndFilter {
	if len(regexPattern) == 0 {
		return nil
	}
	return &AndFilter{Filters: []string{k + `=~"` + regexPattern + `"`}}
}

func PatternFilter(pattern, v string) *AndFilter {
	return &AndFilter{Filters: []string{pattern + `"` + v + `"`}}
}
