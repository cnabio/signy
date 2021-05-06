package main

import (
	"fmt"
	"os"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"github.com/cnabio/signy/pkg/tuf"
)

var (
	Version   = ""
	Commit    = ""
	BuildTime = ""
)

var (
	trustServer string
	tlscacert   string
	trustDir    string
	logLevel    string
	timeout     string
)
var rootCmd = &cobra.Command{
	Use:   "signy",
	Short: "Signy is a tool for exercising the TUF specification in order to sign various cloud-native artifacts",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		l, err := log.ParseLevel(logLevel)
		if err != nil {
			return err
		}
		log.SetLevel(l)
		return nil
	},
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show the version, commit and buildtime",
	Run: func(cmd *cobra.Command, args []string) {
		if Version == "" {
			Version = "unknown"
		}
		if Commit == "" {
			Commit = "unknown"
		}
		if BuildTime == "" {
			BuildTime = "unknown"
		}
		fmt.Printf("Version: %v\nCommit: %v\nBuilt on: %v\n", Version, Commit, BuildTime)
	},
}

func init() {
	rootCmd.AddCommand(
		newListCmd(),
		newSignCmd(),
		newVerifyCmd(),
		versionCmd,
	)

	rootCmd.PersistentFlags().StringVarP(&trustServer, "server", "", tuf.DockerNotaryServer, "The trust server used")
	rootCmd.PersistentFlags().StringVarP(&tlscacert, "tlscacert", "", "", "Trust certs signed only by this CA")
	rootCmd.PersistentFlags().StringVarP(&trustDir, "dir", "d", tuf.DefaultTrustDir(), "Directory where the trust data is persisted to")
	rootCmd.PersistentFlags().StringVar(&logLevel, "log", "info", `Set the logging level ("debug"|"info"|"warn"|"error"|"fatal")`)
	rootCmd.PersistentFlags().StringVarP(&timeout, "timeout", "t", "5s", `Timeout for the trust server`)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
