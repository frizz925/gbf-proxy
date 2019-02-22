package cmd

import (
	"fmt"
	"log"
	"os"
	"strconv"

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
		if _, found := os.LookupEnv("GBF_PROXY_DEBUG"); !found {
			log.SetOutput(new(devNull))
		}

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
		port := 0
		if len(args) >= 1 {
			port, err = strconv.Atoi(args[0])
			if err != nil {
				panic(err)
			}
		}

		proxyAddr := fmt.Sprintf("127.0.0.1:%d", port)
		l, err = proxyServer.Open(proxyAddr)
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
