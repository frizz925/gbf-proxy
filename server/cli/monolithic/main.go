package main

import (
	"gbf-proxy/applications"
	"gbf-proxy/cli"
	"gbf-proxy/lib/logger"

	"github.com/spf13/cobra"
)

var (
	webHost       = "localhost"
	webAddr       = "127.0.0.1:80"
	memcachedAddr = "127.0.0.1:11211"

	version   string = "undefined"
	buildTime string = "0"
)

var (
	log     = logger.DefaultLogger
	rootCmd = &cobra.Command{
		Use:   "gbf-proxy <listener-address>",
		Short: "Start the monolithic Granblue Proxy service",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			listenerAddr := args[0]
			err := (applications.MonolithicApp{
				Version:       version,
				WebHost:       webHost,
				WebAddr:       webAddr,
				ListenerAddr:  listenerAddr,
				MemcachedAddr: memcachedAddr,
			}).Start()
			if err != nil {
				log.Fatal(err)
			}
		},
	}
)

func main() {
	rootCmd.AddCommand(cli.NewVersionCmd(version, buildTime))
	rootCmd.PersistentFlags().StringVar(&webHost, "web-hostname", webHost, "Web server hostname")
	rootCmd.PersistentFlags().StringVar(&webAddr, "web-address", webAddr, "Web server address")
	rootCmd.PersistentFlags().StringVarP(&memcachedAddr, "memcached", "m", memcachedAddr, "Memcached address")
	rootCmd.Execute()
}
