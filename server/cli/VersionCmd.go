package cli

import (
	"fmt"
	"runtime"
	"strconv"
	"time"

	"github.com/spf13/cobra"
)

func NewVersionCmd(version string, buildTime string) *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the build version",
		Run: func(*cobra.Command, []string) {
			fmt.Println("Granblue Proxy")
			printVersion("Version", version)
			printVersion("Go Version", runtime.Version())
			printVersion("Built", formatTime(buildTime))
			printVersion("OS/Arch", getPlatform())
		},
	}
}

func printVersion(name string, value string) {
	fmt.Printf("  %-12s %s\n", name+":", value)
}

func formatTime(timestamp string) string {
	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return timestamp
	}
	return time.Unix(ts, 0).
		UTC().
		Format("2006-01-02T15:04:05")
}

func getPlatform() string {
	return fmt.Sprintf("%s/%s", runtime.GOOS, runtime.GOARCH)
}
