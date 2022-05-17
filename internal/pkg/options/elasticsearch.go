package options

import (
	"github.com/spf13/pflag"
)

type ElasticsearchOptions struct {
	Addrs    []string `json:"addrs" mapstructure:"addrs"`
	Username string   `json:"username" mapstructure:"username"`
	Password string   `json:"password" mapstructure:"password"`
}

func NewElasticsearchOptions() *ElasticsearchOptions {
	return &ElasticsearchOptions{
		Addrs:    []string{"127.0.0.1:9200"},
		Username: "root",
		Password: "123456",
	}
}

func (o *ElasticsearchOptions) Validate() []error {
	return []error{}
}

func (o *ElasticsearchOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&o.Addrs, "elasticsearch.addrs", o.Addrs, "Addrs of elasticsearch cluster.")
	fs.StringVar(&o.Username, "elasticsearch.username", o.Username, "Username of elasticsearch cluster.")
	fs.StringVar(&o.Password, "elasticsearch.password", o.Password, "Password of elasticsearch cluster.")
}
