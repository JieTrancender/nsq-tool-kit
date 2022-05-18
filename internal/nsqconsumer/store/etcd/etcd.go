package etcd

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/marmotedu/errors"
	"github.com/marmotedu/iam/pkg/log"
	"go.etcd.io/etcd/api/v3/mvccpb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"

	"github.com/JieTrancender/nsq-tool-kit/internal/nsqconsumer/store"
	genericoptions "github.com/JieTrancender/nsq-tool-kit/internal/pkg/options"
)

// EtcdWatcher defines an etcd watcher.
type EtcdWatcher struct {
	watcher clientv3.Watcher
	cancel  context.CancelFunc
}

type datastore struct {
	cli             *clientv3.Client
	requestTimeout  time.Duration
	leaseTTLTimeout int

	leaseID             clientv3.LeaseID
	onKeeypaliveFailure func()
	leaseLiving         bool

	watchers  map[string]*EtcdWatcher
	namespace string
}

func (ds *datastore) Nsqs() store.NsqStore {
	return newNsqs(ds)
}

// Close closes the etcd store client.
func (ds *datastore) Close() error {
	if ds.cli != nil {
		return ds.cli.Close()
	}
	return nil
}

var (
	etcdFactory store.Factory
	once        sync.Once
)

func defaultOnKeepAliveFailed() {
	log.Warn("etcd store keepalive failed")
}

// GetEtcdFactoryOr creates an etcd factory store with given options.
func GetEtcdFactoryOr(opt *genericoptions.EtcdOptions, onKeeypaliveFailure func()) (store.Factory, error) {
	if opt == nil && etcdFactory == nil {
		return nil, fmt.Errorf("failed to get etcd store factory")
	}

	var err error
	once.Do(func() {
		var (
			cli *clientv3.Client
		)
		if opt.UseTLS {
			err = fmt.Errorf("enable etcd factory tls but not support now")
			return
		}

		ds := &datastore{}
		if onKeeypaliveFailure == nil {
			onKeeypaliveFailure = defaultOnKeepAliveFailed
		}
		ds.onKeeypaliveFailure = onKeeypaliveFailure
		cli, err = clientv3.New(clientv3.Config{
			Endpoints:   opt.Endpoints,
			DialTimeout: time.Duration(opt.Timeout) * time.Second,
			Username:    opt.Username,
			Password:    opt.Password,

			DialOptions: []grpc.DialOption{
				grpc.WithBlock(),
			},
		})
		if err != nil {
			return
		}

		ds.cli = cli
		ds.requestTimeout = time.Duration(opt.RequestTimeout) * time.Second
		ds.leaseTTLTimeout = opt.LeaseExpire
		ds.watchers = make(map[string]*EtcdWatcher)
		ds.namespace = opt.Namespece

		err = ds.startSession()
		if err != nil {
			if e := ds.Close(); e != nil {
				log.Errorf("etcd store client close failed %s", e)
			}
			return
		}
		etcdFactory = ds
	})

	if etcdFactory == nil || err != nil {
		return nil, fmt.Errorf("failed to get etcd store factory, etcdFactory: %+v, error: %w", etcdFactory, err)
	}

	return etcdFactory, nil
}

func (ds *datastore) startSession() error {
	ctx := context.TODO()
	resp, err := ds.cli.Grant(ctx, int64(ds.leaseTTLTimeout))
	if err != nil {
		return errors.Wrap(err, "creates new lease failed")
	}
	ds.leaseID = resp.ID

	ch, err := ds.cli.KeepAlive(ctx, ds.leaseID)
	if err != nil {
		return errors.Wrap(err, "keep alive failed")
	}
	ds.leaseLiving = true

	go func() {
		for {
			if _, ok := <-ch; !ok {
				ds.leaseLiving = false
				log.Errorf("failed to keepalive session")
				if ds.onKeeypaliveFailure != nil {
					ds.onKeeypaliveFailure()
				}
				break
			}
		}
	}()
	return nil
}

func (ds *datastore) Client() *clientv3.Client {
	return ds.cli
}

func (ds *datastore) SessionLiving() bool {
	return ds.leaseLiving
}

func (ds *datastore) RestartSession() error {
	if ds.leaseLiving {
		return fmt.Errorf("session is living, can't restart")
	}
	return ds.startSession()
}

func (ds *datastore) GetKey(key string) string {
	if len(ds.namespace) > 0 {
		return fmt.Sprintf("%s%s", ds.namespace, key)
	}
	return key
}

func (ds *datastore) Get(ctx context.Context, key string) ([]byte, error) {
	nctx, cancel := context.WithTimeout(ctx, ds.requestTimeout)
	defer cancel()

	key = ds.GetKey(key)
	resp, err := ds.cli.Get(nctx, key)
	if err != nil {
		return nil, errors.Wrap(err, "get key from etcd failed")
	}
	if len(resp.Kvs) == 0 {
		return nil, fmt.Errorf("no such key")
	}
	return resp.Kvs[0].Value, nil
}

// Cancel cancels etcd client.
func (w *EtcdWatcher) Cancel() {
	w.watcher.Close()
	w.cancel()
}

// Watch watchs etcd.
func (ds *datastore) Watch(
	ctx context.Context,
	prefix string,
	onModify store.EtcdModifyEventFunc,
) error {
	if _, ok := ds.watchers[prefix]; ok {
		return fmt.Errorf("watch prefix %s already registered", prefix)
	}

	watcher := clientv3.NewWatcher(ds.cli)
	nctx, cancel := context.WithCancel(ctx)
	ds.watchers[prefix] = &EtcdWatcher{
		watcher: watcher,
		cancel:  cancel,
	}

	prefix = ds.GetKey(prefix)
	rch := watcher.Watch(nctx, prefix, clientv3.WithPrefix(), clientv3.WithPrevKV())
	go func() {
		for wresp := range rch {
			for _, ev := range wresp.Events {
				key := ev.Kv.Key[len(ds.namespace):]
				if ev.PrevKv != nil {
					switch ev.Type {
					case mvccpb.PUT:
						onModify(nctx, key, ev.PrevKv.Value, ev.Kv.Value)
					}
				}
			}
		}
		log.Infof("stop watching %s", prefix)
	}()

	return nil
}

func (ds *datastore) Unwatch(prefix string) {
	watcher, ok := ds.watchers[prefix]
	if ok {
		log.Debugf("unwatch %s", prefix)
		watcher.Cancel()
		delete(ds.watchers, prefix)
	} else {
		log.Debugf("prefix %s not watched!", prefix)
	}
}
