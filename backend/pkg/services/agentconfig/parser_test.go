package agentconfig

import "testing"

func TestExtractAndRender(t *testing.T) {
	vars := ExtractVariables("a=${FOO} b=${BAR:-x} c=${FOO}")
	if len(vars) != 2 || vars[0].Name != "BAR" || vars[1].Name != "FOO" {
		t.Fatalf("unexpected %#v", vars)
	}
	got, e := Render("${FOO}-${BAR:-x}", map[string]string{"FOO": "ok"})
	if e != nil || got != "ok-x" {
		t.Fatalf("%q %v", got, e)
	}
}
