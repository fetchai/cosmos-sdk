package pruning

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/store/rootmulti"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"
)

const FlagAppDBBackend = "app-db-backend"

// PruningCmd prunes the sdk root multi store history versions based on the pruning options
// specified by command flags.
func PruningCmd(appCreator servertypes.AppCreator) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "prune",
		Short: "Prune app history states by keeping the recent heights and deleting old heights",
		Long: `Prune app history states by keeping the recent heights and deleting old heights.
		The pruning strategy is provided via the '--pruning' flag or alternatively with '--pruning-keep-recent'
		
		For '--pruning' the options are as follows:
		
		default: the last 362880 states are kept
		nothing: all historic states will be saved, nothing will be deleted (i.e. archiving node)
		everything: 2 latest states will be kept
		custom: allow pruning options to be manually specified through the 'app.toml' config file or through CLI flags.
		besides pruning options, database home directory and database backend type should also be specified via flags
		'--home' and '--app-db-backend'.
		valid app-db-backend type includes 'goleveldb', 'cleveldb', 'rocksdb', 'boltdb', and 'badgerdb'.
		`,
		Example: `prune --home './' --app-db-backend 'goleveldb' --pruning 'custom' --pruning-keep-recent 100 --
		pruning-keep-every 10, --pruning-interval 10`,
		RunE: func(cmd *cobra.Command, _ []string) error {
			vp := viper.New()

			// Bind flags to the Context's Viper so we can get pruning options.
			if err := vp.BindPFlags(cmd.Flags()); err != nil {
				return err
			}
			pruningOptions, err := server.GetPruningOptionsFromFlags(vp)
			if err != nil {
				return err
			}
			fmt.Printf("get pruning options from command flags, keep-recent: %v\n",
				pruningOptions.KeepRecent,
			)

			home := vp.GetString(flags.FlagHome)
			db, err := openDB(home)
			if err != nil {
				return err
			}

			logger := log.NewTMLogger(log.NewSyncWriter(os.Stdout))
			app := appCreator(logger, db, nil, vp)
			cms := app.CommitMultiStore()

			rootMultiStore, ok := cms.(*rootmulti.Store)
			if !ok {
				return fmt.Errorf("currently only support the pruning of rootmulti.Store type")
			}
			latestHeight := rootmulti.GetLatestVersion(db)
			// valid heights should be greater than 0.
			if latestHeight <= 0 {
				return fmt.Errorf("the database has no valid heights to prune, the latest height: %v", latestHeight)
			}

			effectiveLastHeight := latestHeight - int64(pruningOptions.KeepRecent)

			if effectiveLastHeight < 1 {
				fmt.Printf("no heights to prune\n")
				return nil
			}

			lenPH := effectiveLastHeight + 1
			pruningHeights := make([]int64, lenPH)
			for height := int64(1); height < lenPH; height++ {
				pruningHeights[height] = height
			}
			pruningHeights = pruningHeights[1:]

			fmt.Printf(
				"pruning heights start from %v, end at %v\n",
				pruningHeights[0],
				pruningHeights[len(pruningHeights)-1],
			)

			rootMultiStore.PruneStores(false, pruningHeights)
			if err != nil {
				return err
			}
			fmt.Printf("successfully pruned the application root multi stores\n")
			return nil
		},
	}

	cmd.Flags().String(flags.FlagHome, "", "The database home directory")
	cmd.Flags().String(FlagAppDBBackend, "", "The type of database for application and snapshots databases")
	cmd.Flags().String(server.FlagPruning, storetypes.PruningOptionDefault, "Pruning strategy (default|nothing|everything|custom)")
	cmd.Flags().Uint64(server.FlagPruningKeepRecent, 0, "Number of recent heights to keep on disk (ignored if pruning is not 'custom')")
	cmd.Flags().Uint64(server.FlagPruningKeepEvery, 0,
		`Offset heights to keep on disk after 'keep-every' (ignored if pruning is not 'custom'),
		this is not used by this command but kept for compatibility with the complete pruning options`)
	cmd.Flags().Uint64(server.FlagPruningInterval, 10,
		`Height interval at which pruned heights are removed from disk (ignored if pruning is not 'custom'), 
		this is not used by this command but kept for compatibility with the complete pruning options`)

	return cmd
}

func openDB(rootDir string) (dbm.DB, error) {
	dataDir := filepath.Join(rootDir, "data")
	return sdk.NewLevelDB("application", dataDir)
}
