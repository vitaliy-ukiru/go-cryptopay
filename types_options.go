package cryptopay

import (
	"math/big"
	"net/url"
	"strconv"
	"strings"

	"github.com/vitaliy-ukiru/go-cryptopay/internal"
)

// CreateInvoiceOptions for `createInvoice` api method.
type CreateInvoiceOptions struct {
	Asset  Asset      // Currency code.
	Amount *big.Float // Amount of the invoice in float.

	Description    *string     // Optional. Description for the invoice. User will see this description when they pay the invoice. Up to 1024 characters.
	HiddenMessage  *string     // Optional. Text of the message that will be shown to a user after the invoice is paid. Up to 2o48 characters.
	PaidButtonName *PaidButton // Optional. Name of the button that will be shown to a user after the invoice is paid.
	PaidButtonUrl  *string     // Optional. Required if PaidButtonName is used. URL to be opened when the button is pressed. You can set any success link (for example, a link to your bot). Starts with https or http.
	Payload        *string     // Optional. Any data you want to attach to the invoice (for example, user ID, payment ID, ect). Up to 4kb.
	ExpiresIn      int64       // Optional. You can set a payment time limit for the invoice in seconds. Values between 1-2678400 are accepted
	AllowComments  *bool       // Optional. Allow a user to add a comment to the payment. Default is true.
	AllowAnonymous *bool       // Optional. Allow a user to pay the invoice anonymously. Default is true.
}

// DoTransferOptions for `transfer` (DoTransfer) api method.
type DoTransferOptions struct {
	UserId  int64      // Telegram user ID. User must have previously used @CryptoBot (@CryptoTestnetBot for testnet).
	SpendId string     // Unique ID to make your request idempotent and ensure that only one of the transfers with the same spend_id will be accepted by Crypto Pay API. More https://telegra.ph/Crypto-Pay-API-11-25#transfer
	Asset   Asset      // Currency code.
	Amount  *big.Float // Amount of the invoice in float.

	Comment                 *string // Optional. Comment for the transfer. Users will see this comment when they receive a notification about the transfer. Up to 1024 symbols.
	DisableSendNotification *bool   // Optional. Pass true if the user should not receive a notification about the transfer. Default is false.
}

// GetInvoicesOptions for `getInvoices` api method.
type GetInvoicesOptions struct {
	Asset Asset // Currency code.

	Status     InvoiceStatus // Optional. Status of invoices to be returned. Defaults to all statuses.
	InvoiceIds []int64       // Optional. Invoice IDs
	Offset     int           // Optional. Offset needed to return a specific subset of invoices. Default is 0.
	Count      int           // Optional. Number of invoices to be returned. Values between 1-1000 are accepted. Defaults to 100.
}

// Pointer returns pointer to value.
// This functions helps create pointer values to primitive types.
func Pointer[T any](value T) *T {
	return &value
}

// QueryParams encode options to query params for `createInvoice` method.
func (opt CreateInvoiceOptions) QueryParams() url.Values {
	values := internal.NewValues(url.Values{
		"asset":         []string{opt.Asset.String()},
		"amount":        []string{opt.Amount.String()},
		"paid_btn_name": []string{opt.PaidButtonName.String()},
	})
	if opt.ExpiresIn != 0 {
		values.SetInt64("expires_in", opt.ExpiresIn)
	}

	values.SetPtr("description", opt.Description)
	values.SetPtr("hidden_message", opt.HiddenMessage)
	values.SetPtr("paid_btn_url", opt.PaidButtonUrl)
	values.SetPtr("payload", opt.Payload)

	values.SetBoolPtr("allow_comments", opt.AllowComments)
	values.SetBoolPtr("allow_anonymous", opt.AllowAnonymous)

	return values.Values
}

// QueryParams encode options to query params for `transfer` method.
func (opt DoTransferOptions) QueryParams() url.Values {
	values := internal.NewValues(url.Values{
		"user_id":  []string{strconv.FormatInt(opt.UserId, 10)},
		"asset":    []string{opt.Asset.String()},
		"amount":   []string{opt.Amount.String()},
		"spend_id": []string{opt.SpendId},
	})
	values.SetPtr("comment", opt.Comment)
	values.SetBoolPtr("disable_send_notification", opt.DisableSendNotification)
	return values.Values
}

// QueryParams encode options to query params for `getInvoices` method.
func (opt GetInvoicesOptions) QueryParams() url.Values {
	params := internal.NewValues(url.Values{
		"asset":        []string{opt.Asset.String()},
		"status":       []string{opt.Status.String()},
		"invoices_ids": []string{joinInt64(opt.InvoiceIds, ",")},
	})
	params.SetInt("offset", opt.Offset)
	// Values between 1-1000 are accepted. Defaults to 100.
	if (0 < opt.Count && opt.Count < 1000) && opt.Count != 100 {
		params.SetInt("count", opt.Count)
	}
	return params.Values
}

func joinInt64(items []int64, sep string) string {
	var sb strings.Builder
	sb.WriteString(strconv.FormatInt(items[0], 10))
	for _, i := range items[1:] {
		sb.WriteString(sep)
		sb.WriteString(strconv.FormatInt(i, 10))
	}
	return sb.String()
}
