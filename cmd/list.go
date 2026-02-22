package cmd

import (
	"fmt"
	"time"

	"github.com/seapy/dartcli/internal/api"
	"github.com/seapy/dartcli/internal/render"
	"github.com/spf13/cobra"
)

var (
	listDays  int
	listStart string
	listEnd   string
	listType  string
	listLimit int
)

var listCmd = &cobra.Command{
	Use:   "list <회사명 또는 종목코드>",
	Short: "기업의 공시 목록을 조회합니다",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		corpCode, corpName, err := resolveCorpCode(args[0])
		if err != nil {
			return err
		}

		// Resolve date range
		endDate := listEnd
		if endDate == "" {
			endDate = time.Now().Format("20060102")
		}
		startDate := listStart
		if startDate == "" {
			startDate = time.Now().AddDate(0, 0, -listDays).Format("20060102")
		}

		pageCount := listLimit
		if pageCount <= 0 {
			pageCount = 20
		}

		client := api.New(cfg.APIKey)
		resp, err := client.GetList(api.ListOptions{
			CorpCode:  corpCode,
			StartDate: startDate,
			EndDate:   endDate,
			PblntfTy:  listType,
			PageNo:    1,
			PageCount: pageCount,
		})
		if err != nil {
			return fmt.Errorf("공시 목록 조회 실패: %w", err)
		}

		if len(resp.Items) == 0 {
			fmt.Printf("%s: %s ~ %s 기간에 공시된 내역이 없습니다.\n", corpName, startDate, endDate)
			return nil
		}

		items := resp.Items
		if len(items) > listLimit && listLimit > 0 {
			items = items[:listLimit]
		}

		md := render.ListMarkdown(corpName, items)
		return renderer.Print(md)
	},
}

func init() {
	rootCmd.AddCommand(listCmd)
	listCmd.Flags().IntVar(&listDays, "days", 1825, "최근 N일")
	listCmd.Flags().StringVar(&listStart, "start", "", "시작일 YYYYMMDD")
	listCmd.Flags().StringVar(&listEnd, "end", "", "종료일 YYYYMMDD (기본: 오늘)")
	listCmd.Flags().StringVar(&listType, "type", "", "공시유형 코드 (A=정기공시, B=주요사항...)")
	listCmd.Flags().IntVar(&listLimit, "limit", 20, "최대 결과 수")
}
