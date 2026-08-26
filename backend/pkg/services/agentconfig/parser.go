package agentconfig

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
)

var variablePattern = regexp.MustCompile(`\{%\s*([A-Z_][A-Z0-9_]*)\s*%\}`)

type VariableSpec struct {
	Name         string `json:"name"`
	DefaultValue string `json:"defaultValue"`
	Required     bool   `json:"required"`
}

func ExtractVariables(content string) []VariableSpec {
	seen := map[string]VariableSpec{}
	for _, m := range variablePattern.FindAllStringSubmatch(content, -1) {
		name := m[1]
		if _, ok := seen[name]; !ok {
			seen[name] = VariableSpec{Name: name, Required: true}
		}
	}
	res := make([]VariableSpec, 0, len(seen))
	for _, v := range seen {
		res = append(res, v)
	}
	sort.Slice(res, func(i, j int) bool { return res[i].Name < res[j].Name })
	return res
}

func Render(content string, values map[string]string) (string, error) {
	missing := []string{}
	out := variablePattern.ReplaceAllStringFunc(content, func(token string) string {
		m := variablePattern.FindStringSubmatch(token)
		name := m[1]
		if v, ok := values[name]; ok && v != "" {
			return v
		}
		missing = append(missing, name)
		return token
	})
	if len(missing) > 0 {
		return "", fmt.Errorf("missing variables: %s", strings.Join(missing, ", "))
	}
	return out, nil
}
