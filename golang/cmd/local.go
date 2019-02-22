package cmd

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

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
		extProxyAddr, err := cmd.PersistentFlags().GetString("external-proxy")
		if err != nil {
			panic(err)
		}
		httpClient := http.DefaultClient
		if extProxyAddr != "" {
			proxyURL, err := url.Parse(extProxyAddr)
			if err != nil {
				panic(err)
			}
			httpClient = &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(proxyURL),
				},
			}
			fmt.Printf("Using external proxy at %s\n", proxyURL.String())
		}

		cacheServer := cache.New(&cache.ServerConfig{
			HttpClient: httpClient,
		})
		l, err := cacheServer.Open("127.0.0.1:0")
		if err != nil {
			panic(err)
		}

		controllerServer := controller.New(&controller.ServerConfig{
			CacheAddr:     l.Addr().String(),
			DefaultClient: httpClient,
		})
		l, err = controllerServer.Open("127.0.0.1:0")
		if err != nil {
			panic(err)
		}

		host, err := cmd.PersistentFlags().GetString("host")
		if err != nil {
			panic(err)
		}
		port, err := cmd.PersistentFlags().GetInt("port")
		if err != nil {
			panic(err)
		}

		proxyAddr := fmt.Sprintf("%s:%d", host, port)
		proxyServer := proxy.New(&proxy.ServerConfig{
			BackendAddr: l.Addr().String(),
		})
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
	localCmd.PersistentFlags().String("host", "localhost", "Host the local proxy should listen at")
	localCmd.PersistentFlags().IntP("port", "p", 8088, "Port the local proxy should listen at")
	localCmd.PersistentFlags().StringP("external-proxy", "e", "", "External proxy address to use")
	localCmd.PersistentFlags()
	rootCmd.AddCommand(localCmd)
}
