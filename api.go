package go_crypto_pay

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// Aliases for officials hosts
const (
	MainNetHost = "https://pay.crypt.bot"
	TestNetHost = "https://testnet-pay.crypt.bot"
)

const (
	emptyQuery             = ""
	getMeMethod            = "getMe"
	createInvoiceMethod    = "createInvoice"
	transferMethod         = "transfer"
	getInvoicesMethod      = "getInvoices"
	getBalanceMethod       = "getBalance"
	getExchangeRatesMethod = "getExchangeRates"
	getCurrenciesMethod    = "getCurrencies"
	headerTokenName        = "Crypto-Pay-API-Token"
)

type (
	// CreateInvoiceOptions for `createInvoice` api method.
	CreateInvoiceOptions struct {
		Asset          Asset      // Currency code.
		Amount         string     // Amount of the invoice in float.
		Description    string     // Optional. Description for the invoice. User will see this description when they pay the invoice. Up to 1024 characters.
		HiddenMessage  string     // Optional. Text of the message that will be shown to a user after the invoice is paid. Up to 2o48 characters.
		PaidButtonName PaidButton // Optional. Name of the button that will be shown to a user after the invoice is paid.
		PaidButtonUrl  string     // Optional. Required if PaidButtonName is used. URL to be opened when the button is pressed. You can set any success link (for example, a link to your bot). Starts with https or http.
		Payload        string     // Optional. Any data you want to attach to the invoice (for example, user ID, payment ID, ect). Up to 4kb.
		AllowComments  bool       // Optional. Allow a user to add a comment to the payment. Default is true.
		AllowAnonymous bool       // Optional. Allow a user to pay the invoice anonymously. Default is true.
		ExpiresIn      int        // Optional. You can set a payment time limit for the invoice in seconds. Values between 1-2678400 are accepted
	}
	// DoTransferOptions for `transfer` (DoTransfer) api method.
	DoTransferOptions struct {
		UserId                  int    // Telegram user ID. User must have previously used @CryptoBot (@CryptoTestnetBot for testnet).
		Asset                   Asset  // Currency code.
		Amount                  string // Amount of the invoice in float.
		SpendId                 string // Unique ID to make your request idempotent and ensure that only one of the transfers with the same spend_id will be accepted by Crypto Pay API. More https://telegra.ph/Crypto-Pay-API-11-25#transfer
		Comment                 string // Optional. Comment for the transfer. Users will see this comment when they receive a notification about the transfer. Up to 1024 symbols.
		DisableSendNotification bool   // Optional. Pass true if the user should not receive a notification about the transfer. Default is false.
	}
	// GetInvoicesOptions for `getInvoices` api method.
	GetInvoicesOptions struct {
		Asset      Asset         // Currency code.
		InvoiceIds []string      // Optional. Invoice IDs
		Status     InvoiceStatus // Optional. Status of invoices to be returned. Defaults to all statuses.
		Offset     int           // Optional. Offset needed to return a specific subset of invoices. Default is 0.
		Count      int           // Optional. Number of invoices to be returned. Values between 1-1000 are accepted. Defaults to 100.
	}
)

type (
	// BaseApiResponse  is contained in all api responses .
	BaseApiResponse struct {
		// Ok indicates whether the request was successfully executed.
		Ok bool `json:"ok"`
		// Error from API, nil on successfully.
		Error *ApiError `json:"error,omitempty"`
	}
	// GetMeResponse  for `getMe` method
	GetMeResponse struct {
		BaseApiResponse
		Result *AppInfo `json:"result,omitempty"`
	}
	// CreateInvoiceResponse  for `createInvoice` method
	CreateInvoiceResponse struct {
		BaseApiResponse
		Result *Invoice `json:"result,omitempty"`
	}
	// DoTransferResponse for `transfer` method
	DoTransferResponse struct {
		BaseApiResponse
		Result *Transfer `json:"result,omitempty"`
	}
	// GetInvoicesResponse for `getInvoices` method
	GetInvoicesResponse struct {
		BaseApiResponse
		Result struct {
			Items []Invoice `json:"items"`
		} `json:"result,omitempty"`
	}
	// GetBalanceResponse for `getBalance` method
	GetBalanceResponse struct {
		BaseApiResponse
		Result []BalanceCurrency `json:"result,omitempty"`
	}
	// GetExchangeRatesResponse for `getExchangeRates` method
	GetExchangeRatesResponse struct {
		BaseApiResponse
		Result []ExchangeRate `json:"result,omitempty"`
	}
	// GetCurrenciesResponse for `getCurrencies` method
	GetCurrenciesResponse struct {
		BaseApiResponse
		Result []CurrencyInfo `json:"result,omitempty"`
	}
)

type ApiCore struct {
	token      string
	url        string
	httpClient *http.Client
}

// NewApi is constructor function for ApiCore
func NewApi(token, url string, httpClient *http.Client) *ApiCore {
	return &ApiCore{token: token, url: url, httpClient: httpClient}
}

// urlFmt formatting URL, paste query params
func (c ApiCore) urlFmt(method string, queryParams string) string {
	methodUrl := fmt.Sprintf("%s/api/%s", c.url, method)
	if queryParams == emptyQuery {
		return methodUrl
	}
	return methodUrl + "?" + queryParams
}

