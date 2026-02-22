package cmd

import (
	"fmt"

	"github.com/charmbracelet/huh"
	"github.com/seapy/dartcli/internal/cache"
)

// selectCorp presents an interactive list for the user to pick from multiple results.
func selectCorp(results []*cache.CorpInfo) (corpCode, corpName string, err error) {
	if len(results) == 0 {
		return "", "", fmt.Errorf("결과가 없습니다")
	}

	options := make([]huh.Option[string], 0, len(results))
	for _, r := range results {
		label := r.CorpName
		if r.StockCode != "" {
			label += " (" + r.StockCode + ")"
		}
		options = append(options, huh.NewOption(label, r.CorpCode))
	}

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("여러 기업이 검색되었습니다. 선택해 주세요:").
				Options(options...).
				Value(&selected),
		),
	)

	if err := form.Run(); err != nil {
		return "", "", fmt.Errorf("선택 취소: %w", err)
	}

	for _, r := range results {
		if r.CorpCode == selected {
			return r.CorpCode, r.CorpName, nil
		}
	}
	return "", "", fmt.Errorf("선택 실패")
}
