package config

import (
	"errors"
	"flag"
	"github.com/mitchellh/cli"
	"strings"
)

type Config struct {
	Ui               cli.Ui
	ZoneFileName     string
	ListenAddr       string
	HttpTimeoutSecs  int
	HttpAuthRealm    string
	HttpAuthFile     string
	User             string
	Password         string
	TrustProxy       bool
	TlsCertFilename  string
	TlsKeyFilename   string
	UrlPrefix        string
	RobotsTxt        bool
	TestMode         bool
	SequentialSerial bool
}

func NewConfig(ui cli.Ui) *Config {
	return &Config{
		Ui: ui,
	}
}

func (config *Config) AgentFlagSet() *flag.FlagSet {
	flags := flag.NewFlagSet("agent", flag.ContinueOnError)

	flags.StringVar(&config.ListenAddr, "listen", ":8080", "Where to listen for HTTP(S) connections")
	flags.IntVar(&config.HttpTimeoutSecs, "http-timeout", 60, "HTTP Request timeout")
	flags.StringVar(&config.HttpAuthRealm, "http-auth-realm", "zone-update", "Realm for HTTP Basic Auth")
	flags.StringVar(&config.User, "http-user", "", "HTTP User to allow access")
	flags.StringVar(&config.Password, "http-password", "", "HTTP Password to allow access")
	flags.StringVar(&config.HttpAuthFile, "http-auth-file", "", "A file of users and passwords, plaintext, whitespace delimited")
	flags.BoolVar(&config.TrustProxy, "trust-proxy", false, "Trust X-Real-IP/X-Forwarded-For")
	flags.StringVar(&config.TlsCertFilename, "tls-cert", "", "TLS certificate chain file")
	flags.StringVar(&config.TlsKeyFilename, "tls-key", "", "TLS certificate key file")
	flags.StringVar(&config.UrlPrefix, "url-prefix", "/zone-update", "URL prefix to serve")
	flags.BoolVar(&config.RobotsTxt, "robots-txt", false, "Serve /robots.txt to block indexing")
	flags.BoolVar(&config.SequentialSerial, "sequential-serial", false, "Use a simple incrementing serial number (not date based)")
	flags.BoolVar(&config.TestMode, "test", false, "Testing Mode - Only update temp file")

	return flags
}

func (conf *Config) ValidateConfig() error {
	if !strings.HasPrefix(conf.UrlPrefix, "/") {
		conf.UrlPrefix = "/" + conf.UrlPrefix
	}

	if (conf.User == "") != (conf.Password == "") {
		return errors.New("must supply both user and password or neither")
	}

	if conf.HttpAuthFile != "" && conf.User != "" {
		return errors.New("cannot specify both an auth file and a user")
	}

	if (conf.TlsCertFilename == "") != (conf.TlsKeyFilename == "") {
		return errors.New("must supply both TLS cert AND key files or neither")
	}

	return nil
}

func (conf *Config) UseHttps() bool {
	return conf.TlsCertFilename != "" && conf.TlsKeyFilename != ""
}
