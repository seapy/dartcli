package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:   "search <검색어>",
	Short: "기업명 또는 종목코드로 기업을 검색합니다",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := loadCorpStore(); err != nil {
			return err
		}

		query := args[0]
		results := corpStore.Search(query)

		if len(results) == 0 {
			fmt.Printf("검색 결과 없음: %q\n", query)
			return nil
		}

		var sb strings.Builder
		fmt.Fprintf(&sb, "# 검색 결과: %q\n\n", query)
		fmt.Fprintf(&sb, "총 **%d**건\n\n", len(results))
		sb.WriteString("| 기업명 | 종목코드 | Corp Code | 수정일 |\n")
		sb.WriteString("|--------|----------|-----------|--------|\n")

		for _, r := range results {
			stock := r.StockCode
			if stock == "" {
				stock = "-"
			}
			fmt.Fprintf(&sb, "| %s | %s | `%s` | %s |\n",
				r.CorpName, stock, r.CorpCode, r.ModifyDate)
		}

		return renderer.Print(sb.String())
	},
}

func init() {
	rootCmd.AddCommand(searchCmd)
}
