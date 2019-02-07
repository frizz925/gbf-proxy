package cmd

import (
	"errors"
	"log"

	"github.com/Frizz925/gbf-proxy/golang/proxy"
	"github.com/spf13/cobra"
)

// proxyCmd represents the proxy command
var proxyCmd = &cobra.Command{
	Use:   "proxy <listen-address> <controller-address>",
	Short: "Start the Granblue Proxy service",
	Long: `Arguments:
  listen-address      The address this service should listen at
  controller-address  The address for the controller service
`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		nargs := len(args)
		if nargs < 1 {
			return errors.New("Missing listen-address argument")
		} else if nargs < 2 {
			return errors.New("Missing controller-address argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		s := proxy.New(&proxy.ServerConfig{
			BackendAddr: args[1],
		})
		_, err := s.Open(args[0])
		if err != nil {
			panic(err)
		}
		log.Printf("Proxy at %s -> Controller at %s", args[0], args[1])
		s.WaitGroup().Wait()
	},
}

func init() {
	rootCmd.AddCommand(proxyCmd)
}
