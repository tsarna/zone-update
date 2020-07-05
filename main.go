package main

import (
	"zoneupdated/config"
	"zoneupdated/restapi"
)

func main() {
	conf, err := config.Init()

	if err == nil {
		restapi.ServeHttp(conf)
	}
}
