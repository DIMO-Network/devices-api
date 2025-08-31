package main

import (
	"bytes"
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/url"
	"os"
	"strconv"

	"github.com/ericlagergren/decimal"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"github.com/volatiletech/sqlboiler/v4/types"

	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/pkg/cipher"
	"github.com/DIMO-Network/shared/pkg/db"
	"github.com/DIMO-Network/tesla-go"
)

type teslaFleetStatusCmd struct {
	logger   zerolog.Logger
	settings config.Settings
	pdb      db.Store
	cipher   cipher.Cipher
}

func (*teslaFleetStatusCmd) Name() string     { return "tesla-fleet-status" }
func (*teslaFleetStatusCmd) Synopsis() string { return "populate-sd-info-topic args to stdout." }
func (*teslaFleetStatusCmd) Usage() string {
	return `populate-sd-info-topic:
	populate-sd-info-topic args.
  `
}

// nolint
func (p *teslaFleetStatusCmd) SetFlags(f *flag.FlagSet) {

}

func (p *teslaFleetStatusCmd) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	baseURL, err := url.Parse(p.settings.TeslaFleetURL)
	if err != nil {
		panic(err)
	}

	client := tesla.New(tesla.WithBaseURL(baseURL))

	vidStr := os.Args[2]
	vid, err := strconv.ParseInt(vidStr, 10, 64)
	if err != nil {
		panic(err)
	}

	ud, err := models.UserDevices(
		models.UserDeviceWhere.TokenID.EQ(types.NewNullDecimal(decimal.New(vid, 0))),
		qm.Load(models.UserDeviceRels.UserDeviceAPIIntegrations, models.UserDeviceAPIIntegrationWhere.IntegrationID.EQ("26A5Dk3vvvQutjSyF0Jka2DP5lg")),
	).One(ctx, p.pdb.DBS().Reader)
	if err != nil {
		panic(err)
	}

	atPlain, err := p.cipher.Decrypt(ud.R.UserDeviceAPIIntegrations[0].AccessToken.String)
	if err != nil {
		panic(err)
	}

	b, err := client.GetFleetStatus(ctx, atPlain, ud.VinIdentifier.String)
	if err != nil {
		var buf bytes.Buffer
		json.Indent(&buf, b, "", "  ")

		fmt.Println(buf.String())
		panic(err)
	}

	var buf bytes.Buffer
	json.Indent(&buf, b, "", "  ")

	fmt.Println(buf.String())

	return subcommands.ExitSuccess
}
