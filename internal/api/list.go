package api

import (
	"fmt"
	"net/url"
)

// ListOptions configures the disclosure list query.
type ListOptions struct {
	CorpCode  string
	StartDate string // YYYYMMDD
	EndDate   string // YYYYMMDD
	PblntfTy  string // disclosure type code
	PageNo    int
	PageCount int
}

// GetList fetches the disclosure list for the given options.
func (c *Client) GetList(opts ListOptions) (*ListResponse, error) {
	params := url.Values{}
	params.Set("corp_code", opts.CorpCode)
	params.Set("bgn_de", opts.StartDate)
	params.Set("end_de", opts.EndDate)
	if opts.PblntfTy != "" {
		params.Set("pblntf_ty", opts.PblntfTy)
	}
	if opts.PageNo > 0 {
		params.Set("page_no", fmt.Sprintf("%d", opts.PageNo))
	}
	if opts.PageCount > 0 {
		params.Set("page_count", fmt.Sprintf("%d", opts.PageCount))
	}

	var result ListResponse
	if err := c.get("/api/list.json", params, &result); err != nil {
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
