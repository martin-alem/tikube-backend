package kafka_client

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"os"
	"path/filepath"
)

func KafkaClient(groupId string) (sarama.AsyncProducer, sarama.ConsumerGroup, error) {

	var dir, _ = os.Getwd()

	var kafkaHost = os.Getenv("KAFKA_HOST")
	var kafkaPort = os.Getenv("KAFKA_PORT")

	var keypair, err = tls.LoadX509KeyPair(filepath.Join(dir, "/shared/kafka_client/certs/user-access-certificate.crt"), filepath.Join(dir, "/shared/kafka_client/certs/user-access-key.key"))
	if err != nil {
		return nil, nil, err
	}

	caCert, err := os.ReadFile(filepath.Join(dir, "/shared/kafka_client/certs/ca-certificate.crt"))
	if err != nil {
		return nil, nil, err
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{keypair},
		RootCAs:      caCertPool,
	}

	if kafkaHost == "" || kafkaPort == "" {
		log.Fatal("Kafka credentials are not set in environment variables")
	}

	// init config, enable errors and notifications
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	config.Net.TLS.Enable = true
	config.Net.TLS.Config = tlsConfig
	config.Version = sarama.V3_6_0_0

	brokers := []string{fmt.Sprintf("%s:%s", kafkaHost, kafkaPort)}

	producer, err := sarama.NewAsyncProducer(brokers, config)
	if err != nil {
		log.Fatal(err)
	}

	consumer, err := sarama.NewConsumerGroup(brokers, groupId, config)
	if err != nil {
		log.Fatal(err)
	}

	return producer, consumer, nil
}

func SendLogToKafka(logMessage string, topic string, producer sarama.AsyncProducer) {

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Value: sarama.StringEncoder(logMessage),
	}

	select {
	case producer.Input() <- msg:
	case err := <-producer.Errors():
		log.Println("Failed to write log to Kafka:", err)
	}
}

type ConsumerGroupHandler struct {
	Processor func(ctx context.Context, message *sarama.ConsumerMessage) error
}

func (ConsumerGroupHandler) Setup(_ sarama.ConsumerGroupSession) error   { return nil }
func (ConsumerGroupHandler) Cleanup(_ sarama.ConsumerGroupSession) error { return nil }
func (h ConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		_ = h.Processor(context.Background(), msg)
		sess.MarkMessage(msg, "")
	}
	return nil
}
