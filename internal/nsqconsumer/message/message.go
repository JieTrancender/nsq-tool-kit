package message

import "github.com/nsqio/go-nsq"

type Message struct {
	data  *nsq.Message
	topic string
}

func NewMessage(data *nsq.Message, topic string) *Message {
	return &Message{data: data, topic: topic}
}

func (m *Message) GetData() *nsq.Message {
	return m.data
}

func (m *Message) GetTopic() string {
	return m.topic
}
