package cryptopay

import (
	"net/http"
	"strconv"
)

type (
	// ExchangeRateArray alias for slice of ExchangeRate.
	ExchangeRateArray []ExchangeRate
	// BalanceInfo alias for slice of BalanceCurrency.
	BalanceInfo []BalanceCurrency
	// CurrencyInfoArray alias for slice of CurrencyInfo
	CurrencyInfoArray []CurrencyInfo
)

// WebhookSettings for configure webhook in ClientSettings.
type WebhookSettings struct {
	// OnError is handler for error in webhook.
	OnError func(r *http.Request, err error)
	// DefaultHandlers is set of default handlers. Default creates new set.
	DefaultHandlers map[UpdateType][]Handler
}

// ClientSettings for easy configure NewClient.
type ClientSettings struct {
	// Token of CryptoPay App.
	Token string
	// ApiHost url to api host. Default mainnet (MainNetHost).
	ApiHost string
	// HttpClient for make requests. Default http.DefaultClient.
	HttpClient *http.Client
	// Webhook settings. If set default value webhook can correct work.
	Webhook WebhookSettings
}

// Client is high-level API.
//
// Methods that call API and return error can return ApiError.
// For get ApiError use GetApiError.
//
// If you want set regular params in opt parameter - set regular parameters default value (empty string for Asset & string, 0 for numbers).
type Client struct {
	// api for requests to CryptoBot API..
	api *Api
	// w is Webhook instance for get update from API.
	// Client just have Webhook object & aliases for Webhook methods.
	w *Webhook
}

// NewClient returns new Client.
func NewClient(settings ClientSettings) *Client {
	if settings.Token == "" {
		panic("invalid token")
	}
	httpClient := settings.HttpClient
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	apiHost := settings.ApiHost
	if apiHost == "" {
		apiHost = MainNetHost
	}
	api := NewApi(settings.Token, apiHost, httpClient)

	w := NewWebhook(settings.Token, settings.Webhook.DefaultHandlers, settings.Webhook.OnError)
	return &Client{
		api: api,
		w:   w,
	}
}

// Api return instance of Api
func (c Client) Api() *Api { return c.api }

// Webhook return instance of Webhook
func (c Client) Webhook() Webhook { return *c.w }

// GetMe is representation of api/getMe.
func (c *Client) GetMe() (*AppInfo, error) {
	return doRequest[*AppInfo](c.api, GetMeRequest{})
}

// CreateInvoice is representation for api/createInvoice.
//
func (c *Client) CreateInvoice(asset Asset, amount float64, params *CreateInvoiceOptions) (*Invoice, error) {
	var opt CreateInvoiceOptions
	if params != nil {
		opt = *params
	}
	if asset != "" {
		opt.Asset = asset
	}
	if amount != 0 {
		opt.Amount = new(big.Float).SetFloat64(amount)
	}
	return doRequest[*Invoice](c.api, CreateInvoiceRequest{Options: opt})
}

// DoTransfer is representation for api/transfer. Error regular or ApiError.
//
// If you want set regular params in opt - set regular parameters default value (empty string for Asset & string, 0 for numbers)
// spendId must be unique for every operation.
func (c *Client) DoTransfer(userId int64, asset Asset, amount float64, spendId string, params *DoTransferOptions) (*Transfer, error) {
	var opt DoTransferOptions
	if params != nil {
		opt = *params
	}
	if userId != 0 {
		opt.UserId = userId
	}
	if asset != "" {
		opt.Asset = asset
	}
	if spendId != "" {
		opt.SpendId = spendId
	}
	if amount != 0 {
		opt.Amount = new(big.Float).SetFloat64(amount)
	}
	return doRequest[*Transfer](c.api, DoTransferRequest{Options: opt})
}

// GetInvoices is representation for api/getInvoices.
// Set opt parameter as nil for empty API params.
func (c *Client) GetInvoices(opt *GetInvoicesOptions) ([]Invoice, error) {
	resp, err := doRequest[struct {
		Items []Invoice `json:"items"`
	}](c.api, GetInvoicesRequest{Options: opt})
	return resp.Items, err
}

// GetBalance is representation for api/getBalance.
func (c *Client) GetBalance() (BalanceInfo, error) {
	return doRequest[BalanceInfo](c.api, GetBalanceRequest{})
}

// GetExchangeRates is representation for api/getExchangeRates.
func (c *Client) GetExchangeRates() (ExchangeRates, error) {
	return doRequest[ExchangeRates](c.api, GetExchangeRatesRequest{})

}

// GetCurrencies is representation for api/getCurrencies.
func (c *Client) GetCurrencies() (CurrencyInfoArray, error) {
	return doRequest[CurrencyInfoArray](c.api, GetCurrenciesRequest{})
}

// doRequest needed because in go 1.18 methods cannot have type params
// but this function solving this problem.
func doRequest[T any](api *Api, request Request) (T, error) {
	resp := new(Response[T])
	var err error
	if err = api.Do(request, resp); err != nil {
		return resp.Result, err
	}

	if resp.Error != nil {
		err = resp.Error
	}
	return resp.Result, err
}
