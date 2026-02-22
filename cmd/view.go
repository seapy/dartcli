package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/pkg/browser"
	"github.com/seapy/dartcli/internal/api"
	"github.com/seapy/dartcli/internal/render"
	"github.com/spf13/cobra"
)

var (
	viewBrowser  bool
	viewDownload bool
	viewOutput   string
)

const dartViewBaseURL = "https://dart.fss.or.kr/dsaf001/main.do?rcpNo="

var viewCmd = &cobra.Command{
	Use:   "view <접수번호>",
	Short: "공시 원문을 터미널에서 조회합니다",
	Long: `공시 원문을 다운로드하여 터미널에서 마크다운으로 렌더링합니다.

DART XML 원문을 파싱하여 다음 구조로 출력합니다.
  - 표지(회사명, 사업연도, 제출일 등) → 표 형식
  - 목차 → 표 형식
  - 본문 섹션(I. II. III. …) → ## 헤딩
  - 하위 섹션(1. 2. 3. …)    → ### 헤딩
  - 재무제표·통계 데이터      → 마크다운 표
  - 본문 서술                 → 단락 텍스트

사업보고서 기준 수천 줄 분량이므로 파이프로 페이저를 연결하거나
grep으로 원하는 섹션만 필터링하는 것을 권장합니다.
  dartcli view <접수번호> | less -R
  dartcli view <접수번호> | grep -A 30 "사업의 개요"

출력 옵션:
  --browser  DART 웹사이트에서 브라우저로 열기
  --download ZIP 원문 파일로 저장`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		rceptNo := args[0]

		if viewBrowser {
			url := dartViewBaseURL + rceptNo
			fmt.Printf("브라우저에서 열기: %s\n", url)
			if err := browser.OpenURL(url); err != nil {
				return fmt.Errorf("브라우저 열기 실패: %w", err)
			}
			return nil
		}

		if err := requireAPIKey(); err != nil {
			return err
		}

		client := api.New(cfg.APIKey)
		data, err := client.GetDocumentZIP(rceptNo)
		if err != nil {
			return fmt.Errorf("문서 다운로드 실패: %w", err)
		}

		if viewDownload {
			outPath := viewOutput
			if outPath == "" {
				outPath = rceptNo + ".zip"
			}
			outPath, _ = filepath.Abs(outPath)
			if err := os.WriteFile(outPath, data, 0644); err != nil {
				return fmt.Errorf("파일 저장 실패: %w", err)
			}
			fmt.Printf("저장 완료: %s (%d bytes)\n", outPath, len(data))
			return nil
		}

		// Default: render in terminal
		md, err := render.DocumentFromZIP(data, rceptNo)
		if err != nil {
			return fmt.Errorf("문서 렌더링 실패: %w", err)
		}
		return renderer.PrintWide(md)
	},
}

func init() {
	rootCmd.AddCommand(viewCmd)
	viewCmd.Flags().BoolVar(&viewBrowser, "browser", false, "브라우저로 열기")
	viewCmd.Flags().BoolVar(&viewDownload, "download", false, "ZIP 파일로 저장")
	viewCmd.Flags().StringVarP(&viewOutput, "output", "o", "", "저장 경로 (--download 와 함께 사용)")
}
