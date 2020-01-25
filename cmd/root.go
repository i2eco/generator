package cmd

import (
	"log"
	"os"

	"github.com/spf13/cobra"
)

// RootCmd is the root command line
var RootCmd = &cobra.Command{
	Use: "generator",
}

// Execute the commands
func Execute() {
	if err := RootCmd.Execute(); err != nil {
		log.Println(err)
		os.Exit(-1)
	}
}
