package main

import (
	"zoneupdated/config"
	"zoneupdated/restapi"
	"zoneupdated/updater"
)

func main() {
	conf, err := config.Init()

	if err == nil {
		restapi.ServeHttp(conf, updater.New(conf))
	}
}
