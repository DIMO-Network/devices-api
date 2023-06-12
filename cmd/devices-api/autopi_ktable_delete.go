package main

import (
	"context"
	"flag"

	"github.com/DIMO-Network/devices-api/internal/services"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
)

type autoPiKTableDeleteCmd struct {
	logger    zerolog.Logger
	container dependencyContainer
}

func (*autoPiKTableDeleteCmd) Name() string     { return "autopi-ktable-delete" }
func (*autoPiKTableDeleteCmd) Synopsis() string { return "Remove unit id-to-user device id mapping." }
func (*autoPiKTableDeleteCmd) Usage() string {
	return `autopi-ktable-delete <unit id>:
  Remove unit id-to-user device id mapping.
`
}

func (c *autoPiKTableDeleteCmd) SetFlags(_ *flag.FlagSet) {}

func (c *autoPiKTableDeleteCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	unitID := f.Arg(0)

	c.logger.Info().Msgf("Removing KTable entry for unit %s.", unitID)

	reg := services.NewIngestRegistrar(c.container.getKafkaProducer())

	if err := reg.Deregister(unitID, "", ""); err != nil {
		c.logger.Err(err).Msg("Failed to emit tombstone.")
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
