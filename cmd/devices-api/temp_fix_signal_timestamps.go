package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/DIMO-Network/devices-api/models"
	"github.com/DIMO-Network/shared/db"
	"github.com/google/subcommands"
	"github.com/rs/zerolog"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type fixSignalTimestamps struct {
	logger   zerolog.Logger
	settings config.Settings
	pdb      db.Store
}

func (*fixSignalTimestamps) Name() string { return "fix-signal-timestamps" }
func (*fixSignalTimestamps) Synopsis() string {
	return "fixes the format of the timestamps in user_device_data signals"
}
func (*fixSignalTimestamps) Usage() string {
	return `fix-signal-timestamps`
}

// nolint
func (p *fixSignalTimestamps) SetFlags(f *flag.FlagSet) {

}

func (p *fixSignalTimestamps) Execute(ctx context.Context, _ *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	all, err := models.UserDeviceData(models.UserDeviceDatumWhere.Signals.IsNotNull()).All(ctx, p.pdb.DBS().Reader)
	if err != nil {
		fmt.Println(err.Error())
		return subcommands.ExitFailure
	}
	tx, err := p.pdb.DBS().Writer.BeginTx(ctx, &sql.TxOptions{
		Isolation: sql.LevelSerializable,
		ReadOnly:  false,
	})
	defer func(err error) {
		if err != nil {
			fmt.Println("Error:", err)
			err2 := tx.Rollback()
			if err2 != nil {
				fmt.Println("Error:", err)
				return
			}
		}
	}(err)

	if err != nil {
		return subcommands.ExitFailure
	}
	for _, datum := range all {
		var data map[string]interface{}
		err := json.Unmarshal(datum.Signals.JSON, &data)
		if err != nil {
			return subcommands.ExitFailure
		}

		for _, value := range data {
			if m, ok := value.(map[string]interface{}); ok {
				if timestamp, ok := m["timestamp"].(string); ok {
					lastCharTs := timestamp[len(timestamp)-1]
					if string(lastCharTs) != "Z" {
						m["timestamp"] = timestamp + "Z"
					}
				}
			}
		}

		updatedJSON, err := json.Marshal(data)
		if err != nil {
			return subcommands.ExitFailure
		}

		datum.Signals = null.JSONFrom(updatedJSON)
		_, err = datum.Update(ctx, tx, boil.Whitelist(models.UserDeviceDatumColumns.Signals, "updated_at"))
		if err != nil {
			return subcommands.ExitFailure
		}
	}
	err = tx.Commit()
	if err != nil {
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
