package issuer

import (
	"context"
	"encoding/json"

	"github.com/Shopify/sarama"
	"github.com/rs/zerolog"
)

type Consumer struct {
	logger          *zerolog.Logger
	vinCredentialer *Issuer
}

func (c *Consumer) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

func (c *Consumer) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	for {
		select {
		case message := <-claim.Messages():
			c.logger.Info().Int32("partition", message.Partition).Int64("offset", message.Offset).RawJSON("value", message.Value).Msg("Got fingerprint message")
			event := ADVinCredentialEvent{}
			err := json.Unmarshal(message.Value, &event)
			if err != nil {
				c.logger.Err(err).Int32("partition", message.Partition).Int64("offset", message.Offset).Msg("Failed to parse vin credentialer event.")
			} else {
				err := c.vinCredentialer.Handle(session.Context(), &event)
				if err != nil {
					c.logger.Err(err).Int32("partition", message.Partition).Int64("offset", message.Offset).Msg("Failed to update vin credential status.")
				}
			}
			session.MarkMessage(message, "")
		case <-session.Context().Done():
			return nil
		}
	}
}

func RunConsumer(ctx context.Context, client sarama.Client, logger *zerolog.Logger, i *Issuer) error {
	group, err := sarama.NewConsumerGroupFromClient("aftermarket-device-vin-credential", client)
	if err != nil {
		return err
	}

	c := &Consumer{logger: logger, vinCredentialer: i}

	logger.Info().Msg("Starting transaction request status listener.")

	go func() {
		for {
			err := group.Consume(ctx, []string{"topic.device.fingerprint"}, c)
			if err != nil {
				logger.Warn().Err(err).Msg("Consumer group session ended.")
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	return nil
}
