package cmd

import (
	"errors"
	"fmt"
	"net/url"
	"os"

	"github.com/Frizz925/gbf-proxy/golang/lib/logging"
	"github.com/Frizz925/gbf-proxy/golang/multiplexer"
	"github.com/spf13/cobra"
)

var multiplexerCmd = &cobra.Command{
	Use:   "multiplexer <endpoint-url>",
	Short: "Start the local Granblue Proxy services with multiplexing",
	Long: `Arguments:
  endpoint-url  WebSocket endpoint for multiplexing
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

		s := multiplexer.New(&multiplexer.ServerConfig{
			MultiplexerURL: u,
		})
		l, err := s.Open(addr)
		if err != nil {
			panic(err)
		}
		fmt.Printf("Granblue Proxy is listening at %s\n", l.Addr().String())
		fmt.Printf("Multiplexing to %s\n", u.String())
		s.WaitGroup().Wait()
	},
}

func init() {
	multiplexerCmd.PersistentFlags().String("host", "localhost", "Host the local proxy should listen at")
	multiplexerCmd.PersistentFlags().IntP("port", "p", 8088, "Port the local proxy should listen at")
	rootCmd.AddCommand(multiplexerCmd)
}
