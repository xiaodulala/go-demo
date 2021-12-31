package options

import (
	"fmt"
	"github.com/spf13/pflag"
)

var apiServerDefaults = ApiServerOptions{
	BindAddress: "127.0.0.1",
	BindPort:    9099,
}

type ApiServerOptions struct {
	BindAddress string `json:"bind-address" mapstructure:"bind-address"`
	BindPort    int    `json:"bind-port"    mapstructure:"bind-port"`
}

func NewApiServerOptions() *ApiServerOptions {
	return &ApiServerOptions{
		BindAddress: apiServerDefaults.BindAddress,
		BindPort:    apiServerDefaults.BindPort,
	}
}

func (s *ApiServerOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringVar(&s.BindAddress, "apiserver.bind-address", s.BindAddress, "set to 0.0.0.0 for all IPv4 interfaces.")
	fs.IntVar(&s.BindPort, "apiserver.bind-port", s.BindPort, "set port with bind.address.")
}

func (s *ApiServerOptions) Validate() []error {
	var errors []error

	if s.BindPort < 0 || s.BindPort > 65535 {
		errors = append(errors,
			fmt.Errorf("--apiserver.bind-port %v must be between 0 and 65535, inclusive. 0 for turning off insecure (HTTP) port", s.BindPort))
	}

	return errors
}
