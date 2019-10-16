package main

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	trustServer string
	tlscacert   string
	trustDir    string
	logLevel    string
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

func init() {
	rootCmd.AddCommand(
		newListCmd(),
		newSignCmd(),
		newVerifyCmd(),
	)

	rootCmd.PersistentFlags().StringVarP(&trustServer, "server", "", "https://notary.docker.io", "The trust server used")
	rootCmd.PersistentFlags().StringVarP(&tlscacert, "tlscacert", "", "", "Trust certs signed only by this CA")
	rootCmd.PersistentFlags().StringVarP(&trustDir, "dir", "d", defaultTrustDir(), "Directory where the trust data is persisted to")

	rootCmd.PersistentFlags().StringVar(&logLevel, "log", "info", `Set the logging level ("debug"|"info"|"warn"|"error"|"fatal")`)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func defaultTrustDir() string {
	homeEnvPath := os.Getenv("HOME")
	if homeEnvPath == "" && runtime.GOOS == "windows" {
		homeEnvPath = os.Getenv("USERPROFILE")
	}

	return filepath.Join(homeEnvPath, ".signy")
}
