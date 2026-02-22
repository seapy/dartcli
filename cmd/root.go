package cmd

import (
	"fmt"
	"os"

	"github.com/charmbracelet/lipgloss"
	"github.com/seapy/dartcli/internal/api"
	"github.com/seapy/dartcli/internal/cache"
	"github.com/seapy/dartcli/internal/config"
	"github.com/seapy/dartcli/internal/render"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	cfgFile string
	apiKey  string
	noColor bool
	style   string

	cfg      *config.Config
	apiClient *api.Client
	renderer  *render.Renderer
	corpStore *cache.Store
)

var errStyle = lipgloss.NewStyle().
	Bold(true).
	Foreground(lipgloss.Color("9")).
	Border(lipgloss.RoundedBorder()).
	Padding(0, 1)

var warnStyle = lipgloss.NewStyle().
	Foreground(lipgloss.Color("11")).
	Border(lipgloss.RoundedBorder()).
	Padding(0, 1)

// rootCmd is the base command.
var rootCmd = &cobra.Command{
	Use:   "dartcli",
	Short: "한국 금융감독원 DART 전자공시 CLI 도구",
	Long: `dartcli - DART OpenAPI CLI

기업 공시 정보를 터미널에서 마크다운 형식으로 조회합니다.

API 키 설정:
  dartcli setup          대화형으로 키 입력 후 저장 (권장)
  --api-key <키>         일회성 플래그
  DART_API_KEY 환경변수  스크립트/CI 환경`,
	SilenceUsage:  true,
	SilenceErrors: true,
}

// Execute is the entry point called from main.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		printError(err)
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "설정파일 경로 (기본: ~/.dartcli/config.yaml)")
	rootCmd.PersistentFlags().StringVar(&apiKey, "api-key", "", "DART API 키")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "색상 비활성화")
	rootCmd.PersistentFlags().StringVar(&style, "style", "auto", "Glamour 스타일 (auto|dark|light|notty)")

	viper.BindPFlag("api_key", rootCmd.PersistentFlags().Lookup("api-key"))
}

func initConfig() {
	var err error
	cfg, err = config.Load(cfgFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "설정 로드 실패: %v\n", err)
	}

	// Flag overrides config
	if apiKey != "" {
		cfg.APIKey = apiKey
	}
	if cfg.Style == "" {
		cfg.Style = style
	}

	renderer = render.New(cfg.Style, noColor)
}

// requireAPIKey ensures an API key is available, printing a helpful message if not.
func requireAPIKey() error {
	if cfg.APIKey == "" {
		msg := `DART API 키가 설정되지 않았습니다.

API 키 발급: https://opendart.fss.or.kr/uss/umt/EgovMberInsertView.do

설정 방법:
  dartcli setup              대화형으로 키 저장 (권장)
  export DART_API_KEY=<키>   환경변수
  dartcli --api-key <키> …   플래그`
		fmt.Fprintln(os.Stderr, warnStyle.Render(msg))
		return fmt.Errorf("API 키가 필요합니다")
	}
	return nil
}

// loadCorpStore loads the corp code store (auto-refreshes if stale).
func loadCorpStore() error {
	if corpStore != nil {
		return nil
	}
	if err := requireAPIKey(); err != nil {
		return err
	}
	var err error
	var refreshed bool
	corpStore, refreshed, err = cache.Load(cfg.APIKey)
	if err != nil {
		return fmt.Errorf("corp code 캐시 로드 실패: %w", err)
	}
	if refreshed {
		fmt.Fprintln(os.Stderr, "Corp code 캐시를 갱신했습니다.")
	}
	return nil
}

// resolveCorpCode resolves a company name or stock code to a DART corp code.
// Uses huh.Select if multiple results are found.
func resolveCorpCode(query string) (string, string, error) {
	if err := loadCorpStore(); err != nil {
		return "", "", err
	}

	results := corpStore.Search(query)
	if len(results) == 0 {
		return "", "", fmt.Errorf("기업을 찾을 수 없습니다: %q", query)
	}
	if len(results) == 1 {
		return results[0].CorpCode, results[0].CorpName, nil
	}

	// Multiple results → interactive selection
	return selectCorp(results)
}

func printError(err error) {
	fmt.Fprintln(os.Stderr, errStyle.Render("오류: "+err.Error()))
}

func printWarning(msg string) {
	fmt.Fprintln(os.Stderr, warnStyle.Render(msg))
}
