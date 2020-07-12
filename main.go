package main

import (
  "log"
  "zoneupdated/config"
  "zoneupdated/restapi"
  "zoneupdated/updater"
)

func main() {
  conf, err := config.Init()

  if err == nil {
    err = restapi.ServeHttp(conf, updater.New(conf))

    if err != nil {
      log.Fatal(err)
    }
  } else {
    log.Fatal(err)
  }
}
