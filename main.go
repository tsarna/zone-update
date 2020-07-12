package main

import (
  "log"
  "os"
  "os/signal"
  "syscall"
  "zoneupdated/config"
  "zoneupdated/restapi"
  "zoneupdated/updater"
)

func main() {
  conf, err := config.Init()

  api := restapi.New(conf, updater.New(conf))

  // Listen for SIGHUP and reload
  sigchan := make(chan os.Signal, 1)
  signal.Notify(sigchan, syscall.SIGHUP)
  go func(){
    for range sigchan {
      log.Printf("Received SIGHUP, Reloading")
      err := api.Reload()
      if err != nil {
        log.Printf("Reload failed: %s", err)
      }
    }
  }()

  if err == nil {
    err = api.ServeHttp()
    if err != nil {
      log.Fatal(err)
    }
  } else {
    log.Fatal(err)
  }
}
