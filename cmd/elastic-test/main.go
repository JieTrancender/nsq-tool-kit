package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jehiah/go-strftime"
	"github.com/nsqio/go-nsq"
	"github.com/olivere/elastic/v7"
)

type client struct {
	client    *elastic.Client
	addrs     []string
	indexType string
	indexName string
	username  string
	password  string

	mux sync.Mutex
}

func newElasticsearchClient(config *Config) (*client, error) {
	c := &client{
		addrs:    config.Addrs,
		username: config.Username,
		password: config.Password,
	}
	return c, nil
}

func (c *client) Connect() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	log.Printf("connect: %v\n", c.addrs)
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

func (c *client) Close() error {
	log.Println("Close")
	return nil
}

type Consumer struct {
	topic    string
	consumer *nsq.Consumer
	done     chan struct{}
	msgChan  chan *nsq.Message
}

func (c *Consumer) HandleMessage(m *nsq.Message) error {
	m.DisableAutoResponse()
	c.msgChan <- m
	return nil
}

func (c *Consumer) indexName() string {
	now := time.Now()
	return strftime.Format(fmt.Sprintf("%s-%%y.%%m.%%d", c.topic), now)
}

func (c *Consumer) Stop() {
	c.consumer.Stop()
	<-c.consumer.StopChan
}

func (c *Consumer) Run(reqChan chan<- *elastic.BulkIndexRequest) {
	fmt.Println("Consumer Run")
	for {
		select {
		case <-c.done:
			close(reqChan)
			return
		case m := <-c.msgChan:
			data := make(map[string]interface{})
			err := json.Unmarshal(m.Body, &data)
			m.Finish()
			if err != nil {
				fmt.Printf("Unmarshal fail: %v", err)
			} else {
				req := elastic.NewBulkIndexRequest().
					Index(c.indexName()).
					Doc(data)
				reqChan <- req
			}
		}
	}
}

func main() {
	config := &Config{
		Addrs:    []string{"http://127.0.0.1:9200"},
		Username: "root",
		Password: "123456",
	}

	client, err := newElasticsearchClient(config)
	if err != nil {
		panic(err)
	}

	err = client.Connect()
	if err != nil {
		panic(err)
	}
	defer client.Close()

	cfg := nsq.NewConfig()
	cfg.UserAgent = fmt.Sprintf("nsq-tool-kit/%s go-nsq/%s", "0.0.1", nsq.VERSION)
	cfg.DialTimeout = 6 * time.Second
	cfg.ReadTimeout = 60 * time.Second
	cfg.WriteTimeout = 6 * time.Second
	cfg.MaxInFlight = 200

	topic := "dev_test"
	nsqConsumer, err := nsq.NewConsumer(topic, "nsq_tool_kit", cfg)
	if err != nil {
		panic(err)
	}

	consumer := &Consumer{
		topic:    topic,
		consumer: nsqConsumer,
		done:     make(chan struct{}),
		msgChan:  make(chan *nsq.Message),
	}

	nsqConsumer.AddHandler(consumer)
	err = nsqConsumer.ConnectToNSQLookupds([]string{"http://127.0.0.1:4161"})
	if err != nil {
		panic(err)
	}
	defer consumer.Stop()

	reqChan := make(chan *elastic.BulkIndexRequest)
	go func(consumer *Consumer, reqChan chan<- *elastic.BulkIndexRequest) {
		consumer.Run(reqChan)
	}(consumer, reqChan)

	q := make(chan struct{})
	go pubDoc(client, reqChan, q)

	<-q
}

func pubDoc(client *client, reqChan <-chan *elastic.BulkIndexRequest, q chan<- struct{}) {
	bulkReq := elastic.NewBulkService(client.client)
	timeout := time.Millisecond * 500
	var timer *time.Timer
	maxCount := 100
	count := 0
	for {
		if timer == nil {
			timer = time.NewTimer(timeout)
		} else {
			if count == maxCount {
				count = 0
				timer.Reset(timeout)
				execRequest(bulkReq)
			}
		}

		select {
		case req, ok := <-reqChan:
			if !ok {
				fmt.Println("End.")
				q <- struct{}{}
				return
			}
			bulkReq = bulkReq.Add(req)
			count = count + 1
		case <-timer.C:
			if count > 0 {
				execRequest(bulkReq)
				count = 0
			}
			timer.Reset(timeout)
		}
	}
}

func execRequest(bulkReq *elastic.BulkService) bool {
	bulkResp, err := bulkReq.Do(context.Background())
	defer bulkReq.Reset()
	if err != nil {
		fmt.Printf("do bulk req fail %v\n", err)
		return false
	}
	fmt.Println("耗时: ", bulkResp.Took, "索引了: ", len(bulkResp.Items))
	return true
}
