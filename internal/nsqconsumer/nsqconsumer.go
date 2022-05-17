package nsqconsumer

import (
	"encoding/json"

	"github.com/marmotedu/iam/pkg/log"
	"github.com/nsqio/go-nsq"

	"github.com/JieTrancender/nsq-tool-kit/internal/nsqconsumer/message"
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

func (c *Consumer) Stop() {
	c.consumer.Stop()
	<-c.consumer.StopChan
}

func (c *Consumer) Run(msgChan chan<- *message.Message) {
	log.Infof("%s consumer running", c.topic)
	for {
		select {
		case <-c.done:
			return
		case m := <-c.msgChan:
			data := make(map[string]interface{})
			err := json.Unmarshal(m.Body, &data)
			// m.Finish()
			if err != nil {
				log.Infof("Unmarshal nsq message fail: %v", err)
				m.Finish()
			} else {
				msgChan <- message.NewMessage(m, c.topic)
			}
		}
	}
}
