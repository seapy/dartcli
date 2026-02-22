package api

import (
	"fmt"
	"net/url"
)

// FinanceOptions configures the single account query.
type FinanceOptions struct {
	CorpCode  string
	BsnsYear  string // 4-digit year
	ReprtCode string // 11011=annual, 11013=q1, 11012=half, 11014=q3
	FsDiv     string // CFS=연결, OFS=개별
}

// GetFinance fetches single account financial statements.
func (c *Client) GetFinance(opts FinanceOptions) (*FinanceResponse, error) {
	params := url.Values{}
	params.Set("corp_code", opts.CorpCode)
	params.Set("bsns_year", opts.BsnsYear)
	params.Set("reprt_code", opts.ReprtCode)
	if opts.FsDiv != "" {
		params.Set("fs_div", opts.FsDiv)
	}

	var result FinanceResponse
	if err := c.get("/api/fnlttSinglAcnt.json", params, &result); err != nil {
		return nil, err
	}
	// 013 = 조회된 데이터 없음 → 빈 결과로 처리
	if result.Status == "013" {
		return &result, nil
	}
	if err := checkStatus(result.BaseResponse); err != nil {
		return nil, err
	}
	return &result, nil
}

// ReprtCode maps period string to DART report code.
func ReprtCode(period string) string {
	switch period {
	case "q1":
		return "11013"
	case "half":
		return "11012"
	case "q3":
		return "11014"
	default: // annual
		return "11011"
	}
}

// PeriodLabel returns a human-readable label for a period string.
func PeriodLabel(period string) string {
	switch period {
	case "q1":
		return "1분기"
	case "half":
		return "반기"
	case "q3":
		return "3분기"
	default:
		return "연간"
	}
}

// FsDivLabel returns a human-readable label for fs_div.
func FsDivLabel(fsDiv string) string {
	switch fsDiv {
	case "OFS":
		return "개별"
	default:
		return "연결"
	}
}

// FsDivCode normalises the fs_div string.
func FsDivCode(s string) string {
	switch s {
	case "ofs", "OFS":
		return "OFS"
	default:
		return "CFS"
	}
}

// GetFinanceMultiAccount fetches multi-account financial statements.
func (c *Client) GetFinanceMultiAccount(opts FinanceOptions) (*FinanceResponse, error) {
	params := url.Values{}
	params.Set("corp_code", opts.CorpCode)
	params.Set("bsns_year", opts.BsnsYear)
	params.Set("reprt_code", opts.ReprtCode)
	if opts.FsDiv != "" {
		params.Set("fs_div", opts.FsDiv)
	}

	var result FinanceResponse
	if err := c.get("/api/fnlttMultiAcnt.json", params, &result); err != nil {
		return nil, err
	}
	// Some accounts return 013 (no data) which is not a fatal error
	if result.Status != "000" && result.Status != "013" {
		return nil, fmt.Errorf("DART API error %s: %s", result.Status, result.Message)
	}
	return &result, nil
}
