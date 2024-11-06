package main

import (
	"context"
	"encoding/json"
	"flag"
	"strings"

	"github.com/IBM/sarama"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"

	"github.com/DIMO-Network/shared"
	"github.com/DIMO-Network/shared/db"
	"github.com/DIMO-Network/shared/sdtask"

	"github.com/DIMO-Network/devices-api/internal/config"
)

type findOldStyleTasks struct {
	logger   zerolog.Logger
	settings config.Settings
	pdb      db.Store
}

func (*findOldStyleTasks) Name() string     { return "find-old-style-tasks" }
func (*findOldStyleTasks) Synopsis() string { return "xdd" }
func (*findOldStyleTasks) Usage() string {
	return `xpp`
}

// nolint
func (p *findOldStyleTasks) SetFlags(f *flag.FlagSet) {

}

func (fost *findOldStyleTasks) Execute(_ context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {

	kc := sarama.NewConfig()
	kc.Version = sarama.V3_6_0_0

	client, err := sarama.NewClient(strings.Split(fost.settings.KafkaBrokers, ","), kc)
	if err != nil {
		panic(err)
	}

	cons, err := sarama.NewConsumer(strings.Split(fost.settings.KafkaBrokers, ","), kc)
	if err != nil {
		panic(err)
	}

	topic := fost.settings.TaskCredentialTopic

	ps, err := cons.Partitions(topic)
	if err != nil {
		panic(err)
	}

	missing := make(map[string]struct{})

	for _, p := range ps {
		p := p
		hwm, err := client.GetOffset(topic, p, sarama.OffsetNewest)
		if err != nil {
			panic(err)
		}

		pc, err := cons.ConsumePartition(topic, p, sarama.OffsetOldest)
		if err != nil {
			panic(err)
		}

	MsgLoop:
		for m := range pc.Messages() {
			if m.Offset >= hwm {
				fost.logger.Info().Msgf("Finished processing parition %d.", p)
				continue MsgLoop
			}

			if m.Value == nil {
				continue
			}

			var out shared.CloudEvent[sdtask.CredentialData]

			err := json.Unmarshal(m.Value, &out)
			if err != nil {
				panic(err)
			}

			key := string(m.Key)

			if out.Data.SyntheticDevice == nil {
				missing[key] = struct{}{}
				fost.logger.Warn().Str("userDeviceId", out.Data.UserDeviceID).Str("integrationId", out.Data.IntegrationID).Str("taskId", out.Data.TaskID).Str("key", key).Int64("offset", m.Offset).Msg("Found a bad one.")
			} else {
				if _, ok := missing[key]; ok {
					fost.logger.Info().Str("userDeviceId", out.Data.UserDeviceID).Str("integrationId", out.Data.IntegrationID).Str("taskId", out.Data.TaskID).Str("key", key).Int64("offset", m.Offset).Msg("Bad one was later replaced with a good one.")
					delete(missing, key)
				}
			}
		}
	}

	fost.logger.Info().Msgf("Finished examining %d partitions.", len(ps))

	return subcommands.ExitSuccess
}
