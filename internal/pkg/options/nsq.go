package options

import (
	"github.com/spf13/pflag"
)

type NsqOptions struct {
	LookupdHttpAddresses []string `json:"lookupd-http-addresses" mapstructure:"lookupd-http-addresses"`
	Topics               []string `json:"topics" mapstructure:"topics"`
	Channel              string   `json:"channel" mapstructure:"channel"`
	DialTimeout          int      `json:"dial-timeout" mapstructure:"dial-timeout"`
	ReadTimeout          int      `json:"read-timeout" mapstructure:"read-timeout"`
	WriteTimeout         int      `json:"write-timeout" mapstructure:"write-timeout"`
	MaxInFight           int      `json:"max-in-fight" mapstructure:"max-in-fight"`
}

func NewNsqOptionsOptions() *NsqOptions {
	return &NsqOptions{
		LookupdHttpAddresses: []string{"http://127.0.0.1:4161"},
		Topics:               []string{"dev_test"},
		Channel:              "nsq_tool_kit",
		DialTimeout:          6,
		ReadTimeout:          60,
		WriteTimeout:         5,
		MaxInFight:           200,
	}
}

func (o *NsqOptions) Validate() []error {
	return []error{}
}

func (o *NsqOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&o.LookupdHttpAddresses, "nsq.lookupd-http-addresses", o.LookupdHttpAddresses, "addresses of lookupd cluster.")
	fs.StringSliceVar(&o.Topics, "nsq.topics", o.Topics, "Topics.")
	fs.StringVar(&o.Channel, "nsq.channel", o.Channel, "Name of consumer.")
	fs.IntVar(&o.DialTimeout, "nsq.dial-timeout", o.DialTimeout, "Nsq dial timeout in seconds.")
	fs.IntVar(&o.ReadTimeout, "nsq.read-timeout", o.ReadTimeout, "Nsq read timeout in seconds.")
	fs.IntVar(&o.WriteTimeout, "nsq.write-timeout", o.WriteTimeout, "Nsq write timeout in seconds.")
	fs.IntVar(&o.MaxInFight, "nsq.max-in-fight", o.DialTimeout, "Nsq dial timeout in seconds.")
}
