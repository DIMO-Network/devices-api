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
func (*findOldStyleTasks) SetFlags(f *flag.FlagSet) {

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

	for _, p := range ps {
		missing := make(map[string]shared.CloudEvent[sdtask.CredentialData])

		hwm, err := client.GetOffset(topic, p, sarama.OffsetNewest)
		if err != nil {
			panic(err)
		}

		pc, err := cons.ConsumePartition(topic, p, sarama.OffsetOldest)
		if err != nil {
			panic(err)
		}

		for m := range pc.Messages() {
			if m.Offset >= hwm-1 {
				break
			}

			key := string(m.Key)

			if m.Value == nil {
				delete(missing, key)
				continue
			}

			var out shared.CloudEvent[sdtask.CredentialData]

			err := json.Unmarshal(m.Value, &out)
			if err != nil {
				panic(err)
			}

			if out.Data.SyntheticDevice == nil {
				missing[key] = out
			} else {
				delete(missing, key)
			}
		}

		for key, task := range missing {
			fost.logger.Warn().Str("userDeviceId", task.Data.UserDeviceID).Str("integrationId", task.Data.IntegrationID).Str("taskId", task.Data.TaskID).Str("key", key).Msg("Task is missing synthetic data.")
		}

		_ = pc.Close()
	}

	_ = cons.Close()

	return subcommands.ExitSuccess
}
