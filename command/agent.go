package command

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/mitchellh/cli"
	"github.com/tsarna/envy"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"zone-update/config"
	"zone-update/restapi"
	"zone-update/updater"
)

type AgentCommand struct {
	Ui cli.Ui
}

func (a *AgentCommand) Run(args []string) int {
	conf := config.NewConfig(a.Ui)
	flags := a.getFlagSet(conf)

	if err := flags.Parse(args); err != nil {
		a.Ui.Error(fmt.Sprintf("Error parsing arguments: %q", err))
		return 1
	}

	envy.ParseFlagSet("ZUPD", flags)

	if flag.NArg() == 1 {
		conf.ZoneFileName = flags.Arg(0)
	} else {
		a.usage()
		return 1
	}

	if err := conf.ValidateConfig(); err != nil {
		a.Ui.Error(fmt.Sprintf("Invalid configuration: %q", err))
	}

	api := restapi.New(conf, updater.New(conf))

	listenForReload(api)

	err := api.ServeHttp()
	if err != nil {
		a.Ui.Error(fmt.Sprintf("error: %q", err))
		return 1
	}

	return 0
}

// Listen for SIGHUP and reload
func listenForReload(api restapi.RestApi) {
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGHUP)
	go func() {
		for range sigchan {
			log.Printf("Received SIGHUP, Reloading")
			api.Reload()
		}
	}()
}

func (a *AgentCommand) usage() {
	a.Ui.Output(a.Help())
}

func (a *AgentCommand) Synopsis() string {
	return "Run the zone-update server agent"
}

func (a *AgentCommand) Help() string {
	conf := config.NewConfig(a.Ui)
	flags := a.getFlagSet(conf)
	var buffer bytes.Buffer
	flags.SetOutput(&buffer)
	envy.ParseFlagSet("ZUPD", flags)
	flags.PrintDefaults()

	help := `
Usage: zone-update agent [options] zone-file-name

  Run the zone-update rest server agent.

Options:

` + buffer.String()

	return strings.TrimSpace(help)
}

func (a *AgentCommand) getFlagSet(conf *config.Config) *flag.FlagSet {
	flags := conf.AgentFlagSet()
	flags.Usage = a.usage

	return flags
}
