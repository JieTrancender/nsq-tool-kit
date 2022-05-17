package elasticsearch

import (
	"context"
	"sync"
	"time"

	"github.com/marmotedu/iam/pkg/log"
	"github.com/olivere/elastic/v7"

	"github.com/JieTrancender/nsq-tool-kit/internal/nsqconsumer/message"
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
	c.client.Stop()
	return nil
}

func (c *Client) Run(msgChan <-chan *message.Message, q chan<- struct{}) {
	log.Infof("elasticsearch %v publish", c.addrs)

	timeout := time.Second * 1
	var timer *time.Timer
	maxCount := 100
	msgList := make([]*message.Message, 0)
	for {
		if timer == nil {
			timer = time.NewTimer(timeout)
		} else {
			if len(msgList) == maxCount {
				timer.Reset(timeout)
				c.Publish(msgList)
				msgList = make([]*message.Message, 0)
			}
		}

		select {
		case m, ok := <-msgChan:
			if !ok {
				log.Infof("elasticsearch %v close", c.addrs)
				q <- struct{}{}
				return
			}
			msgList = append(msgList, m)
		case <-timer.C:
			if len(msgList) > 0 {
				c.Publish(msgList)
				msgList = make([]*message.Message, 0)
			}
			timer.Reset(timeout)
		}
	}
}

func (c *Client) Publish(msgList []*message.Message) {
	bulkReq := elastic.NewBulkService(c.client)
	defer bulkReq.Reset()
	log.Infof("len msgList = %d", len(msgList))
	for i, m := range msgList {
		log.Infof("publish %d %v", i, m)
		req := elastic.NewBulkIndexRequest().Index(m.GetTopic()).Doc(m.GetData())
		bulkReq = bulkReq.Add(req)
	}

	bulkResp, err := bulkReq.Do(context.Background())
	if err != nil {
		log.Infof("Do bulk request fail: %v", err)
		for _, message := range msgList {
			message.GetData().Requeue(-1)
		}
		return
	}
	for _, m := range msgList {
		m.GetData().Finish()
	}
	log.Infof("耗时: %v, 索引数目: %d", bulkResp.Took, len(bulkResp.Items))
}
