package cmd

import (
	"github.com/seapy/dartcli/internal/api"
	"github.com/seapy/dartcli/internal/render"
	"github.com/spf13/cobra"
)

var companyCmd = &cobra.Command{
	Use:   "company <회사명 또는 종목코드>",
	Short: "기업 개황 정보를 조회합니다",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		corpCode, _, err := resolveCorpCode(args[0])
		if err != nil {
			return err
		}

		client := api.New(cfg.APIKey)
		info, err := client.GetCompany(corpCode)
		if err != nil {
			return err
		}

		md := render.CompanyMarkdown(info)
		return renderer.Print(md)
	},
}

func init() {
	rootCmd.AddCommand(companyCmd)
}
