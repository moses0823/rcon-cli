package config

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"time"
)

// Allowed protocols.
const (
	ProtocolRCON    = "rcon"
	ProtocolTELNET  = "telnet"
	ProtocolWebRCON = "web"
)

// DefaultProtocol contains the default protocol for connecting to a
// remote server.
const DefaultProtocol = ProtocolRCON

// DefaultTimeout contains the default dial and execute timeout.
const DefaultTimeout = 10 * time.Second

// Session contains details for making a request on a remote server.
type Session struct {
	Address  string         `json:"address" yaml:"address"`
	Password string         `json:"password" yaml:"password"`
	RCON     RCONConfig     `json:"rcon" yaml:"rcon"`
	Security SecurityConfig `json:"security" yaml:"security"`
	// Log is the name of the file to which requests will be logged.
	// If not specified, no logging will be performed.
	Log        string        `json:"log" yaml:"log"`
	Type       string        `json:"type" yaml:"type"`
	SkipErrors bool          `json:"skip_errors" yaml:"skip_errors"`
	Timeout    time.Duration `json:"timeout" yaml:"timeout"`
	Variables  bool          `json:"-" yaml:"-"`
}

type RCONConfig struct {
	Address  string `json:"address" yaml:"address"`
	Password string `json:"password" yaml:"password"`
}

type SecurityConfig struct {
	Enabled    bool   `json:"enabled" yaml:"enabled"`
	Address    string `json:"address" yaml:"address"`
	ClientID   string `json:"client-id" yaml:"client-id"`
	Secret     string `json:"secret" yaml:"secret"`
	SecretFile string `json:"secret-file" yaml:"secret-file"`
}

func (s SecurityConfig) SecretBytes() ([]byte, error) {
	if s.SecretFile != "" {
		return os.ReadFile(s.SecretFile)
	}
	return []byte(s.Secret), nil
}

func (s Session) NativeRCONAddress() string {
	if s.RCON.Address != "" {
		return s.RCON.Address
	}

	return s.Address
}

func (s Session) NativeRCONPassword() string {
	if s.RCON.Password != "" {
		return s.RCON.Password
	}

	return s.Password
}

func (s Session) SecurityAddress() string {
	if s.Security.Address != "" {
		return s.Security.Address
	}

	return s.Address
}

func (s *Session) Print(w io.Writer) error {
	js, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}

	_, _ = fmt.Fprint(w, "Print session:\n")
	_, _ = fmt.Fprint(w, string(js)+"\n")

	return nil
}
