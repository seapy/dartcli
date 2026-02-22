package cmd

import (
	"fmt"

	dartcli "github.com/seapy/dartcli/pkg/dartcli"
	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "버전 정보를 출력합니다",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("dartcli %s (commit: %s, built: %s)\n",
			dartcli.Version, dartcli.Commit, dartcli.BuildDate)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
