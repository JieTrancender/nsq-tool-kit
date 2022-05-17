package nsqconsumer

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/jehiah/go-strftime"
	"github.com/nsqio/go-nsq"
	"github.com/olivere/elastic/v7"
)

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
