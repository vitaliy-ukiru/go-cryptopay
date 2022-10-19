package cryptopay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"time"
)

type (
	// Handler is signature of handler for Webhook
	Handler func(update *WebhookUpdate) error
	// ErrorHandler is signature of Webhook.OnInHandlerError
	ErrorHandler         func(update *WebhookUpdate, err error)
	InternalErrorHandler func(rw io.Writer, err error)
)

// UpdateType is type of webhook update.
type UpdateType string

const (
	// UpdateInvoicePaid update type, indicates that invoice was paid.
	UpdateInvoicePaid UpdateType = "invoice_paid"
)

const HeaderSignatureName = "crypto-pay-api-signature"

// ErrorWrongSignature is returned if webhook don't verify update.
// For example, this can happen if someone who knows webhook's path endpoint and sends fake requests.
// Also, it could happen when request body was changed  from outside.
//
// If this happens, the update is not processed, but Webhook.OnInHandlerError is called
var ErrorWrongSignature = errors.New("crypto-pay/webhook:wrong request signature")

type WebhookRequest struct {
	Signature  string
	Body       []byte
	BodyWriter io.Writer
}

type Listener interface {
	Listen(updates chan<- WebhookRequest, stop chan struct{}) error
}

type Dispatcher interface {
	Dispatch(update *WebhookUpdate) error
}

// WebhookUpdate is object of update from request body.
type WebhookUpdate struct {
	// Id is Non-unique update ID.
	Id int `json:"update_id"`

	// UpdateType is webhook update type.
	UpdateType UpdateType `json:"update_type"`

	// RequestDate is date the request was sent in ISO 8601 format.
	RequestDate time.Time `json:"request_date"`

	// Payload is base invoice information.
	Payload Invoice `json:"payload"`
}

//Webhook representation http.Handler for works with CryptoPay updates
type Webhook struct {
	OnWebhookError   InternalErrorHandler
	OnInHandlerError ErrorHandler

	l       Listener
	d       Dispatcher
	updates chan WebhookRequest

	// tokenHash is SHA256 hash of app's token.
	// Webhook getting only hash because it minimizes calls hash functions and process time for verifyUpdate.
	tokenHash []byte
}

// NewWebhook returns new Webhook.
//func NewWebhook(token string, defaultHandlers map[UpdateType][]Handler, onError ErrorHandler) *Webhook {
//	handlers := defaultHandlers
//	if handlers == nil {
//		handlers = make(map[UpdateType][]Handler)
//	}
//	hash := sha256.New()
//	hash.Write([]byte(token))
//	return &Webhook{handlers: handlers, OnInHandlerError: onError, tokenHash: hash.Sum(nil)}
//}

func (w Webhook) Run() error {
	if w.updates != nil {
		return errors.New("webhook already running")
	}
	w.updates = make(chan WebhookRequest)

	stop := make(chan struct{})
	go func() {
		for {
			select {
			case req := <-w.updates:
				upd, err := w.verify(req)
				if err != nil {
					w.error(err, req.BodyWriter)
				}
				if err := w.d.Dispatch(upd); err != nil {
					w.error(err, nil)
				}
			case <-stop:
				close(w.updates)
				return
			}
		}
	}()
	return w.l.Listen(w.updates, stop)
}

// verify comparing HMAC-SHA-256 signature of request body
// with a secret key that is SHA256 hash of app's token and header parameter
// in requestSignature argument.
func (w Webhook) verify(request WebhookRequest) (*WebhookUpdate, error) {
	signature, err := hex.DecodeString(request.Signature)
	if err != nil {
		return nil, fmt.Errorf("cannot decode request signature: %w", err)
	}

	mac := hmac.New(sha256.New, w.tokenHash)
	mac.Write(request.Body)
	if !hmac.Equal(mac.Sum(nil), signature) {
		return nil, ErrorWrongSignature
	}

	update := new(WebhookUpdate)
	if err = json.Unmarshal(request.Body, update); err != nil {
		return nil, fmt.Errorf("cannot unmarshal body to WebhookUpdate: %w", err)
	}
	return update, nil
}

func (w Webhook) error(err error, other any) {
	switch obj := other.(type) {
	case io.Writer:
		if w.OnWebhookError != nil {
			w.OnWebhookError(obj, err)
		}
	case *WebhookUpdate:
		if w.OnInHandlerError != nil {
			w.OnInHandlerError(obj, err)
		}
	}
}
