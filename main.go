package main

import (
	"github.com/mitchellh/cli"
	"github.com/tsarna/zone-update/command"
	"github.com/tsarna/zone-update/version"
	"log"
	"os"
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

	c := &cli.CLI{
		Name:     "zone-update",
		Version:  version.GetVersion().FullVersionNumber(true),
		Args:     args,
		Commands: command.Commands(ui),
	}

	exitStatus, err := c.Run()
	if err != nil {
		log.Println(err)
	}

	return exitStatus
}
