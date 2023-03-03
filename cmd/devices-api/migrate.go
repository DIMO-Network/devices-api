package main

import (
	"context"
	"database/sql"
	"flag"
	"fmt"
	"github.com/DIMO-Network/devices-api/internal/config"
	"github.com/google/subcommands"
	_ "github.com/lib/pq"
	"github.com/pressly/goose/v3"
	"github.com/rs/zerolog"
)

type migrateDBCmd struct {
	logger   zerolog.Logger
	settings config.Settings

	up   bool
	down bool
}

func (*migrateDBCmd) Name() string     { return "migrate" }
func (*migrateDBCmd) Synopsis() string { return "migrate args to stdout." }
func (*migrateDBCmd) Usage() string {
	return `migrate [-up | -down]:
	migrates database up or down accordingly. No argument default is up.
  `
}

func (p *migrateDBCmd) SetFlags(f *flag.FlagSet) {
	f.BoolVar(&p.up, "up", false, "up database")
	f.BoolVar(&p.down, "down", false, "down database")
}

func (p *migrateDBCmd) Execute(ctx context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	var db *sql.DB
	// setup database
	db, err := sql.Open("postgres", p.settings.DB.BuildConnectionString(true))
	defer func() {
		if err := db.Close(); err != nil {
			p.logger.Fatal().Msgf("goose: failed to close DB: %v\n", err)
		}
	}()
	if err != nil {
		p.logger.Fatal().Msgf("failed to open db connection: %v\n", err)
	}
	if err = db.Ping(); err != nil {
		p.logger.Fatal().Msgf("failed to ping db: %v\n", err)
	}
	// set default
	command := "up"
	if p.down {
		command = "down"
	}
	fmt.Printf("migrate command received is: %s \n", command)

	// todo manually run sql to create devices_api schema
	_, err = db.Exec("CREATE SCHEMA IF NOT EXISTS devices_api;")
	if err != nil {
		p.logger.Fatal().Err(err).Msg("could not create schema")
	}
	goose.SetTableName("devices_api.migrations")
	if err := goose.Run(command, db, "migrations"); err != nil {
		p.logger.Fatal().Msgf("failed to apply go code migrations: %v\n", err)
	}
	// if we add any code migrations import _ "github.com/DIMO-Network/devices-api/migrations" // migrations won't work without this

	return subcommands.ExitSuccess
}
