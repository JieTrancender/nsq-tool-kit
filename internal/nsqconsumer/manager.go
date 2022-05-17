package nsqconsumer

import (
	"context"
	"fmt"
	"time"

	"github.com/marmotedu/iam/pkg/log"
	"github.com/nsqio/go-nsq"
	"github.com/olivere/elastic/v7"

	"github.com/JieTrancender/nsq-tool-kit/internal/nsqconsumer/config"
	es "github.com/JieTrancender/nsq-tool-kit/internal/nsqconsumer/outputs/elasticsearch"
)

type manager struct {
	esConfig *es.Config
}

func createConsumerManager(cfg *config.Config) (*manager, error) {
	return &manager{
		esConfig: &es.Config{
			Addrs:    cfg.Elasticsearch.Addrs,
			Username: cfg.Elasticsearch.Username,
			Password: cfg.Elasticsearch.Password,
		},
	}, nil
}

func (m *manager) Run() error {
	client, err := es.NewClient(m.esConfig)
	if err != nil {
		log.Errorf("New elasticsearch client fail: %v", err)
		return err
	}

	err = client.Connect()
	if err != nil {
		return err
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
	return nil
}

func pubDoc(client *es.Client, reqChan <-chan *elastic.BulkIndexRequest, q chan<- struct{}) {
	bulkReq := elastic.NewBulkService(client.Client())
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
