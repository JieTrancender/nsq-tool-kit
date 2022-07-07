package nsqconsumer

import (
	"context"
	"encoding/json"
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
	"github.com/JieTrancender/nsq-tool-kit/internal/nsqconsumer/store"
	"github.com/JieTrancender/nsq-tool-kit/internal/nsqconsumer/store/etcd"
	genericoptions "github.com/JieTrancender/nsq-tool-kit/internal/pkg/options"
)

type manager struct {
	gs  *shutdown.GracefulShutdown
	cfg *config.Config

	esConfig *es.Config
	esClient *es.Client

	nsqConfig *nsq.Config

	topics map[string]*Consumer

	msgChan chan *message.Message

	storeIns store.Factory
}

func createConsumerManager(cfg *config.Config) (*manager, error) {
	gs := shutdown.New()
	gs.AddShutdownManager(posixsignal.NewPosixSignalManager())

	nsqConfig := nsq.NewConfig()
	nsqConfig.UserAgent = fmt.Sprintf("nsq-tool-kit/%s go-nsq/%s", "0.0.1", nsq.VERSION)
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

	storeIns, err := etcd.GetEtcdFactoryOr(m.cfg.Etcd, nil)
	if err != nil {
		return err
	}
	store.SetClient(storeIns)

	o, err := storeIns.Nsqs().Get(context.Background(), m.cfg.Etcd.Path)
	if err != nil {
		return err
	}
	m.cfg.Nsq = o

	err = storeIns.Watch(context.Background(), "", m.updateNsqConfig)
	if err != nil {
		return err
	}
	m.storeIns = storeIns

	return nil
}

func (m *manager) updateNsqConfig(ctx context.Context, key, oldvalue, value []byte) {
	log.Infof("manager update nsq conifg %s", string(key))
	log.Infof("%s %s", string(key), m.storeIns.Nsqs().GetKey(m.cfg.Etcd.Path))
	if string(key) == m.storeIns.Nsqs().GetKey(m.cfg.Etcd.Path) {
		var o genericoptions.NsqOptions
		if err := json.Unmarshal(value, &o); err != nil {
			log.Errorf("failed to unmarshal to nsq options struct, data: %v", string(value))
			return
		}
		m.cfg.Nsq = &o
		m.updateTopics()
	}
}

func (m *manager) updateTopics() {
	m.nsqConfig.DialTimeout = time.Duration(m.cfg.Nsq.DialTimeout) * time.Second
	m.nsqConfig.ReadTimeout = time.Duration(m.cfg.Nsq.ReadTimeout) * time.Second
	m.nsqConfig.WriteTimeout = time.Duration(m.cfg.Nsq.WriteTimeout) * time.Second
	m.nsqConfig.MaxInFlight = m.cfg.Nsq.MaxInFlight

	for _, topic := range m.cfg.Nsq.Topics {
		if _, ok := m.topics[topic]; ok {
			continue
		}
		log.Infof("launch topic %s", topic)
		nsqConsumer, err := nsq.NewConsumer(topic, m.cfg.Nsq.Channel, m.nsqConfig)
		if err != nil {
			log.Errorf("nsq.NewConsumer fail: %v", err)
			continue
		}
		nsqConsumer.SetLogger(log.StdInfoLogger(), nsq.LogLevelInfo)
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

		go func(consumer *Consumer, msgChan chan<- *message.Message) {
			consumer.Run(msgChan)
		}(consumer, m.msgChan)
	}
}

func (m *manager) launch() error {
	msgChan := make(chan *message.Message)
	m.msgChan = msgChan
	go m.esClient.Run(msgChan)

	m.updateTopics()

	stopCh := make(chan struct{})
	if err := m.gs.Start(); err != nil {
		log.Fatalf("start shutdown manager failed: %s", err.Error())
	}

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
