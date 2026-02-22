package api

import "net/url"

// GetCompany fetches company overview for the given corp code.
func (c *Client) GetCompany(corpCode string) (*CompanyInfo, error) {
	params := url.Values{}
	params.Set("corp_code", corpCode)

	var result CompanyInfo
	if err := c.get("/api/company.json", params, &result); err != nil {
		return nil, err
	}
	if err := checkStatus(result.BaseResponse); err != nil {
		return nil, err
	}
	return &result, nil
}
