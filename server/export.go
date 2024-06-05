package server

// DONTCOVER

import (
	"errors"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmstore "github.com/tendermint/tendermint/store"
	tmtypes "github.com/tendermint/tendermint/types"
	tmdb "github.com/tendermint/tm-db"
	"os"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

const (
	FlagHeight           = "height"
	FlagForZeroHeight    = "for-zero-height"
	FlagJailAllowedAddrs = "jail-allowed-addrs"
	RoundingInterval     = 5 * time.Minute
)

// GetLatestBlockHeaderFromDB returns the latest block header from the blockstore database
func GetLatestBlockHeaderFromDB(blockStoreDB tmdb.DB) (*tmtypes.Header, error) {
	// Load the blockstore
	blockStore := tmstore.NewBlockStore(blockStoreDB)
	defer blockStoreDB.Close()

	// Get the latest block meta
	latestHeight := blockStore.Height()
	blockMeta := blockStore.LoadBlockMeta(latestHeight)
	if blockMeta == nil {
		return nil, errors.New("latest block metadata not found in blockstore")
	}

	return &blockMeta.Header, nil
}

// ExportCmd dumps app state to JSON.
func ExportCmd(appExporter types.AppExporter, defaultNodeHome string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "export",
		Short: "Export state to JSON",
		RunE: func(cmd *cobra.Command, args []string) error {
			serverCtx := GetServerContextFromCmd(cmd)
			config := serverCtx.Config

			homeDir, _ := cmd.Flags().GetString(flags.FlagHome)
			config.SetRoot(homeDir)

			if _, err := os.Stat(config.GenesisFile()); os.IsNotExist(err) {
				return err
			}

			db, err := openDB(config.RootDir)
			if err != nil {
				return err
			}

			if appExporter == nil {
				if _, err := fmt.Fprintln(os.Stderr, "WARNING: App exporter not defined. Returning genesis file."); err != nil {
					return err
				}

				genesis, err := os.ReadFile(config.GenesisFile())
				if err != nil {
					return err
				}

				fmt.Println(string(genesis))
				return nil
			}

			traceWriterFile, _ := cmd.Flags().GetString(flagTraceStore)
			traceWriter, err := openTraceWriter(traceWriterFile)
			if err != nil {
				return err
			}

			height, _ := cmd.Flags().GetInt64(FlagHeight)
			forZeroHeight, _ := cmd.Flags().GetBool(FlagForZeroHeight)
			jailAllowedAddrs, _ := cmd.Flags().GetStringSlice(FlagJailAllowedAddrs)

			exported, err := appExporter(serverCtx.Logger, db, traceWriter, height, forZeroHeight, jailAllowedAddrs, serverCtx.Viper)
			if err != nil {
				return fmt.Errorf("error exporting state: %v", err)
			}

			doc, err := tmtypes.GenesisDocFromFile(serverCtx.Config.GenesisFile())
			if err != nil {
				return err
			}

			latestBlockHeader, err := GetLatestBlockHeaderFromDB(db)
			if err != nil {
				return err
			}
			doc.GenesisTime = latestBlockHeader.Time.Round(RoundingInterval)

			doc.AppState = exported.AppState
			doc.Validators = exported.Validators
			doc.InitialHeight = exported.Height
			doc.ConsensusParams = &tmproto.ConsensusParams{
				Block: tmproto.BlockParams{
					MaxBytes:   exported.ConsensusParams.Block.MaxBytes,
					MaxGas:     exported.ConsensusParams.Block.MaxGas,
					TimeIotaMs: doc.ConsensusParams.Block.TimeIotaMs,
				},
				Evidence: tmproto.EvidenceParams{
					MaxAgeNumBlocks: exported.ConsensusParams.Evidence.MaxAgeNumBlocks,
					MaxAgeDuration:  exported.ConsensusParams.Evidence.MaxAgeDuration,
					MaxBytes:        exported.ConsensusParams.Evidence.MaxBytes,
				},
				Validator: tmproto.ValidatorParams{
					PubKeyTypes: exported.ConsensusParams.Validator.PubKeyTypes,
				},
			}

			// NOTE: Tendermint uses a custom JSON decoder for GenesisDoc
			// (except for stuff inside AppState). Inside AppState, we're free
			// to encode as protobuf or amino.
			encoded, err := tmjson.Marshal(doc)
			if err != nil {
				return err
			}

			cmd.Println(string(sdk.MustSortJSON(encoded)))
			return nil
		},
	}

	cmd.Flags().String(flags.FlagHome, defaultNodeHome, "The application home directory")
	cmd.Flags().Int64(FlagHeight, -1, "Export state from a particular height (-1 means latest height)")
	cmd.Flags().Bool(FlagForZeroHeight, false, "Export state to start at height zero (perform preproccessing)")
	cmd.Flags().StringSlice(FlagJailAllowedAddrs, []string{}, "Comma-separated list of operator addresses of jailed validators to unjail")

	return cmd
}
