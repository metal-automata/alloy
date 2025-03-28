package cmd

import (
	"log"
	"time"

	"github.com/equinix-labs/otel-init-go/otelinit"
	"github.com/metal-automata/alloy/internal/app"
	"github.com/metal-automata/alloy/internal/collector"
	"github.com/metal-automata/alloy/internal/helpers"
	"github.com/metal-automata/alloy/internal/metrics"
	"github.com/metal-automata/alloy/internal/model"
	"github.com/metal-automata/alloy/internal/worker"
	"github.com/metal-automata/rivets/events"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
)

var (
	// interval is passed to the alloy controller along with collectSplay to set the interval
	// at which the inventory data in the store is refreshed.
	interval time.Duration

	// splay when set adds additional time to the collectInterval
	// the additional time added is between zero and the collectSplay time.Duration value
	splay time.Duration

	// assetIDs is the list of asset IDs to lookup out of band inventory for.
	assetIDs []string

	// csvfile holds the path to the csv file
	csvFile string

	// facilityCode to limit Alloy to when running as a worker.
	facilityCode string

	// asWorker when true runs Alloy as a worker listening on the NATS JS for Conditions to act on.
	asWorker bool

	// The number of replicaCount to use NATS KV data
	replicaCount int
)

// outofband inventory, bios configuration collection command
var cmdOutofband = &cobra.Command{
	Use:   "outofband",
	Short: "Collect inventory data, bios configuration data through the BMC",
	Run: func(cmd *cobra.Command, _ []string) {
		alloy, err := app.New(model.AppKindInband, model.StoreKind(storeKind), cfgFile, model.LogLevel(logLevel))
		if err != nil {
			log.Fatal(err)
		}

		alloy.Config.CsvFile = csvFile

		// profiling endpoint
		if enableProfiling {
			helpers.EnablePProfile()
		}

		// serve metrics endpoint
		metrics.ListenAndServe()

		ctx, otelShutdown := otelinit.InitOpenTelemetry(cmd.Context(), "alloy")
		defer otelShutdown(ctx)

		// setup cancel context with cancel func
		ctx, cancelFunc := context.WithCancel(ctx)

		// routine listens for termination signal and cancels the context
		go func() {
			<-alloy.TermCh
			cancelFunc()
		}()

		switch {
		case asWorker:
			runWorker(ctx, alloy)
			return

		case len(assetIDs) > 0:
			runOnAssets(ctx, alloy)
			return
		}

		log.Fatal("either --asset-ids OR --controller-mode was expected")
	},
}

func runWorker(ctx context.Context, alloy *app.App) {
	stream, err := events.NewStream(*alloy.Config.NatsOptions)
	if err != nil {
		alloy.Logger.Fatal(err)
	}

	w, err := worker.New(ctx, facilityCode, replicaCount, stream, alloy.Config, alloy.SyncWg, alloy.Logger)
	if err != nil {
		alloy.Logger.Fatal(err)
	}

	w.Run(ctx)
}

func runOnAssets(ctx context.Context, alloy *app.App) {
	c, err := collector.NewDeviceCollector(
		ctx,
		model.StoreKind(storeKind),
		model.AppKindOutOfBand,
		alloy.Config,
		alloy.Logger,
	)

	if err != nil {
		log.Fatal(err)
	}

	for _, assetID := range assetIDs {
		asset := &model.Asset{ID: assetID}
		if err := c.CollectOutofband(ctx, asset, outputStdout); err != nil {
			alloy.Logger.Warn(err)
		}
	}
}

// install command flags
func init() {
	cmdOutofband.PersistentFlags().DurationVar(&interval, "collect-interval", app.DefaultCollectInterval, "interval sets the periodic data collection interval")
	cmdOutofband.PersistentFlags().DurationVar(&splay, "collect-splay", app.DefaultCollectSplay, "splay adds jitter to the collection interval")
	cmdOutofband.PersistentFlags().StringSliceVar(&assetIDs, "asset-ids", []string{}, "Collect inventory for the given comma separated list of asset IDs.")
	cmdOutofband.PersistentFlags().StringVar(&csvFile, "csv-file", "assets.csv", "CSV file containing BMC credentials for assets.")
	cmdOutofband.PersistentFlags().StringVar(&facilityCode, "facility-code", "sandbox", "The facility code this Alloy instance is associated with")
	cmdOutofband.PersistentFlags().BoolVar(&asWorker, "worker", false, "Run Alloy as a worker listening for conditions on NATS")
	cmdOutofband.PersistentFlags().IntVarP(&replicaCount, "replica-count", "r", 3, "The number of replicaCount to use for NATS KV data") // nolint:gomnd // obvious int is obvious

	rootCmd.AddCommand(cmdOutofband)
}
