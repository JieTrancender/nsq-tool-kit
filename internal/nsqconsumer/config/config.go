package config

import (
	"github.com/JieTrancender/nsq-tool-kit/internal/nsqconsumer/options"
)

type Config struct {
	*options.Options
}

func CreateConfigFromOptions(opts *options.Options) (*Config, error) {
	return &Config{opts}, nil
}
