package nsqconsumer

import (
	"fmt"
	"runtime"
	"time"

	"github.com/marmotedu/iam/pkg/log"
	"github.com/marmotedu/iam/pkg/shutdown"
	"github.com/marmotedu/iam/pkg/shutdown/shutdownmanagers/posixsignal"
	"github.com/nsqio/go-nsq"

	"github.com/JieTrancender/nsq-tool-kit/internal/nsqconsumer/config"
	"github.com/JieTrancender/nsq-tool-kit/internal/nsqconsumer/message"
	es "github.com/JieTrancender/nsq-tool-kit/internal/nsqconsumer/outputs/elasticsearch"
)

type manager struct {
	gs  *shutdown.GracefulShutdown
	cfg *config.Config

	esConfig *es.Config
	esClient *es.Client

	nsqConfig *nsq.Config

	topics map[string]*Consumer

	msgChan chan *message.Message
}

func createConsumerManager(cfg *config.Config) (*manager, error) {
	gs := shutdown.New()
	gs.AddShutdownManager(posixsignal.NewPosixSignalManager())

	nsqConfig := nsq.NewConfig()
	nsqConfig.UserAgent = fmt.Sprintf("nsq-tool-kit/%s go-nsq/%s", "0.0.1", nsq.VERSION)
	nsqConfig.DialTimeout = time.Duration(cfg.Nsq.DialTimeout) * time.Second
	nsqConfig.ReadTimeout = time.Duration(cfg.Nsq.ReadTimeout) * time.Second
	nsqConfig.WriteTimeout = time.Duration(cfg.Nsq.WriteTimeout) * time.Second
	nsqConfig.MaxInFlight = cfg.Nsq.MaxInFlight
	return &manager{
		gs:  gs,
		cfg: cfg,
		esConfig: &es.Config{
			Addrs:    cfg.Elasticsearch.Addrs,
			Username: cfg.Elasticsearch.Username,
			Password: cfg.Elasticsearch.Password,
		},
		nsqConfig: nsqConfig,
		topics:    make(map[string]*Consumer),
	}, nil
}

func (m *manager) initialize() error {
	client, err := es.NewClient(m.esConfig)
	if err != nil {
		log.Errorf("New elasticsearch client fail: %v", err)
		return err
	}

	err = client.Connect()
	if err != nil {
		return err
	}
	m.esClient = client

	return nil
}

func (m *manager) launch() error {
	for _, topic := range m.cfg.Nsq.Topics {
		log.Infof("launch topic %s", topic)
		nsqConsumer, err := nsq.NewConsumer(topic, m.cfg.Nsq.Channel, m.nsqConfig)
		if err != nil {
			log.Errorf("nsq.NewConsumer fail: %v", err)
			continue
		}
		consumer := &Consumer{
			topic:    topic,
			consumer: nsqConsumer,
			done:     make(chan struct{}),
			msgChan:  make(chan *nsq.Message),
		}
		nsqConsumer.AddConcurrentHandlers(consumer, runtime.NumCPU())
		err = nsqConsumer.ConnectToNSQLookupds(m.cfg.Nsq.LookupdHttpAddresses)
		if err != nil {
			log.Errorf("ConnectToNSQLookupd fail: %v", err)
			continue
		}
		m.topics[topic] = consumer
	}

	stopCh := make(chan struct{})
	if err := m.gs.Start(); err != nil {
		log.Fatalf("start shutdown manager failed: %s", err.Error())
	}

	msgChan := make(chan *message.Message)
	m.msgChan = msgChan
	for _, consumer := range m.topics {
		go func(consumer *Consumer, msgChan chan<- *message.Message) {
			consumer.Run(msgChan)
		}(consumer, msgChan)
	}

	go m.esClient.Run(msgChan)

	m.gs.AddShutdownCallback(shutdown.ShutdownFunc(func(string) error {
		m.Stop()
		return nil
	}))

	<-stopCh
	return nil
}

func (m *manager) Run() error {
	err := m.initialize()
	if err != nil {
		return err
	}

	return m.launch()
}

func (m *manager) Stop() {
	log.Info("manager Stopping")
	for _, consumer := range m.topics {
		consumer.Stop()
	}

	close(m.msgChan)

	// 最后关闭elasticsearch
	m.esClient.Close()
	log.Info("manager stopped")
}
