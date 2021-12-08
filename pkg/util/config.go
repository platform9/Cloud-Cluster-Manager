package util

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/kelseyhightower/envconfig"
)

// Config stores information to contact with the pf9 controller.
type Config struct {
	Fqdn          string        `json:"fqdn"`
	Username      string        `json:"username"`
	Password      string        `json:"password"`
	Tenant        string        `json:"tenant"`
	Region        string        `json:"region"`
	WaitPeriod    time.Duration `json:"wait_period"`
	AllowInsecure bool          `json:"allow_insecure"`
	ProxyURL      string        `json:"proxy_url"`
	MfaToken      string        `json:"mfa_token"`
}

var Context Config

// LoadConfig returns the information for communication with PF9 controller.
func LoadConfig(loc string) (Config, error) {

	f, err := os.Open(loc)
	ctx := Config{}

	// We will execute it if no config found or if config found but have invalid credentials
	if err != nil {
		if err := envconfig.Process("", &ctx); err != nil {
			return Config{}, err
		}
	} else {

		defer f.Close()

		//ctx = Config{WaitPeriod: time.Duration(60), AllowInsecure: false}
		err = json.NewDecoder(f).Decode(&ctx)

		if err != nil {
			fmt.Println("An error has occured", err)
			return Config{}, err
		}

	}

	// decode the password
	// Decoding base64 encoded password
	decodedBytePassword, err := base64.StdEncoding.DecodeString(ctx.Password)
	if err != nil {
		return ctx, err
	}
	ctx.Password = string(decodedBytePassword)
	//s.Stop()

	if ctx.ProxyURL != "" {
		if err = os.Setenv("https_proxy", ctx.ProxyURL); err != nil {
			return Config{}, errors.New("Error setting proxy as environment variable")
		}
	}

	return ctx, err
}
