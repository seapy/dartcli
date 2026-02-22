package cmd

import (
	"fmt"
	"strings"

	"github.com/seapy/dartcli/internal/cache"
	"github.com/spf13/cobra"
)

var cacheCmd = &cobra.Command{
	Use:   "cache <refresh|status|clear>",
	Short: "Corp code 캐시를 관리합니다",
	Args:  cobra.ExactArgs(1),
	ValidArgs: []string{"refresh", "status", "clear"},
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "refresh":
			return cacheRefresh()
		case "status":
			return cacheStatus()
		case "clear":
			return cacheClear()
		default:
			return fmt.Errorf("알 수 없는 하위 명령어: %s (refresh|status|clear)", args[0])
		}
	},
}

func cacheRefresh() error {
	if err := requireAPIKey(); err != nil {
		return err
	}
	fmt.Println("Corp code 캐시를 갱신하는 중...")
	store, err := cache.Refresh(cfg.APIKey)
	if err != nil {
		return fmt.Errorf("캐시 갱신 실패: %w", err)
	}
	fmt.Printf("완료: %d개 기업 정보가 캐시되었습니다.\n", len(store.All))
	return nil
}

func cacheStatus() error {
	exists, modTime, stale, err := cache.Status()
	if err != nil {
		return err
	}

	var sb strings.Builder
	sb.WriteString("# 캐시 상태\n\n")

	if !exists {
		sb.WriteString("캐시 파일이 없습니다. `dartcli cache refresh`로 초기화하세요.\n")
	} else {
		path, _ := cache.CorpCodePath()
		staleLabel := "최신"
		if stale {
			staleLabel = "**갱신 필요**"
		}
		fmt.Fprintf(&sb, "| 항목 | 값 |\n|------|----|\n")
		fmt.Fprintf(&sb, "| 파일 | `%s` |\n", path)
		fmt.Fprintf(&sb, "| 최종 갱신 | %s |\n", modTime.Format("2006-01-02 15:04:05"))
		fmt.Fprintf(&sb, "| 상태 | %s |\n", staleLabel)
	}

	return renderer.Print(sb.String())
}

func cacheClear() error {
	if err := cache.Clear(); err != nil {
		return fmt.Errorf("캐시 삭제 실패: %w", err)
	}
	fmt.Println("캐시가 삭제되었습니다.")
	return nil
}

func init() {
	rootCmd.AddCommand(cacheCmd)
}
