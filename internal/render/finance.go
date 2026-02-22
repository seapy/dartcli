package render

import (
	"fmt"
	"strings"

	"github.com/seapy/dartcli/internal/api"
)

// FinanceMarkdown renders financial statements as markdown tables.
func FinanceMarkdown(corpName, year, periodLabel, fsDivLabel string, items []api.FinanceAccount) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# %s 재무정보\n\n", corpName)
	fmt.Fprintf(&sb, "**%s년 %s %s 기준**\n\n", year, fsDivLabel, periodLabel)

	// Group by sj_div: BS, IS, CF, etc.
	groups := map[string][]api.FinanceAccount{}
	order := []string{}
	seen := map[string]bool{}

	for _, item := range items {
		key := item.SjDiv + "|" + item.SjNm
		if !seen[key] {
			seen[key] = true
			order = append(order, key)
		}
		groups[key] = append(groups[key], item)
	}

	for _, key := range order {
		parts := strings.SplitN(key, "|", 2)
		sectionName := parts[1]
		accts := groups[key]

		fmt.Fprintf(&sb, "## %s\n\n", sectionName)
		sb.WriteString("| 계정과목 | 당기 | 전기 | 증감률 |\n")
		sb.WriteString("|----------|------|------|--------|\n")

		for _, a := range accts {
			growth := GrowthRate(a.Thstrm_amount, a.Frmtrm_amount)
			fmt.Fprintf(&sb, "| %s | %s | %s | %s |\n",
				a.AccountNm,
				FormatAmount(a.Thstrm_amount),
				FormatAmount(a.Frmtrm_amount),
				growth,
			)
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
