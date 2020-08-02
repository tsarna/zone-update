package command

import (
	"github.com/mitchellh/cli"
	"github.com/tsarna/zone-update/version"
	"strings"
)

type VersionCommand struct {
	Ui      cli.Ui
	Version *version.VersionInfo
}

func (v *VersionCommand) Help() string {
	help := `
Usage: zone-update version

	Print version information about this build of zone-update.
`

	return strings.TrimSpace(help)
}

func (v VersionCommand) Run(_ []string) int {
	v.Ui.Output(v.Version.FullVersionNumber(true))
	return 0
}

func (v *VersionCommand) Synopsis() string {
	return "Print the zone-update version"
}
