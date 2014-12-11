package goformkeeper

import (
	"fmt"
	"strings"
)

type FilterRule interface {
	GetFilterNames() []string
}

var filters map[string]func(string) string

func init() {
	filters = make(map[string]func(string) string)
	setDefaultFilters()
}

func AddFilterFunc(funcName string, f func(string) string) {
	filters[funcName] = f
}

func setDefaultFilters() {
	AddFilterFunc("trim", strings.TrimSpace)
	AddFilterFunc("lowercase", strings.ToLower)
	AddFilterFunc("uppercase", strings.ToUpper)
}

func filter(filterRule FilterRule, value string) (string, error) {
	for _, filterName := range filterRule.GetFilterNames() {
		f, found := filters[filterName]
		if !found {
			return "", fmt.Errorf("Unknown filter %s", filterName)
		}
		value = f(value)
	}
	return value, nil
}
