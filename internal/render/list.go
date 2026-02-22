package render

import (
	"fmt"
	"strings"

	"github.com/seapy/dartcli/internal/api"
)

// ListMarkdown returns a markdown table for disclosure list items.
func ListMarkdown(corpName string, items []api.DisclosureItem) string {
	var sb strings.Builder

	fmt.Fprintf(&sb, "# %s 공시 목록\n\n", corpName)
	fmt.Fprintf(&sb, "총 **%d**건\n\n", len(items))

	sb.WriteString("| 접수일 | 공시명 | 제출인 | 접수번호 |\n")
	sb.WriteString("|--------|--------|--------|----------|\n")

	for _, item := range items {
		date := FormatDate(item.RceptDt)
		name := item.ReportNm
		if item.RmFlag != "" {
			name += " [" + item.RmFlag + "]"
		}
		fmt.Fprintf(&sb, "| %s | %s | %s | `%s` |\n",
			date, name, item.Flr, item.RceptNo)
	}

	sb.WriteString("\n")
	return sb.String()
}
