package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var Version = "latest"

// localCmd represents the local command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the Granblue Proxy version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
