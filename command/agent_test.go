package command

import (
	"github.com/mitchellh/cli"
	"strings"
	"testing"
)

func TestAgentCommand_NoFile(t *testing.T) {
	testAgentCommand_Usage(t, []string{})
}

func TestAgentCommand_TooManyArgs(t *testing.T) {
	testAgentCommand_Usage(t, []string{"foo", "bar"})
}

func TestAgentCommand_BadFlag(t *testing.T) {
	testAgentCommand_Usage(t, []string{"-foo", "bar"})
}

func testAgentCommand_Usage(t *testing.T, args []string) {
	ui := cli.NewMockUi()
	defer ui.OutputWriter.Reset()

	cmd := &AgentCommand{Ui: ui}

	if code := cmd.Run(args); code != 1 {
		t.Fatalf("expected exit 1, got: %d", code)
	}

	out := ui.OutputWriter.String()
	if !strings.Contains(out, "Usage:") {
		t.Errorf("Output '%s' did not contain usage", out)
	}

	if !strings.Contains(out, "-trust-proxy") {
		t.Errorf("Output '%s' did not contain options information", out)
	}

	if !strings.Contains(out, "ZUPD_TRUST_PROXY") {
		t.Errorf("Output '%s' did not contain environment variables information", out)
	}
}
