package options

import (
	cliflag "github.com/marmotedu/component-base/pkg/cli/flag"
	"github.com/marmotedu/iam/pkg/log"

	genericoptions "github.com/JieTrancender/nsq-tool-kit/internal/pkg/options"
)

// Options runs a iam api server.
type Options struct {
	Log           *log.Options                         `json:"log" mapstructure:"log"`
	Elasticsearch *genericoptions.ElasticsearchOptions `json:"elasticsearch" mapstructure:"elasticsearch"`
	Nsq           *genericoptions.NsqOptions           `json:"nsq" mapstructure:"nsq"`
	Etcd          *genericoptions.EtcdOptions          `json:"etcd" mapstructure:"etcd"`
}

// NewOptions creates a new Options object with default parameters.
func NewOptions() *Options {
	o := Options{
		Log:           log.NewOptions(),
		Elasticsearch: genericoptions.NewElasticsearchOptions(),
		Nsq:           genericoptions.NewNsqOptionsOptions(),
		Etcd:          genericoptions.NewEtcdOptions(),
	}

	return &o
}

// Flags returns flags for a specific APIServer by section name.
func (o *Options) Flags() (fss cliflag.NamedFlagSets) {
	o.Log.AddFlags(fss.FlagSet("logs"))
	o.Elasticsearch.AddFlags(fss.FlagSet("elasticsearch"))
	o.Nsq.AddFlags(fss.FlagSet("nsq"))
	o.Etcd.AddFlags(fss.FlagSet("etcd"))
	return fss
}
