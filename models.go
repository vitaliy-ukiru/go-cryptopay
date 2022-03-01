package go_crypto_pay

import (
	"errors"
	"fmt"
	"time"
)

// Asset is currency code.
type Asset string

//goland:noinspection ALL
const (
	BTC  Asset = "BTC"
	TON  Asset = "TON"
	ETH  Asset = "ETH" // ETH ONLY TESTNET
	USDT Asset = "USDT"
	USDC Asset = "USDC"
	BUSD Asset = "BUSD"
)

// PaidButton is name of the button that will be shown to a user after the invoice is paid.
type PaidButton string

//goland:noinspection ALL
const (
	ButtonViewItem    PaidButton = "viewItem"
	ButtonOpenChannel PaidButton = "openChannel"
	ButtonOpenBot     PaidButton = "openBot"
	ButtonCallback    PaidButton = "callback"
)

// InvoiceStatus is status of the invoice.
type InvoiceStatus string

//goland:noinspection ALL
const (
	StatusActive  InvoiceStatus = "active"
	StatusPaid    InvoiceStatus = "paid"
	StatusExpired InvoiceStatus = "expired"
)

type (
	// ApiError is error of CryptoPay API.
	ApiError struct {
		Code int    `json:"code"` // HTTP Code of error.
		Name string `json:"name"` // Name of error.
	}

	// AppInfo basic information about an app.
	AppInfo struct {
		Id                 int    `json:"app_id"`                          // ID of application.
		Name               string `json:"name"`                            // Name of application (sets on create app).
		PaymentBotUsername string `json:"payment_processing_bot_username"` // Telegram username of the bot that processing payments.
	}
	UpdateInvoice struct {
		Id              int           `json:"invoice_id"`                 // Unique ID for this invoice.
		Status          InvoiceStatus `json:"status"`                     // Status of the invoice, can be either .
		Hash            string        `json:"hash,omitempty"`             // Hash of the invoice.
		Asset           Asset         `json:"asset"`                      // Currency code.
		Amount          string        `json:"amount"`                     // Amount of the invoice.
		PayUrl          string        `json:"pay_url,omitempty"`          // URL should be presented to the user to pay the invoice.
		CreatedAt       time.Time     `json:"created_at"`                 // Date the invoice was created in ISO 8601 format.
		AllowComments   bool          `json:"allow_comments,omitempty"`   // True, if the user can add comment to the payment.
		AllowAnonymous  bool          `json:"allow_anonymous,omitempty"`  // True, if the user can pay the invoice anonymously.
		PaidAt          time.Time     `json:"paid_at,omitempty"`          // Optional. Date the invoice was paid in Unix time.
		PaidAnonymously bool          `json:"paid_anonymously,omitempty"` // Optional. Text of the hidden message for this invoice.

	}

	// Invoice object.
	Invoice struct {
		UpdateInvoice
		Description    string     `json:"description,omitempty"`     // Optional. Description for this invoice.
		ExpirationDate string     `json:"expiration_date,omitempty"` // Optional. Date the invoice expires in Unix time. (not timestamp)
		Comment        string     `json:"comment,omitempty"`         // Optional. Comment to the payment from the user.
		HiddenMessage  string     `json:"hidden_message,omitempty"`  // Optional. Text of the hidden message for this invoice.
		Payload        string     `json:"payload,omitempty"`         // Optional. Previously provided data for this invoice.
		PaidBtnName    PaidButton `json:"paid_btn_name,omitempty"`   // Optional. Name of the button.
		PaidBtnUrl     string     `json:"paid_btn_url,omitempty"`    // Optional. URL of the button.
	}
	// Transfer object
	Transfer struct {
		Id          int       `json:"transfer_id"`       // Unique ID for this transfer.
		UserId      int       `json:"user_id"`           // Telegram user ID the transfer was sent to.
		Asset       Asset     `json:"asset"`             // Currency code.
		Amount      string    `json:"amount"`            // Amount of the transfer.
		Status      string    `json:"status"`            // Status of the transfer, can be “completed”.
		CompletedAt time.Time `json:"completed_at"`      // Date the transfer was completed in ISO 8601 format.
		Comment     string    `json:"comment,omitempty"` // Optional. Comment for this transfer.
	}
	// BalanceCurrency  contains information about available funds for a particular currency.
	BalanceCurrency struct {
		CurrencyCode Asset  `json:"currency_code"`
		Available    string `json:"available"` // Balance
	}
	ExchangeRate struct {
		IsValid bool   `json:"is_valid"` // Indicates valid exchange
		Source  Asset  `json:"source"`   // Source currency
		Target  Asset  `json:"target"`   // Target currency
		Rate    string `json:"rate"`     // Cost Target in Source currency
	}
	CurrencyInfo struct {
		IsBlockchain bool   `json:"is_blockchain"` // Indicates what currency is crypto.
		IsStablecoin bool   `json:"is_stablecoin"` // Indicates what currency is stablecoin.
		IsFiat       bool   `json:"is_fiat"`       // Indicates what currency is fiat (real)
		Name         string `json:"name"`          // Name of currency
		Code         Asset  `json:"code"`          // Currency code
		Url          string `json:"url"`           // Url to currency homepage
		Decimals     int    `json:"decimals"`      // I don't know what is
	}
)

// GetApiError retrieves the ApiError from given error. If unsuccessfully returns nil.
func GetApiError(err error) *ApiError {
	var apiErr *ApiError
	if errors.As(err, &apiErr) {
		return apiErr
	}
	return nil
}

func (a ApiError) Error() string {
	return fmt.Sprintf("crypto-pay/api: api response %d - %s", a.Code, a.Name)
}

func (a Asset) String() string {
	return string(a)
}
func (p PaidButton) String() string {
	return string(p)
}
func (i InvoiceStatus) String() string {
	return string(i)
}
