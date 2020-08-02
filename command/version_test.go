package command

import (
	"github.com/mitchellh/cli"
	"github.com/tsarna/zone-update/version"
	"strings"
	"testing"
)

func TestVersionCommand_Run(t *testing.T) {
	ver := version.GetVersion()

	ui := cli.NewMockUi()
	defer ui.OutputWriter.Reset()

	cmd := &VersionCommand{
		Ui:      ui,
		Version: ver,
	}

	if code := cmd.Run([]string{}); code != 0 {
		t.Fatalf("expected exit 0, got: %d", code)
	}

	out := ui.OutputWriter.String()
	if !strings.Contains(out, ver.String()) {
		t.Errorf("Version output '%s' did not contain version string", out)
	}
}

func TestVersionCommand_Help(t *testing.T) {
	ver := version.GetVersion()

	ui := cli.NewMockUi()
	defer ui.OutputWriter.Reset()

	cmd := &VersionCommand{
		Ui:      ui,
		Version: ver,
	}

	ui.Output("get help")

	help := cmd.Help()
	if !strings.Contains(help, "Usage:") {
		t.Errorf("Version subcommand help output '%s' did not contain usage", help)
	}
}
