package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/Frizz925/gbf-proxy/golang/lib/logging"
	"github.com/Frizz925/gbf-proxy/golang/tunnel"
	"github.com/spf13/cobra"
)

var tunnelCmd = &cobra.Command{
	Use:   "tunnel <endpoint-url>",
	Short: "Start the local Granblue Proxy services with tunneling",
	Long: `Arguments:
  endpoint-url  WebSocket endpoint for tunneling
`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		nargs := len(args)
		if nargs < 1 {
			return errors.New("Missing endpoint-url argument")
		}
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		if _, found := os.LookupEnv("GBF_PROXY_DEBUG"); !found {
			logging.DefaultWriter = logging.NullWriter
			logging.DefaultErrWriter = logging.NullWriter
		}

		u, err := url.Parse(args[0])
		if err != nil {
			panic(err)
		}

		h, err := cmd.PersistentFlags().GetString("host")
		if err != nil {
			panic(err)
		}
		p, err := cmd.PersistentFlags().GetInt("port")
		if err != nil {
			panic(err)
		}
		addr := fmt.Sprintf("%s:%d", h, p)

		s := tunnel.New(&tunnel.ServerConfig{
			TunnelURL: u,
		})
		l, err := s.Open(addr)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Granblue Proxy is listening at %s\n", l.Addr().String())
		fmt.Printf("Tunneling to %s\n", u.String())
		s.WaitGroup().Wait()
	},
}

func init() {
	tunnelCmd.PersistentFlags().String("host", "localhost", "Host the local proxy should listen at")
	tunnelCmd.PersistentFlags().IntP("port", "p", 8088, "Port the local proxy should listen at")
	rootCmd.AddCommand(tunnelCmd)
}
