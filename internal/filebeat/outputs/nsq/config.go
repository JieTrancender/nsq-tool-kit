package nsq

import (
	"fmt"
	"time"

	"github.com/elastic/beats/v7/libbeat/common"
	"github.com/elastic/beats/v7/libbeat/outputs/codec"
)

type nsqConfig struct {
	Nsqd         string        `config:"nsqd"`
	Topic        string        `config:"topic"`
	BulkMaxSize  int           `config:"bulk_max_size"`
	MaxRetries   int           `config:"max_retries"`
	WriteTimeout time.Duration `config:"write_timeout"`
	DialTimeout  time.Duration `config:"dial_timeout"`
	Codec        codec.Config  `config:"codec"`
	FilterKeys   []string      `config:"filter_keys"`
	IgnoreKeys   []string      `config:"ignore_keys"`
}

func defaultConfig() nsqConfig {
	return nsqConfig{
		Nsqd:         "127.0.0.1:4150",
		Topic:        "nsqbeat",
		BulkMaxSize:  256,
		MaxRetries:   3,
		WriteTimeout: 3 * time.Second,
		DialTimeout:  4 * time.Second,
		FilterKeys:   nil,
		IgnoreKeys:   nil,
	}
}

func readConfig(cfg *common.Config) (*nsqConfig, error) {
	c := defaultConfig()
	if err := cfg.Unpack(&c); err != nil {
		return nil, err
	}

	return &c, nil
}

func (c *nsqConfig) Validate() error {
	if c.Topic == "" {
		return fmt.Errorf("Topic can not be empty")
	}

	return nil
}
