package go_crypto_pay

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

type (
	// Handler is signature of handler for Webhook
	Handler func(update *WebhookUpdate)
	// ErrorHandler is signature of Webhook.OnError
	ErrorHandler func(r *http.Request, err error)
)

// UpdateType is type of webhook update.
type UpdateType string

const (
	// UpdateInvoicePaid update type, indicates that invoice was paid.
	UpdateInvoicePaid UpdateType = "invoice_paid"
)

const (
	headerSignatureName = "crypto-pay-api-signature"
	wrongSignature      = "wrong request signature"
)

// ErrorWrongSignature is returned if webhook don't verify update.
// For example, this can happen if someone who knows webhook's path endpoint and sends fake requests.
// Also, it could happen when request body was changed  from outside.
//
// If this happens, the update is not processed, but Webhook.OnError is called
var ErrorWrongSignature = fmt.Errorf("crypto-pay/webhook: %s", wrongSignature)

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
	// handlers set of split by update type
	handlers map[UpdateType][]Handler
	// OnError handler for errors
	OnError ErrorHandler
	// tokenHash is SHA256 hash of app's token.
	// Webhook getting only hash because it minimizes calls hash functions and process time for verifyUpdate.
	tokenHash []byte
}

// NewWebhook returns new Webhook.
func NewWebhook(token string, defaultHandlers map[UpdateType][]Handler, onError ErrorHandler) *Webhook {
	handlers := defaultHandlers
	if handlers == nil {
		handlers = make(map[UpdateType][]Handler)
	}
	hash := sha256.New()
	hash.Write([]byte(token))
	return &Webhook{handlers: handlers, OnError: onError, tokenHash: hash.Sum(nil)}
}

// Bind add handler given update type. Returns handler index.
func (w *Webhook) Bind(updateType UpdateType, handler Handler) int {
	v, ok := w.handlers[updateType]
	if !ok {
		v = []Handler{handler}
	} else {
		v = append(v, handler)
	}
	w.handlers[updateType] = v
	return len(w.handlers[updateType]) - 1
}

// DeleteHandlers deletes all handlers given type. Also, can delete all handlers for all update types.
// If you want to delete all handlers for all types, then pass "*" as a parameter
func (w *Webhook) DeleteHandlers(updateType UpdateType) {
	if updateType == "*" {
		w.handlers = make(map[UpdateType][]Handler)
		return
	}
	delete(w.handlers, updateType)
}

// DeleteHandlerByIndex deletes handler given type and index.
func (w *Webhook) DeleteHandlerByIndex(updateType UpdateType, i int) {
	a := w.handlers[updateType]
	w.handlers[updateType] = append(a[:i], a[i+1:]...)
}

// ServeHTTP implementing http.Handler.
//
// Thanks to this, you can transmit a webhook as a handler for a "net/http" server.
// By recommended use this like as http.Handler parameter for function http.Handle.
//
// If you use other router you can adapt. For this you must create handler that call this method.
// Examples of adapt see in README.md file
func (w Webhook) ServeHTTP(rw http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	data, _ := io.ReadAll(r.Body)
	signature, _ := hex.DecodeString(r.Header.Get(headerSignatureName))
	if !w.verifyUpdate(data, signature) {
		w.badRequestError(rw, r, ErrorWrongSignature, wrongSignature)
		return
	}
	update := new(WebhookUpdate)
	if err := json.Unmarshal(data, &update); err != nil {
		w.badRequestError(rw, r, err, "")
		return
	}
	rw.WriteHeader(http.StatusOK)
	if v, ok := w.handlers[update.UpdateType]; ok {
		for _, handler := range v {
			go handler(update)
		}
	}
}

// verifyUpdate comparing HMAC-SHA-256 signature of request body with a secret key that is SHA256 hash of app's token and header parameter in requestSignature argument.
func (w Webhook) verifyUpdate(requestBody, requestSignature []byte) bool {
	mac := hmac.New(sha256.New, w.tokenHash)
	mac.Write(requestBody)
	return hmac.Equal(mac.Sum(nil), requestSignature)
}

func (w Webhook) badRequestError(rw http.ResponseWriter, r *http.Request, err error, msg string) {
	if msg != "" {
		badRequest(rw, msg)
	} else {
		badRequest(rw, err.Error())
	}
	if w.OnError != nil {
		w.OnError(r, err)
	}
}

// badRequest helper for response with 400 code.
func badRequest(rw http.ResponseWriter, message string) {
	rw.WriteHeader(http.StatusBadRequest)
	rw.Write([]byte(message))
}
