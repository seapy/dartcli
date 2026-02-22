package api

// BaseResponse is the common DART API response envelope.
type BaseResponse struct {
	Status  string `json:"status"`
	Message string `json:"message"`
}

// CompanyInfo holds GET /api/company.json response fields.
type CompanyInfo struct {
	BaseResponse
	CorpCode       string `json:"corp_code"`
	CorpName       string `json:"corp_name"`
	CorpNameEng    string `json:"corp_name_eng"`
	StockName      string `json:"stock_name"`
	StockCode      string `json:"stock_code"`
	CEO            string `json:"ceo_nm"`
	CorpCls        string `json:"corp_cls"`
	JurisdictionNo string `json:"jurir_no"`
	BusinessNo     string `json:"bizr_no"`
	Address        string `json:"adres"`
	HomepageURL    string `json:"hm_url"`
	IRHomepageURL  string `json:"ir_url"`
	Phone          string `json:"phn_no"`
	Fax            string `json:"fax_no"`
	Industry       string `json:"induty_code"`
	EstDate        string `json:"est_dt"`
	AccountMonth   string `json:"acc_mt"`
}

// DisclosureItem is one item in GET /api/list.json.
type DisclosureItem struct {
	CorpCode    string `json:"corp_code"`
	CorpName    string `json:"corp_name"`
	StockCode   string `json:"stock_code"`
	CorpCls     string `json:"corp_cls"`
	ReportNm    string `json:"report_nm"`
	RceptNo     string `json:"rcept_no"`
	Flr         string `json:"flr_nm"`
	RceptDt     string `json:"rcept_dt"`
	RmFlag      string `json:"rm"`
}

// ListResponse wraps GET /api/list.json.
type ListResponse struct {
	BaseResponse
	TotalCount int              `json:"total_count"`
	Items      []DisclosureItem `json:"list"`
}

// FinanceAccount is one row in GET /api/fnlttSinglAcnt.json.
type FinanceAccount struct {
	RceptNo     string `json:"rcept_no"`
	ReprtCode   string `json:"reprt_code"`
	BsnsYear    string `json:"bsns_year"`
	CorpCode    string `json:"corp_code"`
	SjDiv       string `json:"sj_div"`
	SjNm        string `json:"sj_nm"`
	AccountId   string `json:"account_id"`
	AccountNm   string `json:"account_nm"`
	AccountDetail string `json:"account_detail"`
	Thstrm_dt   string `json:"thstrm_dt"`
	Thstrm_amount string `json:"thstrm_amount"`
	Frmtrm_dt   string `json:"frmtrm_dt"`
	Frmtrm_amount string `json:"frmtrm_amount"`
	Bfefrmtrm_dt string `json:"bfefrmtrm_dt"`
	Bfefrmtrm_amount string `json:"bfefrmtrm_amount"`
	OrdNo       string `json:"ord"`
	Currency    string `json:"currency"`
}

// FinanceResponse wraps GET /api/fnlttSinglAcnt.json.
type FinanceResponse struct {
	BaseResponse
	Items []FinanceAccount `json:"list"`
}
