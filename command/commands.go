package command

import (
	"github.com/mitchellh/cli"
	"github.com/tsarna/zone-update/version"
)

func Commands(ui cli.Ui) map[string]cli.CommandFactory {
	return map[string]cli.CommandFactory{
		"version": func() (cli.Command, error) {
			return &VersionCommand{
				Ui:      ui,
				Version: version.GetVersion(),
			}, nil
		},
		"agent": func() (cli.Command, error) {
			return &AgentCommand{
				Ui: ui,
			}, nil
		},
	}
}
