package kafka_test

import (
	"context"
	"fmt"
	"time"

	"github.com/Shopify/sarama"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

var _ = Describe("Kafka", func() {
	Context("admin test", func() {
		It("create topic", func() {
			topicDetail := &sarama.TopicDetail{
				NumPartitions:     1,
				ReplicationFactor: 1}
			err = kafkaAdmin.CreateTopic(kafkaTopicName4Test, topicDetail, false)
			Expect(err).NotTo(HaveOccurred())
		})
		It("list topic", func() {
			var result map[string]sarama.TopicDetail
			result, err = kafkaAdmin.ListTopics()
			Expect(err).NotTo(HaveOccurred())
			Expect(result).To(HaveKey(kafkaTopicName4Test))
		})
		It("describe topic", func() {
			var result []*sarama.TopicMetadata
			result, err = kafkaAdmin.DescribeTopics([]string{kafkaTopicName4Test})
			Expect(err).NotTo(HaveOccurred())
			Expect(result[0].Name).To(Equal(kafkaTopicName4Test))
		})
		It("delete topic", func() {
			err = kafkaAdmin.DeleteTopic(kafkaTopicName4Test)
			Expect(err).NotTo(HaveOccurred())
		})
	})
	Context("producer test", func() {
		It("sync producer", func() {
			_, _, err = syncProducer.SendMessage(&sarama.ProducerMessage{
				Topic: kafkaTopicName,
				Value: sarama.StringEncoder("sync producer"),
			})
			Expect(err).NotTo(HaveOccurred())
		})
		It("async producer", func() {
			asyncProducer.Input() <- &sarama.ProducerMessage{
				Topic: kafkaTopicName,
				Value: sarama.StringEncoder("async producer"),
			}
			select {
			case msg := <-asyncProducer.Successes():
				Expect(msg.Offset).NotTo(BeZero())
			case err = <-asyncProducer.Errors():
				Fail("async producer failed")
			case <-time.After(time.Second):
				Fail("timed out waiting for output")
			}
		})
	})
	Context("consumer test", func() {
		It("consumer messages", func() {
			var partitions []int32
			partitions, err = kafkaConsumer.Partitions(kafkaTopicName)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(partitions)).To(Equal(1))

			var pc sarama.PartitionConsumer
			pc, err = kafkaConsumer.ConsumePartition(kafkaTopicName, 0, sarama.OffsetOldest)
			Expect(err).NotTo(HaveOccurred())

			results := []string{}
		CONSUME:
			for {
				select {
				case tmp := <-pc.Messages():
					results = append(results, string(tmp.Value))
				case <-time.After(5 * time.Second):
					break CONSUME
				}
			}

			Expect(len(results)).To(Equal(2))
			Expect(results).To(ContainElement("sync producer"))
			Expect(results).To(ContainElement("async producer"))
		})
	})
	Context("consumer group test", func() {
		It("consumer messages", func() {
			topics := []string{kafkaTopicName}
			ctx := context.Background()

			handler := &ConsumerGroupHandler{
				messageChannel: make(chan *sarama.ConsumerMessage, 10),
			}

			go func() {
				err = kafkaConsumerGroup.Consume(ctx, topics, handler)
				Expect(err).NotTo(HaveOccurred())
			}()

			results := []string{}
		CONSUME:
			for {
				select {
				case tmp := <-handler.messageChannel:
					results = append(results, string(tmp.Value))
				case <-time.After(5 * time.Second):
					break CONSUME
				}
			}
			Expect(len(results)).To(Equal(2))
			Expect(results).To(ContainElement("sync producer"))
			Expect(results).To(ContainElement("async producer"))
		})
	})
})

type ConsumerGroupHandler struct {
	messageChannel chan *sarama.ConsumerMessage
}

func (ConsumerGroupHandler) Setup(s sarama.ConsumerGroupSession) error {
	fmt.Println("Partition allocation -", s.Claims())
	return nil
}
func (ConsumerGroupHandler) Cleanup(s sarama.ConsumerGroupSession) error {
	fmt.Println("Consumer group clean up initiated")
	return nil
}
func (h ConsumerGroupHandler) ConsumeClaim(sess sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for msg := range claim.Messages() {
		h.messageChannel <- msg
		fmt.Printf("Message topic:%q partition:%d offset:%d\n", msg.Topic, msg.Partition, msg.Offset)
		if err != nil {
			logrus.Error("Fail to process message", err)
			continue
		}
		sess.MarkMessage(msg, "")
	}
	return nil
}
