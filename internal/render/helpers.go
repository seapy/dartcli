package render

import (
	"fmt"
	"math"
	"strconv"
	"strings"
)

// FormatDate converts YYYYMMDD → YYYY-MM-DD.
func FormatDate(s string) string {
	if len(s) == 8 {
		return s[:4] + "-" + s[4:6] + "-" + s[6:]
	}
	return s
}

// FormatAmount converts a raw amount string to 억원 with comma-separated thousands.
func FormatAmount(s string) string {
	if s == "" || s == "-" {
		return "-"
	}
	// Remove commas already present
	s = strings.ReplaceAll(s, ",", "")
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return s
	}
	// Convert to 억원 (1억 = 100,000,000)
	uck := float64(v) / 1e8
	if math.Abs(uck) >= 1 {
		return fmt.Sprintf("%.1f억", uck)
	}
	// Fallback to 만원
	man := float64(v) / 1e4
	return fmt.Sprintf("%.0f만", man)
}

// FormatAmountKRW returns a plain comma-formatted KRW string.
func FormatAmountKRW(s string) string {
	if s == "" || s == "-" {
		return "-"
	}
	s = strings.ReplaceAll(s, ",", "")
	v, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return s
	}
	return commaInt(v)
}

func commaInt(v int64) string {
	negative := v < 0
	if negative {
		v = -v
	}
	str := strconv.FormatInt(v, 10)
	var result strings.Builder
	for i, c := range str {
		if i > 0 && (len(str)-i)%3 == 0 {
			result.WriteRune(',')
		}
		result.WriteRune(c)
	}
	if negative {
		return "-" + result.String()
	}
	return result.String()
}

// GrowthRate computes (current - previous) / |previous| * 100.
func GrowthRate(current, previous string) string {
	current = strings.ReplaceAll(current, ",", "")
	previous = strings.ReplaceAll(previous, ",", "")
	c, err1 := strconv.ParseFloat(current, 64)
	p, err2 := strconv.ParseFloat(previous, 64)
	if err1 != nil || err2 != nil || p == 0 {
		return "-"
	}
	rate := (c - p) / math.Abs(p) * 100
	if rate > 0 {
		return fmt.Sprintf("+%.1f%%", rate)
	}
	return fmt.Sprintf("%.1f%%", rate)
}

// CorpClassLabel converts corp_cls code to Korean.
func CorpClassLabel(cls string) string {
	switch cls {
	case "Y":
		return "유가증권시장"
	case "K":
		return "코스닥"
	case "N":
		return "코넥스"
	case "E":
		return "기타"
	default:
		return cls
	}
}
