package model

import (
	"net/url"
	"time"
)

// Types for the Nigerian Exchange (NGX) endpoints. Prices are in NGN. NGX
// symbols are case-insensitive.
type (
	// Company is one row of the NGX company list, with its current price.
	Company struct {
		ID                   int        `json:"id"`
		Symbol               string     `json:"symbol"`
		Name                 string     `json:"name"`
		LogoURL              string     `json:"logo_url"`
		Sector               string     `json:"sector"`
		SubSector            string     `json:"sub_sector"`
		MarketClassification string     `json:"market_classification"`
		SharesOutstanding    float64    `json:"shares_outstanding"`
		Website              string     `json:"website"`
		Price                float64    `json:"price"`
		PrevClose            float64    `json:"prev_close"`
		DayHigh              *float64   `json:"day_high"`
		DayLow               *float64   `json:"day_low"`
		Volume               float64    `json:"volume"`
		MarketCap            float64    `json:"market_cap"`
		PriceChange          float64    `json:"price_change"`
		PriceChangePercent   float64    `json:"price_change_percent"`
		Change7DPercent      float64    `json:"change_7d_percent"`
		Change1MPercent      float64    `json:"change_1m_percent"`
		Change52WPercent     float64    `json:"change_52w_percent"`
		ChangeYTDPercent     float64    `json:"change_ytd_percent"`
		High52Wk             float64    `json:"high_52wk"`
		Low52Wk              float64    `json:"low_52wk"`
		LastUpdated          *time.Time `json:"last_updated"`
	}

	// CompanyList is a page of companies.
	CompanyList struct {
		Companies  []Company
		Pagination Pagination
		Meta       Meta
	}

	// CompanySort is a sortable field of the company list.
	CompanySort string

	// ListCompaniesParams filters and pages the company list. The zero value
	// requests the first page with the API defaults (50 per page, by market cap).
	ListCompaniesParams struct {
		Page         int
		Limit        int
		Sector       string
		Search       string // matches name or symbol
		Sort         CompanySort
		Order        SortOrder
		MinMarketCap *float64
		MaxMarketCap *float64
	}

	// CompanyIdentifier is a lightweight symbol entry.
	CompanyIdentifier struct {
		ID                 int    `json:"id"`
		Symbol             string `json:"symbol"`
		Name               string `json:"name"`
		InternationalSecID string `json:"international_sec_id"`
		LogoURL            string `json:"logo_url"`
	}

	// CompanyIdentifiers is the full NGX symbol list.
	CompanyIdentifiers struct {
		Companies []CompanyIdentifier
		Count     int
		Meta      Meta
	}

	// CompanyDetail is the full profile and quote of one company (Hobby plan).
	CompanyDetail struct {
		ID                 int      `json:"id"`
		Symbol             string   `json:"symbol"`
		Name               string   `json:"name"`
		LogoURL            string   `json:"logo_url"`
		InternationalSecID string   `json:"international_sec_id"`
		Sector             string   `json:"sector"`
		SubSector          string   `json:"sub_sector"`
		Classification     string   `json:"market_classification"`
		SharesOutstanding  float64  `json:"shares_outstanding"`
		DateListed         string   `json:"date_listed"`
		About              string   `json:"about"`
		Website            string   `json:"website"`
		NatureOfBusiness   string   `json:"nature_of_business"`
		CurrentPrice       float64  `json:"current_price"`
		PrevClose          float64  `json:"prev_close"`
		OpenPrice          float64  `json:"open_price"`
		DayHigh            float64  `json:"day_high"`
		DayLow             float64  `json:"day_low"`
		Volume             float64  `json:"volume"`
		ValueTraded        float64  `json:"value_traded"`
		MarketCap          float64  `json:"market_cap"`
		PriceChange        float64  `json:"price_change"`
		PriceChangePercent float64  `json:"price_change_percent"`
		High52Wk           float64  `json:"high52wk"`
		Low52Wk            float64  `json:"low52wk"`
		TTMEPS             *float64 `json:"ttm_eps"`
		PBRatio            *float64 `json:"pb_ratio"`
		DividendYield      *float64 `json:"dividend_yield"`
		LastUpdated        string   `json:"last_updated"`
		Meta               Meta     `json:"-"`
	}

	// NGXChartParams selects the history window. Use Period for a named window,
	// or From/To for explicit dates (UTC calendar days). A From older than the
	// plan's history depth (Hobby 2y, Starter 5y, Pro+ full) is silently clamped
	// by the API.
	NGXChartParams struct {
		Period string // "7d", "30d", "90d", "1y", "5y" or "all"
		From   time.Time
		To     time.Time
	}

	// NGXChartPoint is one trading day. Close is the only field the API
	// guarantees; the others are nil when no intraday data exists for that day.
	NGXChartPoint struct {
		Date          string   `json:"date"`
		Timestamp     int64    `json:"timestamp"`
		Close         *float64 `json:"close"`
		Price         *float64 `json:"price"`
		Open          *float64 `json:"open"`
		High          *float64 `json:"high"`
		Low           *float64 `json:"low"`
		Volume        *float64 `json:"volume"`
		ValueTraded   *float64 `json:"value_traded"`
		VWAP          *float64 `json:"vwap"`
		TradeCount    *float64 `json:"trade_count"`
		Change        *float64 `json:"change"`
		ChangePercent *float64 `json:"change_percent"`
	}

	// NGXChartStats summarises the returned range.
	NGXChartStats struct {
		FirstPrice         float64  `json:"first_price"`
		LastPrice          float64  `json:"last_price"`
		MinPrice           float64  `json:"min_price"`
		MaxPrice           float64  `json:"max_price"`
		PriceChange        float64  `json:"price_change"`
		PriceChangePercent *float64 `json:"price_change_percent"`
		StartDate          string   `json:"start_date"`
		EndDate            string   `json:"end_date"`
	}

	// NGXChart is a company's daily price history.
	NGXChart struct {
		Symbol      string          `json:"symbol"`
		CompanyName string          `json:"company_name"`
		Count       int             `json:"count"`
		Points      []NGXChartPoint `json:"data"`
		Statistics  NGXChartStats   `json:"statistics"`
		Meta        Meta            `json:"-"`
	}
)

