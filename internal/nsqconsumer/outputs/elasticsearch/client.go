package elasticsearch

import (
	"sync"

	"github.com/marmotedu/iam/pkg/log"
	"github.com/olivere/elastic/v7"
)

type Client struct {
	client   *elastic.Client
	addrs    []string
	username string
	password string

	mux sync.Mutex
}

func NewClient(config *Config) (*Client, error) {
	c := &Client{
		addrs:    config.Addrs,
		username: config.Username,
		password: config.Password,
	}
	return c, nil
}

func (c *Client) Client() *elastic.Client {
	return c.client
}

func (c *Client) Connect() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	log.Infof("connect: %v\n", c.addrs)
	optionFuncs := []elastic.ClientOptionFunc{elastic.SetURL(c.addrs...)}
	if c.username != "" {
		optionFuncs = append(optionFuncs, elastic.SetBasicAuth(c.username, c.password))
	}
	client, err := elastic.NewClient(optionFuncs...)
	if err != nil {
		return err
	}

	c.client = client
	return nil
}

func (c *Client) Close() error {
	log.Info("Close")
	return nil
}
