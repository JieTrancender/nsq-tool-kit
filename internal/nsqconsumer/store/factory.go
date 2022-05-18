package store

import "context"

var client Factory

// EtcdModifyEventFunc defines etcd update event function handler.
type EtcdModifyEventFunc func(ctx context.Context, key, oldvalue, value []byte)

// Factory defines the consumer storage interface.
type Factory interface {
	Nsqs() NsqStore
	GetKey(string) string
	Watch(context.Context, string, EtcdModifyEventFunc) error
	Close() error
}

// Client returns the store client instance.
func Client() Factory {
	return client
}

// SetClient sets the consumer store client.
func SetClient(factory Factory) {
	client = factory
}
