package render

import (
	"fmt"
	"strings"

	"github.com/seapy/dartcli/internal/api"
)

// CompanyMarkdown returns a markdown string for the company overview.
func CompanyMarkdown(info *api.CompanyInfo) string {
	var sb strings.Builder

	cls := CorpClassLabel(info.CorpCls)
	if info.StockCode != "" {
		fmt.Fprintf(&sb, "# %s (%s)\n\n", info.CorpName, info.StockCode)
	} else {
		fmt.Fprintf(&sb, "# %s\n\n", info.CorpName)
	}

	if info.CorpNameEng != "" {
		fmt.Fprintf(&sb, "> %s\n\n", info.CorpNameEng)
	}

	sb.WriteString("## 기업 개요\n\n")
	sb.WriteString("| 항목 | 내용 |\n")
	sb.WriteString("|------|------|\n")

	rows := []struct{ key, val string }{
		{"법인구분", cls},
		{"대표이사", info.CEO},
		{"설립일", FormatDate(info.EstDate)},
		{"결산월", info.AccountMonth + "월"},
		{"사업자번호", info.BusinessNo},
		{"법인등록번호", info.JurisdictionNo},
		{"주소", info.Address},
		{"홈페이지", info.HomepageURL},
		{"IR 홈페이지", info.IRHomepageURL},
		{"전화번호", info.Phone},
		{"팩스번호", info.Fax},
	}

	for _, r := range rows {
		if r.val != "" && r.val != "월" {
			fmt.Fprintf(&sb, "| %s | %s |\n", r.key, r.val)
		}
	}

	sb.WriteString("\n")
	return sb.String()
}
