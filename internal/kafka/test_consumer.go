package kafka

import (
	"github.com/Shopify/sarama"
	"github.com/rs/zerolog"
)

// TestConsumer represents a Sarama consumer group consumer
type TestConsumer struct {
	ready  chan bool
	logger *zerolog.Logger
}

func NewTestConsumer(ready chan bool, logger *zerolog.Logger) *TestConsumer {
	return &TestConsumer{ready: ready, logger: logger}
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer TestConsumer) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer TestConsumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer TestConsumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/main/consumer_group.go#L27-L29
	for message := range claim.Messages() {
		consumer.logger.Info().Msgf("Message claimed: value = %s, timestamp = %v, topic = %s", string(message.Value), message.Timestamp, message.Topic)
		session.MarkMessage(message, "")
	}

	return nil
}
