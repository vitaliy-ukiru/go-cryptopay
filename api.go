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
//goland:noinspection ALL
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
	// CreateInvoiceOptions is params for `createInvoice` api method.
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
	// DoTransferOptions is params for `transfer` (DoTransfer) api method.
	DoTransferOptions struct {
		UserId                  int    // Telegram user ID. User must have previously used @CryptoBot (@CryptoTestnetBot for testnet).
		Asset                   Asset  // Currency code.
		Amount                  string // Amount of the invoice in float.
		SpendId                 string // Unique ID to make your request idempotent and ensure that only one of the transfers with the same spend_id will be accepted by Crypto Pay API. More https://telegra.ph/Crypto-Pay-API-11-25#transfer
		Comment                 string // Optional. Comment for the transfer. Users will see this comment when they receive a notification about the transfer. Up to 1024 symbols.
		DisableSendNotification bool   // Optional. Pass true if the user should not receive a notification about the transfer. Default is false.
	}
	// GetInvoicesOptions is params for `getInvoices` api method.
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
		// Error from CryptoPay API, nil on successfully.
		Error *ApiError `json:"error,omitempty"`
	}
	// GetMeResponse is response object for `getMe` method
	GetMeResponse struct {
		BaseApiResponse
		Result *AppInfo `json:"result,omitempty"`
	}
	// CreateInvoiceResponse is response object for `createInvoice` method
	CreateInvoiceResponse struct {
		BaseApiResponse
		Result *Invoice `json:"result,omitempty"`
	}
	// DoTransferResponse is response object for `transfer` method
	DoTransferResponse struct {
		BaseApiResponse
		Result *Transfer `json:"result,omitempty"`
	}
	// GetInvoicesResponse is response object for `getInvoices` method
	GetInvoicesResponse struct {
		BaseApiResponse
		Result struct {
			Items []Invoice `json:"items"`
		} `json:"result,omitempty"`
	}
	// GetBalanceResponse is response object for `getBalance` method
	GetBalanceResponse struct {
		BaseApiResponse
		Result []BalanceCurrency `json:"result,omitempty"`
	}
	// GetExchangeRatesResponse is response object for `getExchangeRates` method
	GetExchangeRatesResponse struct {
		BaseApiResponse
		Result []ExchangeRate `json:"result,omitempty"`
	}
	// GetCurrenciesResponse is response object for `getCurrencies` method
	GetCurrenciesResponse struct {
		BaseApiResponse
		Result []CurrencyInfo `json:"result,omitempty"`
	}
)

// ApiCore is low-level (in library context) client for CryptoPay API
// without convenient public interface and additional checks.
//
// Recommended using it only for specific operations.
type ApiCore struct {
	token      string
	url        string
	httpClient *http.Client
}

// NewApi returns new ApiCore
func NewApi(token, url string, httpClient *http.Client) *ApiCore {
	return &ApiCore{token: token, url: url, httpClient: httpClient}
}

// urlFmt formats URL, paste query params
func (c ApiCore) urlFmt(method string, queryParams string) string {
	methodUrl := fmt.Sprintf("%s/api/%s", c.url, method)
	if queryParams == emptyQuery {
		return methodUrl
	}
	return methodUrl + "?" + queryParams
}

// apiCall makes request to API and deserialization response body in dest argument
func (c ApiCore) apiCall(apiMethod, queryParams string, dest interface{}) error {
	req, err := http.NewRequest("GET", c.urlFmt(apiMethod, queryParams), nil)
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

// GetMe calls api/getMe.
func (c ApiCore) GetMe() (*GetMeResponse, error) {
	var appInfo GetMeResponse
	if err := c.apiCall(getMeMethod, emptyQuery, &appInfo); err != nil {
		return nil, err
	}
	return &appInfo, nil
}

// CreateInvoice calls api/createInvoice.
func (c ApiCore) CreateInvoice(opt CreateInvoiceOptions) (*CreateInvoiceResponse, error) {
	var newInvoice CreateInvoiceResponse
	err := c.apiCall(createInvoiceMethod, opt.QueryParams(), &newInvoice)
	if err != nil {
		return nil, err
	}
	return &newInvoice, nil
}

// DoTransfer calls api/transfer.
func (c ApiCore) DoTransfer(opt DoTransferOptions) (*DoTransferResponse, error) {
	var newTransfer DoTransferResponse
	err := c.apiCall(transferMethod, opt.QueryParams(), &newTransfer)
	if err != nil {
		return nil, err
	}
	return &newTransfer, nil
}

// GetInvoices calls api/getInvoices. Set opt as nil for empty API params.
func (c ApiCore) GetInvoices(opt *GetInvoicesOptions) (*GetInvoicesResponse, error) {
	var invoices GetInvoicesResponse
	var queryParams string
	if opt != nil {
		queryParams = opt.QueryParams()
	}
	err := c.apiCall(getInvoicesMethod, queryParams, &invoices)
	if err != nil {
		return nil, err
	}
	return &invoices, nil
}

// GetBalance calls api/getBalance.
func (c ApiCore) GetBalance() (*GetBalanceResponse, error) {
	var balanceInfo GetBalanceResponse
	err := c.apiCall(getBalanceMethod, emptyQuery, &balanceInfo)
	if err != nil {
		return nil, err
	}
	return &balanceInfo, nil
}

// GetExchangeRates calls api/getExchangeRates.
func (c ApiCore) GetExchangeRates() (*GetExchangeRatesResponse, error) {
	var exchangesInfo GetExchangeRatesResponse
	err := c.apiCall(getExchangeRatesMethod, emptyQuery, &exchangesInfo)
	if err != nil {
		return nil, err
	}
	return &exchangesInfo, nil
}

// GetCurrencies calls api/getCurrencies.
func (c ApiCore) GetCurrencies() (*GetCurrenciesResponse, error) {
	var currencyInfo GetCurrenciesResponse
	err := c.apiCall(getCurrenciesMethod, emptyQuery, &currencyInfo)
	if err != nil {
		return nil, err
	}
	return &currencyInfo, nil
}

// createEncodeQuery creates url.Values from given map and encode to string.
func createEncodeQuery(params map[string]string) string {
	values := url.Values{}
	for k, v := range params {
		if v != emptyQuery {
			values.Add(k, v)
		}
	}
	return values.Encode()
}

// QueryParams encodes options to query params for `createInvoice` method.
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

// QueryParams encodes options to query params for `transfer` method.
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

// QueryParams encodes options to query params for `getInvoices` method.
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
