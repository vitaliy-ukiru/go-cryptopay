package cryptopay

import "net/url"

const (
	getMeMethod            = "getMe"
	createInvoiceMethod    = "createInvoice"
	transferMethod         = "transfer"
	getInvoicesMethod      = "getInvoices"
	getBalanceMethod       = "getBalance"
	getExchangeRatesMethod = "getExchangeRates"
	getCurrenciesMethod    = "getCurrencies"
)

// CreateInvoiceRequest ...
type CreateInvoiceRequest struct {
	Options CreateInvoiceOptions
}

func (CreateInvoiceRequest) Endpoint() string    { return createInvoiceMethod }
func (c CreateInvoiceRequest) Query() url.Values { return c.Options.QueryParams() }

// DoTransferRequest ...
type DoTransferRequest struct {
	Options DoTransferOptions
}

func (DoTransferRequest) Endpoint() string    { return transferMethod }
func (r DoTransferRequest) Query() url.Values { return r.Options.QueryParams() }

// GetInvoicesRequest ...
type GetInvoicesRequest struct {
	Options *GetInvoicesOptions
}

func (GetInvoicesRequest) Endpoint() string    { return getInvoicesMethod }
func (r GetInvoicesRequest) Query() url.Values { return r.Options.QueryParams() }

// GetMeRequest ...
type GetMeRequest struct{}

func (GetMeRequest) Endpoint() string    { return getMeMethod }
func (g GetMeRequest) Query() url.Values { return nil }

// GetBalanceRequest ...
type GetBalanceRequest struct{}

func (GetBalanceRequest) Endpoint() string    { return getBalanceMethod }
func (g GetBalanceRequest) Query() url.Values { return nil }

// GetExchangeRatesRequest ...
type GetExchangeRatesRequest struct{}

func (GetExchangeRatesRequest) Endpoint() string    { return getExchangeRatesMethod }
func (g GetExchangeRatesRequest) Query() url.Values { return nil }

// GetCurrenciesRequest ...
type GetCurrenciesRequest struct{}

func (GetCurrenciesRequest) Endpoint() string    { return getCurrenciesMethod }
func (g GetCurrenciesRequest) Query() url.Values { return nil }
