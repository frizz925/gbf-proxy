package cmd

import (
	"errors"

	"github.com/Frizz925/gbf-proxy/golang/cache"

	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache <listen-address>",
	Short: "Start the Granblue Proxy cache service",
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
		config := &cache.ServerConfig{}
		redisAddr, err := cmd.PersistentFlags().GetString("redis-address")
		if err == nil && redisAddr != "" {
			config.RedisAddr = redisAddr
		}

		s := cache.New(config)
		_, err = s.Open(listenAddr)
		if err != nil {
			panic(err)
		}
		s.WaitGroup().Wait()
	},
}

func init() {
	cacheCmd.PersistentFlags().StringP("redis-address", "r", "localhost:6379", "The address for the Redis server")
	rootCmd.AddCommand(cacheCmd)
}
