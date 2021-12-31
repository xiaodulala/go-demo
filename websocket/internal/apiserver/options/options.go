package options

import (
	cliflag "github.com/xiaodulala/component-tools/component/cli/flag"
	"github.com/xiaodulala/component-tools/component/json"
	"socket-demo/internal/pkg/options"
)

type Options struct {
	ApiServerOptions *ApiServerOptions         `json:"api" mapstructure:"api"`
	ServerRunOptions *options.ServerRunOptions `json:"server" mapstructure:"server"`
	LogOptions       *options.LogOptions       `json:"log"      mapstructure:"log"`
}

func NewOptions() *Options {
	o := Options{
		ApiServerOptions: NewApiServerOptions(),
		ServerRunOptions: options.NewServerRunOptions(),
		LogOptions:       options.NewLogOptions(),
	}
	return &o
}

func (o *Options) Flags() (fss cliflag.NamedFlagSets) {
	o.ApiServerOptions.AddFlags(fss.FlagSet("api"))
	o.ServerRunOptions.AddFlags(fss.FlagSet("server"))
	o.LogOptions.AddFlags(fss.FlagSet("logs"))

	return fss
}

func (o *Options) Validate() []error {
	var errors []error
	errors = append(errors, o.ApiServerOptions.Validate()...)
	errors = append(errors, o.ServerRunOptions.Validate()...)
	errors = append(errors, o.LogOptions.Validate()...)
	return errors
}

func (o *Options) String() string {
	data, _ := json.Marshal(o)

	return string(data)
}
