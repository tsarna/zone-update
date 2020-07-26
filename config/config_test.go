package config

import "testing"

func TestValidateConfig_UserPass(t *testing.T) {
	err := ValidateConfig(Config{User: "bob"})
	if err == nil {
		t.Error("User without password should have thrown an error")
	}

	err = ValidateConfig(Config{Password: "12345"})
	if err == nil {
		t.Error("Password without username should have thrown an error")
	}

	err = ValidateConfig(Config{User: "bob", Password: "12345"})
	if err != nil {
		t.Errorf("User together plus password should be allowed, but got %s", err)
	}

	err = ValidateConfig(Config{})
	if err != nil {
		t.Errorf("No User and no password should be allowed, but got %s", err)
	}
}

func TestValidateConfig_HttpAuthFile(t *testing.T) {
	err := ValidateConfig(Config{User: "bob", HttpAuthFile: "passwd"})
	if err == nil {
		t.Error("User should not be allowed together with a password file")
	}

	err = ValidateConfig(Config{HttpAuthFile: "passwd"})
	if err != nil {
		t.Errorf("Password file without user should be allowed, but got %s", err)
	}
}

func TestValidateConfig_TLS(t *testing.T) {
	err := ValidateConfig(Config{TlsCertFilename: "cert.pem"})
	if err == nil {
		t.Error("TLS Cert without key should have thrown an error")
	}

	err = ValidateConfig(Config{TlsKeyFilename: "key.pem"})
	if err == nil {
		t.Error("TLS Key without certificate should have thrown an error")
	}

	err = ValidateConfig(Config{TlsCertFilename: "cert.pem", TlsKeyFilename: "key.pem"})
	if err != nil {
		t.Errorf("TLS cert and key together should be allowed, but got %s", err)
	}

	err = ValidateConfig(Config{})
	if err != nil {
		t.Errorf("No TLS cert and no TLS key should be allowed, but got %s", err)
	}
}
