// Copyright 2025 CloudDetail
// SPDX-License-Identifier: Apache-2.0

package prometheus

import (
	"fmt"
	"strings"
)

// PQLFilter provides fine-grained filtering, primarily used for data group scenarios.
//
// All conditions will be expanded into DNF, which may cause combinatorial explosion.
// It is recommended to use RegexFilter instead of OrFilter to reduce complexity
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

	// HACK remove related filter directly
	SplitFilters(keys []string) (remain PQLFilter, removed PQLFilter)

	// return PQL using VM-extended syntax syntax
	//
	// e.g. {a=1 or b=2}ss
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
	v = strings.ReplaceAll(v, `unknown`, ``)
	if len(v) == 0 || len(k) == 0 {
		return f
	}
	f.Filters = append(f.Filters, k+`="`+v+`"`)
	return f
}

func (f *AndFilter) Equal(k, v string) PQLFilter {
	v = strings.ReplaceAll(v, `unknown`, ``)
	f.Filters = append(f.Filters, k+`="`+v+`"`)
	return f
}

func (f *AndFilter) NotEqual(k, v string) PQLFilter {
	v = strings.ReplaceAll(v, `unknown`, ``)
	f.Filters = append(f.Filters, k+`!="`+v+`"`)
	return f
}

func (f *AndFilter) RegexMatch(k, regexPattern string) PQLFilter {
	regexPattern = strings.ReplaceAll(regexPattern, `unknown`, ``)
	f.Filters = append(f.Filters, k+`=~"`+regexPattern+`"`)
	return f
}

func (f *AndFilter) AddPatternFilter(pattern, v string) PQLFilter {
	v = strings.ReplaceAll(v, `unknown`, ``)
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
	if f == nil {
		return nil
	}

	newFilters := make([]string, len(f.Filters))
	copy(newFilters, f.Filters)
	return &AndFilter{
		Filters: newFilters,
	}
}

func (f *AndFilter) SplitFilters(keys []string) (PQLFilter, PQLFilter) {
	var removed []string
	for _, key := range keys {
		for i := len(f.Filters) - 1; i >= 0; i-- {
			if strings.HasPrefix(f.Filters[i], key+"=") ||
				strings.HasPrefix(f.Filters[i], key+"!=") ||
				strings.HasPrefix(f.Filters[i], key+"=~") {
				removed = append(removed, f.Filters[i])
				f.Filters = append(f.Filters[:i], f.Filters[i+1:]...)
			}
		}
	}

	if len(removed) == 0 {
		return f, nil
	}
	return f, &AndFilter{Filters: removed}
}

type OrFilter struct {
	Filters []AndFilter
}

func Clone(filter PQLFilter) PQLFilter {
	if filter == nil {
		return &AndFilter{}
	}
	switch f := filter.(type) {
	case *AndFilter:
		return f.Clone()
	case *OrFilter:
		return f.Clone()
	}
	return nil
}

func Or(filters ...PQLFilter) *OrFilter {
	if len(filters) == 0 {
		return &OrFilter{
			Filters: []AndFilter{*AlwaysFalseFilter},
		}
	}

	var options []AndFilter
	for _, f := range filters {
		if f == nil {
			continue
		}
		switch ft := f.(type) {
		case *AndFilter:
			if len(ft.Filters) == 0 {
				continue
			}
			options = append(options, *ft.Clone().(*AndFilter))
		case *OrFilter:
			for _, andf := range ft.Filters {
				options = append(options, *andf.Clone().(*AndFilter))
			}
		}
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
			if len(f.Filters) == 0 {
				continue
			}

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
			if len(f.Filters) == 0 {
				continue
			}
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
	for i := 0; i < len(f.Filters); i++ {
		f.Filters[i].Equal(k, v)
	}
	return f
}

func (f *OrFilter) NotEqual(k, v string) PQLFilter {
	for i := 0; i < len(f.Filters); i++ {
		f.Filters[i].NotEqual(k, v)
	}
	return f
}

func (f *OrFilter) RegexMatch(k, regexPattern string) PQLFilter {
	for i := 0; i < len(f.Filters); i++ {
		f.Filters[i].RegexMatch(k, regexPattern)
	}
	return f
}

func (f *OrFilter) AddPatternFilter(pattern, v string) PQLFilter {
	for i := 0; i < len(f.Filters); i++ {
		f.Filters[i].AddPatternFilter(pattern, v)
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
	if o == nil {
		return nil
	}

	newFilters := make([]AndFilter, len(o.Filters))
	for i, f := range o.Filters {
		newFilters[i] = *f.Clone().(*AndFilter)
	}
	return &OrFilter{Filters: newFilters}
}

func (o *OrFilter) SplitFilters(keys []string) (PQLFilter, PQLFilter) {
	var removed []AndFilter
	for i := 0; i < len(o.Filters); i++ {
		_, r := o.Filters[i].SplitFilters(keys)
		if r != nil {
			aF := r.(*AndFilter)
			if len(aF.Filters) > 0 {
				removed = append(removed, *aF)
			}
		}
	}
	return o, &OrFilter{
		Filters: removed,
	}
}

// ############## Fast Filter ##############

// Fast Filter
func EqualFilter(k, v string) *AndFilter {
	v = strings.ReplaceAll(v, `unknown`, ``)
	return &AndFilter{Filters: []string{k + `="` + v + `"`}}
}

func EqualIfNotEmptyFilter(k, v string) *AndFilter {
	if len(v) == 0 {
		return nil
	}
	v = strings.ReplaceAll(v, `unknown`, ``)
	return &AndFilter{Filters: []string{k + `="` + v + `"`}}
}

func NotEqualFilter(k, v string) *AndFilter {
	v = strings.ReplaceAll(v, `unknown`, ``)
	return &AndFilter{Filters: []string{k + `!="` + v + `"`}}
}

func RegexMatchFilter(k, regexPattern string) *AndFilter {
	regexPattern = strings.ReplaceAll(regexPattern, `unknown`, ``)
	return &AndFilter{Filters: []string{k + `=~"` + regexPattern + `"`}}
}

func RegexMatchIfNotEmptyFilter(k, regexPattern string) *AndFilter {
	if len(regexPattern) == 0 {
		return nil
	}
	regexPattern = strings.ReplaceAll(regexPattern, `unknown`, ``)
	return &AndFilter{Filters: []string{k + `=~"` + regexPattern + `"`}}
}

func PatternFilter(pattern, v string) *AndFilter {
	v = strings.ReplaceAll(v, `unknown`, ``)
	return &AndFilter{Filters: []string{pattern + `"` + v + `"`}}
}

var AlwaysFalseFilter = &AndFilter{Filters: []string{"apo_filter=\"never_match\""}}
