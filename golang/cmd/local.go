package cmd

import (
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/Frizz925/gbf-proxy/golang/lib/logging"

	"github.com/Frizz925/gbf-proxy/golang/local"
	"github.com/spf13/cobra"
)

type devNull struct{}

func (d devNull) Write(b []byte) (int, error) {
	return len(b), nil
}

// localCmd represents the local command
var localCmd = &cobra.Command{
	Use:   "local",
	Short: "Start the local Granblue Proxy services",
	Run: func(cmd *cobra.Command, args []string) {
		if _, found := os.LookupEnv("GBF_PROXY_DEBUG"); !found {
			logging.DefaultWriter = logging.NullWriter
			logging.DefaultErrWriter = logging.NullWriter
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

		host, err := cmd.PersistentFlags().GetString("host")
		if err != nil {
			panic(err)
		}
		port, err := cmd.PersistentFlags().GetInt("port")
		if err != nil {
			panic(err)
		}
		addr := fmt.Sprintf("%s:%d", host, port)

		s := local.New(&local.ServerConfig{
			HttpClient: httpClient,
		})
		l, err := s.Open(addr)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Granblue Proxy is listening at %s\n", l.Addr().String())
		s.WaitGroup().Wait()
	},
}

func init() {
	localCmd.PersistentFlags().String("host", "localhost", "Host the local proxy should listen at")
	localCmd.PersistentFlags().IntP("port", "p", 8088, "Port the local proxy should listen at")
	localCmd.PersistentFlags().StringP("external-proxy", "e", "", "External proxy address to use")
	localCmd.PersistentFlags()
	rootCmd.AddCommand(localCmd)
}
