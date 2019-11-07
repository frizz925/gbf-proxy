package main

import (
	"gbf-proxy/applications"
	"gbf-proxy/lib/logger"

	"github.com/spf13/cobra"
)

var (
	webHost       = "localhost"
	webAddr       = "127.0.0.1:80"
	memcachedAddr = "127.0.0.1:11211"
)

var (
	log     = logger.DefaultLogger
	rootCmd = &cobra.Command{
		Use:   "gbf-proxy <listener-address>",
		Short: "Start the monolithic Granblue Proxy service",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			listenerAddr := args[0]
			return (applications.MonolithicApp{
				WebHost:       webHost,
				WebAddr:       webAddr,
				ListenerAddr:  listenerAddr,
				MemcachedAddr: memcachedAddr,
			}).Start()
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringVar(&webHost, "web-hostname", webHost, "Web server hostname")
	rootCmd.PersistentFlags().StringVar(&webAddr, "web-address", webAddr, "Web server address")
	rootCmd.PersistentFlags().StringVarP(&memcachedAddr, "memcached", "m", memcachedAddr, "Memcached address")
}

func main() {
	rootCmd.Execute()
}
