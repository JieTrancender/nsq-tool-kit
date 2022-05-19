package nsq

import (
	"context"
	"encoding/json"
	"strings"
	"sync"
	"time"

	"github.com/elastic/beats/v7/libbeat/logp"
	"github.com/elastic/beats/v7/libbeat/outputs"
	"github.com/elastic/beats/v7/libbeat/outputs/codec"
	"github.com/elastic/beats/v7/libbeat/publisher"
	"github.com/nsqio/go-nsq"
)

type client struct {
	log      *logp.Logger
	observer outputs.Observer
	outputs.NetworkClient
	codec codec.Codec
	index string

	// for nsq
	nsqd       string
	topic      string
	producer   *nsq.Producer
	config     *nsq.Config
	filterKeys []string
	ignoreKeys []string

	mux sync.Mutex
	// wg  sync.WaitGroup
}

func newNsqClient(
	observer outputs.Observer,
	nsqd string,
	index string,
	topic string,
	writer codec.Codec,
	writeTimeout time.Duration,
	dialTimeout time.Duration,
	filterKeys []string,
	ignoreKeys []string,
) (*client, error) {
	cfg := nsq.NewConfig()
	cfg.WriteTimeout = writeTimeout
	cfg.DialTimeout = dialTimeout
	c := &client{
		log:        logp.NewLogger(logSelector),
		observer:   observer,
		nsqd:       nsqd,
		topic:      topic,
		index:      strings.ToLower(index),
		codec:      writer,
		config:     cfg,
		filterKeys: filterKeys,
		ignoreKeys: ignoreKeys,
	}

	return c, nil
}

func (c *client) Connect() error {
	c.mux.Lock()
	defer c.mux.Unlock()

	c.log.Debugf("connect: %v", c.nsqd)
	producer, err := nsq.NewProducer(c.nsqd, c.config)
	if err != nil {
		c.log.Errorf("nsq connect fails with: %+v", err)
		return err
	}

	// todo: set logger
	// pruducer.SetLogger(c.log, LogLevelInfo)
	c.producer = producer

	// c.wg.Add(2)
	// go c.successWorker(producer.Successes())
	// go c.errorWorker(producer.Errors())

	return nil
}

func (c *client) Publish(_ context.Context, batch publisher.Batch) error {
	events := batch.Events()
	c.observer.NewBatch(len(events))

	st := c.observer
	msgs, dealed, err := c.buildNsqMessages(events)
	dropped := len(events) - dealed
	// c.log.Info("events=%v msgs=%v", len(events), len(msgs))
	if err != nil {
		c.log.Errorf("[main:nsq] c.buildNsqMessages %v", err)
		c.observer.Failed(len(events))
		batch.RetryEvents(events)
		return nil
	}

	if len(msgs) > 0 {
		// nsq send failed do retry...
		err = c.producer.MultiPublish(c.topic, msgs)
		if err != nil {
			c.observer.Failed(len(events))
			batch.RetryEvents(events)
			return err
		}
	}
	batch.ACK()

	st.Dropped(dropped)
	st.Acked(dealed)
	return err
}

func (c *client) buildNsqMessages(events []publisher.Event) ([][]byte, int, error) {
	length := len(events)
	msgs := make([][]byte, length)
	var count int
	var err error
	var isIgnore bool
	var isFilter bool
	var msg string
	var dealed int

	for idx := 0; idx < length; idx++ {
		event := events[idx].Content
		serializedEvent, nerr := c.codec.Encode(c.index, &event)
		if nerr != nil {
			c.log.Errorf("[main:nsq] c.codec.Encode fail %v", nerr)
			err = nerr
			continue
		}

		data := make(map[string]interface{})
		nerr = json.Unmarshal(serializedEvent, &data)
		if nerr != nil {
			c.log.Errorf("[main:nsq] json.Unmarshal fail %v", nerr)
			err = nerr
			continue
		}

		// isIgnore = false
		isFilter = false
		msg = data["message"].(string)
		for _, value := range c.filterKeys {
			if strings.Contains(msg, value) {
				isFilter = true

				// 被忽略的消息也认为是处理掉了
				// Ignored msgs are thinked success
				dealed++
				break
			}
		}

		// if not set filter keys, all msg can filter
		if !isFilter && len(c.filterKeys) > 0 {
			continue
		}

		isIgnore = false
		for _, value := range c.ignoreKeys {
			if strings.Contains(msg, value) {
				isIgnore = true
				dealed++
				break
			}
		}
		if isIgnore {
			continue
		}

		tmp := string(serializedEvent)
		msgs[count] = []byte(tmp)
		count++
		dealed++
	}

	return msgs[:count], dealed, err
}

func (c *client) String() string {
	return "NSQD"
}