// apiCall make request to API and deserialization response body in dest argument
func (c ApiCore) apiCall(method, queryParams string, dest interface{}) error {
	req, err := http.NewRequest("GET", c.urlFmt(method, queryParams), nil)
	if err != nil {
		return err
	}
	req.Header.Set(headerTokenName, c.token)
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}
	return json.NewDecoder(resp.Body).Decode(&dest)

}

// GetMe call api/getMe.
func (c ApiCore) GetMe() (*GetMeResponse, error) {
	appInfo := new(GetMeResponse)
	if err := c.apiCall(getMeMethod, emptyQuery, appInfo); err != nil {
		return nil, err
	}
	return appInfo, nil
}

// CreateInvoice call api/createInvoice.
func (c ApiCore) CreateInvoice(opt CreateInvoiceOptions) (*CreateInvoiceResponse, error) {
	newInvoice := new(CreateInvoiceResponse)
	if err := c.apiCall(createInvoiceMethod, opt.QueryParams(), newInvoice); err != nil {
		return nil, err
	}
	return newInvoice, nil
}

// DoTransfer call api/transfer.
func (c ApiCore) DoTransfer(opt DoTransferOptions) (*DoTransferResponse, error) {
	newTransfer := new(DoTransferResponse)
	if err := c.apiCall(transferMethod, opt.QueryParams(), newTransfer); err != nil {
		return nil, err
	}
	return newTransfer, nil
}

// GetInvoices call api/getInvoices. Set opt as nil for empty API params.
func (c ApiCore) GetInvoices(opt *GetInvoicesOptions) (*GetInvoicesResponse, error) {
	invoices := new(GetInvoicesResponse)
	var queryParams string
	if opt != nil {
		queryParams = opt.QueryParams()
	}
	if err := c.apiCall(getInvoicesMethod, queryParams, invoices); err != nil {
		return nil, err
	}
	return invoices, nil
}

// GetBalance call api/getBalance.
func (c ApiCore) GetBalance() (*GetBalanceResponse, error) {
	balanceInfo := new(GetBalanceResponse)
	if err := c.apiCall(getBalanceMethod, emptyQuery, balanceInfo); err != nil {
		return nil, err
	}
	return balanceInfo, nil
}

// GetExchangeRates call api/getExchangeRates.
func (c ApiCore) GetExchangeRates() (*GetExchangeRatesResponse, error) {
	exchangesInfo := new(GetExchangeRatesResponse)
	if err := c.apiCall(getExchangeRatesMethod, emptyQuery, exchangesInfo); err != nil {
		return nil, err
	}
	return exchangesInfo, nil
}

// GetCurrencies call api/getCurrencies.
func (c ApiCore) GetCurrencies() (*GetCurrenciesResponse, error) {
	currencyInfo := new(GetCurrenciesResponse)
	if err := c.apiCall(getCurrenciesMethod, emptyQuery, currencyInfo); err != nil {
		return nil, err
	}
	return currencyInfo, nil
}

// createEncodeQuery create url.Values from given map and encode to string.
func createEncodeQuery(params map[string]string) string {
	values := url.Values{}
	for k, v := range params {
		if v != emptyQuery {
			values.Add(k, v)
		}
	}
	return values.Encode()
}

// QueryParams encode options to query params for `createInvoice` method.
func (opt CreateInvoiceOptions) QueryParams() string {
	params := map[string]string{
		"asset":           opt.Asset.String(),
		"amount":          opt.Amount,
		"description":     opt.Description,
		"hidden_message":  opt.HiddenMessage,
		"paid_btn_url":    opt.PaidButtonUrl,
		"payload":         opt.Payload,
		"paid_btn_name":   opt.PaidButtonName.String(),
		"allow_comments":  strconv.FormatBool(opt.AllowComments),
		"allow_anonymous": strconv.FormatBool(opt.AllowAnonymous),
	}
	if opt.ExpiresIn != 0 {
		params["expires_in"] = strconv.Itoa(opt.ExpiresIn)
	}
	return createEncodeQuery(params)

}

// QueryParams encode options to query params for `transfer` method.
func (opt DoTransferOptions) QueryParams() string {
	return createEncodeQuery(map[string]string{
		"user_id":                   strconv.Itoa(opt.UserId),
		"asset":                     opt.Asset.String(),
		"amount":                    opt.Amount,
		"spend_id":                  opt.SpendId,
		"comment":                   opt.Comment,
		"disable_send_notification": strconv.FormatBool(opt.DisableSendNotification),
	})

}

// QueryParams encode options to query params for `getInvoices` method.
func (opt GetInvoicesOptions) QueryParams() string {
	params := map[string]string{
		"asset":        opt.Asset.String(),
		"status":       opt.Status.String(),
		"offset":       strconv.Itoa(opt.Offset),
		"invoices_ids": strings.Join(opt.InvoiceIds, ","),
	}
	// Values between 1-1000 are accepted. Defaults to 100.
	if (0 < opt.Count && opt.Count < 1000) && opt.Count != 100 {
		params["count"] = strconv.Itoa(opt.Count)
	}
	return createEncodeQuery(params)
}
