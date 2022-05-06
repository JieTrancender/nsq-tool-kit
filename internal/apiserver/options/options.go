package options

import (
	cliflag "github.com/marmotedu/component-base/pkg/cli/flag"
)

// Options runs a iam api server.
type Options struct {
}

// NewOptions creates a new Options object with default parameters.
func NewOptions() *Options {
	o := Options{}

	return &o
}

// Flags returns flags for a specific APIServer by section name.
func (o *Options) Flags() (fss cliflag.NamedFlagSets) {
	return fss
}
