package services

import (
	"encoding/json"
	"testing"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/shared"
	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	"github.com/stretchr/testify/assert"
)

func kafkaEventChecker[A any](t *testing.T, topic, key string, data A) mocks.MessageChecker {
	return func(msg *sarama.ProducerMessage) error {
		assert.Equal(t, topic, msg.Topic)

		akeyb, err := msg.Key.Encode()
		assert.NoError(t, err)
		assert.Equal(t, key, string(akeyb))

		avalb, err := msg.Value.Encode()
		assert.NoError(t, err)

		aval := &shared.CloudEvent[A]{}
		err = json.Unmarshal(avalb, aval)
		assert.NoError(t, err)

		assert.Equal(t, data, aval.Data)

		return nil
	}
}

func TestDDEventSuccess(t *testing.T) {
	sp := mocks.NewSyncProducer(t, nil)

	sp.ExpectSendMessageWithMessageCheckerFunctionAndSucceed(kafkaEventChecker(t, "ud_to_dd_table", "SomeId", DeviceDefinitionIDEventData{
		UserDeviceID:       "SomeId",
		DeviceDefinitionID: "DDMockId",
	}))
	sp.ExpectSendMessageWithMessageCheckerFunctionAndSucceed(kafkaEventChecker(t, "dd_to_mmy_table", "DDMockId", DeviceDefinitionMetadataEventData{
		Make:      "Tesla",
		Model:     "Model Y",
		Year:      2021,
		Region:    "someRegion",
		MakeSlug:  "sommeMakeSlug",
		ModelSlug: "someModelSlug",
	}))

	d := NewDeviceDefinitionRegistrar(sp, &config.Settings{
		DeviceDefinitionTopic:         "ud_to_dd_table",
		DeviceDefinitionMetadataTopic: "dd_to_mmy_table",
	})

	ddDTO := DeviceDefinitionDTO{
		UserDeviceID:       "SomeId",
		DeviceDefinitionID: "DDMockId",
		Make:               "Tesla",
		Model:              "Model Y",
		Year:               2021,
		IntegrationID:      "MockIntegrationID",
		Region:             "someRegion",
		MakeSlug:           "sommeMakeSlug",
		ModelSlug:          "someModelSlug",
	}

	assert.NoError(t, d.Register(ddDTO))

	assert.NoError(t, sp.Close())
}

func TestDDEventFailure(t *testing.T) {
	sp := mocks.NewSyncProducer(t, nil)

	sp.ExpectSendMessageAndSucceed()
	sp.ExpectSendMessageAndFail(sarama.ErrOutOfBrokers)

	d := NewDeviceDefinitionRegistrar(sp, &config.Settings{
		DeviceDefinitionTopic:         "ud_to_dd_table",
		DeviceDefinitionMetadataTopic: "dd_to_mmy_table",
	})

	ddDTO := DeviceDefinitionDTO{
		UserDeviceID:       "SomeId",
		DeviceDefinitionID: "DDMockId",
		Make:               "Tesla",
		Model:              "Model Y",
		Year:               2021,
		IntegrationID:      "MockIntegrationID",
		Region:             "someRegion",
		MakeSlug:           "sommeMakeSlug",
		ModelSlug:          "someModelSlug",
	}

	assert.ErrorIs(t, d.Register(ddDTO), sarama.ErrOutOfBrokers)

	assert.NoError(t, sp.Close())
}
