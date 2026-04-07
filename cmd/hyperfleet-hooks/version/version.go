package version

import (
	"fmt"

	"github.com/openshift-hyperfleet/hyperfleet-hooks/pkg/version"
	"github.com/spf13/cobra"
)

// NewCommand creates the version command
func NewCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Run: func(cmd *cobra.Command, args []string) {
			info := version.Get()
			fmt.Println(info.String())
		},
	}
}
