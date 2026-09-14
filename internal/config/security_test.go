package config_test

import (
	"os"
	"testing"

	"github.com/gorcon/rcon-cli/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestSecureSessionConfiguration(t *testing.T) {
	name := "rcon-secure-test.yaml"
	body := "" +
		"servers:\n" +
		"  6b7t:\n" +
		"    address: gateway.example:25576\n" +
		"    rcon:\n" +
		"      address: 127.0.0.1:25575\n" +
		"      password: native-password\n" +
		"    security:\n" +
		"      enabled: true\n" +
		"      client-id: moses\n" +
		"      secret: configured-secret\n"
	if err := os.WriteFile(name, []byte(body), 0600); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(name)

	cfg, err := config.NewConfig(name)
	assert.NoError(t, err)
	session := (*cfg)["6b7t"]
	assert.True(t, session.Security.Enabled)
	assert.Equal(t, "gateway.example:25576", session.SecurityAddress())
	assert.Equal(t, "127.0.0.1:25575", session.NativeRCONAddress())
	assert.Equal(t, "native-password", session.NativeRCONPassword())
}

func TestSecretFile(t *testing.T) {
	name := "rcon-secret-test.key"
	if err := os.WriteFile(name, []byte("file-secret"), 0600); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(name)

	secret, err := (config.SecurityConfig{SecretFile: name}).SecretBytes()
	assert.NoError(t, err)
	assert.Equal(t, []byte("file-secret"), secret)
}

func TestLegacySessionConfigurationRemainsNativeRCON(t *testing.T) {
	session := config.Session{Address: "example.com:25575", Password: "password"}
	assert.False(t, session.Security.Enabled)
	assert.Equal(t, session.Address, session.NativeRCONAddress())
	assert.Equal(t, session.Password, session.NativeRCONPassword())
}
