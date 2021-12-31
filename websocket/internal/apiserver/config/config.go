package config

import "socket-demo/internal/apiserver/options"

type Config struct {
	*options.Options
}

// CreateConfigFromOptions 将选项转换为Config
func CreateConfigFromOptions(opts *options.Options) (*Config, error) {
	return &Config{opts}, nil
}
