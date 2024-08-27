package publisher

import (
	"github.com/IBM/sarama"

	"github.com/theopenlane/core/pkg/events/kafka/kafkaconfig"
)

// KafkaPublisher is a publisher that sends messages to a Kafka topic
type KafkaPublisher struct {
	// Broker is a list of Kafka brokers
	Broker []string
	// Config is the configuration for the Kafka event source
	Config kafkaconfig.Config
}

// NewKafkaPublisher creates a new KafkaPublisher
func NewKafkaPublisher(broker []string) *KafkaPublisher {
	return &KafkaPublisher{
		Broker: broker,
	}
}

// Publisher is an interface for publishing messages
type Publisher interface {
	Publish(topic string, message []byte) error
}

// Publish satisfies the Publisher interface
func (kp *KafkaPublisher) Publish(topic string, message []byte) error {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = 5
	config.Producer.Return.Successes = true

	producer, err := sarama.NewSyncProducer(kp.Broker, config)
	if err != nil {
		return err
	}
	defer producer.Close()

	_, _, err = producer.SendMessage(&sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.ByteEncoder(message),
	})

	return err
}
