package cmd

import (
	"fmt"

	"github.com/metal-automata/alloy/internal/version"
	"github.com/spf13/cobra"
)

var cmdVersion = &cobra.Command{
	Use:   "version",
	Short: "Print Alloy version along with dependency - ironlib, bmclib version information.",
	Run: func(_ *cobra.Command, _ []string) {
		fmt.Printf(
			"commit: %s\nbranch: %s\ngit summary: %s\nbuildDate: %s\nversion: %s\nGo version: %s\nironlib version: %s\nbmclib version: %s\nserverservice version: %s",
			version.GitCommit, version.GitBranch, version.GitSummary, version.BuildDate, version.AppVersion, version.GoVersion, version.IronlibVersion, version.BmclibVersion, version.FleetDBAPIVersion)

	},
}

func init() {
	rootCmd.AddCommand(cmdVersion)
}
