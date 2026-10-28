package model

import "time"

// ForexRates is the object inside "data".
type ForexRates struct {
	Target string      `json:"target"` // the base currency, "NGN"
	Date   *time.Time  `json:"date"`
	Rates  []ForexRate `json:"rates"`
	Meta   Meta        `json:"-"`
}

// ForexRate is one currency's rate against NGN.
type ForexRate struct {
	Currency           string     `json:"currency"`
	Rate               float64    `json:"rate"`
	InverseRate        float64    `json:"inverse_rate"`
	DailyChange        float64    `json:"daily_change"`
	DailyChangePercent float64    `json:"daily_change_percent"`
	LastUpdated        *time.Time `json:"last_updated"`
}

// NGXTradingSession holds the NGX trading hours.
type NGXTradingSession struct {
	OpenTime  string `json:"open_time"`
	CloseTime string `json:"close_time"`
	Timezone  string `json:"timezone"`
}

// NGXMarketStatus is the response from GET /market/status.
type NGXMarketStatus struct {
	IsOpen   bool              `json:"is_open"`
	Status   string            `json:"status"` // "open" or "closed"
	Reason   string            `json:"reason"` // open, weekend, holiday, pre_market, after_hours
	Session  NGXTradingSession `json:"session"`
	Holiday  *NGXHoliday       `json:"holiday"`   // nil unless today is a public holiday
	NextOpen *NGXNextOpen      `json:"next_open"` // nil while the market is open
	ClosesIn *NGXClosesIn      `json:"closes_in"` // nil unless the market is open
	AsOf     time.Time         `json:"as_of"`
	Meta     Meta              `json:"-"` // copy this line from ForexRates
}

type NGXHoliday struct {
	Name string `json:"name"`
	Date string `json:"date"`
}

type NGXNextOpen struct {
	Date     string    `json:"date"`
	Datetime time.Time `json:"datetime"`
	Label    string    `json:"label"`
}

type NGXClosesIn struct {
	Hours        int `json:"hours"`
	Minutes      int `json:"minutes"`
	TotalMinutes int `json:"total_minutes"`
}

// NGXMarketSnapshot is the response from GET /market/snapshot.
type NGXMarketSnapshot struct {
	Date                  string            `json:"date"`
	ASI                   float64           `json:"asi"`
	ASIChange             float64           `json:"asi_change"`
	ASIChangePercent      float64           `json:"asi_change_percent"`
	YTDASIChangePercent   *float64          `json:"ytd_asi_change_percent"`
	Deals                 int64             `json:"deals"`
	Volume                int64             `json:"volume"`
	ValueTraded           float64           `json:"value_traded"`
	TurnoverRate          *float64          `json:"turnover_rate"`
	MarketCap             NGXMarketCap      `json:"market_cap"`
	Breadth               NGXMarketBreadth  `json:"breadth"`
	TotalListedSecurities *int              `json:"total_listed_securities"`
	Session               NGXTradingSession `json:"session"`
	UpdatedAt             time.Time         `json:"updated_at"`
	Meta                  Meta              `json:"-"`
}

type NGXMarketCap struct {
	Equity float64 `json:"equity"`
	Bonds  float64 `json:"bonds"`
	ETFs   float64 `json:"etfs"`
	Total  float64 `json:"total"`
}

type NGXMarketBreadth struct {
	Advancers   int      `json:"advancers"`
	Decliners   int      `json:"decliners"`
	Unchanged   int      `json:"unchanged"`
	Total       int      `json:"total"`
	AdvDecRatio *float64 `json:"adv_dec_ratio"` // nil when decliners is zero
}

// NGXIndex is one entry from GET /indices.
type NGXIndex struct {
	Symbol    string  `json:"symbol"`
	Name      string  `json:"name"`
	Value     float64 `json:"value"`
	ChangePct float64 `json:"change_pct"`
}

// USTradingSession holds the US trading hours.
type USTradingSession struct {
	OpenTime   string `json:"open_time"`
	CloseTime  string `json:"close_time"`
	Timezone   string `json:"timezone"`
	EarlyClose bool   `json:"early_close"`
}

// USMarketStatus is the response from GET /us/market/status.
type USMarketStatus struct {
	IsOpen   bool             `json:"is_open"`
	Status   string           `json:"status"` // "open" or "closed"
	Reason   string           `json:"reason"` // open, weekend, holiday, pre_market, after_hours
	Session  USTradingSession `json:"session"`
	Holiday  *USHoliday       `json:"holiday"`   // nil unless today is a holiday
	NextOpen *USNextOpen      `json:"next_open"` // nil while the market is open
	ClosesIn *USClosesIn      `json:"closes_in"` // nil unless the market is open
	AsOf     time.Time        `json:"as_of"`
	Meta     Meta             `json:"-"`
}

type USHoliday struct {
	Name       string `json:"name"`
	Date       string `json:"date"`
	EarlyClose bool   `json:"early_close"`
}

type USNextOpen struct {
	Date     string    `json:"date"`
	Datetime time.Time `json:"datetime"`
	Label    string    `json:"label"`
}

type USClosesIn struct {
	Hours        int `json:"hours"`
	Minutes      int `json:"minutes"`
	TotalMinutes int `json:"total_minutes"`
}

// USMarketSnapshot is the response from GET /us/market/snapshot.
type USMarketSnapshot struct {
	Indices     []USSnapshotIndex `json:"indices"`
	Covered     int               `json:"covered"` // tracked stocks/ADRs with a current quote
	TotalVolume *int64            `json:"total_volume"`
	Oldest      *time.Time        `json:"oldest"`
	Newest      *time.Time        `json:"newest"`
	Meta        Meta              `json:"-"`
}

type USSnapshotIndex struct {
	Symbol    string     `json:"symbol"`
	Name      string     `json:"name"`
	LastPrice *float64   `json:"last_price"`
	ChangeAbs *float64   `json:"change_abs"`
	ChangePct *float64   `json:"change_pct"`
	QuoteTime *time.Time `json:"quote_time"`
}

// USIndexList is the response from GET /us/indices.
type USIndexList struct {
	Data  []USIndex `json:"data"`
	Count int       `json:"count"`
	Meta  Meta      `json:"-"`
}

// USIndex is one US market index.
type USIndex struct {
	Symbol       string     `json:"symbol"`
	Name         string     `json:"name"`
	LastPrice    *float64   `json:"last_price"`
	ChangeAbs    *float64   `json:"change_abs"`
	ChangePct    *float64   `json:"change_pct"`
	OpenPrice    *float64   `json:"open_price"`
	HighPrice    *float64   `json:"high_price"`
	LowPrice     *float64   `json:"low_price"`
	PrevClose    *float64   `json:"prev_close"`
	Volume       *int64     `json:"volume"`
	High52Wk     *float64   `json:"high_52wk"`
	High52WkDate *string    `json:"high_52wk_date"`
	Low52Wk      *float64   `json:"low_52wk"`
	Low52WkDate  *string    `json:"low_52wk_date"`
	Return52W    *float64   `json:"return_52w"`
	Volatility   *float64   `json:"volatility"`
	QuoteTime    *time.Time `json:"quote_time"`
}
