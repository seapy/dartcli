package cmd

import (
	"fmt"
	"strconv"
	"time"

	"github.com/seapy/dartcli/internal/api"
	"github.com/seapy/dartcli/internal/render"
	"github.com/spf13/cobra"
)

var (
	financeYear   int
	financePeriod string
	financeType   string
)

var financeCmd = &cobra.Command{
	Use:   "finance <회사명 또는 종목코드>",
	Short: "기업의 재무정보를 조회합니다",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		corpCode, corpName, err := resolveCorpCode(args[0])
		if err != nil {
			return err
		}

		year := financeYear
		if year == 0 {
			year = time.Now().Year() - 1
		}
		yearStr := strconv.Itoa(year)

		reprtCode := api.ReprtCode(financePeriod)
		fsDiv := api.FsDivCode(financeType)

		client := api.New(cfg.APIKey)
		resp, err := client.GetFinance(api.FinanceOptions{
			CorpCode:  corpCode,
			BsnsYear:  yearStr,
			ReprtCode: reprtCode,
			FsDiv:     fsDiv,
		})
		if err != nil {
			return fmt.Errorf("재무정보 조회 실패: %w", err)
		}

		if len(resp.Items) == 0 {
			fmt.Printf("%s: %s년 %s %s 재무정보가 없습니다.\n",
				corpName, yearStr, api.FsDivLabel(fsDiv), api.PeriodLabel(financePeriod))
			return nil
		}

		md := render.FinanceMarkdown(
			corpName, yearStr,
			api.PeriodLabel(financePeriod),
			api.FsDivLabel(fsDiv),
			resp.Items,
		)
		return renderer.Print(md)
	},
}

func init() {
	rootCmd.AddCommand(financeCmd)
	financeCmd.Flags().IntVar(&financeYear, "year", 0, "사업연도 (기본: 작년)")
	financeCmd.Flags().StringVar(&financePeriod, "period", "annual", "기간 (annual|q1|half|q3)")
	financeCmd.Flags().StringVar(&financeType, "type", "cfs", "재무제표 구분 (cfs=연결|ofs=개별)")
}