// Sortable fields for ListCompaniesParams.Sort.
const (
	CompanySortSymbol             CompanySort = "symbol"
	CompanySortName               CompanySort = "company_name"
	CompanySortSector             CompanySort = "sector"
	CompanySortPrice              CompanySort = "current_price"
	CompanySortMarketCap          CompanySort = "market_cap"
	CompanySortVolume             CompanySort = "volume"
	CompanySortPriceChangePercent CompanySort = "price_change_percent"
)

// Values encodes the parameters as a query string. Zero values are omitted.
func (p *ListCompaniesParams) Values() url.Values {
	q := url.Values{}
	if p == nil {
		return q
	}
	setInt(q, "page", p.Page)
	setInt(q, "limit", p.Limit)
	setStr(q, "sector", p.Sector)
	setStr(q, "search", p.Search)
	setStr(q, "sort", string(p.Sort))
	setStr(q, "order", string(p.Order))
	setFloat(q, "minMarketCap", p.MinMarketCap)
	setFloat(q, "maxMarketCap", p.MaxMarketCap)
	return q
}

// Values encodes the parameters as a query string. It always asks for the
// "detailed" row format so the response shape is stable.
func (p *NGXChartParams) Values() url.Values {
	q := url.Values{"format": {"detailed"}}
	if p == nil {
		return q
	}
	setStr(q, "period", p.Period)
	setDate(q, "from", p.From)
	setDate(q, "to", p.To)
	return q
}

// ClosePrice returns the day's close, falling back to the legacy price alias.
func (p NGXChartPoint) ClosePrice() (float64, bool) {
	switch {
	case p.Close != nil:
		return *p.Close, true
	case p.Price != nil:
		return *p.Price, true
	}
	return 0, false
}

// Day returns the point's calendar date ("YYYY-MM-DD"), derived from Timestamp
// when Date is empty. It returns "" if neither is present.
func (p NGXChartPoint) Day() string {
	if p.Date != "" {
		return p.Date
	}
	if p.Timestamp == 0 {
		return ""
	}
	ts := p.Timestamp
	if ts > 1e12 { // milliseconds
		ts /= 1000
	}
	return time.Unix(ts, 0).UTC().Format(DateLayout)
}
