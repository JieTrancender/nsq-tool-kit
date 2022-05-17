package options

import (
	cliflag "github.com/marmotedu/component-base/pkg/cli/flag"
	"github.com/marmotedu/iam/pkg/log"
)

// Options runs a iam api server.
type Options struct {
	Log *log.Options `json:"log" mapstructure:"log"`
}

// NewOptions creates a new Options object with default parameters.
func NewOptions() *Options {
	o := Options{
		Log: log.NewOptions(),
	}

	return &o
}

// Flags returns flags for a specific APIServer by section name.
func (o *Options) Flags() (fss cliflag.NamedFlagSets) {
	o.Log.AddFlags(fss.FlagSet("logs"))
	return fss
}
