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
	app, err := c.api.GetMe()
	if err != nil {
		return nil, err
	}
	if !app.IsSuccessfully() {
		return nil, app.Error
	}
	return app.Result, nil
}

// CreateInvoice is representation for api/createInvoice.
//
func (c *Client) CreateInvoice(asset Asset, amount float64, opt CreateInvoiceOptions) (*Invoice, error) {
	if asset != "" {
		opt.Asset = asset
	}
	if amount != 0 {
		opt.Amount = strconv.FormatFloat(amount, 'f', -1, 64)
	}
	invoice, err := c.api.CreateInvoice(opt)
	if err != nil {
		return nil, err
	}
	if !invoice.IsSuccessfully() {
		return nil, invoice.Error
	}
	return invoice.Result, nil
}

// DoTransfer is representation for api/transfer. Error regular or ApiError.
//
// If you want set regular params in opt - set regular parameters default value (empty string for Asset & string, 0 for numbers)
// spendId must be unique for every operation.
func (c *Client) DoTransfer(userId int, asset Asset, amount float64, spendId string, opt DoTransferOptions) (*Transfer, error) {
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
		opt.Amount = strconv.FormatFloat(amount, 'f', -1, 64)
	}
	transfer, err := c.api.DoTransfer(opt)
	if err != nil {
		return nil, err
	}
	if !transfer.IsSuccessfully() {
		return nil, transfer.Error
	}
	return transfer.Result, nil
}

// GetInvoices is representation for api/getInvoices.
// Set opt parameter as nil for empty API params.
func (c *Client) GetInvoices(opt *GetInvoicesOptions) ([]Invoice, error) {
	invoices, err := c.api.GetInvoices(opt)
	if err != nil {
		return nil, err
	}
	if !invoices.IsSuccessfully() {
		return nil, invoices.Error
	}
	return invoices.Result.Items, nil
}

// GetBalance is representation for api/getBalance.
func (c *Client) GetBalance() (BalanceInfo, error) {
	balance, err := c.api.GetBalance()
	if err != nil {
		return nil, err
	}
	if !balance.IsSuccessfully() {
		return nil, balance.Error
	}
	return balance.Result, nil
}

// GetExchangeRates is representation for api/getExchangeRates.
func (c *Client) GetExchangeRates() (ExchangeRateArray, error) {
	exchangeInfo, err := c.api.GetExchangeRates()
	if err != nil {
		return nil, err
	}
	if !exchangeInfo.IsSuccessfully() {
		return nil, exchangeInfo.Error
	}
	return exchangeInfo.Result, nil
}

// GetCurrencies is representation for api/getCurrencies.
func (c *Client) GetCurrencies() (CurrencyInfoArray, error) {
	currencies, err := c.api.GetCurrencies()
	if err != nil {
		return nil, err
	}
	if !currencies.IsSuccessfully() {
		return nil, currencies.Error
	}
	return currencies.Result, nil
}

// On alias for Webhook.Bind. Add handler to slice for given update type. Return index of new handler
func (c *Client) On(updateType UpdateType, handler Handler) int {
	return c.w.Bind(updateType, handler)
}

// OnInvoicePaid is shortcut for Client.On with update type "invoice_paid".
func (c *Client) OnInvoicePaid(handler Handler) int {
	return c.w.Bind(UpdateInvoicePaid, handler)
}

// DeleteAllHandlersFor alias for Webhook.DeleteHandlers.
//
// Delete all handlers for given update type.
// Also if update type is "*" reset handlers to empty value for EVERY UPDATE TYPE
func (c *Client) DeleteAllHandlersFor(updateType UpdateType) {
	c.w.DeleteHandlers(updateType)
}

func (c *Client) DeleteHandler(updateType UpdateType, i int) {
	c.w.DeleteHandlerByIndex(updateType, i)
}

// Once add handler that will call once.
func (c *Client) Once(updateType UpdateType, handler Handler) {
	// Index for adding handler
	i := len(c.w.handlers[updateType])

	c.w.Bind(updateType, func(update *WebhookUpdate) {
		handler(update)
		a := c.w.handlers[updateType]
		c.w.handlers[updateType] = append(a[:i], a[i+1:]...)
	})
}

// IsSuccessfully indicates whether API request success.
func (r BaseApiResponse) IsSuccessfully() bool {
	return r.Ok && r.Error == nil
}

// AsMap returns transformed BalanceInfo ([]BalanceCurrency) into map,
// key - currency code (Asset), value - balance for Asset as string
func (b BalanceInfo) AsMap() map[Asset]string {
	balances := make(map[Asset]string)
	for _, currency := range b {
		balances[currency.CurrencyCode] = currency.Available
	}
	return balances
}

// AsMapFloat returns transformed BalanceInfo ([]BalanceCurrency) into map,
// key - currency code (Asset), value - balance for Asset as float64
func (b BalanceInfo) AsMapFloat() (map[Asset]float64, error) {
	balances := make(map[Asset]float64)
	for _, currency := range b {
		balance, err := strconv.ParseFloat(currency.Available, 64)
		if err != nil {
			return nil, err
		}
		balances[currency.CurrencyCode] = balance
	}
	return balances, nil
}

// AsMap returns transformed CurrencyInfoArray ([]CurrencyInfo) into map.
func (c CurrencyInfoArray) AsMap() map[Asset]CurrencyInfo {
	currencies := make(map[Asset]CurrencyInfo)
	for _, currency := range c {
		currencies[currency.Code] = currency
	}
	return currencies
}

// RatesKey is two-value key for ExchangeRateArray.
type RatesKey struct {
	Source Asset
	Target Asset
}

// AsMap returns transformed ExchangeRateArray ([]ExchangeRate) into map.
func (e ExchangeRateArray) AsMap() map[RatesKey]ExchangeRate {
	rates := make(map[RatesKey]ExchangeRate)
	for _, exchangeRate := range e {
		key := RatesKey{exchangeRate.Source, exchangeRate.Target}
		rates[key] = exchangeRate
	}
	return rates
}

// Get returns exchange rate of target currency in source currency and the success indicator.
func (e ExchangeRateArray) Get(source, target Asset) (string, bool) {
	rate, ok := e.AsMap()[RatesKey{source, target}]
	if !ok || !rate.IsValid {
		return "", false
	}
	return rate.Rate, true
}
