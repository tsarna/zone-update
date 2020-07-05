package config

import (
  "flag"
  "fmt"
  "github.com/jamiealquiza/envy"
  "os"
)

type Config struct {
  HttpAddr        string
  HttpPort        int
  HttpTimeoutSecs int
  ZoneFileName    string
  TestMode        bool
  HttpAuthRealm   string
  HttpAuthFile    string
  User            string
  Password        string
}

func Init() (Config, error) {
  var config Config

  flag.StringVar(&config.HttpAddr, "http-addr", "", "HTTP listen address")
  flag.IntVar(&config.HttpPort, "http-port", 8080, "HTTP listen port")
  flag.IntVar(&config.HttpTimeoutSecs, "http-timeout", 60, "HTTP Request timeout")
  flag.StringVar(&config.HttpAuthRealm, "http-auth-realm", "zoneupdated", "Realm for HTTP Basic Auth")
  flag.StringVar(&config.User, "http-user", "", "HTTP User to allow access")
  flag.StringVar(&config.Password, "http-password", "", "HTTP Password to allow access")
  flag.StringVar(&config.HttpAuthFile, "http-auth-file", "", "A file of users and passwords, plaintext, whitespace delimited")
  flag.BoolVar(&config.TestMode, "test", false, "Testing Mode")

  envy.Parse("ZUPD") // Expose environment variables.

  flag.Usage = usage
  flag.Parse()

  if (len(flag.Args()) != 1) {
    flag.Usage()
    return Config{}, fmt.Errorf("Incorrect arguments")
  }

  config.ZoneFileName = flag.Arg(0)

  return config, nil
}

func usage() {
  fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s zone-file-name\n\n", os.Args[0])
  flag.PrintDefaults()
}