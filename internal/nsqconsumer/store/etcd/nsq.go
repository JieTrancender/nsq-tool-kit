package etcd

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/marmotedu/errors"

	genericoptions "github.com/JieTrancender/nsq-tool-kit/internal/pkg/options"
)

type nsqs struct {
	ds *datastore
}

func newNsqs(ds *datastore) *nsqs {
	return &nsqs{ds: ds}
}

var keyNsq = "/%v"

func (n *nsqs) GetKey(name string) string {
	return fmt.Sprintf(keyNsq, name)
}

func (n *nsqs) Get(ctx context.Context, key string) (*genericoptions.NsqOptions, error) {
	resp, err := n.ds.Get(ctx, n.GetKey(key))
	if err != nil {
		return nil, err
	}

	var o genericoptions.NsqOptions
	if err := json.Unmarshal(resp, &o); err != nil {
		return nil, errors.Wrap(err, "unmarshal to nsq options struct failed")
	}

	return &o, nil
}
