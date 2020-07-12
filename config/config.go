package config

import (
  "flag"
  "fmt"
  "github.com/jamiealquiza/envy"
  "os"
  "strings"
)

type Config struct {
  ZoneFileName    string
  ListenAddr      string
  HttpTimeoutSecs int
  HttpAuthRealm   string
  HttpAuthFile    string
  User            string
  Password        string
  TrustProxy      bool
  TlsCertFilename string
  TlsKeyFilename  string
  UrlPrefix       string
  RobotsTxt       bool
  TestMode        bool
}

func Init() (Config, error) {
  var config Config

  flag.StringVar(&config.ListenAddr, "listen", ":8080", "Where to listen for HTTP(S) connections")
  flag.IntVar(&config.HttpTimeoutSecs, "http-timeout", 60, "HTTP Request timeout")
  flag.StringVar(&config.HttpAuthRealm, "http-auth-realm", "zoneupdated", "Realm for HTTP Basic Auth")
  flag.StringVar(&config.User, "http-user", "", "HTTP User to allow access")
  flag.StringVar(&config.Password, "http-password", "", "HTTP Password to allow access")
  flag.StringVar(&config.HttpAuthFile, "http-auth-file", "", "A file of users and passwords, plaintext, whitespace delimited")
  flag.BoolVar(&config.TrustProxy, "trust-proxy", false, "Trust X-Real-IP/X-Forwarded-For")
  flag.StringVar(&config.TlsCertFilename, "tls-cert", "", "TLS certificate chain file")
  flag.StringVar(&config.TlsKeyFilename, "tls-key", "", "TLS certificate key file")
  flag.StringVar(&config.UrlPrefix, "url-prefix", "/zone-update", "URL prefix to serve")
  flag.BoolVar(&config.RobotsTxt, "robots-txt", false, "Serve /robots.txt to block indexing")
  flag.BoolVar(&config.TestMode, "test", false, "Testing Mode - Only update temp file")

  envy.Parse("ZUPD") // Expose environment variables.

  flag.Usage = usage
  flag.Parse()

  if len(flag.Args()) != 1 {
    flag.Usage()
    return Config{}, fmt.Errorf("Incorrect arguments")
  }

  if (config.TlsCertFilename == "") != (config.TlsKeyFilename == "") {
    return Config{}, fmt.Errorf("Must supply both TLS cert AND key files or neither.")
  }

  if !strings.HasPrefix(config.UrlPrefix, "/") {
    config.UrlPrefix = "/" + config.UrlPrefix
  }

  config.ZoneFileName = flag.Arg(0)

  return config, nil
}

func usage() {
  _, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage: %s zone-file-name\n\n", os.Args[0])
  flag.PrintDefaults()
}
