package cmd

import (
	"fmt"
	"log"

	"github.com/Frizz925/gbf-proxy/golang/controller"

	"github.com/Frizz925/gbf-proxy/golang/cache"
	"github.com/Frizz925/gbf-proxy/golang/proxy"
	"github.com/spf13/cobra"
)

type devNull struct{}

func (d devNull) Write(b []byte) (int, error) {
	return len(b), nil
}

// localCmd represents the local command
var localCmd = &cobra.Command{
	Use:   "local",
	Short: "Start the local Granblue Proxy service",
	Run: func(cmd *cobra.Command, args []string) {
		log.SetOutput(new(devNull))
		cacheServer := cache.New(&cache.ServerConfig{})
		l, err := cacheServer.Open("127.0.0.1:0")
		if err != nil {
			panic(err)
		}
		controllerServer := controller.New(&controller.ServerConfig{
			CacheAddr: l.Addr().String(),
		})
		l, err = controllerServer.Open("127.0.0.1:0")
		if err != nil {
			panic(err)
		}
		proxyServer := proxy.New(&proxy.ServerConfig{
			BackendAddr: l.Addr().String(),
		})
		l, err = proxyServer.Open("127.0.0.1:0")
		if err != nil {
			panic(err)
		}
		fmt.Printf("Granblue Proxy is listening at %s\n", l.Addr().String())
		cacheServer.WaitGroup().Wait()
		controllerServer.WaitGroup().Wait()
		proxyServer.WaitGroup().Wait()
	},
}

func init() {
	rootCmd.AddCommand(localCmd)
}
