package cmd

import (
	"fmt"

	"github.com/Frizz925/gbf-proxy/golang/consts"
	"github.com/spf13/cobra"
)

// localCmd represents the local command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the Granblue Proxy version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println(consts.Version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
