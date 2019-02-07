package cmd

import (
	"errors"

	"github.com/Frizz925/gbf-proxy/golang/controller"

	"github.com/spf13/cobra"
)

var controllerCmd = &cobra.Command{
	Use:   "controller <listen-address>",
	Short: "Start the Granblue Proxy controller service",
	Long: `Arguments:
  listen-address  The address this service should listen at
`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		nargs := len(args)
		if nargs < 1 {
			return errors.New("Missing listen-address argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		listenAddr := args[0]
		config := &controller.ServerConfig{}
		cacheAddr, err := cmd.PersistentFlags().GetString("cache-address")
		if err == nil && cacheAddr != "" {
			config.CacheAddr = cacheAddr
		}
		webAddr, err := cmd.PersistentFlags().GetString("web-address")
		if err == nil && webAddr != "" {
			config.WebAddr = webAddr
		}
		webHost, err := cmd.PersistentFlags().GetString("web-hostname")
		if err == nil && webHost != "" {
			config.WebHost = webHost
		}

		s := controller.New(config)
		_, err = s.Open(listenAddr)
		if err != nil {
			panic(err)
		}
		s.WaitGroup().Wait()
	},
}

func init() {
	controllerCmd.PersistentFlags().StringP("cache-address", "c", "localhost:8001", "The address for the cache service")
	controllerCmd.PersistentFlags().StringP("web-address", "w", "localhost:8080", "The address for the web server serving static files")
	controllerCmd.PersistentFlags().String("web-hostname", "", "The hostname for the web server")
	rootCmd.AddCommand(controllerCmd)
}
