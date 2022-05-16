package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/jehiah/go-strftime"
	"github.com/olivere/elastic/v7"
)

func genIndexName(topic string) string {
	now := time.Now()
	return strftime.Format(fmt.Sprintf("%s-%%y.%%m.%%d", topic), now)
}

type message struct {
	GamePlatform string `json:"gamePlatform"`
	NodeName     string `json:"nodeName"`
	Message      string `json:"message"`
}

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
		addrs:     config.Addrs,
		indexType: config.IndexType,
		indexName: config.IndexName,
		username:  config.Username,
		password:  config.Password,
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

func main() {
	config := &Config{
		Addrs:     []string{"http://127.0.0.1:9200"},
		IndexName: "dev_test_v1-%Y.%m.%d",
		IndexType: "doc",
		Username:  "root",
		Password:  "123456",
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

	reqChan := make(chan *elastic.BulkIndexRequest, 0)
	q := make(chan struct{})
	go genDoc(reqChan)
	go pubDoc(client, reqChan, q)

	<-q
}

func genDoc(reqChan chan<- *elastic.BulkIndexRequest) {
	topic := "dev_test_v1"
	for i := 1; i <= 10000; i++ {
		data := &message{
			GamePlatform: "dev_test",
			NodeName:     "worldd",
			Message:      fmt.Sprintf("i am %d", i),
		}
		req := elastic.NewBulkIndexRequest().
			Index(genIndexName(topic)).
			Doc(data)
		reqChan <- req
	}
	close(reqChan)
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
