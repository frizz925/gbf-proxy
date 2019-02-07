package cmd

import (
	"errors"
	"log"

	"github.com/Frizz925/gbf-proxy/golang/cache"
	"github.com/go-redis/redis"

	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache <listen-address> <redis-address>",
	Short: "Start the Granblue Proxy cache service",
	Long: `Arguments:
  listen-address  The address this service should listen at
  redis-address   The address for the Redis server
`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		nargs := len(args)
		if nargs < 1 {
			return errors.New("Missing listen-address argument")
		} else if nargs < 2 {
			return errors.New("Missing redis-address argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		redis := redis.NewClient(&redis.Options{
			Addr:     args[1],
			Password: "",
			DB:       0,
		})
		s := cache.New(&cache.ServerConfig{
			Redis: redis,
		})
		_, err := s.Open(args[0])
		if err != nil {
			panic(err)
		}
		log.Printf("Cache at %s -> Redis server at %s", args[0], args[1])
		s.WaitGroup().Wait()
	},
}

func init() {
	rootCmd.AddCommand(cacheCmd)
}
