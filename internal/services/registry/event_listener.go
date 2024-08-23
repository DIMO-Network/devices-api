package registry

import (
	"context"
	"encoding/json"

	"github.com/DIMO-Network/shared"
	"github.com/Shopify/sarama"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/rs/zerolog"
)

var failureCount = promauto.NewCounter(
	prometheus.CounterOpts{
		Namespace: "devices_api",
		Subsystem: "meta_transaction_consumer",
		Name:      "failures_total",
		Help:      "Metatransactions intended for devices-api on which the service errored.",
	},
)

type ceLog struct {
	Address common.Address `json:"address"`
	Topics  []common.Hash  `json:"topics"`
	Data    hexutil.Bytes  `json:"data"`
}

type ceTx struct {
	Hash       string  `json:"hash"`
	Successful *bool   `json:"successful,omitempty"`
	Logs       []ceLog `json:"logs,omitempty"`
}

type ceReason struct {
	Data string `json:"data"`
}

// Just using the same struct for all three event types. Lazy.
type ceData struct {
	RequestID   string   `json:"requestId"`
	Type        string   `json:"type"`
	Transaction ceTx     `json:"transaction"`
	Reason      ceReason `json:"reason"`
}

type Consumer struct {
	logger  *zerolog.Logger
	storage StatusProcessor
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
		case <-session.Context().Done():
			return nil
		default:
		}

		select {
		case message := <-claim.Messages():
			c.logger.Info().Int32("partition", message.Partition).Int64("offset", message.Offset).RawJSON("value", message.Value).Msg("Got message")
			event := shared.CloudEvent[ceData]{}
			err := json.Unmarshal(message.Value, &event)
			if err != nil {
				c.logger.Err(err).Int32("partition", message.Partition).Int64("offset", message.Offset).Msg("Failed to parse transaction event.")
			} else {
				err := c.storage.Handle(session.Context(), &event.Data)
				if err != nil {
					failureCount.Inc()
					c.logger.Err(err).Str("requestId", event.Data.RequestID).Msg("Failed to process meta-transaction status update.")
				}
			}
			session.MarkMessage(message, "")
		default:
		}
	}
}

// TODO(elffjs): Proper cleanup.
func RunConsumer(ctx context.Context, client sarama.Client, logger *zerolog.Logger, s StatusProcessor) error {
	group, err := sarama.NewConsumerGroupFromClient("devices-api-transaction-consumer", client)
	if err != nil {
		return err
	}

	c := &Consumer{logger: logger, storage: s}

	logger.Info().Msg("Starting transaction request status listener.")

	go func() {
		for {
			err := group.Consume(ctx, []string{"topic.transaction.request.status"}, c)
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
