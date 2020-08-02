package config

import "testing"

func TestValidateConfig_UserPass(t *testing.T) {
	conf := &Config{User: "bob"}
	err := conf.ValidateConfig()
	if err == nil {
		t.Error("User without password should have thrown an error")
	}

	conf = &Config{Password: "12345"}
	err = conf.ValidateConfig()
	if err == nil {
		t.Error("Password without username should have thrown an error")
	}

	conf = &Config{User: "bob", Password: "12345"}
	err = conf.ValidateConfig()
	if err != nil {
		t.Errorf("User together plus password should be allowed, but got %s", err)
	}

	conf = &Config{}
	err = conf.ValidateConfig()
	if err != nil {
		t.Errorf("No User and no password should be allowed, but got %s", err)
	}
}

func TestValidateConfig_HttpAuthFile(t *testing.T) {
	conf := &Config{User: "bob", Password: "12345", HttpAuthFile: "passwd"}
	err := conf.ValidateConfig()
	if err == nil {
		t.Error("User should not be allowed together with a password file")
	}

	conf = &Config{HttpAuthFile: "passwd"}
	err = conf.ValidateConfig()
	if err != nil {
		t.Errorf("Password file without user should be allowed, but got %s", err)
	}
}

func TestValidateConfig_TLS(t *testing.T) {
	conf := &Config{TlsCertFilename: "cert.pem"}
	err := conf.ValidateConfig()
	if err == nil {
		t.Error("TLS Cert without key should have thrown an error")
	}

	conf = &Config{TlsKeyFilename: "key.pem"}
	err = conf.ValidateConfig()
	if err == nil {
		t.Error("TLS Key without certificate should have thrown an error")
	}

	conf = &Config{TlsCertFilename: "cert.pem", TlsKeyFilename: "key.pem"}
	err = conf.ValidateConfig()
	if err != nil {
		t.Errorf("TLS cert and key together should be allowed, but got %s", err)
	}
	if conf.UseHttps() != true {
		t.Error("Expected UseHttps true, got false")
	}

	conf = &Config{}
	err = conf.ValidateConfig()
	if err != nil {
		t.Errorf("No TLS cert and no TLS key should be allowed, but got %s", err)
	}
	if conf.UseHttps() != false {
		t.Error("Expected UseHttps false, got true")
	}
}

func TestValidateConfig_Prefix(t *testing.T) {
	conf := &Config{UrlPrefix: "foo"}
	err := conf.ValidateConfig()
	if err != nil {
		t.Errorf("URL prefix only should have been allowed, but got %s", err)
	}

	if conf.UrlPrefix != "/foo" {
		t.Errorf("Expected URL Prefix '/foo' but got '%s'", conf.UrlPrefix)
	}
}
