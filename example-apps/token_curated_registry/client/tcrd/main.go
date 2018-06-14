package main

import (
	"encoding/json"
	"os"

	"github.com/spf13/cobra"

	abci "github.com/tendermint/abci/types"
	"github.com/tendermint/tmlibs/cli"
	dbm "github.com/tendermint/tmlibs/db"
	"github.com/tendermint/tmlibs/log"

	"github.com/AdityaSripal/token_curated_registry/app"
	"github.com/cosmos/cosmos-sdk/server"
)

func main() {
	cdc := app.MakeCodec()
	ctx := server.NewDefaultContext()

	rootCmd := &cobra.Command{
		Use:               "tcrd",
		Short:             "Token-Curated Registry Daemon (server)",
		PersistentPreRunE: server.PersistentPreRunEFn(ctx),
	}

	server.AddCommands(ctx, cdc, rootCmd, server.DefaultAppInit,
		server.ConstructAppCreator(newApp, "tcr"),
		server.ConstructAppExporter(exportAppState, "basecoin"))

	// prepare and add flags
	rootDir := os.ExpandEnv("$HOME/.tcrd")
	executor := cli.PrepareBaseCmd(rootCmd, "TCR", rootDir)
	executor.Execute()
}

func newApp(logger log.Logger, db dbm.DB) abci.Application {
	return app.NewRegistryApp(logger, db, 100, 10, 10, 10, 0.5, 0.5)
}

func exportAppState(logger log.Logger, db dbm.DB) (json.RawMessage, error) {
	rapp := app.NewRegistryApp(logger, db, 100, 10, 10, 10, 0.5, 0.5)
	return rapp.ExportAppStateJSON()
}