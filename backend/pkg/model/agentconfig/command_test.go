package agentconfig

import "testing"

func TestCommandTransitionPreventsRestartRedelivery(t *testing.T) {
	if err := ValidateCommandTransition("CLAIMED", "STARTED"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateCommandTransition("STARTED", "STARTED"); err != nil {
		t.Fatal("STARTED ack retry must be idempotent", err)
	}
	if err := ValidateCommandTransition("STARTED", "SUCCEEDED"); err != nil {
		t.Fatal(err)
	}
	if err := ValidateCommandTransition("CLAIMED", "SUCCEEDED"); err == nil {
		t.Fatal("terminal result before STARTED must be rejected")
	}
	if !IsCommandTerminal("SUCCEEDED") || IsCommandTerminal("STARTED") {
		t.Fatal("terminal classification is wrong")
	}
}
