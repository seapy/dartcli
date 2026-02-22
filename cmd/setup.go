package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/huh"
	"github.com/seapy/dartcli/internal/config"
	"github.com/spf13/cobra"
)

var setupCmd = &cobra.Command{
	Use:   "setup",
	Short: "API 키를 대화형으로 설정합니다",
	Long: `DART OpenAPI 키를 입력받아 ~/.dartcli/config.yaml 에 저장합니다.

API 키 발급: https://opendart.fss.or.kr/uss/umt/EgovMberInsertView.do
(개인회원 즉시 발급, 기업회원 1~2 영업일)`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		// 현재 키가 있으면 안내
		if cfg.APIKey != "" {
			fmt.Fprintln(os.Stderr, "이미 API 키가 설정되어 있습니다. 새 키를 입력하면 덮어씁니다.")
		}

		var key string
		form := huh.NewForm(
			huh.NewGroup(
				huh.NewInput().
					Title("DART API 키").
					Description("https://opendart.fss.or.kr 에서 발급받은 키를 입력하세요.").
					EchoMode(huh.EchoModePassword).
					Validate(func(s string) error {
						if len(s) < 10 {
							return fmt.Errorf("올바른 API 키를 입력해 주세요")
						}
						return nil
					}).
					Value(&key),
			),
		)

		if err := form.Run(); err != nil {
			return fmt.Errorf("취소됨")
		}

		saveCfg := &config.Config{
			APIKey: key,
			Style:  cfg.Style,
		}
		if err := config.Save(saveCfg); err != nil {
			return fmt.Errorf("설정 저장 실패: %w", err)
		}

		fmt.Printf("✓ API 키를 저장했습니다: %s\n", config.DefaultConfigPath)
		fmt.Println()
		fmt.Println("이제 시작하세요:")
		fmt.Println("  dartcli search 삼성전자")
		return nil
	},
}

func init() {
	rootCmd.AddCommand(setupCmd)
}
