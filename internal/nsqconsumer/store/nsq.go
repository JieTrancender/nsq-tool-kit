package store

import (
	"context"

	genericoptions "github.com/JieTrancender/nsq-tool-kit/internal/pkg/options"
)

type NsqStore interface {
	Get(ctx context.Context, key string) (*genericoptions.NsqOptions, error)
	GetKey(key string) string
}
