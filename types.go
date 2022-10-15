package cryptopay

import (
	"math/big"
	"time"
)

// Asset is currency code.

// AppInfo basic information about an app.
type AppInfo struct {
	Id                 int    `json:"app_id"`                          // ID of application.
	Name               string `json:"name"`                            // Name of application (sets on create app).
	PaymentBotUsername string `json:"payment_processing_bot_username"` // Telegram username of the bot that processing payments.
}

// Invoice object.
type Invoice struct {
	Id              int64         `json:"invoice_id"`                 // Unique ID for this invoice.
	Hash            string        `json:"hash,omitempty"`             // Hash of the invoice.
	Status          InvoiceStatus `json:"status"`                     // Status of the invoice, can be either .
	Asset           Asset         `json:"asset"`                      // Currency code.
	Amount          *big.Float    `json:"amount"`                     // Amount of the invoice.
	USDRate         *big.Float    `json:"usd_rate"`                   // Optional. Price of the Asset in USD at the time the invoice was paid.
	Fee             *big.Float    `json:"fee"`                        // Optional. Amount of charged service fees.
	PayUrl          string        `json:"pay_url,omitempty"`          // URL should be presented to the user to pay the invoice.
	CreatedAt       *time.Time    `json:"created_at"`                 // Date the invoice was created in ISO 8601 format.
	ExpirationDate  *time.Time    `json:"expiration_date,omitempty"`  // Optional. Date the invoice expires in Unix time. (not timestamp)
	Description     *string       `json:"description,omitempty"`      // Optional. Description for this invoice.
	Comment         *string       `json:"comment,omitempty"`          // Optional. Comment to the payment from the user.
	HiddenMessage   *string       `json:"hidden_message,omitempty"`   // Optional. Text of the hidden message for this invoice.
	Payload         *string       `json:"payload,omitempty"`          // Optional. Previously provided data for this invoice.
	PaidBtnName     *PaidButton   `json:"paid_btn_name,omitempty"`    // Optional. Name of the button.
	PaidBtnUrl      *string       `json:"paid_btn_url,omitempty"`     // Optional. URL of the button.
	PaidAt          *time.Time    `json:"paid_at,omitempty"`          // Optional. Date the invoice was paid in Unix time.
	PaidAnonymously *bool         `json:"paid_anonymously,omitempty"` // Optional. Text of the hidden message for this invoice.
	AllowComments   bool          `json:"allow_comments,omitempty"`   // True, if the user can add comment to the payment.
	AllowAnonymous  bool          `json:"allow_anonymous,omitempty"`  // True, if the user can pay the invoice anonymously.
}

// Transfer object
type Transfer struct {
	Id          int64      `json:"transfer_id"`       // Unique ID for this transfer.
	UserId      int64      `json:"user_id"`           // Telegram user ID the transfer was sent to.
	Asset       Asset      `json:"asset"`             // Currency code.
	Amount      *big.Float `json:"amount"`            // Amount of the transfer.
	Status      string     `json:"status"`            // Status of the transfer, can be “completed”.
	Comment     *string    `json:"comment,omitempty"` // Optional. Comment for this transfer.
	CompletedAt time.Time  `json:"completed_at"`      // Date the transfer was completed in ISO 8601 format.
}

// BalanceCurrency  contains information about available funds for a particular currency.
type BalanceCurrency struct {
	CurrencyCode Asset      `json:"currency_code"`
	Available    *big.Float `json:"available"` // Balance
}
type ExchangeRate struct {
	Source  Asset      `json:"source"`   // Source currency
	Target  Asset      `json:"target"`   // Target currency
	Rate    *big.Float `json:"rate"`     // Cost Target in Source currency
	IsValid bool       `json:"is_valid"` // Indicates valid exchange
}
type CurrencyInfo struct {
	Name     string `json:"name"`     // Name of currency
	Code     Asset  `json:"code"`     // Currency code
	Url      string `json:"url"`      // Url to currency homepage
	Decimals int    `json:"decimals"` // CryptoPay missing description this Field

	IsBlockchain bool `json:"is_blockchain"` // Indicates what currency is crypto.
	IsStablecoin bool `json:"is_stablecoin"` // Indicates what currency is stablecoin.
	IsFiat       bool `json:"is_fiat"`       // Indicates what currency is fiat (real)
}

func (c CurrencyInfo) Type() CurrencyType {
	switch {
	case c.IsBlockchain:
		return CurrencyBlockchain
	case c.IsStablecoin:
		return CurrencyStablecoin
	case c.IsFiat:
		return CurrencyFiat
	default:
		return CurrencyNone
	}
}

type CurrencyType uint8

const (
	CurrencyNone       CurrencyType = 'n'
	CurrencyBlockchain CurrencyType = 'b'
	CurrencyStablecoin CurrencyType = 's'
	CurrencyFiat       CurrencyType = 'f'
)
