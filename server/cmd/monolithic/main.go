package main

import (
	"gbf-proxy/applications"
	"gbf-proxy/lib/logger"

	"github.com/spf13/cobra"
)

var (
	log     = logger.DefaultLogger
	rootCmd = &cobra.Command{
		Use:   "gbf-proxy <listener-address>",
		Short: "Start the monolithic Granblue Proxy service",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			listenerAddr := args[0]
			memcachedAddr, err := cmd.PersistentFlags().GetString("memcached")
			if err != nil {
				return err
			}
			return (applications.MonolithicApp{
				ListenerAddr:  listenerAddr,
				MemcachedAddr: memcachedAddr,
			}).Start()
		},
	}
)

func init() {
	rootCmd.PersistentFlags().StringP("memcached", "m", "127.0.0.1:11211", "Memcached address")
}

func main() {
	rootCmd.Execute()
}
