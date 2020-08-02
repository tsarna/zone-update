package main

import (
	"github.com/mitchellh/cli"
	"log"
	"os"
	"zoneupdated/command"
	"zoneupdated/version"
)

func main() {
	os.Exit(Run(os.Args[1:]))
}

func Run(args []string) int {
	ui := &cli.BasicUi{
		Reader:      os.Stdin,
		Writer:      os.Stdout,
		ErrorWriter: os.Stderr,
	}

	cli := &cli.CLI{
		Name: "zone-update",
		Version: version.GetVersion().FullVersionNumber(true),
		Args: args,
		Commands: command.Commands(ui),
	}

	exitStatus, err := cli.Run()
	if err != nil {
		log.Println(err)
	}

	return exitStatus
}
