package kafka_test

import (
	"testing"

	"github.com/Shopify/sarama"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

func TestStorage(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Storage Suite")
}

var (
	kafkaClient          sarama.Client
	syncProducer         sarama.SyncProducer
	asyncProducer        sarama.AsyncProducer
	kafkaConsumer        sarama.Consumer
	kafkaConsumerGroup   sarama.ConsumerGroup
	kafkaAdmin           sarama.ClusterAdmin
	err                  error
	kafkaBrokers         []string = []string{"127.0.0.1:9092"}
	kafkaTopicName       string   = "topic4test"
	kafkaTopicName4Test  string   = "topic4test4test"
	kafkaConsumerGroupID string   = "testgroup"
)

var _ = BeforeSuite(func() {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll // Wait for all in-sync replicas to ack the message
	config.Producer.Retry.Max = 10                   // Retry up to 10 times to produce the message
	config.Producer.Return.Successes = true
	config.Version = sarama.V2_3_0_0
	config.Consumer.Offsets.Initial = sarama.OffsetOldest

	kafkaClient, err = sarama.NewClient(kafkaBrokers, config)
	Expect(err).NotTo(HaveOccurred())

	syncProducer, err = sarama.NewSyncProducerFromClient(kafkaClient)
	Expect(err).NotTo(HaveOccurred())

	asyncProducer, err = sarama.NewAsyncProducerFromClient(kafkaClient)
	Expect(err).NotTo(HaveOccurred())

	kafkaConsumer, err = sarama.NewConsumerFromClient(kafkaClient)
	Expect(err).NotTo(HaveOccurred())

	kafkaAdmin, err = sarama.NewClusterAdminFromClient(kafkaClient)
	Expect(err).NotTo(HaveOccurred())

	kafkaConsumerGroup, err = sarama.NewConsumerGroupFromClient(kafkaConsumerGroupID, kafkaClient)
	Expect(err).NotTo(HaveOccurred())

	topicDetail := &sarama.TopicDetail{
		NumPartitions:     1,
		ReplicationFactor: 1}
	err = kafkaAdmin.CreateTopic(kafkaTopicName, topicDetail, false)
	Expect(err).NotTo(HaveOccurred())
})

var _ = AfterSuite(func() {
	err = kafkaAdmin.DeleteTopic(kafkaTopicName)
	Expect(err).NotTo(HaveOccurred())

	err = kafkaClient.Close()
	Expect(err).NotTo(HaveOccurred())
})
