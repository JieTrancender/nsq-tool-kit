package options

import (
	"fmt"

	"github.com/spf13/pflag"
)

// EtcdOptions defines options for etcd cluster.
type EtcdOptions struct {
	Endpoints      []string `json:"endpoints" mapstructure:"endpoints"`
	Timeout        int      `json:"timeout" mapstructure:"timeout"`
	RequestTimeout int      `json:"request-time" mapstructure:"request-time"`
	LeaseExpire    int      `json:"lease-expire" mapstructure:"lease-expire"`
	Username       string   `json:"username" mapstructure:"username"`
	Password       string   `json:"password" mapstructure:"password"`
	UseTLS         bool     `json:"use-tls" mapstructure:"use-tls"`
	Namespece      string   `json:"namespace" mapstructure:"namespace"`
	Path           string   `json:"path" mapstructure:"path"`
}

// NewEtcdOptions creates a `zero` value instance.
func NewEtcdOptions() *EtcdOptions {
	return &EtcdOptions{
		Timeout:     5,
		LeaseExpire: 5,
		Path:        "dev",
	}
}

func (o *EtcdOptions) Validate() []error {
	errs := []error{}
	if len(o.Endpoints) == 0 {
		errs = append(errs, fmt.Errorf("etcd endpoints can not be empty"))
	}

	return errs
}

// AddFlags adds flags related to etcd storage to the specified FlagSet.
func (o *EtcdOptions) AddFlags(fs *pflag.FlagSet) {
	fs.StringSliceVar(&o.Endpoints, "etcd.endpoints", o.Endpoints, "Endpoints of etcd cluster.")
	fs.StringVar(&o.Username, "etcd.username", o.Username, "Username of etcd cluster.")
	fs.StringVar(&o.Password, "etcd.password", o.Password, "Password of etcd cluster.")
	fs.IntVar(&o.LeaseExpire, "etcd.lease-expire", o.LeaseExpire, "Etcd expire timeout in seconds.")
	fs.StringVar(&o.Namespece, "etcd.namespace", o.Namespece, "Etcd storage namespace.")
	fs.StringVar(&o.Path, "etcd.path", o.Path, "Path of config.")
}
