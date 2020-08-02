package command

import (
	"github.com/mitchellh/cli"
	"log"
	"os"
	"os/signal"
	"syscall"
	"zone-update/config"
	"zone-update/restapi"
	"zone-update/updater"
)

type AgentCommand struct {
	Ui cli.Ui
}

func (a AgentCommand) Run(args []string) int {
	conf, err := config.Init(args)

	api := restapi.New(conf, updater.New(conf))

	listenForReload(api)

	if err == nil {
		err = api.ServeHttp()
		if err != nil {
			log.Print(err)
			return 1
		}
	} else {
		log.Print(err)
		return 1
	}

	return 0
}

func listenForReload(api restapi.RestApi) {
	// Listen for SIGHUP and reload
	sigchan := make(chan os.Signal, 1)
	signal.Notify(sigchan, syscall.SIGHUP)
	go func() {
		for range sigchan {
			log.Printf("Received SIGHUP, Reloading")
			api.Reload()
		}
	}()
}

func (a AgentCommand) Synopsis() string {
	return "Run the zone-update server agent"
}

func (a AgentCommand) Help() string {
	return "TODO -- help here"
}
