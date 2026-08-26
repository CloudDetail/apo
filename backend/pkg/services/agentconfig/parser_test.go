package agentconfig

import (
	"encoding/json"
	"strings"
	"testing"

	ac "github.com/CloudDetail/apo/backend/pkg/model/agentconfig"
)

func TestExtractAndRender(t *testing.T) {
	vars := ExtractVariables("a={% FOO %} b={%BAR%} c={% FOO %}")
	if len(vars) != 2 || vars[0].Name != "BAR" || vars[1].Name != "FOO" {
		t.Fatalf("unexpected %#v", vars)
	}
	got, e := Render("{% FOO %}-{% BAR %}", map[string]string{"FOO": "ok", "BAR": "x"})
	if e != nil || got != "ok-x" {
		t.Fatalf("%q %v", got, e)
	}
}

func TestShellVariablesAreNotServerTemplateVariables(t *testing.T) {
	content := "PATH=${PATH}\nCACHE=${CACHE:-30}\nvalue={% SERVER_VALUE %}\n"
	vars := ExtractVariables(content)
	if len(vars) != 1 || vars[0].Name != "SERVER_VALUE" {
		t.Fatalf("unexpected variables %#v", vars)
	}
	got, err := Render(content, map[string]string{"SERVER_VALUE": "managed"})
	if err != nil || got != "PATH=${PATH}\nCACHE=${CACHE:-30}\nvalue=managed\n" {
		t.Fatalf("rendered %q: %v", got, err)
	}
}

func TestRenderRequiresServerTemplateVariableValue(t *testing.T) {
	_, err := Render("value={% REQUIRED_VALUE %}", nil)
	if err == nil || !strings.Contains(err.Error(), "REQUIRED_VALUE") {
		t.Fatalf("missing variable was accepted: %v", err)
	}
}

func TestTemplateProbeConfigsDropsLegacyShellVariableMetadata(t *testing.T) {
	configs, err := json.Marshal([]ac.TemplateProbeConfig{{
		ProbeType: "ebpf-agent",
		Files: []ac.InitialConfigFile{{
			Path:    "load-env.sh",
			Content: "export LD_LIBRARY_PATH=${LD_LIBRARY_PATH:-}\nLEVEL={% LEVEL %}\n",
		}},
	}})
	if err != nil {
		t.Fatal(err)
	}
	template := &ac.Template{
		Configs: ac.JSON(configs),
		Variables: []ac.TemplateVariable{
			{Name: "LD_LIBRARY_PATH", Required: true},
			{Name: "LEVEL", DefaultValue: "info"},
		},
	}
	if _, err := templateProbeConfigs(template); err != nil {
		t.Fatal(err)
	}
	if len(template.Variables) != 1 || template.Variables[0].Name != "LEVEL" || template.Variables[0].DefaultValue != "info" || template.Variables[0].Required {
		t.Fatalf("unexpected effective variables: %#v", template.Variables)
	}
}
